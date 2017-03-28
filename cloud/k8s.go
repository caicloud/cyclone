package cloud

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/zoumo/logdog"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/rest"
)

// K8SCloud ...
type K8SCloud struct {
	name        string
	host        string
	bearerToken string
	namespace   string
	insecure    bool
	client      *kubernetes.Clientset
}

// NewK8SCloud ...
func NewK8SCloud(opts Options) (Cloud, error) {

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
		Insecure:    opts.Insecure,
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
	_, err := cloud.client.CoreV1().Pods(cloud.namespace).List(apiv1.ListOptions{})
	return err
}

// Resource returns the limit and used quotas of the cloud
func (cloud *K8SCloud) Resource() (*Resource, error) {
	quotas, err := cloud.client.CoreV1().ResourceQuotas(cloud.namespace).List(apiv1.ListOptions{})
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
	can, err := cloud.CanProvision(wopts.Quota)
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
		ObjectMeta: apiv1.ObjectMeta{
			Namespace: cloud.namespace,
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
				},
			},
		},
	}

	// pod, err = cloud.Client.CoreV1().Pods(cloud.namespace).Create(pod)
	// if err != nil {
	// 	return nil, err
	// }

	worker := &K8SPodWorker{
		K8SCloud: cloud,
		pod:      pod,
	}

	return worker, nil
}

// LoadWorker rebuilds a worker from worker info
func (cloud *K8SCloud) LoadWorker(info WorkerInfo) (Worker, error) {

	if cloud.Kind() != info.CloudKind {
		return nil, fmt.Errorf("K8SCloud: can not load worker with another cloud kind %s", info.CloudKind)
	}

	pod, err := cloud.client.CoreV1().Pods(cloud.namespace).Get(info.PodName)
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
		pod, err = worker.Client().CoreV1().Pods(worker.namespace).Get(worker.pod.Name)
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
		return err
	}

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
	return client.Delete(
		worker.pod.Name,
		&apiv1.DeleteOptions{
			GracePeriodSeconds: &GracePeriodSeconds,
		})
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
			Name:  ClairDisable,
			Value: strconv.FormatBool(opts.WorkerEnvs.ClairDisable),
		},
		{
			Name:  ClairServer,
			Value: opts.WorkerEnvs.ClairServer,
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
	}

	return env
}
