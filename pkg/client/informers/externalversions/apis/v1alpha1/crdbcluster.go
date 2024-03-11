/*
Copyright 2024 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	apisv1alpha1 "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	versioned "github.com/cockroachdb/cockroach-operator/pkg/client/clientset/versioned"
	internalinterfaces "github.com/cockroachdb/cockroach-operator/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/cockroachdb/cockroach-operator/pkg/client/listers/apis/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// CrdbClusterInformer provides access to a shared informer and lister for
// CrdbClusters.
type CrdbClusterInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.CrdbClusterLister
}

type crdbClusterInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewCrdbClusterInformer constructs a new informer for CrdbCluster type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCrdbClusterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCrdbClusterInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredCrdbClusterInformer constructs a new informer for CrdbCluster type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCrdbClusterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CrdbV1alpha1().CrdbClusters(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CrdbV1alpha1().CrdbClusters(namespace).Watch(context.TODO(), options)
			},
		},
		&apisv1alpha1.CrdbCluster{},
		resyncPeriod,
		indexers,
	)
}

func (f *crdbClusterInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCrdbClusterInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *crdbClusterInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apisv1alpha1.CrdbCluster{}, f.defaultInformer)
}

func (f *crdbClusterInformer) Lister() v1alpha1.CrdbClusterLister {
	return v1alpha1.NewCrdbClusterLister(f.Informer().GetIndexer())
}
