/*
Copyright 2020 caicloud authors. All rights reserved.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	cyclonev1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	clientset "github.com/caicloud/cyclone/pkg/k8s/clientset"
	internalinterfaces "github.com/caicloud/cyclone/pkg/k8s/informers/internalinterfaces"
	v1alpha1 "github.com/caicloud/cyclone/pkg/k8s/listers/cyclone/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// WorkflowInformer provides access to a shared informer and lister for
// Workflows.
type WorkflowInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.WorkflowLister
}

type workflowInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewWorkflowInformer constructs a new informer for Workflow type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewWorkflowInformer(client clientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredWorkflowInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredWorkflowInformer constructs a new informer for Workflow type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredWorkflowInformer(client clientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CycloneV1alpha1().Workflows(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CycloneV1alpha1().Workflows(namespace).Watch(context.TODO(), options)
			},
		},
		&cyclonev1alpha1.Workflow{},
		resyncPeriod,
		indexers,
	)
}

func (f *workflowInformer) defaultInformer(client clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredWorkflowInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *workflowInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&cyclonev1alpha1.Workflow{}, f.defaultInformer)
}

func (f *workflowInformer) Lister() v1alpha1.WorkflowLister {
	return v1alpha1.NewWorkflowLister(f.Informer().GetIndexer())
}
