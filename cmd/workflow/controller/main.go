package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/common/signals"
	"github.com/caicloud/cyclone/pkg/leaderelection"
	utilk8s "github.com/caicloud/cyclone/pkg/util/k8s"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/controllers"
	"github.com/caicloud/cyclone/pkg/workflow/controller/store"
)

var (
	kubeConfigPath  = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	configMap       = flag.String("configmap", "workflow-controller-config", "ConfigMap that configures workflow controller")
	healthCheckPort = flag.Int("health-check-port", 8080, "Workflow controller health check port")
)

func main() {
	flag.Parse()

	// Print Cyclone ascii art logo
	fmt.Println(common.CycloneLogo)

	// Create k8s clientset and registry system signals for exit.
	client, err := utilk8s.GetClient(*kubeConfigPath)
	if err != nil {
		log.Fatal("Create k8s clientset error: ", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	signals.GracefulShutdown(cancel)

	// Load configuration from ConfigMap.
	systemNamespace := common.GetSystemNamespace()
	cm, err := client.CoreV1().ConfigMaps(systemNamespace).Get(*configMap, metav1.GetOptions{})
	if err != nil {
		log.WithField("configmap", *configMap).Fatal("Get ConfigMap error: ", err)
	}
	if err = controller.LoadConfig(cm); err != nil {
		log.WithField("configmap", *cm).Fatal("Load config from ConfigMap error: ", err)
	}

	// Init logging
	controller.InitLogger(&controller.Config.Logging)

	// create CRD
	v1alpha1.EnsureCRDCreated("", *kubeConfigPath)

	// Init control cluster, ExecutionCluster for control cluster will be created.
	if err = controller.InitControlCluster(client); err != nil {
		log.Fatal("Init control cluster error: ", err)
	}

	runController := func(ctx context.Context) {
		// Watch configure changes in ConfigMap.
		cmController := controllers.NewConfigMapController(client, systemNamespace, *configMap)
		go cmController.Run(1, ctx.Done())

		// Watch workflowTrigger who will start workflowRun on schedule
		wftController := controllers.NewWorkflowTriggerController(client)
		go wftController.Run(controller.Config.WorkersNumber.WorkflowTrigger, ctx.Done())

		// Create and start WorkflowRun controller.
		wfrController := controllers.NewWorkflowRunController(client)
		go wfrController.Run(controller.Config.WorkersNumber.WorkflowRun, ctx.Done())

		// Create and start execution cluster controller.
		clusterController := controllers.NewExecutionClusterController(client)
		go clusterController.Run(controller.Config.WorkersNumber.ExecutionCluster, ctx.Done())

		// Watch for execution cluster, start pod controller for it.
		for {
			select {
			case <-ctx.Done():
				return
			case cluster := <-store.NewClusterChan:
				podController := controllers.NewPodController(cluster.Client, client)
				go podController.Run(controller.Config.WorkersNumber.Pod, cluster.StopCh)
			}
		}
	}

	leaseLockNamespace := os.Getenv("POD_NAMESPACE")
	if leaseLockNamespace == "" {
		leaseLockNamespace = "default"
	}
	leaderelection.RunOrDie(leaderelection.Option{
		LeaseLockName:      "cyclone-workflow-controller",
		LeaseLockNamespace: leaseLockNamespace,
		KubeClient:         client,
		Run:                runController,
		Port:               *healthCheckPort,
		StopCh:             ctx.Done(),
	})
}
