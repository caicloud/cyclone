package coordinator

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	fileutil "github.com/caicloud/cyclone/pkg/util/file"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/k8sapi"
)

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
	WaitContainers(state common.ContainerState, selectors ...common.ContainerSelector) error
	CollectLog(container, wrorkflowrun, stage string) error
	GetStageOutputs(name string) (v1alpha1.Outputs, error)
	CopyFromContainer(container, path, dst string) error
	GetPod() (*core_v1.Pod, error)
}

// NewCoordinator create a coordinator instance.
func NewCoordinator(client clientset.Interface, kubecfg string) *Coordinator {
	return &Coordinator{
		runtimeExec:       k8sapi.NewK8sapiExecutor(getNamespace(), getPodName(), client, getCycloneServerAddr(), kubecfg),
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
		go func(container, workflowrun, stage string) {
			errl := co.runtimeExec.CollectLog(container, workflowrun, stage)
			if errl != nil {
				log.Errorf("Collect %s log failed:%v", c, errl)
			}
		}(c, co.workflowrunName, co.stageName)
	}

}

// WaitRunning waits all containers to start run.
func (co *Coordinator) WaitRunning() {
	err := co.runtimeExec.WaitContainers(common.ContainerStateInitialized)
	if err != nil {
		log.Errorf("Wait containers to running error: %v", err)
		return
	}
}

// WaitWorkloadTerminate waits all workload containers to be Terminated status.
func (co *Coordinator) WaitWorkloadTerminate() {
	err := co.runtimeExec.WaitContainers(common.ContainerStateTerminated, common.OnlyWorkload)
	if err != nil {
		log.Errorf("Wait containers to completion error: %v", err)
		return
	}
}

// WaitAllOthersTerminate waits all containers except for
// the coordinator container itself to become Terminated status.
func (co *Coordinator) WaitAllOthersTerminate() {
	err := co.runtimeExec.WaitContainers(common.ContainerStateTerminated, common.NonWorkloadSidecar, common.NonCoordinator)
	if err != nil {
		log.Errorf("Wait containers to completion error: %v", err)
		return
	}
}

// IsStageSuccess checks if the workload and resolver containers are succeeded.
func (co *Coordinator) IsStageSuccess() bool {
	ws, err := co.GetExitCodes()
	if err != nil {
		log.Errorf("Get Exit Codes failed: %v", err)
		return false
	}

	log.WithField("codes", ws).Debug("Get containers exit codes")

	for _, code := range ws {
		if code != 0 {
			return false
		}
	}

	return true
}

// Get workload and resolver containers' exit code.
func (co *Coordinator) GetExitCodes() (map[string]int32, error) {
	ws := make(map[string]int32)

	pod, err := co.runtimeExec.GetPod()
	if err != nil {
		return ws, err
	}

	log.WithField("container statuses", pod.Status.ContainerStatuses).Debug()

	for _, cs := range pod.Status.ContainerStatuses {
		if common.NonCoordinator(cs.Name) && common.NonWorkloadSidecar(cs.Name) {
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
	fileutil.CreateDirectory(common.CoordinatorArtifactsPath)

	log.WithField("artifacts", outputs.Artifacts).Info("start to collect.")
	for _, artifact := range outputs.Artifacts {
		dst := path.Join(common.CoordinatorArtifactsPath, artifact.Name)
		fileutil.CreateDirectory(dst)

		id, err := co.getContainerID(co.workloadContainer)
		if err != nil {
			log.Errorf("get container %s's id failed: %v", co.workloadContainer, err)
			return err
		}

		erra := co.runtimeExec.CopyFromContainer(id, artifact.Path, dst)
		if erra != nil {
			log.Errorf("Copy container %s artifact %s failed: %v", co.workloadContainer, artifact.Name, erra)
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
	fileutil.CreateDirectory(common.CoordinatorResourcesPath)

	log.WithField("resources", outputs.Resources).Info("start to collect.")
	for _, resource := range outputs.Resources {
		dst := path.Join(common.CoordinatorResourcesPath, resource.Name)
		fileutil.CreateDirectory(dst)

		id, err := co.getContainerID(co.workloadContainer)
		if err != nil {
			log.Errorf("get container %s's id failed: %v", co.workloadContainer, err)
			return err
		}

		erra := co.runtimeExec.CopyFromContainer(id, resource.Path, dst)
		if erra != nil {
			log.Errorf("Copy container %s resources %s failed: %v", co.workloadContainer, resource.Name, erra)
			return err
		}

	}
	return nil

}

// NotifyResolvers create a file to notify output resolvers to start working.
func (co *Coordinator) NotifyResolvers() error {
	exist := fileutil.CreateDirectory(common.CoordinatorResolverNotifyPath)
	log.WithField("exist", exist).WithField("notifydir", common.CoordinatorResolverNotifyPath).Info()

	_, err := os.Create(common.CoordinatorResolverNotifyOkPath)
	if err != nil {
		log.WithField("file", common.CoordinatorResolverNotifyOkPath).Error("Create ok file error: ", err)
		return err
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
		if c.Name != common.CoordinatorSidecarName {
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

	return "", fmt.Errorf("container %s not found", name)
}
