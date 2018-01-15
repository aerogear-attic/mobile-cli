package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"testing"
)

var namespace = flag.String("namespace", "", "Openshift namespace (most often Project) to run our integration tests in")
var name = flag.String("name", "", "Client name to be created")
var executable = flag.String("executable", "", "Executable under test")

type MobileClientSpec struct {
	Name       string
	ApiKey     string
	ClientType string
}

type MobileClientJson struct {
	Spec MobileClientSpec
}

func TestPositive(t *testing.T) {

	actions := []struct {
		Name     string
		Args     func(clientType string) []string
		Validate func(clientType string, output []byte)
	}{
		{
			Name: "Create",
			Args: func(clientType string) []string {
				return []string{"create", "client", *name, clientType}
			},
			Validate: func(clientType string, output []byte) {
				var parsed MobileClientJson
				errJson := json.Unmarshal([]byte(output), &parsed)
				if errJson != nil {
					t.Fatal(errJson)
				}
				if parsed.Spec.ClientType != clientType {
					t.Fatal(fmt.Sprintf("Expected the ClientType to be %s, but got %s", clientType, parsed.Spec.ClientType))
				}
				if parsed.Spec.Name != *name {
					t.Fatal(fmt.Sprintf("Expected the Name to be %s, but got %s", *name, parsed.Spec.Name))
				}
			},
		},
		/*{
			Name: "Get",
			Args: func(clientType string) []string {
				return []string{"get", "client", fmt.Sprintf("%s-%s", *name, clientType)}
			},
			Validate: func(clientType string, output []byte) {
				var parsed MobileClientJson
				errJson := json.Unmarshal([]byte(output), &parsed)
				if errJson != nil {
					t.Fatal(errJson)
				}
				if parsed.Spec.ClientType != clientType {
					t.Fatal(fmt.Sprintf("Expected the ClientType to be %s, but got %s", clientType, parsed.Spec.ClientType))
				}
				if parsed.Spec.Name != *name {
					t.Fatal(fmt.Sprintf("Expected the Name to be %s, but got %s", *name, parsed.Spec.Name))
				}
			},
		},
		{
			Name: "Delete",
		},*/
	}

	clientTypes := []string{
		"cordova",
		"iOS",
		"android",
	}

	for _, clientType := range clientTypes {
		for _, action := range actions {
			t.Run(fmt.Sprintf("%s/%s", clientType, action.Name), func(t *testing.T) {
				args := append(action.Args(clientType), fmt.Sprintf("--namespace=%s", *namespace), "-o=json")
				cmd := exec.Command(*executable, args...)
				output, errCommand := cmd.CombinedOutput()
				t.Log(fmt.Sprintf("%s\n", output))
				if errCommand != nil {
					t.Fatal(errCommand)
				}
				action.Validate(clientType, output)
			})
		}
	}
}
