package coordinator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

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
	runtimeExec RuntimeExecutor
	// workloadContainer represents name of the workload container.
	workloadContainer string
	// Stage related to this pod.
	Stage *v1alpha1.Stage
	// Wfr represents the WorkflowRun which triggered this pod.
	Wfr *v1alpha1.WorkflowRun
	// OutputResources represents output resources the related stage configured.
	OutputResources []*v1alpha1.Resource
}

// RuntimeExecutor is an interface defined some methods
// to communicate with k8s container runtime.
type RuntimeExecutor interface {
	// WaitContainers waits selected containers to state.
	WaitContainers(state common.ContainerState, selectors ...common.ContainerSelector) error
	// CollectLog collects container logs to cyclone server.
	CollectLog(container, wrorkflowrun, stage string) error
	// CopyFromContainer copy a file or directory from container:path to dst.
	CopyFromContainer(container, path, dst string) error
	// GetPod get the stage related pod.
	GetPod() (*core_v1.Pod, error)
	// SetResults set results (key-values) to the pod, workflow controller would sync this result
	// to WorkflowRun status.
	SetResults(values []v1alpha1.KeyValue) error
}

// NewCoordinator create a coordinator instance.
func NewCoordinator(client clientset.Interface) (*Coordinator, error) {
	// Get stage from Env
	var stage *v1alpha1.Stage
	stageInfo := os.Getenv(common.EnvStageInfo)
	if stageInfo == "" {
		return nil, fmt.Errorf("get stage info from env failed")
	}

	err := json.Unmarshal([]byte(stageInfo), &stage)
	if err != nil {
		return nil, fmt.Errorf("unmarshal stage info error %s", err)
	}

	// Get workflowrun from Env
	var wfr *v1alpha1.WorkflowRun
	wfrInfo := os.Getenv(common.EnvWorkflowRunInfo)
	if stageInfo == "" {
		return nil, fmt.Errorf("get workflowrun info from env failed")
	}

	err = json.Unmarshal([]byte(wfrInfo), &wfr)
	if err != nil {
		return nil, fmt.Errorf("unmarshal workflowrun info error %s", err)
	}

	// Get output resources from Env
	var rscs []*v1alpha1.Resource
	rscInfo := os.Getenv(common.EnvOutputResourcesInfo)
	if rscInfo == "" {
		return nil, fmt.Errorf("get output resources info from env failed")
	}

	err = json.Unmarshal([]byte(rscInfo), &rscs)
	if err != nil {
		return nil, fmt.Errorf("unmarshal output resources info error %s", err)
	}

	return &Coordinator{
		runtimeExec:       k8sapi.NewK8sapiExecutor(client, wfr.Namespace, getNamespace(), getPodName(), getCycloneServerAddr()),
		workloadContainer: getWorkloadContainer(),
		Stage:             stage,
		Wfr:               wfr,
		OutputResources:   rscs,
	}, nil
}

// CollectLogs collects all containers' logs.
func (co *Coordinator) CollectLogs() error {
	cs, err := co.getAllContainers()
	if err != nil {
		return err
	}

	for _, c := range cs {
		go func(container, workflowrun, stage string) {
			err := co.runtimeExec.CollectLog(container, workflowrun, stage)
			if err != nil {
				log.Errorf("Collect %s log failed:%v", container, err)
			}
		}(c, co.Wfr.Name, co.Stage.Name)
	}

	return nil
}

// WaitRunning waits all containers to start run.
func (co *Coordinator) WaitRunning() error {
	err := co.runtimeExec.WaitContainers(common.ContainerStateInitialized)
	if err != nil {
		log.Errorf("Wait containers to running error: %v", err)
	}
	return err
}

// WaitWorkloadTerminate waits all workload containers to be Terminated status.
func (co *Coordinator) WaitWorkloadTerminate() error {
	err := co.runtimeExec.WaitContainers(common.ContainerStateTerminated, common.OnlyWorkload)
	if err != nil {
		log.Errorf("Wait containers to completion error: %v", err)
	}
	return err
}

