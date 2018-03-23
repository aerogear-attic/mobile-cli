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
	"github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/internalversion/typed/servicecatalog/internalversion"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

type Servicecatalog struct {
	*testing.Fake
}

func (c *Servicecatalog) ClusterServiceBrokers() internalversion.ClusterServiceBrokerInterface {
	return &ClusterServiceBrokers{c}
}

func (c *Servicecatalog) ClusterServiceClasses() internalversion.ClusterServiceClassInterface {
	return &ClusterServiceClasses{c}
}

func (c *Servicecatalog) ClusterServicePlans() internalversion.ClusterServicePlanInterface {
	return &ClusterServicePlans{c}
}

func (c *Servicecatalog) ServiceBindings(namespace string) internalversion.ServiceBindingInterface {
	return &ServiceBindings{c, namespace}
}

func (c *Servicecatalog) ServiceInstances(namespace string) internalversion.ServiceInstanceInterface {
	return &ServiceInstances{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *Servicecatalog) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
