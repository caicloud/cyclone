package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/common"
	utilk8s "github.com/caicloud/cyclone/pkg/util/k8s"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator"
)

var kubeConfigPath = flag.String("kubeconfig", "", "Path to kubeconfig. Only required if out-of-cluster.")

func main() {
	flag.Parse()

	// Print Cyclone ascii art logo
	fmt.Println(common.CycloneLogo)

	var err error
	var message string

	// Create k8s clientset and registry system signals for exit.
	client, err := utilk8s.GetClient(*kubeConfigPath)
	if err != nil {
		log.Errorf("Get k8s client error: %v", err)
		os.Exit(1)
	}

	// New workflow stage coordinator.
	c, err := coordinator.NewCoordinator(client)
	if err != nil {
		log.Errorf("New coornidator failed: %v", err)
		os.Exit(1)
	}

	defer func() {
		if err != nil {
			log.Error(message)
			// Wait for sending event
			time.Sleep(1 * time.Second)
			os.Exit(1)
		} else {
			log.Info(message)
			// Wait for sending event
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}
	}()

	// Wait all containers running, so we can start to collect logs.
	log.Info("Wait all containers running ... ")
	c.WaitRunning()

	// Collect real-time container logs using goroutines.
	log.Info("Start to collect logs.")
	c.CollectLogs()

	// Wait workload containers completion, so we can notify output resolvers.
	log.Info("Wait workload containers completion ... ")
	c.WaitWorkloadTerminate()

	// Check if the workload is succeeded.
	if !c.WorkLoadSuccess() {
		message = fmt.Sprintf("Stage %s failed, workload exit code is not 0", c.Stage.Name)
		err = fmt.Errorf(message)
		return
	}

	// Collect all resources
	log.Info("Start to collect resources.")
	err = c.CollectResources()
	if err != nil {
		message = fmt.Sprintf("Stage %s failed to collect output resource, error: %v.", c.Stage.Name, err)
		return
	}

	// Notify output resolver to start working.
	log.Info("Start to notify resolvers.")
	err = c.NotifyResolvers()
	if err != nil {
		message = fmt.Sprintf("Stage %s failed to notify output resolvers, error: %v", c.Stage.Name, err)
		return
	}

	// Collect all artifacts
	log.Info("Start to collect artifacts.")
	err = c.CollectArtifacts()
	if err != nil {
		message = fmt.Sprintf("Stage %s failed to collect artifacts, error: %v", c.Stage.Name, err)
		return
	}

	// Wait all others container completion. Coordinator will be the last one
	// to quit since it need to collect other containers' logs.
	log.Info("Wait for all other containers completion ... ")
	c.WaitAllOthersTerminate()

	// Check if the workload and resolver containers are succeeded.
	if c.StageSuccess() {
		message = fmt.Sprintf("Stage %s succeeded", c.Stage.Name)
		return
	}

	message = fmt.Sprintf("Stage %s failed, resolver exit code is not 0", c.Stage.Name)
	err = fmt.Errorf(message)
	return
}
