package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"testing"
)

var namespace = flag.String("namespace", "myproject", "Openshift namespace (most often Project) to run our integration tests in")
var name = flag.String("name", "", "Client name to be created")
var executable = flag.String("executable", "mobile", "Executable under test")

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
	}
	outputTypes := []string{
		"-o=json",
		"",
	}
for _, outputType := range outputTypes{
	for _, action := range actions {
		t.Run(fmt.Sprintf("/%s", action.Name), func(t *testing.T) {
			args := append(action.Args(fmt.Sprintf("--namespace=%s", *namespace)),  )
			cmd := exec.Command(*executable, args...)
			output, errCommand := cmd.CombinedOutput()
			t.Log(fmt.Sprintf("%s\n", output))
			if errCommand != nil {
				t.Fatal(errCommand)
			}
			action.Validate(outputType, output)
		})
	}
}

}
