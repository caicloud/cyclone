package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	k8sclient "github.com/caicloud/cyclone/pkg/util/k8s-client"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator"
)

var kubeConfigPath = flag.String("kubeconfig", "", "Path to kubeconfig. Only required if out-of-cluster.")
var logLevel = flag.String("loglevel", "Info", "Log level of workflow assistant.")

func main() {
	flag.Parse()

	// Create k8s clientset and registry system signals for exit.
	client, err := k8sclient.GetClient("", *kubeConfigPath)
	if err != nil {
		log.Errorf("Get k8s client error: %v", err)
		os.Exit(1)
	}

	// New workflow stage coordinator.
	c := coordinator.NewCoordinator(120, client, *kubeConfigPath)

	// Wait all containers running, so we can start to collect logs.
	log.Info("Wait all containers running ... ")
	c.WaitRunning()

	// Collect real-time container logs using goroutines.
	log.Info("Start to collect logs.")
	c.CollectLogs()

	// Wait customized containers completion, so we can notify output resolver.
	//
	// Users need to control ALL their workload containers to quit after work done,
	// otherwise, the coordinator container will wait forever.
	log.Info("Wait customized containers completion ... ")
	c.WaitWorkloadTerminate()

	// Notify output resolver to start working.
	log.Info("Start to notify resolver.")
	err = c.NotifyResolver()
	if err != nil {
		log.Errorf("Notify resolver failed, error: %v", err)
		os.Exit(1)
	}

	// Collect all artifacts
	err = c.CollectArtifacts()
	if err != nil {
		log.Errorf("Collect artifacts error: %v", err)
		os.Exit(1)
	}

	// Collect all resources
	err = c.CollectResources()
	if err != nil {
		log.Errorf("Collect resources error: %v", err)
		os.Exit(1)
	}

	// Wait all others container completion. Coordinator will be the last one
	// to quit since it need to collect other containers' logs.
	log.Info("Wait output resolver containers completion ... ")
	c.WaitAllOthersTerminate()

	// Check if the workload is succeeded.
	if c.IsWorkloadSuccess() {
		log.Info("Coordinator finished.")
		os.Exit(0)
	} else {
		log.Errorf("Workload failed.")
		os.Exit(1)
	}

}
