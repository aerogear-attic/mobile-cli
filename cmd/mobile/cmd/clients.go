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

	"encoding/json"
	"log"
	"os"

	"time"

	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// appsCmd represents the apps command
var appsCmd = &cobra.Command{
	Use:   "clients",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		maps, err := clientset.CoreV1().ConfigMaps(currentNamespace()).List(metav1.ListOptions{LabelSelector: "group=mobileapp"})
		if err != nil {
			panic(err.Error())
		}
		apps := []*App{}
		for _, cm := range maps.Items {
			app := convertConfigMapToMobileApp(cm)
			apps = append(apps, app)
		}

		e := json.NewEncoder(os.Stdout)
		if err := e.Encode(apps); err != nil {
			log.Fatal("failed to encode mobile apps ", err)
		}

	},
}

var createAppCmd = &cobra.Command{
	Use:   "client",
	Short: "create a mobile app",
	Long:  `Sets up the definition of a mobile application in your namespace. create client <name> <clientType>`,
	Run: func(cmd *cobra.Command, args []string) {
		name := mustGetStrFlag(cmd.PersistentFlags(), "name")
		clientType := mustGetStrFlag(cmd.PersistentFlags(), "clientType")
		appKey := uuid.NewV4().String()
		app := &App{Name: name, ClientType: clientType, APIKey: appKey, Labels: map[string]string{}, MetaData: map[string]string{}}
		switch app.ClientType {
		case "android":
			app.MetaData["icon"] = "fa-android"
			break
		case "iOS":
			app.MetaData["icon"] = "fa-apple"
			break
		case "cordova":
			app.MetaData["icon"] = "icon-cordova"
			break
		}
		app.MetaData["created"] = time.Now().Format("2006-01-02 15:04:05")
		app.ID = app.Name + "-" + fmt.Sprintf("%v", time.Now().Unix())
		cmap := convertMobileAppToConfigMap(app)
		if _, err := clientset.CoreV1().ConfigMaps(currentNamespace()).Create(cmap); err != nil {
			log.Fatal("error creating backing configmap", err)
		}

	},
}

func init() {
	createAppCmd.PersistentFlags().String("name", "myapp", "--name=myapp")
	createAppCmd.PersistentFlags().String("clientType", "cordova", "--clientType=cordova")
	appsCmd.PersistentFlags().String("foo", "", "A help for foo")
	getCmd.AddCommand(appsCmd)
	createCmd.AddCommand(createAppCmd)
}
