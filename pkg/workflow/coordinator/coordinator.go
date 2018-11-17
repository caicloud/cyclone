package coordinator

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/common/constants"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/common"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/k8sapi"
)

func init() {
	// create container directory if not exist.
	createDirectory(constants.ContainerLogDir)
}

// Coordinator is a struct which contains infomations
// will be used in workflow sidecar named coordinator.
type Coordinator struct {
	timeoutWait     time.Duration
	runtimeExec     RuntimeExector
	workflowrunName string
	stageName       string
}

// RuntimeExector is an interface defined some methods
// to communicate with k8s container runtime.
type RuntimeExector interface {
	WaitContainers(timeout time.Duration, state common.ContainerState, excepts []string) error
	GetAllContainers() ([]string, error)
	KillContainer(containerName string) error
	CollectLog(containerName string, path string) error
}

// NewCoordinator create an assistant instance.
func NewCoordinator(timeout int, client clientset.Interface) *Coordinator {
	return &Coordinator{
		timeoutWait:     time.Duration(timeout) * time.Minute,
		runtimeExec:     k8sapi.NewK8sapiExector(getNamespace(), getPodName(), client),
		workflowrunName: getWorkflowrunName(),
		stageName:       getStageName(),
	}
}

// CollectLogs collects all containers' logs except for the coordinator container itself.
func (co *Coordinator) CollectLogs() {
	cs, err := co.getAllOtherContainers()
	if err != nil {
		return
	}

	for _, c := range cs {
		path := fmt.Sprintf(constants.FmtContainerLogPath, c)
		go func(containerName string, logPath string) {
			errl := co.runtimeExec.CollectLog(containerName, logPath)
			if errl != nil {
				log.Errorf("Collect %s log failed:%v", c, errl)
			}
		}(c, path)

		go func() {
			// TODO websocket send logs to cyclone-server
		}()
	}

}

// WaitWorkloadTerminate waits all containers to start run.
func (co *Coordinator) WaitRunning() {
	excepts := []string{}

	err := co.runtimeExec.WaitContainers(co.timeoutWait, common.ContainerStateNotWaiting, excepts)
	if err != nil {
		log.Error("Wait containers to running error: %v", err)

		//co.killAllOthers()
		return
	}

}

// WaitWorkloadTerminate waits all customized containers to be Terminated status.
func (co *Coordinator) WaitWorkloadTerminate() {
	excepts := []string{constants.ContainerCoordinatorName, constants.ContainerOutputResolverName}

	err := co.runtimeExec.WaitContainers(co.timeoutWait, common.ContainerStateTerminated, excepts)
	if err != nil {
		log.Error("Wait containers to completion error: %v", err)

		co.killAllOthers()
		return
	}

}

// WaitAllOthersTerminate waits all containers except for
// the coordinator container itself to become Terminated status.
func (co *Coordinator) WaitAllOthersTerminate() {
	excepts := []string{constants.ContainerCoordinatorName}

	err := co.runtimeExec.WaitContainers(co.timeoutWait, common.ContainerStateTerminated, excepts)
	if err != nil {
		log.Error("Wait containers to completion error: %v", err)

		co.killAllOthers()
		return
	}

}

// NotifyResolver create a file to notify output resolver to start working.
func (co *Coordinator) NotifyResolver() error {
	_, err := os.Create(constants.OutputResolverStartFlagPath)
	if err != nil {
		return err
	}
	return nil

}

// killAllOthers kill all containers except for the coordinator container itself.
func (co *Coordinator) killAllOthers() error {
	cs, err := co.getAllOtherContainers()
	if err != nil {
		return err
	}

	for _, c := range cs {
		co.runtimeExec.KillContainer(c)
	}

	return nil
}

func (co *Coordinator) getAllOtherContainers() ([]string, error) {
	var cs []string
	allContainers, err := co.runtimeExec.GetAllContainers()
	if err != nil {

	}

	for _, c := range allContainers {
		if c != constants.ContainerCoordinatorName {
			cs = append(cs, c)
		}
	}

	return cs, nil
}
