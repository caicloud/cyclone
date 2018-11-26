package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/common/signals"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/controllers"
)

var kubeConfigPath = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
var configMap = flag.String("configmap", "workflow-controller-config", "ConfigMap that configures workflow controller")
var namespace = flag.String("namespace", "default", "Namespace that workflow controller will run in")

func main() {
	flag.Parse()

	// Create k8s clientset and registry system signals for exit.
	client, err := common.GetClient("", *kubeConfigPath)
	if err != nil {
		log.Fatal("Create k8s clientset error: ", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	signals.GracefulShutdown(cancel)

	// Load configuration from ConfigMap.
	cm, err := client.CoreV1().ConfigMaps(*namespace).Get(*configMap, metav1.GetOptions{})
	if err != nil {
		log.WithField("configmap", *configMap).Fatal("Get ConfigMap error: ", err)
	}
	if err = controller.LoadConfig(cm); err != nil {
		log.WithField("configmap", *cm).Fatal("Load config from ConfigMap error: ", err)
	}

	// Init logging system.
	controller.InitLogger(&controller.Config.Logging)

	// Watch configure changes in ConfigMap.
	cmController := controllers.NewConfigMapController(client, *namespace, *configMap)
	go cmController.Run(ctx.Done())

	// Watch workflowTrigger who will start workflowRun on schedule
	wftController := controllers.NewWorkflowTriggerController(client)
	go wftController.Run(ctx.Done())

	// Create and start WorkflowRun controller.
	wfrController := controllers.NewWorkflowRunController(client)
	go wfrController.Run(ctx.Done())

	// Create and start Pod controller.
	podController := controllers.NewPodController(client)
	go podController.Run(ctx.Done())

	// Wait forever.
	select {}
}
