package integration

import (
	"fmt"
	"os/exec"
	"regexp"
	"testing"
)

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
