/*
Copyright 2018 caicloud authors. All rights reserved.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/caicloud/cyclone/pkg/k8s/clientset/typed/cyclone/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCycloneV1alpha1 struct {
	*testing.Fake
}

func (c *FakeCycloneV1alpha1) CycloneResources(namespace string) v1alpha1.CycloneResourceInterface {
	return &FakeCycloneResources{c, namespace}
}

func (c *FakeCycloneV1alpha1) Stages(namespace string) v1alpha1.StageInterface {
	return &FakeStages{c, namespace}
}

func (c *FakeCycloneV1alpha1) StageTemplates(namespace string) v1alpha1.StageTemplateInterface {
	return &FakeStageTemplates{c, namespace}
}

func (c *FakeCycloneV1alpha1) Workflows(namespace string) v1alpha1.WorkflowInterface {
	return &FakeWorkflows{c, namespace}
}

func (c *FakeCycloneV1alpha1) WorkflowParams(namespace string) v1alpha1.WorkflowParamInterface {
	return &FakeWorkflowParams{c, namespace}
}

func (c *FakeCycloneV1alpha1) WorkflowRuns(namespace string) v1alpha1.WorkflowRunInterface {
	return &FakeWorkflowRuns{c, namespace}
}

func (c *FakeCycloneV1alpha1) WorkflowTriggers(namespace string) v1alpha1.WorkflowTriggerInterface {
	return &FakeWorkflowTriggers{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCycloneV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
