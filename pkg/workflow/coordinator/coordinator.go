package coordinator

import (
	"fmt"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"

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
	workloadContainer string
	// Stage which this run pod belonged to.
	Stage *v1alpha1.Stage
	// WorkflowRun which triggered this run pod.
	Wfr *v1alpha1.WorkflowRun
	// Event recorder.
	Recorder record.EventRecorder
}

// RuntimeExecutor is an interface defined some methods
// to communicate with k8s container runtime.
type RuntimeExecutor interface {
	WaitContainers(state common.ContainerState, selectors ...common.ContainerSelector) error
	CollectLog(container, wrorkflowrun, stage string) error
	CopyFromContainer(container, path, dst string) error
	GetPod() (*core_v1.Pod, error)
}

// NewCoordinator create a coordinator instance.
func NewCoordinator(client clientset.Interface, kubecfg string) (*Coordinator, error) {
	namespace := getNamespace()
	stageName := getStageName()
	wfrName := getWorkflowrunName()
	var stage *v1alpha1.Stage
	var wfr *v1alpha1.WorkflowRun

	backoff := wait.Backoff{
		Duration: time.Duration(time.Second),
		Factor:   2,
		Jitter:   1,
		Steps:    3,
	}

	err := wait.ExponentialBackoff(backoff, func() (done bool, err error) {
		stage, err = client.CycloneV1alpha1().Stages(namespace).Get(stageName, meta_v1.GetOptions{})
		if err != nil {
			log.WithField("error", err).
				WithField("stageName", stageName).Info("Get stage error")
			return false, nil
		}

		wfr, err = client.CycloneV1alpha1().WorkflowRuns(namespace).Get(wfrName, meta_v1.GetOptions{})
		if err != nil {
			log.WithField("error", err).
				WithField("wfrName", wfrName).Info("Get wfr error")
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		log.WithField("stageName", stageName).WithField("wfrName", wfrName).
			WithField("error", err).Error("Get stage or workflowrun failed")
		return nil, err
	}

	return &Coordinator{
		runtimeExec:       k8sapi.NewK8sapiExecutor(namespace, getPodName(), client, getCycloneServerAddr(), kubecfg),
		workloadContainer: getWorkloadContainer(),
		Stage:             stage,
		Wfr:               wfr,
		Recorder:          common.GetEventRecorder(client, common.EventSourceCoordinator),
	}, nil
}

// CollectLogs collects all containers' logs except for the coordinator container itself.
func (co *Coordinator) CollectLogs() {
	cs, err := co.getAllOtherContainers()
	if err != nil {
		return
	}

	for _, c := range cs {
		go func(container, workflowrun, stage string) {
			err := co.runtimeExec.CollectLog(container, workflowrun, stage)
			if err != nil {
				log.Errorf("Collect %s log failed:%v", c, err)
			}
		}(c, co.Wfr.Name, co.Stage.Name)
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

// StageSuccess checks if the workload and resolver containers are succeeded.
func (co *Coordinator) StageSuccess() bool {
	ws, err := co.GetExitCodes(common.NonCoordinator, common.NonWorkloadSidecar)
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

// WorkLoadSuccess checks if the workload containers are succeeded.
func (co *Coordinator) WorkLoadSuccess() bool {
	ws, err := co.GetExitCodes(common.OnlyWorkload)
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

// Get exit codes of containers passed the selector
func (co *Coordinator) GetExitCodes(selectors ...common.ContainerSelector) (map[string]int32, error) {
	ws := make(map[string]int32)

	pod, err := co.runtimeExec.GetPod()
	if err != nil {
		return ws, err
	}

	log.WithField("container statuses", pod.Status.ContainerStatuses).Debug()

	for _, cs := range pod.Status.ContainerStatuses {
		if common.Pass(cs.Name, selectors) {
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
	if co.Stage.Spec.Pod == nil {
		return fmt.Errorf("Get stage output artifacts failed, stage pod nil.")
	}

	artifacts := co.Stage.Spec.Pod.Outputs.Artifacts
	if len(artifacts) == 0 {
		log.Info("output artifacts empty, no need to collect.")
		return nil
	}

	log.WithField("artifacts", artifacts).Info("start to collect.")

	// Create the artifacts directory if not exist.
	fileutil.CreateDirectory(common.CoordinatorArtifactsPath)

	for _, artifact := range artifacts {
		dst := path.Join(common.CoordinatorArtifactsPath, artifact.Name)
		fileutil.CreateDirectory(dst)

		id, err := co.getContainerID(co.workloadContainer)
		if err != nil {
			log.Errorf("get container %s's id failed: %v", co.workloadContainer, err)
			return err
		}

		err = co.runtimeExec.CopyFromContainer(id, artifact.Path, dst)
		if err != nil {
			log.Errorf("Copy container %s artifact %s failed: %v", co.workloadContainer, artifact.Name, err)
			return err
		}
	}

	return nil
}

// CollectResources collects workload resources.
func (co *Coordinator) CollectResources() error {
	if co.Stage.Spec.Pod == nil {
		return fmt.Errorf("Get stage output resources failed, stage pod nil.")
	}

	reources := co.Stage.Spec.Pod.Outputs.Resources
	if len(reources) == 0 {
		log.Info("output resources empty, no need to collect.")
		return nil
	}

	log.WithField("resources", reources).Info("start to collect.")

	// Create the resources directory if not exist.
	fileutil.CreateDirectory(common.CoordinatorResourcesPath)

	for _, resource := range reources {
		dst := path.Join(common.CoordinatorResourcesPath, resource.Name)
		fileutil.CreateDirectory(dst)

		id, err := co.getContainerID(co.workloadContainer)
		if err != nil {
			log.Errorf("get container %s's id failed: %v", co.workloadContainer, err)
			return err
		}

		err = co.runtimeExec.CopyFromContainer(id, resource.Path, dst)
		if err != nil {
			log.Errorf("Copy container %s resources %s failed: %v", co.workloadContainer, resource.Name, err)
			return err
		}
	}

	return nil
}

// NotifyResolvers create a file to notify output resolvers to start working.
func (co *Coordinator) NotifyResolvers() error {
	if co.Stage.Spec.Pod == nil {
		return fmt.Errorf("Get stage output resources failed, stage pod nil.")
	}

	reources := co.Stage.Spec.Pod.Outputs.Resources
	if len(reources) == 0 {
		log.Info("output resources empty, no need to notify resolver.")
		return nil
	}

	log.WithField("resources", reources).Info("start to notify resolver.")

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
