package integration

import (
	"os/exec"
	"regexp"
	"testing"
)

const testPath = "getServicesTestData/"

func TestGetServices(t *testing.T) {
	//regexes to match dynamic properties in service json, UID resourceVersion and creationTimestamp
	regexes := []*regexp.Regexp{
		regexp.MustCompile("\"uid\".*?,"),
		regexp.MustCompile("\"resourceVersion\".*?,"),
		regexp.MustCompile("\"creationTimestamp\".*?,"),
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
				t.Fatal(err)
			}
			if *update {
				WriteSnapshot(t, testPath+test.fixture, output)
			}

			actual := string(output)

			expected := LoadSnapshot(t, testPath+test.fixture)

			if test.name == "json output" {
				actual = cleanStringByRegex(actual, regexes)
				expected = cleanStringByRegex(expected, regexes)
			}

			if actual != expected {
				t.Fatalf("actual = \n%s, expected = \n%s", actual, expected)
			}
		})
	}

}

func cleanStringByRegex(input string, regexes []*regexp.Regexp) string {
	for _, regex := range regexes {
		input = regex.ReplaceAllString(input, "")
	}
	return input
}
