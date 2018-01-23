// Copyright Red Hat, Inc., and individual contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd_test

import (
	"bytes"
	"testing"

	"encoding/json"

	"fmt"

	"regexp"

	"github.com/aerogear/mobile-cli/pkg/apis/mobile/v1alpha1"
	mc "github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned"
	mcFake "github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned/fake"
	"github.com/aerogear/mobile-cli/pkg/cmd"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kt "k8s.io/client-go/testing"
)

func TestMobileClientsCmd_TestListClients(t *testing.T) {
	cases := []struct {
		Name         string
		MobileClient func() mc.Interface
		ExpectError  bool
		Flags        []string
		Validate     func(t *testing.T, list *v1alpha1.MobileClientList)
		ErrorPattern string
	}{
		{
			Name: "test getting mobile clients returns a list of clients",
			MobileClient: func() mc.Interface {
				cs := &mcFake.Clientset{}
				cs.AddReactor("list", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1alpha1.MobileClientList{
						Items: []v1alpha1.MobileClient{{Spec: v1alpha1.MobileClientSpec{Name: "test", ApiKey: "testkey", ClientType: "cordova"}}, {Spec: v1alpha1.MobileClientSpec{Name: "test2", ApiKey: "testkey", ClientType: "cordova"}}},
					}, nil
				})
				return cs
			},
			Flags: []string{"--namespace=test", "-o=json"},
			Validate: func(t *testing.T, list *v1alpha1.MobileClientList) {
				if nil == list {
					t.Fatal("expected a list of mobileclients but got none")
				}
				if len(list.Items) != 2 {
					t.Fatalf("expected 2 mobile clients but got %v", len(list.Items))
				}
			},
		},
		{
			Name: "test getting mobile clients outputs clear error message on failure",
			MobileClient: func() mc.Interface {
				cs := &mcFake.Clientset{}
				cs.AddReactor("list", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("failed to do something")
				})
				return cs
			},
			Flags:        []string{"--namespace=test", "-o=json"},
			ExpectError:  true,
			ErrorPattern: "^failed to list mobile clients: .*",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var stdOut bytes.Buffer
			root := cmd.NewRootCmd()
			underTest := cmd.NewClientCmd(tc.MobileClient(), &stdOut)
			clientCmd := underTest.ListClientsCmd()
			root.AddCommand(clientCmd)
			if err := clientCmd.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse flags ", err)
			}
			err := clientCmd.RunE(clientCmd, []string{})
			if tc.ExpectError && err == nil {
				t.Fatal("expected an error but got none")
			}
			if !tc.ExpectError && err != nil {
				t.Fatal("did not expect an error but got one ", err)
			}
			if tc.ExpectError && err != nil {
				if m, err := regexp.Match(tc.ErrorPattern, []byte(err.Error())); !m {
					t.Fatal("expected regex to match error ", tc.ErrorPattern, err)
				}
			}
			if nil != tc.Validate {
				mobileClients := &v1alpha1.MobileClientList{}
				if err := json.Unmarshal(stdOut.Bytes(), mobileClients); err != nil {
					t.Fatal("unexpected error unmarshalling json", err)
				}
				tc.Validate(t, mobileClients)
			}
		})
	}
}

