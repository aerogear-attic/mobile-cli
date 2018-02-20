package integration

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

const getServicesTestPath = "getServicesTestData/"

func TestGetServices(t *testing.T) {
	//regexes to match dynamic properties in service json, UID resourceVersion and creationTimestamp
	regexes := []*regexp.Regexp{
		regexp.MustCompile("\"uid\".*?,"),
		regexp.MustCompile("\"resourceVersion\".*?,"),
		regexp.MustCompile("\"creationTimestamp\": \".*?\""),
		regexp.MustCompile("\"removedFromBrokerCatalog\": (true|false)"),
	}

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
			cmd := exec.Command(*executable, test.args...)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("output: %v\n", string(output))
				t.Fatal(err)
			}
			if *update {
				WriteSnapshot(t, getServicesTestPath+test.fixture, output)
			}

			actual := string(output)

			expected := LoadSnapshot(t, getServicesTestPath+test.fixture)

			if test.name == "json output" {
				actual = strings.TrimSpace(CleanStringByRegex(actual, regexes))
				expected = strings.TrimSpace(CleanStringByRegex(expected, regexes))
			}

			if actual != expected {
				t.Fatalf("actual = \n'%s', expected = \n'%s'", actual, expected)
			}
		})
	}
}
