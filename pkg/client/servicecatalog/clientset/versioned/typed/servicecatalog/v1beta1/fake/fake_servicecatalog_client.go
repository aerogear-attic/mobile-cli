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
	"github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned/typed/servicecatalog/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

type ServicecatalogV1beta1 struct {
	*testing.Fake
}

func (c *ServicecatalogV1beta1) ClusterServiceBrokers() v1beta1.ClusterServiceBrokerInterface {
	return &ClusterServiceBrokers{c}
}

func (c *ServicecatalogV1beta1) ClusterServiceClasses() v1beta1.ClusterServiceClassInterface {
	return &ClusterServiceClasses{c}
}

func (c *ServicecatalogV1beta1) ClusterServicePlans() v1beta1.ClusterServicePlanInterface {
	return &ClusterServicePlans{c}
}

func (c *ServicecatalogV1beta1) ServiceBindings(namespace string) v1beta1.ServiceBindingInterface {
	return &ServiceBindings{c, namespace}
}

func (c *ServicecatalogV1beta1) ServiceInstances(namespace string) v1beta1.ServiceInstanceInterface {
	return &ServiceInstances{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ServicecatalogV1beta1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
