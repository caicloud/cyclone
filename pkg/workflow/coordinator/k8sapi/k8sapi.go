package k8sapi

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	cyclone_common "github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/cycloneserver"
)

// Executor ...
type Executor struct {
	client        clientset.Interface
	metaNamespace string
	namespace     string
	podName       string
	cycloneClient cycloneserver.Client
}

// NewK8sapiExecutor ...
func NewK8sapiExecutor(client clientset.Interface, metaNamespace, namespace, pod string, cycloneServer string) *Executor {
	return &Executor{
		metaNamespace: metaNamespace,
		namespace:     namespace,
		podName:       pod,
		client:        client,
		cycloneClient: cycloneserver.NewClient(cycloneServer),
	}
}

// WaitContainers waits containers that pass selectors.
func (k *Executor) WaitContainers(expectState common.ContainerState, selectors ...common.ContainerSelector) error {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	log.Infof("Starting to wait for containers of pod %s to be %s ...", k.podName, expectState)
	for range ticker.C {
		pod, err := k.client.CoreV1().Pods(k.namespace).Get(k.podName, meta_v1.GetOptions{})
		if err != nil {
			log.WithField("ns", k.namespace).WithField("pod", k.podName).Error("get pod failed")
			return err
		}

		var reachGoals = true
		for _, c := range pod.Spec.Containers {
			// Skip containers that are not selected.
			if !common.Pass(c.Name, selectors) {
				continue
			}

			var s *core_v1.ContainerStatus
			for _, cs := range pod.Status.ContainerStatuses {
				if c.Name == cs.Name {
					s = &cs
					break
				}
			}

			switch expectState {
			case common.ContainerStateTerminated:
				if s == nil || s.State.Terminated == nil {
					log.WithField("container", c.Name).WithField("expected", expectState).Debugf("Container not expected status")
					reachGoals = false
				}
			case common.ContainerStateInitialized:
				if s == nil || (s.State.Running == nil && s.State.Terminated == nil) {
					log.WithField("container", c.Name).WithField("expected", expectState).Debugf("Container not in expected status")
					reachGoals = false
				}
			default:
				return fmt.Errorf("Unsupported state: %s, Only support: %s, %s", expectState, common.ContainerStateTerminated, common.ContainerStateInitialized)
			}
		}

		if reachGoals {
			log.WithField("pod", pod.Name).WithField("expected", expectState).Info("All containers reached expected status")
			return nil
		}
	}

	return nil
}

// GetPod get the stage pod.
func (k *Executor) GetPod() (*core_v1.Pod, error) {
	return k.client.CoreV1().Pods(k.namespace).Get(k.podName, meta_v1.GetOptions{})
}

// CollectLog collects container logs.
func (k *Executor) CollectLog(container, workflowrun, stage string) error {
	log.Infof("Start to collect %s log", container)
	stream, err := k.client.CoreV1().Pods(k.namespace).GetLogs(k.podName, &core_v1.PodLogOptions{
		Container: container,
		Follow:    true,
	}).Stream()
	if err != nil {
		return err
	}

	defer func() {
		stream.Close()
	}()

	err = k.cycloneClient.PushLogStream(k.metaNamespace, workflowrun, stage, container, stream)
	if err != nil {
		return err
	}
	return nil
}

// MarkLogEOF marks the end of stage logs
func (k *Executor) MarkLogEOF(workflowrun, stage string) error {
	err := k.cycloneClient.PushLogStream(k.metaNamespace, workflowrun, stage, cyclone_common.FolderEOFFile, strings.NewReader(""))
	if err != nil {
		return err
	}
	return nil
}

// CopyFromContainer copy a file/directory from container:path to dst.
func (k *Executor) CopyFromContainer(container, path, dst string) error {
	args := []string{"cp", fmt.Sprintf("%s:%s", container, path), dst}

	cmd := exec.Command("docker", args...)
	log.WithField("args", args).Info()
	ret, err := cmd.CombinedOutput()
	log.WithField("message", string(ret)).WithField("error", err).Info("copy file result")
	if err != nil {
		return fmt.Errorf("%s, error: %v", string(ret), err)
	}

	return nil
}

// SetResults sets execution results (key-values) to the pod, workflow controller will sync this result to WorkflowRun status.
func (k *Executor) SetResults(values []v1alpha1.KeyValue) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pod, err := k.client.CoreV1().Pods(k.namespace).Get(k.podName, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		b, err := json.Marshal(values)
		if err != nil {
			return err
		}

		pod.Annotations[meta.AnnotationStageResult] = string(b)
		_, err = k.client.CoreV1().Pods(k.namespace).Update(pod)
		return err
	})
}