func TestMobileClientsCmd_TestGetClient(t *testing.T) {
	cases := []struct {
		Name         string
		ClientName   string
		MobileClient func() mc.Interface
		ExpectError  bool
		ErrorPattern string
		Flags        []string
		Validate     func(t *testing.T, c *v1alpha1.MobileClient)
	}{
		{
			Name: "test get client returns only one client with the right name",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("get", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1alpha1.MobileClient{
						Spec: v1alpha1.MobileClientSpec{
							Name: "myapp",
						},
					}, nil
				})
				return mc
			},
			ClientName: "myapp",
			Flags:      []string{"--namespace=myproject", "-o=json"},
			Validate: func(t *testing.T, c *v1alpha1.MobileClient) {
				if nil == c {
					t.Fatal("expected a mobile client but got nil")
				}
				if c.Spec.Name != "myapp" {
					t.Fatalf("expected an app with name %s but got %s ", "myapp", c.Spec.Name)
				}
			},
		},
		{
			Name: "test get client returns a clear error when it fails",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("get", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed to get mobile client")
				})
				return mc
			},
			ClientName:   "myapp",
			Flags:        []string{"--namespace=myproject", "-o=json"},
			ExpectError:  true,
			ErrorPattern: "^failed to get mobile client with clientID (\\w+)+:.*",
		},
		{
			Name: "test get client fails when missing a required argument",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				return mc
			},
			ClientName:   "",
			Flags:        []string{"--namespace=myproject", "-o=json"},
			ExpectError:  true,
			ErrorPattern: "^missing argument .*",
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var stdOut bytes.Buffer
			root := cmd.NewRootCmd()
			clientCmd := cmd.NewClientCmd(tc.MobileClient(), &stdOut)

			getClients := clientCmd.GetClientCmd()
			root.AddCommand(getClients)
			if err := getClients.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse flags ", err)
			}
			var args []string
			if tc.ClientName != "" {
				args = append(args, tc.ClientName)
			}
			err := getClients.RunE(getClients, args)
			if tc.ExpectError && err == nil {
				t.Fatal("expected an error but got none")
			}
			if !tc.ExpectError && err != nil {
				t.Fatal("did not expect an error but got one ", err)
			}
			if tc.ExpectError && err != nil {
				if m, err := regexp.Match(tc.ErrorPattern, []byte(err.Error())); !m {
					t.Fatal("expected error to match pattern "+tc.ErrorPattern, err)
				}
			}
			if nil != tc.Validate {
				mobileClient := &v1alpha1.MobileClient{}
				if err := json.Unmarshal(stdOut.Bytes(), mobileClient); err != nil {
					t.Fatal("failed to unmarshal mobile client")
				}
				tc.Validate(t, mobileClient)
			}
		})
	}
}

func TestMobileClientsCmd_TestDeleteClient(t *testing.T) {
	cases := []struct {
		Name         string
		ClientName   string
		MobileClient func() mc.Interface
		ExpectError  bool
		ErrorPattern string
		Flags        []string
	}{
		{
			Name: "test delete client succeeds with no errors",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				return mc
			},
			ClientName: "myapp",
			Flags:      []string{"--namespace=myproject", "-o=json"},
		},
		{
			Name: "test delete client fails when missing arguments",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				return mc
			},
			Flags:       []string{"--namespace=myproject", "-o=json"},
			ExpectError: true,
		},
		{
			Name: "test delete client returns a clear error when delete fails",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("delete", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed to delete mobileclient")
				})
				return mc
			},
			Flags:       []string{"--namespace=myproject", "-o=json"},
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var stdOut bytes.Buffer
			root := cmd.NewRootCmd()
			clientCmd := cmd.NewClientCmd(tc.MobileClient(), &stdOut)
			deleteClient := clientCmd.DeleteClientCmd()
			root.AddCommand(deleteClient)
			if err := deleteClient.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse flags ", err)
			}
			var args []string
			if tc.ClientName != "" {
				args = append(args, tc.ClientName)
			}
			err := deleteClient.RunE(deleteClient, args)
			if tc.ExpectError && err == nil {
				t.Fatal("expected an error but got none")
			}
			if !tc.ExpectError && err != nil {
				t.Fatal("did not expect an error but got one", err)
			}
			if tc.ExpectError && err != nil {
				if m, err := regexp.Match(tc.ErrorPattern, []byte(err.Error())); !m {
					t.Fatal("expected the error to match the pattern "+tc.ErrorPattern, err)
				}
			}
		})
	}
}

