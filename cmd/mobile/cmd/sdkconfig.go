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

	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// sdkconfigCmd represents the sdkconfig command
var sdkconfigCmd = &cobra.Command{
	Use:   "sdkconfig",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ret := []*ServiceConfig{}
		convertors := map[string]SecretConvertor{
			"fh-sync-server": &syncSecretConvertor{},
			"keycloak":       &keycloakSecretConvertor{},
		}
		ms := listServices()
		for _, svc := range ms {
			var svcConifg *ServiceConfig
			var err error
			if _, ok := convertors[svc.Name]; !ok {

				convertor := defaultSecretConvertor{}
				svcConifg, err = convertor.Convert(svc)
				if err != nil {
					//bail out here as now our config may not be compelete?
					log.Fatal("failed to convert to sdk config ", err)
				}
			} else {
				// we can only convert what is available
				convertor := convertors[svc.Name]
				svcConifg, err = convertor.Convert(svc)
				if err != nil {
					//bail out here as now our config may not be compelete?
					log.Fatal("failed to convert to sdkconfig ", err)
				}
			}
			ret = append(ret, svcConifg)
		}
		encoder := json.NewEncoder(os.Stdout)
		if err := encoder.Encode(ret); err != nil {
			log.Fatal("failed to encode sdk config ", err)
		}

	},
}

func init() {
	getCmd.AddCommand(sdkconfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sdkconfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sdkconfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
