/*
Copyright 2025 The Cockroach Authors

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCrdbClusters implements CrdbClusterInterface
type FakeCrdbClusters struct {
	Fake *FakeCrdbV1alpha1
	ns   string
}

var crdbclustersResource = v1alpha1.SchemeGroupVersion.WithResource("crdbclusters")

var crdbclustersKind = v1alpha1.SchemeGroupVersion.WithKind("CrdbCluster")

// Get takes name of the crdbCluster, and returns the corresponding crdbCluster object, and an error if there is any.
func (c *FakeCrdbClusters) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.CrdbCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(crdbclustersResource, c.ns, name), &v1alpha1.CrdbCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CrdbCluster), err
}

// List takes label and field selectors, and returns the list of CrdbClusters that match those selectors.
func (c *FakeCrdbClusters) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.CrdbClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(crdbclustersResource, crdbclustersKind, c.ns, opts), &v1alpha1.CrdbClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CrdbClusterList{ListMeta: obj.(*v1alpha1.CrdbClusterList).ListMeta}
	for _, item := range obj.(*v1alpha1.CrdbClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested crdbClusters.
func (c *FakeCrdbClusters) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(crdbclustersResource, c.ns, opts))

}

// Create takes the representation of a crdbCluster and creates it.  Returns the server's representation of the crdbCluster, and an error, if there is any.
func (c *FakeCrdbClusters) Create(ctx context.Context, crdbCluster *v1alpha1.CrdbCluster, opts v1.CreateOptions) (result *v1alpha1.CrdbCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(crdbclustersResource, c.ns, crdbCluster), &v1alpha1.CrdbCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CrdbCluster), err
}

// Update takes the representation of a crdbCluster and updates it. Returns the server's representation of the crdbCluster, and an error, if there is any.
func (c *FakeCrdbClusters) Update(ctx context.Context, crdbCluster *v1alpha1.CrdbCluster, opts v1.UpdateOptions) (result *v1alpha1.CrdbCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(crdbclustersResource, c.ns, crdbCluster), &v1alpha1.CrdbCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CrdbCluster), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCrdbClusters) UpdateStatus(ctx context.Context, crdbCluster *v1alpha1.CrdbCluster, opts v1.UpdateOptions) (*v1alpha1.CrdbCluster, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(crdbclustersResource, "status", c.ns, crdbCluster), &v1alpha1.CrdbCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CrdbCluster), err
}

// Delete takes name of the crdbCluster and deletes it. Returns an error if one occurs.
func (c *FakeCrdbClusters) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(crdbclustersResource, c.ns, name, opts), &v1alpha1.CrdbCluster{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCrdbClusters) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(crdbclustersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.CrdbClusterList{})
	return err
}

// Patch applies the patch and returns the patched crdbCluster.
func (c *FakeCrdbClusters) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CrdbCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(crdbclustersResource, c.ns, name, pt, data, subresources...), &v1alpha1.CrdbCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CrdbCluster), err
}
