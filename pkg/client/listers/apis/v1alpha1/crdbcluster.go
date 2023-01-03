/*
Copyright 2023 The Cockroach Authors

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CrdbClusterLister helps list CrdbClusters.
// All objects returned here must be treated as read-only.
type CrdbClusterLister interface {
	// List lists all CrdbClusters in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.CrdbCluster, err error)
	// CrdbClusters returns an object that can list and get CrdbClusters.
	CrdbClusters(namespace string) CrdbClusterNamespaceLister
	CrdbClusterListerExpansion
}

// crdbClusterLister implements the CrdbClusterLister interface.
type crdbClusterLister struct {
	indexer cache.Indexer
}

// NewCrdbClusterLister returns a new CrdbClusterLister.
func NewCrdbClusterLister(indexer cache.Indexer) CrdbClusterLister {
	return &crdbClusterLister{indexer: indexer}
}

// List lists all CrdbClusters in the indexer.
func (s *crdbClusterLister) List(selector labels.Selector) (ret []*v1alpha1.CrdbCluster, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CrdbCluster))
	})
	return ret, err
}

// CrdbClusters returns an object that can list and get CrdbClusters.
func (s *crdbClusterLister) CrdbClusters(namespace string) CrdbClusterNamespaceLister {
	return crdbClusterNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CrdbClusterNamespaceLister helps list and get CrdbClusters.
// All objects returned here must be treated as read-only.
type CrdbClusterNamespaceLister interface {
	// List lists all CrdbClusters in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.CrdbCluster, err error)
	// Get retrieves the CrdbCluster from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.CrdbCluster, error)
	CrdbClusterNamespaceListerExpansion
}

// crdbClusterNamespaceLister implements the CrdbClusterNamespaceLister
// interface.
type crdbClusterNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CrdbClusters in the indexer for a given namespace.
func (s crdbClusterNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.CrdbCluster, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CrdbCluster))
	})
	return ret, err
}

// Get retrieves the CrdbCluster from the indexer for a given namespace and name.
func (s crdbClusterNamespaceLister) Get(name string) (*v1alpha1.CrdbCluster, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("crdbcluster"), name)
	}
	return obj.(*v1alpha1.CrdbCluster), nil
}
