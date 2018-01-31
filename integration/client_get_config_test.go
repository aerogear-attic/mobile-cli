package integration

import (
	"fmt"
	"os/exec"
	"regexp"
	"testing"
)

type Client struct {
	ID        string
	Name      string
	Type      string
	Namespace string
	BundleID  string
}

const getClientTestPath = "getClientConfigTestData/"

func TestGetClientConfig(t *testing.T) {
	//regexes to match dynamic properties in the client config
	regexes := []*regexp.Regexp{
		regexp.MustCompile("\"cluster_name\".*?,"),
		regexp.MustCompile("\"namespace\".*?,"),
	}

	client := &Client{
		ID:        "myapp-cordova",
		Name:      "myapp",
		Type:      "cordova",
		Namespace: fmt.Sprintf("--namespace=%s", *namespace),
		BundleID:  "my.app.org",
	}

	createTestClient(t, client)

	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"json output", []string{"get", "clientconfig", client.ID, client.Namespace, "-o=json"}, "json-output.golden"},
		{"table output", []string{"get", "clientconfig", client.ID, client.Namespace, "-o=table"}, "table-output.golden"},
		{"no clientID", []string{"get", "clientconfig", client.Namespace}, "no-client-id.golden"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.Command(*executable, test.args...)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatal(err)
			}

			actual := string(output)
			expected := LoadSnapshot(t, getClientTestPath+test.fixture)

			if *update {
				WriteSnapshot(t, getClientTestPath+test.fixture, output)
			}

			if test.name == "json output" {
				actual = CleanStringByRegex(actual, regexes)
			}

			if actual != expected {
				t.Fatalf("actual = \n%s, expected = \n%s", actual, expected)
			}
		})
	}

	deleteTestClient(t, client)
}

func createTestClient(t *testing.T, client *Client) {
	createClientCmdArgs := []string{"create", "client", client.Name, client.Type, client.BundleID, client.Namespace}
	createClientCmd := exec.Command(*executable, createClientCmdArgs...)

	output, err := createClientCmd.CombinedOutput()
	if err != nil {
		t.Fatal("Failed to create client: ", string(output))
	}
}

func deleteTestClient(t *testing.T, client *Client) {
	deleteClientCmdArgs := []string{"delete", "client", client.ID, client.Namespace}
	deleteClientCmd := exec.Command(*executable, deleteClientCmdArgs...)

	output, err := deleteClientCmd.CombinedOutput()
	if err != nil {
		t.Fatal("Failed to delete client: ", string(output))
	}
}
