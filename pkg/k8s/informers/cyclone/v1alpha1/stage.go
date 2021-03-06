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

// StageInformer provides access to a shared informer and lister for
// Stages.
type StageInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.StageLister
}

type stageInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewStageInformer constructs a new informer for Stage type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewStageInformer(client clientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredStageInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredStageInformer constructs a new informer for Stage type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredStageInformer(client clientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CycloneV1alpha1().Stages(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CycloneV1alpha1().Stages(namespace).Watch(context.TODO(), options)
			},
		},
		&cyclonev1alpha1.Stage{},
		resyncPeriod,
		indexers,
	)
}

func (f *stageInformer) defaultInformer(client clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredStageInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *stageInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&cyclonev1alpha1.Stage{}, f.defaultInformer)
}

func (f *stageInformer) Lister() v1alpha1.StageLister {
	return v1alpha1.NewStageLister(f.Informer().GetIndexer())
}
