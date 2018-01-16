package integration

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
)

var namespace = flag.String("namespace", "myproject", "Openshift namespace (most often Project) to run our integration tests in")
var executable = flag.String("executable", "mobile", "Executable under test")
var update = flag.Bool("update", false, "update golden files")

const testPath = "getServicesTestData/"

func TestGetServices(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"no arguments", []string{"get", "services"}, "no-args.golden"},
		{"json output", []string{"get", "services", "-o=json"}, "json-output.golden"},
		{"table output", []string{"get", "services", "-o=table"}, "table-output.golden"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(path.Join(dir, *executable), test.args...)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatal(err)
			}
			if *update {
				WriteSnapshot(t, testPath+test.fixture, output)
			}

			actual := string(output)

			expected := LoadSnapshot(t, testPath+test.fixture)

			if actual != expected {
				t.Fatalf("actual = \n%s, expected = \n%s", actual, expected)
			}
		})
	}

}

func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
