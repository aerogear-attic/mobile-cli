package integration

import (
	"fmt"
	"os/exec"
	"regexp"
	"testing"
)

type ValidationResult struct {
	Error   error
	Output  []byte
	Message []string
	Success bool
}

func (v ValidationResult) Test(t *testing.T) (output []byte, err error) {
	t.Log(fmt.Sprintf("%s\n", v.Output))
	if !v.Success {
		t.Fatal(v.Message)
	}
	return v.Output, v.Error
}

type ValidationFunction = func(output []byte, err error) ValidationResult

func SuccessValidation(output []byte, err error) ValidationResult {
	return ValidationResult{nil, output, []string{}, true}
}

func FailureValidation(output []byte, err error) ValidationResult {
	return ValidationResult{err, output, []string{fmt.Sprintf("%s", err)}, false}
}

func NoErr(output []byte, err error) ValidationResult {
	if err == nil {
		return SuccessValidation(output, err)
	}
	return FailureValidation(output, err)
}

func IsErr(output []byte, err error) ValidationResult {
	if err != nil {
		return SuccessValidation(output, err)
	}
	return ValidationResult{nil, output, []string{"Expected error to occur"}, false}
}

func ValidRegex(pattern string) func(output []byte, err error) ValidationResult {
	return func(output []byte, err error) ValidationResult {
		matched, errMatch := regexp.MatchString(pattern, fmt.Sprintf("%s", output))
		if errMatch != nil {
			return ValidationResult{
				Success: false,
				Message: []string{fmt.Sprintf("Error in regexp %s when trying to match %s", errMatch, pattern)},
				Error:   err,
				Output:  output,
			}

		}
		if !matched {
			return ValidationResult{
				Success: false,
				Message: []string{fmt.Sprintf("Expected combined output matching %s", pattern)},
				Error:   nil,
				Output:  output,
			}

		}
		return SuccessValidation(output, err)
	}
}

func All(vs ...ValidationFunction) ValidationFunction {
	return func(output []byte, err error) ValidationResult {
		for _, v := range vs {
			r := v(output, err)
			if !r.Success {
				return r
			}
		}
		return ValidationResult{nil, output, []string{}, true}
	}
}

type CmdDesc struct {
	Executable string
	Arg        []string
	Validator  ValidationFunction
}

func (c CmdDesc) Args(arg ...string) CmdDesc {
	return CmdDesc{c.Executable, append(c.Arg, arg...), c.Validator}
}

func (c CmdDesc) Should(validator ValidationFunction) CmdDesc {
	return CmdDesc{c.Executable, c.Arg, All(c.Validator, validator)}
}

func (c CmdDesc) Run() ValidationResult {
	cmd := exec.Command(c.Executable, c.Arg...)
	output, err := cmd.CombinedOutput()
	return c.Validator(output, err)
}

func ValidatedCmd(executable string, arg ...string) CmdDesc {
	return CmdDesc{executable, arg, SuccessValidation}
}
