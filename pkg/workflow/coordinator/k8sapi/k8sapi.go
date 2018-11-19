package k8sapi

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/common"
)

type K8sapiExector struct {
	client     clientset.Interface
	kubeconfig string
	namespace  string
	podName    string
}

func NewK8sapiExector(n string, pod string, client clientset.Interface, kubecfg string) *K8sapiExector {
	return &K8sapiExector{
		namespace:  n,
		podName:    pod,
		client:     client,
		kubeconfig: kubecfg,
	}
}

// WaitContainersTerminate waits containers except for
// those whose name has prefix 'exceptsPrefix' to be 'expectState' status.
func (k *K8sapiExector) WaitContainers(expectState common.ContainerState, exceptsPrefix []string) error {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	log.Infof("Starting to wait for containers of pod %s to be %s ...", k.podName, expectState)
	for {
		select {
		case <-ticker.C:
			pod, err := k.client.CoreV1().Pods(k.namespace).Get(k.podName, meta_v1.GetOptions{})
			if err != nil {
				return err
			}

			containerNum := len(pod.Spec.Containers)
			expectNum := containerNum
			actualNum := 0
			for _, cs := range pod.Status.ContainerStatuses {
				for _, except := range exceptsPrefix {
					if strings.HasPrefix(cs.Name, except) {
						expectNum--
						continue
					}
				}

				switch expectState {
				case common.ContainerStateTerminated:
					// Check if container is terminated
					if cs.State.Terminated != nil {
						log.Debugf("Container %s is terminated: %v", cs.Name, cs.State.Terminated)
						actualNum++
					}
				case common.ContainerStateInitialized:
					// Check if container is not waiting
					if cs.State.Running != nil || cs.State.Terminated != nil {
						log.Debugf("Container %s is started: %v", cs.Name, cs.State.Terminated)
						actualNum++
					}
				}
			}

			if expectNum == actualNum {
				log.Infof("End of waiting for containers of pod %s to be %s.", k.podName, expectState)
				return nil
			}

		}
	}

	return nil
}

// GetPod get the stage pod.
func (k *K8sapiExector) GetPod() (*core_v1.Pod, error) {
	return k.client.CoreV1().Pods(k.namespace).Get(k.podName, meta_v1.GetOptions{})

}

// KillContainer kills a container of a pod.
func (k *K8sapiExector) KillContainer(name string) error {
	// TODO

	return nil
}

// CollectLog collects container logs.
func (k *K8sapiExector) CollectLog(name, path string) error {
	log.Infof("Start to collect %s log to %s:", name, path)
	stream, err := k.client.CoreV1().Pods(k.namespace).GetLogs(k.podName, &core_v1.PodLogOptions{
		Container: name,
		Follow:    true,
	}).Stream()
	if err != nil {
		return err
	}
	defer stream.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	if err != nil {
		return err
	}
	return nil
}

// GetStageOutputs get outputs of a stage.
func (k *K8sapiExector) GetStageOutputs(name string) (v1alpha1.Outputs, error) {
	stage, err := k.client.CycloneV1alpha1().Stages(k.namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return v1alpha1.Outputs{}, err
	}

	return stage.Spec.Pod.Outputs, nil
}

// CopyFromContainer copy a file/directory frome container:path to dst.
func (k *K8sapiExector) CopyFromContainer(container, path, dst string) error {
	//args := []string{"--kubeconfig", k.kubeconfig, "cp", fmt.Sprintf("%s/%s:%s", k.namespace, k.podName, path), "-c", container, dst}
	//
	//cmd := exec.Command("kubectl", args...)
	//return cmd.Run()

	// Fixme, use docker instead of kubectl since
	// kubectl can not cp a file from a stopped container.
	args := []string{"cp", fmt.Sprintf("%s:%s", container, path), dst}

	cmd := exec.Command("docker", args...)
	log.WithField("args", args).Info()
	ret, err := cmd.CombinedOutput()
	log.WithField("message", string(ret)).WithField("error", err).Info()
	return err
}
