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

package cmd

import (
	"fmt"
	"log"
	"os"

	"flag"
	"path/filepath"

	sc "github.com/feedhenry/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset
var scClientSet *sc.Clientset

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "mobile",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().String("namespace", "", "--namespace=myproject")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	scClientSet, err = sc.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func mustGetStrFlag(flags *pflag.FlagSet, name string) string {
	val, err := flags.GetString(name)
	if err != nil {
		log.Fatal("failed to get flag "+name, err)
	}
	if val == "" {
		log.Fatal("missing required flag argument " + name + "\n")
	}
	return val
}

func currentNamespace() string {
	var ns = os.Getenv("KUBECTL_PLUGINS_CURRENT_NAMESPACE")
	var err error
	if ns == "" {
		ns, err = RootCmd.PersistentFlags().GetString("namespace")
		if err != nil {
			log.Fatal("failed to get flag namespace", err)
		}
	}
	if ns == "" {
		log.Fatal("no namespace present. Cannot continue. Please set the --namespace flat or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var")
	}
	return ns
}
