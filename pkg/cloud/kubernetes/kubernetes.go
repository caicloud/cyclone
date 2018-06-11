/*
Copyright 2018 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"fmt"
	"time"

	log "github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/cloud"
	"github.com/caicloud/cyclone/pkg/wait"
	"github.com/caicloud/cyclone/pkg/worker/scm"
)

func init() {
	if err := cloud.RegistryCloudProvider(api.CloudTypeKubernetes, NewK8sCloud); err != nil {
		log.Errorln(err)
	}
}

type k8sCloud struct {
	client    *kubernetes.Clientset
	namespace string
}

func NewK8sCloud(c *api.Cloud) (cloud.Provider, error) {
	if c.Type != api.CloudTypeKubernetes {
		err := fmt.Errorf("fail to new k8s cloud as cloud type %s is not %s", c.Type, api.CloudTypeKubernetes)
		log.Error(err)
		return nil, err
	}

	var ck *api.CloudKubernetes
	if c.Kubernetes == nil {
		err := fmt.Errorf("k8s cloud %s is empty", c.Name)
		log.Error(err)
		return nil, err
	} else {
		ck = c.Kubernetes
	}

	if ck.InCluster {
		return newInclusterK8sCloud(ck)
	} else {
		return newK8sCloud(ck, c.Insecure)
	}
}

func newK8sCloud(c *api.CloudKubernetes, insecure bool) (cloud.Provider, error) {
	config := &rest.Config{
		Host:        c.Host,
		BearerToken: c.BearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: insecure,
		},
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &k8sCloud{client, c.Namespace}, nil
}

func newInclusterK8sCloud(c *api.CloudKubernetes) (cloud.Provider, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &k8sCloud{client, c.Namespace}, nil
}

func (c *k8sCloud) Resource() (*options.Resource, error) {
	quotas, err := c.client.CoreV1().ResourceQuotas(c.namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	resource := options.NewResource()
	if len(quotas.Items) == 0 {
		// TODO get quota from metrics
		return resource, nil
	}

	quota := quotas.Items[0]
	for k, v := range quota.Status.Hard {
		resource.Limit[string(k)] = options.NewQuantityFor(v)
	}

	for k, v := range quota.Status.Used {
		resource.Used[string(k)] = options.NewQuantityFor(v)
	}

	return resource, nil
}

func (c *k8sCloud) CanProvision(quota options.Quota) (bool, error) {
	resource, err := c.Resource()
	if err != nil {
		return false, err
	}
	if resource.Limit.IsZero() {
		return true, nil
	}

	if resource.Limit.Enough(resource.Used, quota) {
		return true, nil
	}

	return false, nil
}

func (c *k8sCloud) Provision(info *api.WorkerInfo, opts *options.WorkerOptions) (*api.WorkerInfo, error) {
	log.Infof("create worker with info: %v; opts: %v", info, opts)

	can, err := c.CanProvision(opts.Quota)
	if err != nil {
		return nil, err
	}

	if !can {
		// wait
		return nil, cloud.ErrNoEnoughResource
	}

	eventID := opts.EventID
	namespace := c.namespace
	Privileged := true
	pod := &apiv1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: namespace,
			Name:      info.Name,
			Labels: map[string]string{
				"cyclone":    "worker",
				"cyclone/id": eventID,
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:            "cyclone-worker",
					Image:           opts.WorkerImage,
					Env:             buildK8SEnv(eventID, *opts),
					WorkingDir:      scm.GetCloneDir(),
					Resources:       opts.Quota.ToK8SQuota(),
					SecurityContext: &apiv1.SecurityContext{Privileged: &Privileged},
					ImagePullPolicy: apiv1.PullAlways,
				},
			},
			RestartPolicy: apiv1.RestartPolicyNever,
		},
	}

	// Mount the cache volume to the worker.
	cacheVolume := info.CacheVolume
	mountPath := info.MountPath
	if len(cacheVolume) != 0 && len(mountPath) != 0 {
		// Check the existence and status of cache volume.
		if pvc, err := c.client.CoreV1().PersistentVolumeClaims(namespace).Get(cacheVolume, meta_v1.GetOptions{}); err == nil {
			if pvc.Status.Phase == apiv1.ClaimBound {
				volumeName := "cache-dependency"
				pod.Spec.Containers[0].VolumeMounts = []apiv1.VolumeMount{
					apiv1.VolumeMount{
						Name:      volumeName,
						MountPath: mountPath,
					},
				}

				pod.Spec.Volumes = []apiv1.Volume{
					apiv1.Volume{
						Name: volumeName,
					},
				}

				pod.Spec.Volumes[0].PersistentVolumeClaim = &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: cacheVolume,
				}
			} else {
				// Just log error and let the pipeline to run in non-cache mode.
				log.Errorf("Can not use cache volume %s as its status is %v", cacheVolume, pvc.Status.Phase)
			}
		} else {
			// Just log error and let the pipeline to run in non-cache mode.
			log.Errorf("Can not use cache volume %s as fail to get it: %v", cacheVolume, err)
		}
	}

	check := func() (bool, error) {
		// change pod here
		pod, err := c.client.CoreV1().Pods(namespace).Get(pod.Name, meta_v1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case apiv1.PodPending:
			return false, nil
		case apiv1.PodRunning:
			return true, nil
		default:
			return false, fmt.Errorf("get an error pod status phase[%s]", pod.Status.Phase)
		}
	}

	pod, err = c.client.CoreV1().Pods(namespace).Create(pod)
	if err != nil {
		log.Errorf("fail to create worker pod as %v", err)
		return nil, err
	}

	// wait until pod is running
	err = wait.Poll(7*time.Second, 2*time.Minute, check)
	if err != nil {
		log.Errorf("timeout to wait worker pod to be running as %v", err)
		return nil, err
	}

	// add time
	info.StartTime = time.Now()
	info.DueTime = info.StartTime.Add(time.Duration(cloud.WorkerTimeout))

	info.Name = pod.Name
	return info, nil
}

func (c *k8sCloud) Ping() error {
	_, err := c.client.CoreV1().Pods(apiv1.NamespaceDefault).List(meta_v1.ListOptions{})

	return err
}

func (c *k8sCloud) TerminateWorker(name string) error {
	return c.client.CoreV1().Pods(c.namespace).Delete(name, &meta_v1.DeleteOptions{})
}

func buildK8SEnv(id string, opts options.WorkerOptions) []apiv1.EnvVar {

	env := []apiv1.EnvVar{
		{
			Name:  options.EventID,
			Value: id,
		},
		{
			Name:  options.CycloneServer,
			Value: opts.CycloneServer,
		},
		{
			Name:  options.ConsoleWebEndpoint,
			Value: opts.ConsoleWebEndpoint,
		},
		{
			Name:  options.RegistryLocation,
			Value: opts.RegistryLocation,
		},
		{
			Name:  options.RegistryUsername,
			Value: opts.RegistryUsername,
		},
		{
			Name:  options.RegistryPassword,
			Value: opts.RegistryPassword,
		},
		{
			Name:  options.GitlabURL,
			Value: opts.GitlabURL,
		},
		{
			Name:  options.WorkerImage,
			Value: opts.WorkerImage,
		},
		{
			Name:  options.LimitCPU,
			Value: opts.Quota[options.ResourceLimitsCPU].String(),
		},
		{
			Name:  options.LimitMemory,
			Value: opts.Quota[options.ResourceLimitsMemory].String(),
		},
		{
			Name:  options.RequestCPU,
			Value: opts.Quota[options.ResourceRequestsCPU].String(),
		},
		{
			Name:  options.RequestMemory,
			Value: opts.Quota[options.ResourceRequestsMemory].String(),
		},
	}

	return env
}
