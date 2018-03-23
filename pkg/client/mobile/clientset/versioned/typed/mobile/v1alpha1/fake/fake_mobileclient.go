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
	"github.com/aerogear/mobile-cli/pkg/apis/mobile/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"
)

// MobileClients implements MobileClientInterface
type MobileClients struct {
	Fake *MobileV1alpha1
	ns   string
}

var mobileclientsResource = schema.GroupVersionResource{Group: "mobile.k8s.io", Version: "v1alpha1", Resource: "mobileclients"}

var mobileclientsKind = schema.GroupVersionKind{Group: "mobile.k8s.io", Version: "v1alpha1", Kind: "MobileClient"}

// Get takes name of the mobileClient, and returns the corresponding mobileClient object, and an error if there is any.
func (c *MobileClients) Get(name string, options v1.GetOptions) (result *v1alpha1.MobileClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mobileclientsResource, c.ns, name), &v1alpha1.MobileClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MobileClient), err
}

// List takes label and field selectors, and returns the list of MobileClients that match those selectors.
func (c *MobileClients) List(opts v1.ListOptions) (result *v1alpha1.MobileClientList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mobileclientsResource, mobileclientsKind, c.ns, opts), &v1alpha1.MobileClientList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MobileClientList{}
	for _, item := range obj.(*v1alpha1.MobileClientList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mobileClients.
func (c *MobileClients) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mobileclientsResource, c.ns, opts))

}

// Create takes the representation of a mobileClient and creates it.  Returns the server's representation of the mobileClient, and an error, if there is any.
func (c *MobileClients) Create(mobileClient *v1alpha1.MobileClient) (result *v1alpha1.MobileClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mobileclientsResource, c.ns, mobileClient), &v1alpha1.MobileClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MobileClient), err
}

// Update takes the representation of a mobileClient and updates it. Returns the server's representation of the mobileClient, and an error, if there is any.
func (c *MobileClients) Update(mobileClient *v1alpha1.MobileClient) (result *v1alpha1.MobileClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mobileclientsResource, c.ns, mobileClient), &v1alpha1.MobileClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MobileClient), err
}

// Delete takes name of the mobileClient and deletes it. Returns an error if one occurs.
func (c *MobileClients) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(mobileclientsResource, c.ns, name), &v1alpha1.MobileClient{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *MobileClients) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mobileclientsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.MobileClientList{})
	return err
}

// Patch applies the patch and returns the patched mobileClient.
func (c *MobileClients) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.MobileClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mobileclientsResource, c.ns, name, data, subresources...), &v1alpha1.MobileClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MobileClient), err
}
