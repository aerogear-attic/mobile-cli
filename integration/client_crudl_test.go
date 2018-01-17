package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

type MobileClientSpec struct {
	Name       string
	ApiKey     string
	ClientType string
}

type MobileClientJson struct {
	Spec MobileClientSpec
}

var namespace = flag.String("namespace", "", "Openshift namespace (most often Project) to run our integration tests in")
var name = flag.String("name", "", "Client name to be created")
var executable = flag.String("executable", "", "Executable under test")

type ValidationFunction = func(output []byte, err error) (bool, []string)

func EmptyValidation(output []byte, err error) (bool, []string) {
	return true, []string{}
}

func VNoErr(output []byte, err error) (bool, []string) {
	return err == nil, []string{fmt.Sprintf("%s", err)}
}

func VIsErr(output []byte, err error) (bool, []string) {
	return err != nil, []string{fmt.Sprintf("%s", err)}
}

func VRegex(pattern string) func(output []byte, err error) (bool, []string) {
	return func(output []byte, err error) (bool, []string) {
		matched, errMatch := regexp.MatchString(pattern, fmt.Sprintf("%s", output))
		if errMatch != nil {
			return false, []string{fmt.Sprintf("Error in regexp %s when trying to match %s", errMatch, pattern)}
		}
		if !matched {
			return false, []string{fmt.Sprintf("Expected combined output matching %s", pattern)}
		}
		return true, []string{}
	}
}
func VMobileClientJson(name string, clientType string) func(output []byte, err error) (bool, []string) {
	return func(output []byte, err error) (bool, []string) {
		var parsed MobileClientJson
		errJson := json.Unmarshal([]byte(output), &parsed)
		if errJson != nil {
			return false, []string{fmt.Sprintf("%s", err)}
		}
		if parsed.Spec.ClientType != clientType {
			return false, []string{fmt.Sprintf("Expected the ClientType to be %s, but got %s", clientType, parsed.Spec.ClientType)}
		}
		if parsed.Spec.Name != name {
			return false, []string{fmt.Sprintf("Expected the Name to be %s, but got %s", name, parsed.Spec.Name)}
		}
		return true, []string{}
	}
}

func All(vs ...ValidationFunction) ValidationFunction {
	return func(output []byte, err error) (bool, []string) {
		for _, v := range vs {
			r, o := v(output, err)
			if !r {
				return r, o
			}
		}
		return true, []string{}
	}
}

type CmdDesc struct {
	executable string
	Arg        []string
	Validator  ValidationFunction
}

func (c CmdDesc) Add(arg ...string) CmdDesc {
	return CmdDesc{c.executable, append(c.Arg, arg...), c.Validator}
}

func (c CmdDesc) Complying(validator ValidationFunction) CmdDesc {
	return CmdDesc{c.executable, c.Arg, All(c.Validator, validator)}
}

func (c CmdDesc) Run(t *testing.T) ([]byte, error) {
	t.Log(c.Arg)
	cmd := exec.Command(c.executable, c.Arg...)
	output, err := cmd.CombinedOutput()
	t.Log(fmt.Sprintf("%s\n", output))
	v, errs := c.Validator(output, err)
	if !v {
		t.Fatal(errs)
	}
	return output, err
}

func ValidatedCmd(executable string, arg ...string) CmdDesc {
	return CmdDesc{executable, arg, EmptyValidation}
}

func TestPositive(t *testing.T) {

	clientTypes := []string{
		"cordova",
		"iOS",
		"android",
	}

	m := ValidatedCmd(*executable, fmt.Sprintf("--namespace=%s", *namespace), "-o=json")
	for _, clientType := range clientTypes {
		t.Run(clientType, func(t *testing.T) {
			expectedId := strings.ToLower(fmt.Sprintf("%s-%s", *name, clientType))
			notExists := All(VIsErr, VRegex(".*Error: failed to get.*"))
			exists := All(VNoErr, VMobileClientJson(*name, clientType))
			m.Add("get", "client", expectedId).Complying(notExists).Run(t)
			m.Add("create", "client", *name, clientType).Complying(exists).Run(t)
			m.Add("get", "client", expectedId).Complying(exists).Run(t)
			m.Add("delete", "client", expectedId).Complying(VNoErr).Run(t)
			m.Add("get", "client", expectedId).Complying(notExists).Run(t)

		})
	}
}
