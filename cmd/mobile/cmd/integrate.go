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
	"log"

	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// integrateCmd represents the integrate command
var createIntegrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "integrate mobile services together",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatal("expected a service to bind from and to ")
		}
		sc := serviceCatalogClient{
			k8client: clientset,
		}
		from := args[0]
		to := args[1]

		fromSvc := getService(from)
		toSvc := getService(to)
		bindParams := buildBindParams(fromSvc, toSvc)
		if err := sc.BindToService(from, to, bindParams, currentNamespace(), currentNamespace()); err != nil {
			log.Fatal("failed to bind to service ", err)
		}
	},
}

// integrateCmd represents the integrate command
var deleteIntegrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "disintegrate mobile services together",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatal("expected a service to bind from and to ")
		}
		from := args[0]
		to := args[1]
		sc := serviceCatalogClient{}
		if err := sc.UnBindFromService(from, to, currentNamespace()); err != nil {
			log.Fatal(err)
		}
	},
}

func buildBindParams(from *Service, to *Service) map[string]interface{} {
	var p = map[string]interface{}{}
	p["credentials"] = map[string]string{
		"route":          from.Host,
		"service_secret": to.ID,
	}

	for k, v := range from.Params {
		p[k] = v
	}
	if from.Name == ServiceNameThreeScale {
		p["apicast_route"] = from.Params["apicast_route"]
		p["apicast_token"] = from.Params["token"]
		p["apicast_service_id"] = from.Params["service_id"]
		p["service_route"] = to.Host
		p["service_name"] = to.Name
		p["app_key"] = uuid.NewV4().String()
		p["service_secret"] = to.ID
	} else if from.Name == ServiceNameKeycloak {
		p["service_name"] = to.Name
	}
	return p
}

func init() {
	createCmd.AddCommand(createIntegrationCmd)
	deleteCmd.AddCommand(deleteIntegrationCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// integrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// integrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