// WaitAllOthersTerminate waits all containers except for
// the coordinator container itself to become Terminated status.
func (co *Coordinator) WaitAllOthersTerminate() error {
	err := co.runtimeExec.WaitContainers(common.ContainerStateTerminated,
		common.NonWorkloadSidecar, common.NonCoordinator, common.NonDockerInDocker)
	if err != nil {
		log.Errorf("Wait containers to completion error: %v", err)
	}
	return err
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

// GetExitCodes gets exit codes of containers passed the selector
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
		return fmt.Errorf("get stage output artifacts failed, stage pod nil")
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
		return fmt.Errorf("get stage output resources failed, stage pod nil")
	}

	resources := co.Stage.Spec.Pod.Outputs.Resources
	if len(resources) == 0 {
		log.Info("output resources empty, no need to collect.")
		return nil
	}

	log.WithField("resources", resources).Info("start to collect.")

	// Create the resources directory if not exist.
	fileutil.CreateDirectory(common.CoordinatorResourcesPath)

	for _, resource := range resources {
		for _, r := range co.OutputResources {
			if r.Name == resource.Name {
				// If the resource is persisted in PVC, no need to copy here, Cyclone
				// will mount it to resolver container directly.
				if r.Spec.Persistent != nil {
					continue
				}
			}
		}

		if len(resource.Path) == 0 {
			continue
		}

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
		return fmt.Errorf("get stage output resources failed, stage pod nil")
	}

	resources := co.Stage.Spec.Pod.Outputs.Resources
	if len(resources) == 0 {
		log.Info("output resources empty, no need to notify resolver.")
		return nil
	}

	log.WithField("resources", resources).Info("start to notify resolver.")

	exist := fileutil.CreateDirectory(common.CoordinatorResolverNotifyPath)
	log.WithField("exist", exist).WithField("notifydir", common.CoordinatorResolverNotifyPath).Info()

	_, err := os.Create(common.CoordinatorResolverNotifyOkPath)
	if err != nil {
		log.WithField("file", common.CoordinatorResolverNotifyOkPath).Error("Create ok file error: ", err)
		return err
	}

	return nil
}

func (co *Coordinator) getAllContainers() ([]string, error) {
	var cs []string
	pod, err := co.runtimeExec.GetPod()
	if err != nil {
		return cs, err
	}

	for _, c := range pod.Spec.InitContainers {
		cs = append(cs, c.Name)
	}
	for _, c := range pod.Spec.Containers {
		cs = append(cs, c.Name)
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

// CollectExecutionResults collects execution results (key-values) and store them in pod's annotation
func (co *Coordinator) CollectExecutionResults() error {
	pod, err := co.runtimeExec.GetPod()
	if err != nil {
		return err
	}

	for _, c := range pod.Spec.Containers {
		if !common.OnlyWorkload(c.Name) {
			continue
		}

		dst := fmt.Sprintf("/tmp/__result__%s", c.Name)
		containerID, err := co.getContainerID(c.Name)
		if err != nil {
			log.WithField("c", containerID).Error("Get container ID error: ", err)
			return err
		}
		err = co.runtimeExec.CopyFromContainer(containerID, common.ResultFilePath, dst)
		if isFileNotExist(err) {
			continue
		}

		if err != nil {
			return err
		}

		b, err := ioutil.ReadFile(dst)
		if err != nil {
			return err
		}
		log.Info("Result file content: ", string(b))

		var keyValues []v1alpha1.KeyValue
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				log.Warn("Invalid result item: ", line)
				continue
			}
			log.Info("Result item: ", line)

			keyValues = append(keyValues, v1alpha1.KeyValue{
				Key:   parts[0],
				Value: parts[1],
			})
		}

		if len(keyValues) > 0 {
			log.Info("To set execution result")
			if err := co.runtimeExec.SetResults(keyValues); err != nil {
				return err
			}
		}
	}

	return nil
}

func isFileNotExist(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "No such container:path")
}
