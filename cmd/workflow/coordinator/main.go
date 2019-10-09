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

const (
	//  exitDelayTime is the delay time of coordinator exit, coordinator need the time to do
	// something, like collecting logs, to make exit gracefully
	exitDelayTime = 3 * time.Second
)

var kubeConfigPath = flag.String("kubeconfig", "", "Path to kubeconfig. Only required if out-of-cluster.")

func main() {
	flag.Parse()

	// Print Cyclone ascii art logo
	fmt.Println(common.CycloneLogo)

	var err error
	var message string
	defer func() {
		// graceful showdown, need delay time to collect logs of other containers
		time.Sleep(exitDelayTime)
		if err != nil {
			log.Error(message)
			os.Exit(1)
		} else {
			log.Info(message)
			os.Exit(0)
		}
	}()

	// Create k8s clientset
	client, err := utilk8s.GetClient(*kubeConfigPath)
	if err != nil {
		log.Errorf("Get k8s client error: %v", err)
		return
	}

	// New workflow stage coordinator.
	c, err := coordinator.NewCoordinator(client)
	if err != nil {
		log.Errorf("New coornidator failed: %v", err)
		return
	}

	// Wait all containers running, so we can start to collect logs.
	log.Info("Wait all containers running ... ")
	err = c.WaitRunning()
	if err != nil {
		return
	}

	// Collect real-time container logs using goroutines.
	log.Info("Start to collect logs.")
	err = c.CollectLogs()
	if err != nil {
		return
	}
	defer func() {
		go func() {
			if err := c.MarkLogEOF(); err != nil {
				log.Warn("Mark log EOF error: ", err)
			}
		}()
	}()

	// Wait workload containers completion, so we can notify output resolvers.
	log.Info("Wait workload containers completion ... ")
	if err = c.WaitWorkloadTerminate(); err != nil {
		return
	}

	// Collect execution result from the workload container, results are key-value pairs in a
	// specified file, /__result__
	if err = c.CollectExecutionResults(); err != nil {
		message = fmt.Sprintf("Collect execution results error: %v", err)
		return
	}

	// Check if the workload is succeeded.
	if !c.WorkLoadSuccess() {
		message = fmt.Sprintf("Stage %s failed, workload exit code is not 0", c.Stage.Name)
		err = fmt.Errorf(message)
		return
	}

	// Collect all resources
	log.Info("Start to collect resources.")
	if err = c.CollectResources(); err != nil {
		message = fmt.Sprintf("Stage %s failed to collect output resource, error: %v.", c.Stage.Name, err)
		return
	}

	// Notify output resolver to start working.
	log.Info("Start to notify resolvers.")
	if err = c.NotifyResolvers(); err != nil {
		message = fmt.Sprintf("Stage %s failed to notify output resolvers, error: %v", c.Stage.Name, err)
		return
	}

	// Collect all artifacts
	log.Info("Start to collect artifacts.")
	if err = c.CollectArtifacts(); err != nil {
		message = fmt.Sprintf("Stage %s failed to collect artifacts, error: %v", c.Stage.Name, err)
		return
	}

	// Wait all others container completion. Coordinator will be the last one
	// to quit since it need to collect other containers' logs.
	log.Info("Wait for all other containers completion ... ")
	if err = c.WaitAllOthersTerminate(); err != nil {
		return
	}

	// Check if the workload and resolver containers are succeeded.
	if !c.StageSuccess() {
		message = fmt.Sprintf("Stage %s failed, some containers exited with code non-zero", c.Stage.Name)
		err = fmt.Errorf(message)
		return
	}

	message = fmt.Sprintf("Stage %s succeeded", c.Stage.Name)
}
