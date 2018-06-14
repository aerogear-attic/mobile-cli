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

package main

import (
	"flag"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"path/filepath"

	"log"

	"github.com/aerogear/mobile-cli/pkg/cmd"
	m "github.com/aerogear/mobile-crd-client/pkg/client/mobile/clientset/versioned"
	sc "github.com/aerogear/mobile-crd-client/pkg/client/servicecatalog/clientset/versioned"
	restclient "k8s.io/client-go/rest"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	k8Client, mobileClient, scClient := NewClientsOrDie(config)
	var (
		out              = os.Stdout
		rootCmd          = cmd.NewRootCmd()
		clientCmd        = cmd.NewClientCmd(mobileClient, scClient, k8Client, out)
		bindCmd          = cmd.NewIntegrationCmd(scClient, k8Client, out)
		serviceConfigCmd = cmd.NewServiceConfigCommand(k8Client)
		clientCfgCmd     = cmd.NewClientConfigCmd(k8Client, mobileClient, scClient, config.Host, out)
		clientBuilds     = cmd.NewClientBuildsCmd()
		svcCmd           = cmd.NewServicesCmd(scClient, k8Client, out)
	)

	// create
	{
		createCmd := cmd.NewCreateCommand()
		createCmd.AddCommand(svcCmd.CreateServiceInstanceCmd())
		createCmd.AddCommand(bindCmd.CreateIntegrationCmd())
		createCmd.AddCommand(clientCmd.CreateClientCmd())
		createCmd.AddCommand(serviceConfigCmd.CreateServiceConfigCmd())
		createCmd.AddCommand(clientBuilds.CreateClientBuildsCmd())
		rootCmd.AddCommand(createCmd)
	}
	//get
	{
		getCmd := cmd.NewGetCommand()
		getCmd.AddCommand(clientCmd.GetClientCmd())
		getCmd.AddCommand(clientCmd.ListClientsCmd())
		getCmd.AddCommand(serviceConfigCmd.GetServiceConfigCmd())
		getCmd.AddCommand(serviceConfigCmd.ListServiceConfigCmd())
		getCmd.AddCommand(clientCfgCmd.GetClientConfigCmd())
		getCmd.AddCommand(bindCmd.GetIntegrationCmd())
		getCmd.AddCommand(bindCmd.ListIntegrationsCmd())
		getCmd.AddCommand(clientBuilds.GetClientBuildsCmd())
		getCmd.AddCommand(clientBuilds.ListClientBuildsCmd())
		getCmd.AddCommand(svcCmd.ListServicesCmd())
		getCmd.AddCommand(svcCmd.ListServiceInstCmd())
		rootCmd.AddCommand(getCmd)
	}

	//set
	{
		setCmd := cmd.NewSetCommand()
		setCmd.AddCommand(clientCmd.SetClientValueFromJsonCmd())
		setCmd.AddCommand(clientCmd.SetClientSpecValueCmd())
		rootCmd.AddCommand(setCmd)
	}

	// delete
	{
		deleteCmd := cmd.NewDeleteComand()
		deleteCmd.AddCommand(bindCmd.DeleteIntegrationCmd())
		deleteCmd.AddCommand(clientCmd.DeleteClientCmd())
		deleteCmd.AddCommand(serviceConfigCmd.DeleteServiceConfigCmd())
		deleteCmd.AddCommand(clientBuilds.DeleteClientBuildsCmd())
		deleteCmd.AddCommand(svcCmd.DeleteServiceInstanceCmd())
		rootCmd.AddCommand(deleteCmd)
	}

	// stop
	{
		stopCmd := cmd.NewStopCmd()
		stopCmd.AddCommand(clientBuilds.StopClientBuildsCmd())
		rootCmd.AddCommand(stopCmd)
	}

	// start
	{
		startCmd := cmd.NewStartCmd()
		startCmd.AddCommand(clientBuilds.StartClientBuildsCmd())
		rootCmd.AddCommand(startCmd)
	}

	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		// as using pkg/errors lets allow the full stack to be seen if needed
		if os.Getenv("MCP_DEBUG") == "true" {
			log.Fatalf("error: %+v", err)
		}
		os.Exit(1)
	}
}

// NewClientsOrDie creates a new set of clients for Kubernetes, Service Catalog and Mobile Services
// if any of these clients fails to create then the process wil die.
func NewClientsOrDie(config *restclient.Config) (kubernetes.Interface, m.Interface, sc.Interface) {
	// create the K8client
	k8client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	scClientSet, err := sc.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	mobileClientSet, err := m.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return k8client, mobileClientSet, scClientSet
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
