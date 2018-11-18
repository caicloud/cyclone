package main

import (
	"context"
	"flag"

	"github.com/caicloud/cyclone/pkg/common/signals"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/controllers"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeConfigPath = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
var configPath = flag.String("config", "workflow-controller.json", "Path to workflow controller config.")
var cm = flag.String("cm", "workflow-controller-config", "ConfigMap that configures workflow controller")
var namespace = flag.String("namespace", "default", "Namespace that workflow controller will run in")

func main() {
	flag.Parse()

	// Load configuration from config file.
	if err := controller.LoadConfig(configPath, &controller.Config); err != nil {
		log.Fatal("Load config failed.")
	}

	// Init logging system.
	controller.InitLogger(&controller.Config.Logging)

	// Create k8s clientset and registry system signals for exit.
	client := getClients(*kubeConfigPath)
	ctx, cancel := context.WithCancel(context.Background())
	signals.GracefulShutdown(cancel)

	// Watch configure changes in ConfigMap.
	cmController := controllers.NewConfigMapController(client, *namespace, *cm)
	go cmController.Run(ctx.Done())

	// Create and start WorkflowRun controller.
	wfrController := controllers.NewWorkflowRunController(client)
	go wfrController.Run(ctx.Done())

	// Create and start Pod controller.
	podController := controllers.NewPodController(client)
	go podController.Run(ctx.Done())

	// Wait forever.
	select {}
}

func getClients(kubeConfigPath string) (clientset.Interface) {
	var config *rest.Config
	var err error
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Fatalf("create config error: %v", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("create config error: %v", err)
		}
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("create client error: %v", err)
	}

	return client
}