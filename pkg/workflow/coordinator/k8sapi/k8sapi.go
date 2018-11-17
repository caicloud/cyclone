package k8sapi

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/coordinator/common"
)

type K8sapiExector struct {
	client    clientset.Interface
	namespace string
	podName   string
}

func NewK8sapiExector(n string, pod string, client clientset.Interface) *K8sapiExector {
	return &K8sapiExector{
		namespace: n,
		podName:   pod,
		client:    client,
	}
}

// WaitContainersTerminate waits containers except for 'excepts' to be 'expectState' status.
func (k *K8sapiExector) WaitContainers(timeout time.Duration, expectState common.ContainerState, excepts []string) error {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	log.Infof("Starting to wait for containers of pod %s to be %s ...", k.podName, expectState)
	for {
		select {
		case <-ticker.C:
			pod, err := k.client.CoreV1().Pods(k.namespace).Get(k.podName, meta_v1.GetOptions{})
			if err != nil {
				return err
			}

			containerNum := len(pod.Spec.Containers)
			expectTerminatedNum := containerNum
			actualNum := 0
			for _, cs := range pod.Status.ContainerStatuses {
				for _, except := range excepts {
					if cs.Name == except {
						expectTerminatedNum--
						continue
					}
				}

				switch expectState {
				case common.ContainerStateTerminated:
					// Check if container is terminated
					if cs.State.Terminated != nil {
						log.Infof("Container %s is terminated: %v", cs.Name, cs.State.Terminated)
						actualNum++
					}
				case common.ContainerStateNotWaiting:
					// Check if container is not waiting
					if cs.State.Running != nil || cs.State.Terminated != nil {
						log.Infof("Container %s is started: %v", cs.Name, cs.State.Terminated)
						actualNum++
					}
				}
			}

			if expectTerminatedNum == actualNum {
				log.Infof("End of waiting for containers of pod %s to be %s.", k.podName, expectState)
				return nil
			}

		case <-timer.C:
			return fmt.Errorf("Timeout after %s", timeout.String())
		}
	}

	return nil
}

// GetAllContainers get all containers within a pod.
func (k *K8sapiExector) GetAllContainers() ([]string, error) {
	var cs []string
	pod, err := k.client.CoreV1().Pods(k.namespace).Get(k.podName, meta_v1.GetOptions{})
	if err != nil {
		return cs, err
	}

	for _, c := range pod.Spec.Containers {
		cs = append(cs, c.Name)
	}

	return cs, nil
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