func TestMobileClientsCmd_TestCreateClient(t *testing.T) {
	cases := []struct {
		Name         string
		Args         []string
		MobileClient func() mc.Interface
		ExpectError  bool
		ErrorPattern string
		Flags        []string
		Validate     func(t *testing.T, c *v1alpha1.MobileClient)
	}{
		{
			Name: "test create cordova mobile client succeeds without error",
			Args: []string{"test", "cordova"},
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("create", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					ca := action.(kt.CreateAction)
					return true, ca.GetObject(), nil
				})
				return mc
			},
			Flags: []string{"--namespace=myproject", "-o=json"},
			Validate: func(t *testing.T, c *v1alpha1.MobileClient) {
				if nil == c {
					t.Fatal("expected a mobile client but got nil")
				}
				if c.Spec.ClientType != "cordova" {
					t.Fatal("expected the clientType to be cordova but got ", c.Spec.ClientType)
				}
				if c.Spec.Name != "test" {
					t.Fatal("expected the client name to be test but got ", c.Spec.Name)
				}
				if c.Spec.ApiKey == "" {
					t.Fatal("expected an apiKey to be generated but it was empty")
				}
				icon, ok := c.Labels["icon"]
				if !ok {
					t.Fatal("expected an icon to be set but there was none")
				}
				if icon != "icon-cordova" {
					t.Fatal("expected the icon to be icon-cordova but got ", icon)
				}
			},
		},
		{
			Name: "test create android mobile client succeeds without error",
			Args: []string{"test", "android"},
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("create", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					ca := action.(kt.CreateAction)
					return true, ca.GetObject(), nil
				})
				return mc
			},
			Flags: []string{"--namespace=myproject", "-o=json"},
			Validate: func(t *testing.T, c *v1alpha1.MobileClient) {
				if nil == c {
					t.Fatal("expected a mobile client but got nil")
				}
				if c.Spec.ClientType != "android" {
					t.Fatal("expected the clientType to be android but got ", c.Spec.ClientType)
				}
				if c.Spec.Name != "test" {
					t.Fatal("expected the client name to be test but got ", c.Spec.Name)
				}
				if c.Spec.ApiKey == "" {
					t.Fatal("expected an apiKey to be generated but it was empty")
				}
				icon, ok := c.Labels["icon"]
				if !ok {
					t.Fatal("expected an icon to be set but there was none")
				}
				if icon != "fa-android" {
					t.Fatal("expected the icon to be fa-android but got ", icon)
				}
			},
		},
		{
			Name: "test create iOS mobile client succeeds without error",
			Args: []string{"test", "iOS"},
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("create", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					ca := action.(kt.CreateAction)
					return true, ca.GetObject(), nil
				})
				return mc
			},
			Flags: []string{"--namespace=myproject", "-o=json"},
			Validate: func(t *testing.T, c *v1alpha1.MobileClient) {
				if nil == c {
					t.Fatal("expected a mobile client but got nil")
				}
				if c.Spec.ClientType != "iOS" {
					t.Fatal("expected the clientType to be iOS but got ", c.Spec.ClientType)
				}
				if c.Spec.Name != "test" {
					t.Fatal("expected the client name to be test but got ", c.Spec.Name)
				}
				if c.Spec.ApiKey == "" {
					t.Fatal("expected an apiKey to be generated but it was empty")
				}
				icon, ok := c.Labels["icon"]
				if !ok {
					t.Fatal("expected an icon to be set but there was none")
				}
				if icon != "fa-apple" {
					t.Fatal("expected the icon to be fa-apple but got ", icon)
				}
			},
		},
		{
			Name:         "test create mobile client fails with unknown client type",
			Args:         []string{"test", "firefox"},
			ExpectError:  true,
			ErrorPattern: "^Failed validation while creating new mobile client: .*",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("create", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("should not have been called")
				})
				return mc
			},
			Flags: []string{"--namespace=myproject", "-o=json"},
		},
		{
			Name:         "test create mobile client fails when missing required arguments",
			Args:         []string{"test"},
			ExpectError:  true,
			ErrorPattern: "^expected a name and a clientType",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("create", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("should not have been called")
				})
				return mc
			},
			Flags: []string{"--namespace=myproject", "-o=json"},
		},
		{
			Name:         "test create mobile client fails with clear error message",
			Args:         []string{"test", "cordova"},
			ExpectError:  true,
			ErrorPattern: "^failed to create mobile client: something went wrong",
			MobileClient: func() mc.Interface {
				mc := &mcFake.Clientset{}
				mc.AddReactor("create", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("something went wrong")
				})
				return mc
			},
			Flags: []string{"--namespace=myproject", "-o=json"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var stdOut bytes.Buffer
			root := cmd.NewRootCmd()
			clientCmd := cmd.NewClientCmd(tc.MobileClient(), &stdOut)
			createCmd := clientCmd.CreateClientCmd()
			root.AddCommand(createCmd)
			if err := createCmd.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse flags ", err)
			}
			err := createCmd.RunE(createCmd, tc.Args)
			if tc.ExpectError && err == nil {
				t.Fatal("expected an error but got none")
			}
			if !tc.ExpectError && err != nil {
				t.Fatal("did not expect an error but got one", err)
			}
			if tc.ExpectError && err != nil {
				if m, regErr := regexp.Match(tc.ErrorPattern, []byte(err.Error())); !m {
					t.Fatal("expected the error to match the pattern "+tc.ErrorPattern, err, regErr)
				}
			}
			if nil != tc.Validate {
				mobileClient := &v1alpha1.MobileClient{}
				if err := json.Unmarshal(stdOut.Bytes(), mobileClient); err != nil {
					t.Fatal("failed to unmarshal mobile client")
				}
				tc.Validate(t, mobileClient)
			}
		})
	}
}
