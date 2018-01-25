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

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "mobile",
		Short: "A brief description of your application",
		Long:  ``,
	}
	root.PersistentFlags().String("namespace", "", "--namespace=myproject")
	root.PersistentFlags().StringP("output", "o", "table", "-o=json -o=template")
	cobra.OnInitialize(initConfig)
	return root
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
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

func currentNamespace(flags *pflag.FlagSet) (string, error) {
	var err error
	var ns = os.Getenv("KUBECTL_PLUGINS_CURRENT_NAMESPACE")
	if ns == "" {
		if ns, err = flags.GetString("namespace"); ns == "" {
			err = errors.New("no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var")
		}
	}
	return ns, err
}

func outputType(flags *pflag.FlagSet) string {
	o, _ := flags.GetString("output")

	if o == "" {
		o = "json"
	}
	return o
}
