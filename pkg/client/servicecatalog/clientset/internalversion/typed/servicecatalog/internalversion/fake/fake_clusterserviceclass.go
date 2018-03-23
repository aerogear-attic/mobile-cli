/*
Copyright 2018 The Kubernetes Authors.

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

package fake

import (
	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"
)

// ClusterServiceClasses implements ClusterServiceClassInterface
type ClusterServiceClasses struct {
	Fake *Servicecatalog
}

var clusterserviceclassesResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "", Resource: "clusterserviceclasses"}

var clusterserviceclassesKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "", Kind: "ClusterServiceClass"}

// Get takes name of the clusterServiceClass, and returns the corresponding clusterServiceClass object, and an error if there is any.
func (c *ClusterServiceClasses) Get(name string, options v1.GetOptions) (result *servicecatalog.ClusterServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(clusterserviceclassesResource, name), &servicecatalog.ClusterServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ClusterServiceClass), err
}

// List takes label and field selectors, and returns the list of ClusterServiceClasses that match those selectors.
func (c *ClusterServiceClasses) List(opts v1.ListOptions) (result *servicecatalog.ClusterServiceClassList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(clusterserviceclassesResource, clusterserviceclassesKind, opts), &servicecatalog.ClusterServiceClassList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &servicecatalog.ClusterServiceClassList{}
	for _, item := range obj.(*servicecatalog.ClusterServiceClassList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterServiceClasses.
func (c *ClusterServiceClasses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(clusterserviceclassesResource, opts))
}

// Create takes the representation of a clusterServiceClass and creates it.  Returns the server's representation of the clusterServiceClass, and an error, if there is any.
func (c *ClusterServiceClasses) Create(clusterServiceClass *servicecatalog.ClusterServiceClass) (result *servicecatalog.ClusterServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(clusterserviceclassesResource, clusterServiceClass), &servicecatalog.ClusterServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ClusterServiceClass), err
}

// Update takes the representation of a clusterServiceClass and updates it. Returns the server's representation of the clusterServiceClass, and an error, if there is any.
func (c *ClusterServiceClasses) Update(clusterServiceClass *servicecatalog.ClusterServiceClass) (result *servicecatalog.ClusterServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(clusterserviceclassesResource, clusterServiceClass), &servicecatalog.ClusterServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ClusterServiceClass), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *ClusterServiceClasses) UpdateStatus(clusterServiceClass *servicecatalog.ClusterServiceClass) (*servicecatalog.ClusterServiceClass, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(clusterserviceclassesResource, "status", clusterServiceClass), &servicecatalog.ClusterServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ClusterServiceClass), err
}

// Delete takes name of the clusterServiceClass and deletes it. Returns an error if one occurs.
func (c *ClusterServiceClasses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(clusterserviceclassesResource, name), &servicecatalog.ClusterServiceClass{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *ClusterServiceClasses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(clusterserviceclassesResource, listOptions)

	_, err := c.Fake.Invokes(action, &servicecatalog.ClusterServiceClassList{})
	return err
}

// Patch applies the patch and returns the patched clusterServiceClass.
func (c *ClusterServiceClasses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ClusterServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(clusterserviceclassesResource, name, data, subresources...), &servicecatalog.ClusterServiceClass{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ClusterServiceClass), err
}
