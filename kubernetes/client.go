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

package kubernetes

import (
	k8s_core_api "k8s.io/kubernetes/pkg/api"
	batch "k8s.io/kubernetes/pkg/apis/batch"
	//k8s_ext_api "k8s.io/kubernetes/pkg/apis/extensions"
	"github.com/caicloud/cyclone/pkg/log"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	restclient "k8s.io/kubernetes/pkg/client/restclient"
)

// newClientWithToken creates a new kubernetes client using BearerToken.
func newClientWithToken(host, token string) (*clientset.Clientset, error) {
	config := &restclient.Config{
		Host:        host,
		BearerToken: token,
		Insecure:    true,
	}
	return clientset.NewForConfig(config)
}

// CreateJob
func CreateJob(host, token, namespace string, job *batch.Job) error {
	k8sClient, err := newClientWithToken(host, token)
	if err != nil {
		return err
	}

	newjob, err := k8sClient.Jobs(namespace).Create(job)
	if err != nil {
		return err
	}

	log.Infof("Create new job, active pods(%d).", newjob.Status.Active)

	return nil
}

// CreatePod
func CreatePod(host, token, namespace string, pod *k8s_core_api.Pod) (string, error) {
	k8sClient, err := newClientWithToken(host, token)
	if err != nil {
		return "", err
	}

	newPod, err := k8sClient.Pods(namespace).Create(pod)
	if err != nil {
		return "", err
	}

	log.Infof("Create new pod (%s).", newPod.ObjectMeta.Name)

	return newPod.ObjectMeta.Name, nil
}

// DeletePod
func DeletePod(host, token, namespace string, podName string, options *k8s_core_api.DeleteOptions) error {
	k8sClient, err := newClientWithToken(host, token)
	if err != nil {
		return err
	}

	err = k8sClient.Pods(namespace).Delete(podName, options)
	if err != nil {
		return err
	}

	log.Infof("Delete pod (%s).", podName)

	return nil
}
