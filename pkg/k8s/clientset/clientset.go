/*
Copyright 2018 caicloud authors. All rights reserved.
*/

// Code generated by client-gen. DO NOT EDIT.

package clientset

import (
	cyclonev1alpha1 "github.com/caicloud/cyclone/pkg/k8s/clientset/typed/cyclone/v1alpha1"
	kubernetes "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	kubernetes.Interface
	CycloneV1alpha1() cyclonev1alpha1.CycloneV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Cyclone() cyclonev1alpha1.CycloneV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*kubernetes.Clientset
	cycloneV1alpha1 *cyclonev1alpha1.CycloneV1alpha1Client
}

// CycloneV1alpha1 retrieves the CycloneV1alpha1Client
func (c *Clientset) CycloneV1alpha1() cyclonev1alpha1.CycloneV1alpha1Interface {
	return c.cycloneV1alpha1
}

// Deprecated: Cyclone retrieves the default version of CycloneClient.
// Please explicitly pick a version.
func (c *Clientset) Cyclone() cyclonev1alpha1.CycloneV1alpha1Interface {
	return c.cycloneV1alpha1
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.cycloneV1alpha1, err = cyclonev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.Clientset, err = kubernetes.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.cycloneV1alpha1 = cyclonev1alpha1.NewForConfigOrDie(c)

	cs.Clientset = kubernetes.NewForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.cycloneV1alpha1 = cyclonev1alpha1.New(c)

	cs.Clientset = kubernetes.New(c)
	return &cs
}
