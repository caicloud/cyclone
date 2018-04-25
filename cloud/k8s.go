/*
Copyright 2016 caicloud authors. All rights reserved.

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

package cloud

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/zoumo/logdog"
	apiv1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8SCloud ...
type K8SCloud struct {
	name        string
	host        string
	bearerToken string
	namespace   string
	insecure    bool
	inCluster   bool
	client      *kubernetes.Clientset
}

// NewK8SCloud ...
func NewK8SCloud(opts Options) (Cloud, error) {

	if opts.K8SInCluster == true {
		return NewK8SCloudInCluster(opts)
	}

	return newK8SCloud(opts)
}

// newK8SCloud returns a cloud object which uses the Options
func newK8SCloud(opts Options) (Cloud, error) {

	if opts.Name == "" {
		return nil, errors.New("K8SCloud: Invalid cloud name")
	}
	if opts.Host == "" {
		return nil, errors.New("K8SCloud: Invalid cloud host")
	}
	if opts.K8SNamespace == "" {
		opts.K8SNamespace = apiv1.NamespaceDefault
	}

	cloud := &K8SCloud{
		name:        opts.Name,
		host:        opts.Host,
		bearerToken: opts.K8SBearerToken,
		namespace:   opts.K8SNamespace,
		insecure:    opts.Insecure,
	}
	config := &rest.Config{
		Host:        opts.Host,
		BearerToken: opts.K8SBearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: opts.Insecure,
		},
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	cloud.client = clientset
	return cloud, nil
}

// NewK8SCloudInCluster returns a cloud object which uses the service account
// kubernetes gives to pods
func NewK8SCloudInCluster(opts Options) (Cloud, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + apiv1.ServiceAccountNamespaceKey)
	if err != nil {
		return nil, err
	}

	cloud := &K8SCloud{
		name:        opts.Name,
		host:        config.Host,
		bearerToken: config.BearerToken,
		namespace:   string(namespace),
		insecure:    opts.Insecure,
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	cloud.client = clientset
	return cloud, nil
}

// Client returns k8s clientset
func (cloud *K8SCloud) Client() *kubernetes.Clientset {
	return cloud.client
}

// Name returns k8s cloud name
func (cloud *K8SCloud) Name() string {
	return cloud.name
}

// Kind returns cloud type.
func (cloud *K8SCloud) Kind() string {
	return KindK8SCloud
}

// Ping returns nil if cloud is accessible
func (cloud *K8SCloud) Ping() error {
	_, err := cloud.client.CoreV1().Pods(cloud.namespace).List(meta_v1.ListOptions{})
	return err
}

// Resource returns the limit and used quotas of the cloud
func (cloud *K8SCloud) Resource() (*Resource, error) {
	quotas, err := cloud.client.CoreV1().ResourceQuotas(cloud.namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	resource := &Resource{
		Limit: ZeroQuota.DeepCopy(),
		Used:  ZeroQuota.DeepCopy(),
	}

	if len(quotas.Items) == 0 {
		// TODO get quota from metrics
		return resource, nil
	}

	quota := quotas.Items[0]
	for k, v := range quota.Status.Hard {
		resource.Limit[string(k)] = NewQuantityFor(v)
	}

	for k, v := range quota.Status.Used {
		resource.Used[string(k)] = NewQuantityFor(v)
	}

	return resource, nil
}

// CanProvision returns true if the cloud can provision a worker meetting the quota
func (cloud *K8SCloud) CanProvision(quota Quota) (bool, error) {
	resource, err := cloud.Resource()
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

// Provision returns a worker if the cloud can provison
func (cloud *K8SCloud) Provision(id string, wopts WorkerOptions) (Worker, error) {
	logdog.Infof("Create worker %s with options %v", id, wopts)
	var cp *K8SCloud
	// If specify the namespace for worker in worker options, new a cloud pointer and set its namespace.
	if len(wopts.Namespace) != 0 {
		nc := *cloud
		cp = &nc
		cp.namespace = wopts.Namespace
	} else {
		cp = cloud
	}

	can, err := cp.CanProvision(wopts.Quota)
	if err != nil {
		return nil, err
	}

	if !can {
		// wait
		return nil, ErrNoEnoughResource
	}

	name := "cyclone-worker-" + id
	Privileged := true
	pod := &apiv1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: cp.namespace,
			Name:      name,
			Labels: map[string]string{
				"cyclone":    "worker",
				"cyclone/id": id,
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:            "cyclone-worker",
					Image:           wopts.WorkerEnvs.WorkerImage,
					Env:             buildK8SEnv(id, wopts),
					WorkingDir:      WorkingDir,
					Resources:       wopts.Quota.ToK8SQuota(),
					SecurityContext: &apiv1.SecurityContext{Privileged: &Privileged},
					ImagePullPolicy: apiv1.PullAlways,
				},
			},
			RestartPolicy: apiv1.RestartPolicyNever,
		},
	}

	// Mount the cache volume to the worker.
	cacheVolume := wopts.CacheVolume
	mountPath := wopts.MountPath
	if len(cacheVolume) != 0 && len(mountPath) != 0 {
		// Check the existence and status of cache volume.
		if pvc, err := cloud.client.CoreV1().PersistentVolumeClaims(cp.namespace).Get(cacheVolume, meta_v1.GetOptions{}); err == nil {
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
				logdog.Errorf("Can not use cache volume %s as its status is %v", cacheVolume, pvc.Status.Phase)
			}
		} else {
			// Just log error and let the pipeline to run in non-cache mode.
			logdog.Errorf("Can not use cache volume %s as fail to get it: %v", cacheVolume, err)
		}

	}

	// pod, err = cloud.Client.CoreV1().Pods(cloud.namespace).Create(pod)
	// if err != nil {
	// 	return nil, err
	// }

	worker := &K8SPodWorker{
		K8SCloud: cp,
		pod:      pod,
	}

	return worker, nil
}

// LoadWorker rebuilds a worker from worker info
func (cloud *K8SCloud) LoadWorker(info WorkerInfo) (Worker, error) {

	if cloud.Kind() != info.CloudKind {
		return nil, fmt.Errorf("K8SCloud: can not load worker with another cloud kind %s", info.CloudKind)
	}

	pod, err := cloud.client.CoreV1().Pods(info.Namespace).Get(info.PodName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	worker := &K8SPodWorker{
		K8SCloud:   cloud,
		createTime: info.CreateTime,
		dueTime:    info.DueTime,
		pod:        pod,
	}

	return worker, nil
}

// GetOptions ...
func (cloud *K8SCloud) GetOptions() Options {
	return Options{
		Name:           cloud.name,
		Kind:           cloud.Kind(),
		Host:           cloud.host,
		Insecure:       cloud.insecure,
		K8SBearerToken: cloud.bearerToken,
		K8SNamespace:   cloud.namespace,
		K8SInCluster:   cloud.inCluster,
	}
}

// ---------------------------------------------------------------------------------

// K8SPodWorker ...
type K8SPodWorker struct {
	*K8SCloud
	createTime time.Time
	dueTime    time.Time
	pod        *apiv1.Pod
}

// Do starts the worker and do the work
func (worker *K8SPodWorker) Do() error {

	var pod *apiv1.Pod
	var err error

	check := func() (bool, error) {

		// change pod here
		pod, err = worker.Client().CoreV1().Pods(worker.namespace).Get(worker.pod.Name, meta_v1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case apiv1.PodPending:
			return false, nil
		case apiv1.PodRunning:
			return true, nil
		default:
			return false, fmt.Errorf("K8SCloud: get an error pod status phase[%s]", pod.Status.Phase)
		}
	}

	pod, err = worker.Client().CoreV1().Pods(worker.namespace).Create(worker.pod)

	if err != nil {
		return err
	}

	// wait until pod is running
	err = wait.Poll(7*time.Second, 2*time.Minute, check)
	if err != nil {
		logdog.Error("K8SPodWorker: do worker error", logdog.Fields{"err": err})
		return err
	}

	// add time
	worker.createTime = time.Now()
	worker.dueTime = worker.createTime.Add(time.Duration(WorkerTimeout))

	worker.pod = pod

	return nil
}

// GetWorkerInfo returns worker's infomation
func (worker *K8SPodWorker) GetWorkerInfo() WorkerInfo {
	return WorkerInfo{
		CloudName:  worker.Name(),
		CloudKind:  worker.Kind(),
		CreateTime: worker.createTime,
		DueTime:    worker.dueTime,
		PodName:    worker.pod.Name,
		Namespace:  worker.namespace,
	}
}

// IsTimeout returns true if worker is timeout
// and returns the time left until it is due
func (worker *K8SPodWorker) IsTimeout() (bool, time.Duration) {
	now := time.Now()
	if now.After(worker.dueTime) {
		return true, time.Duration(0)
	}
	return false, worker.dueTime.Sub(now)
}

// Terminate terminates the worker and destroy it
func (worker *K8SPodWorker) Terminate() error {
	client := worker.Client().CoreV1().Pods(worker.namespace)
	GracePeriodSeconds := int64(0)
	logdog.Debug("worker terminating...", logdog.Fields{"cloud": worker.Name(), "kind": worker.Kind(), "podName": worker.pod.Name})

	if Debug {
		req := client.GetLogs(worker.pod.Name, &apiv1.PodLogOptions{})
		readCloser, err := req.Stream()
		if err != nil {
			logdog.Error("Can not read log from pod", logdog.Fields{
				"cloud":   worker.Name(),
				"kind":    worker.Kind(),
				"podName": worker.pod.Name,
				"err":     err,
			})
		} else {
			defer readCloser.Close()
			content, _ := ioutil.ReadAll(readCloser)
			logdog.Debug(string(content))
		}
	}

	err := client.Delete(
		worker.pod.Name,
		&meta_v1.DeleteOptions{
			GracePeriodSeconds: &GracePeriodSeconds,
		})

	return err
}

func buildK8SEnv(id string, opts WorkerOptions) []apiv1.EnvVar {

	env := []apiv1.EnvVar{
		{
			Name:  WorkerEventID,
			Value: id,
		},
		{
			Name:  CycloneServer,
			Value: opts.WorkerEnvs.CycloneServer,
		},
		{
			Name:  ConsoleWebEndpoint,
			Value: opts.WorkerEnvs.ConsoleWebEndpoint,
		},
		{
			Name:  RegistryLocation,
			Value: opts.WorkerEnvs.RegistryLocation,
		},
		{
			Name:  RegistryUsername,
			Value: opts.WorkerEnvs.RegistryUsername,
		},
		{
			Name:  RegistryPassword,
			Value: opts.WorkerEnvs.RegistryPassword,
		},
		{
			Name:  GitlabURL,
			Value: opts.WorkerEnvs.GitlabURL,
		},
		{
			Name:  LogServer,
			Value: opts.WorkerEnvs.LogServer,
		},
		{
			Name:  WorkerImage,
			Value: opts.WorkerEnvs.WorkerImage,
		},
		{
			Name:  LimitCPU,
			Value: opts.Quota[ResourceLimitsCPU].String(),
		},
		{
			Name:  LimitMemory,
			Value: opts.Quota[ResourceLimitsMemory].String(),
		},
		{
			Name:  RequestCPU,
			Value: opts.Quota[ResourceRequestsCPU].String(),
		},
		{
			Name:  RequestMemory,
			Value: opts.Quota[ResourceRequestsMemory].String(),
		},
	}

	return env
}
