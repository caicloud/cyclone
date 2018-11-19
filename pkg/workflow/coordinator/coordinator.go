package coordinator

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common/constants"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/common"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/k8sapi"
)

// GetExitCodeRetry defined the max retry times for
// get containers exit code.
var GetExitCodeRetry = 10

func init() {
	// create container directory if not exist.
	createDirectory(constants.ContainerLogDir)
}

// Coordinator is a struct which contains infomations
// will be used in workflow sidecar named coordinator.
type Coordinator struct {
	runtimeExec       RuntimeExecutor
	workflowrunName   string
	stageName         string
	workloadContainer string
}

// RuntimeExecutor is an interface defined some methods
// to communicate with k8s container runtime.
type RuntimeExecutor interface {
	WaitContainers(state common.ContainerState, excepts []string) error
	KillContainer(containerName string) error
	CollectLog(containerName string, path string) error
	GetStageOutputs(name string) (v1alpha1.Outputs, error)
	CopyArtifact(container, path, dst string) error
	GetPod() (*core_v1.Pod, error)
}

// NewCoordinator create a coordinator instance.
func NewCoordinator(client clientset.Interface, kubecfg string) *Coordinator {
	return &Coordinator{
		runtimeExec:       k8sapi.NewK8sapiExector(getNamespace(), getPodName(), client, kubecfg),
		workflowrunName:   getWorkflowrunName(),
		stageName:         getStageName(),
		workloadContainer: getWorkloadContainer(),
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

// WaitRunning waits all containers to start run.
func (co *Coordinator) WaitRunning() {
	excepts := []string{}

	err := co.runtimeExec.WaitContainers(common.ContainerStateInitialized, excepts)
	if err != nil {
		log.Errorf("Wait containers to running error: %v", err)
		return
	}

}

// WaitWorkloadTerminate waits all customized containers to be Terminated status.
func (co *Coordinator) WaitWorkloadTerminate() {
	excepts := []string{constants.ContainerSidecarPrefix}

	err := co.runtimeExec.WaitContainers(common.ContainerStateTerminated, excepts)
	if err != nil {
		log.Errorf("Wait containers to completion error: %v", err)
		return
	}

}

// WaitAllOthersTerminate waits all containers except for
// the coordinator container itself to become Terminated status.
func (co *Coordinator) WaitAllOthersTerminate() {
	excepts := []string{constants.ContainerCoordinatorName}

	err := co.runtimeExec.WaitContainers(common.ContainerStateTerminated, excepts)
	if err != nil {
		log.Errorf("Wait containers to completion error: %v", err)
		return
	}

}

// Check if the workload is succeeded.
func (co *Coordinator) IsWorkloadSuccess() bool {
	ws, err := co.GetExitCodes()
	if err != nil {
		log.Errorf("Get Exit Codes failed: %v", err)
		return false
	}

	log.WithField("codes", ws).Debug()

	for _, code := range ws {
		if code != 0 {
			return false
		}
	}

	return true
}

// Get all other containers exit code.
// Try 'GetExitCodeRetry' times if the exit code not exist.
func (co *Coordinator) GetExitCodes() (map[string]int32, error) {
	ws := make(map[string]int32)

	pod, err := co.runtimeExec.GetPod()
	if err != nil {
		return ws, err
	}

	log.WithField("container statuses", pod.Status.ContainerStatuses).Debug()

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name != constants.ContainerCoordinatorName {
			if cs.State.Terminated == nil {
				log.Warningf("container %s not terminated.", cs.Name)
				continue
			}
			ws[cs.Name] = cs.State.Terminated.ExitCode
		}
	}

	return ws, nil
}

// CollectArtifacts collects workload artifacts.
func (co *Coordinator) CollectArtifacts() error {
	outputs, err := co.runtimeExec.GetStageOutputs(co.stageName)

	if err != nil {
		log.Errorf("Get stage %s output artifacts spec failed", co.stageName)
		return err
	}

	// Create the artifacts directory if not exist.
	createDirectory(constants.ArtifactsDir)

	for _, artifact := range outputs.Artifacts {
		dst := path.Join(constants.ArtifactsDir, artifact.Name)
		createDirectory(dst)
		if artifact.Container == "" {
			artifact.Container = co.workloadContainer
		}

		id, err := co.getContainerID(artifact.Container)
		if err != nil {
			log.Errorf("get container %s's id failed: %v", artifact.Container, err)
			return err
		}

		erra := co.runtimeExec.CopyArtifact(id, artifact.Path, dst)
		if erra != nil {
			log.Errorf("Copy container %s artifact %s failed: %v", artifact.Container, artifact.Name, erra)
			return err
		}

	}
	return nil

}

// CollectResources collects workload resources.
func (co *Coordinator) CollectResources() error {
	outputs, err := co.runtimeExec.GetStageOutputs(co.stageName)

	if err != nil {
		log.Errorf("Get stage %s output artifacts spec failed", co.stageName)
		return err
	}

	// Create the resources directory if not exist.
	createDirectory(constants.ResourcesDir)

	for _, resource := range outputs.Resources {
		dst := path.Join(constants.ResourcesDir, resource.Name)
		createDirectory(dst)

		id, err := co.getContainerID(co.workloadContainer)
		if err != nil {
			log.Errorf("get container %s's id failed: %v", co.workloadContainer, err)
			return err
		}

		erra := co.runtimeExec.CopyArtifact(id, resource.Path, dst)
		if erra != nil {
			log.Errorf("Copy container %s resources %s failed: %v", co.workloadContainer, resource.Name, erra)
			return err
		}

	}
	return nil

}

// NotifyResolvers create a file to notify output resolvers to start working.
func (co *Coordinator) NotifyResolvers() error {
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
	pod, err := co.runtimeExec.GetPod()
	if err != nil {
		return cs, err
	}

	for _, c := range pod.Spec.Containers {
		if c.Name != constants.ContainerCoordinatorName {
			cs = append(cs, c.Name)
		}
	}

	return cs, nil
}

func (co *Coordinator) getContainerID(name string) (string, error) {
	pod, err := co.runtimeExec.GetPod()
	if err != nil {
		return "", err
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == name {
			return refineContainerID(cs.ContainerID), nil
		}
	}

	return "", fmt.Errorf("Container %s not found", name)
}
