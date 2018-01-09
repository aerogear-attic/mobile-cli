package cmd_test

import (
	"bytes"
	"testing"

	"encoding/json"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	"github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	scFake "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/aerogear/mobile-cli/pkg/cmd"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestServicesCmd_ListServicesCmd(t *testing.T) {
	cases := []struct {
		Name             string
		SvcCatalogClient func() versioned.Interface
		K8Client         func() kubernetes.Interface
		ExpectError      bool
		Validate         func(t *testing.T, output []byte)
		Flags            []string
	}{
		{
			Name:  "test list services as expected",
			Flags: []string{"-o=json"},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				fakeClient.AddReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ClusterServiceClassList{
						Items: []v1beta1.ClusterServiceClass{
							{
								Spec: v1beta1.ClusterServiceClassSpec{
									Tags: []string{"mobile-service", "other"},
								},
							},
							{
								Spec: v1beta1.ClusterServiceClassSpec{
									Tags: []string{},
								},
							},
						},
					}, nil
				})
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			Validate: func(t *testing.T, data []byte) {
				var list = &v1beta1.ClusterServiceClassList{}
				if err := json.Unmarshal(data, list); err != nil {
					t.Fatal("failed to unmarshal data", err)
				}

				if nil == list {
					t.Fatal("expected a list but got nil")
				}
				if len(list.Items) != 1 {
					t.Fatalf("expected only one item in the list but got %v", len(list.Items))
				}
			},
		},
		{
			Name:  "test list services returns error on failure",
			Flags: []string{"-o=json"},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				fakeClient.AddReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed for some reason")
				})
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			ExpectError: true,
			Validate: func(t *testing.T, output []byte) {
				err := string(output)
				if err == "" {
					t.Fatal("expected an error message but got none")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var out bytes.Buffer
			serviceCmd := cmd.NewServicesCmd(tc.SvcCatalogClient(), tc.K8Client(), &out)
			listCmd := serviceCmd.ListServicesCmd()
			err := listCmd.RunE(listCmd, tc.Flags)
			if err != nil && !tc.ExpectError {
				t.Fatal("did not expect an error but gone one ", err)
			}
			if err == nil && tc.ExpectError {
				t.Fatal("expected an error but got none")
			}
			if tc.Validate != nil {
				tc.Validate(t, out.Bytes())
			}

		})
	}
}
