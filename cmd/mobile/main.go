// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

	"github.com/feedhenry/mobile-cli/cmd/mobile/cmd"
	m "github.com/feedhenry/mobile-cli/pkg/client/mobile/clientset/versioned"
	sc "github.com/feedhenry/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}

	k8Client, mobileClient, scClient := NewClientsOrDie(*kubeconfig)

	var rootCmd = cmd.NewRootCmd()
	var clientCmd = cmd.NewClientCmd(mobileClient)
	var bindCmd = cmd.NewBindCmd(scClient, k8Client)
	var serviceConfigCmd = cmd.NewServiceConfigCommand(k8Client)
	// create
	{
		createCmd := cmd.NewCreateCommand()

		createCmd.AddCommand(bindCmd.BuildCreateBindCmd())
		createCmd.AddCommand(clientCmd.CreateClientCmd())
		rootCmd.AddCommand(createCmd)

	}
	//get
	{
		getCmd := cmd.NewGetCommand()
		getCmd.AddCommand(clientCmd.GetClientCmd())
		getCmd.AddCommand(clientCmd.ListClientsCmd())
		getCmd.AddCommand(serviceConfigCmd.BuildGetServiceConfigCmd())
		getCmd.AddCommand(serviceConfigCmd.BuildListServiceConfigCmd())
		rootCmd.AddCommand(getCmd)
	}
	// delete
	{
		deleteCmd := cmd.NewDeleteComand()
		deleteCmd.AddCommand(bindCmd.BuildDeleteBindCmd())
		deleteCmd.AddCommand(clientCmd.DeleteClientCmd())
		rootCmd.AddCommand(deleteCmd)
	}

	if err := rootCmd.Execute(); err != nil {

		os.Exit(1)
	}
}

func NewClientsOrDie(configLoc string) (kubernetes.Interface, m.Interface, sc.Interface) {
	config, err := clientcmd.BuildConfigFromFlags("", configLoc)
	if err != nil {
		panic(err.Error())
	}

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
