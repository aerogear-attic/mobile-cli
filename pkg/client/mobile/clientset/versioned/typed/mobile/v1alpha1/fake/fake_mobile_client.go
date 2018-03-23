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
	"github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned/typed/mobile/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

type MobileV1alpha1 struct {
	*testing.Fake
}

func (c *MobileV1alpha1) MobileClients(namespace string) v1alpha1.MobileClientInterface {
	return &MobileClients{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *MobileV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
