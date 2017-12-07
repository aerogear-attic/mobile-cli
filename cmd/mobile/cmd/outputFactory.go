package cmd

import (
	"encoding/json"
	"io"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type OutPutFactory struct {
	out io.Writer
}

func (op *OutPutFactory) Output(cmd, outputType string, data interface{}) error {
	if strings.ToLower(outputType) == "json" {
		encoder := json.NewEncoder(op.out)
		return encoder.Encode(data)
	}
	if strings.ToLower(outputType) == "template" {
		t := template.New(cmd)
		t, err := t.Parse(templates[cmd])
		if err != nil {
			return errors.Wrap(err, "failed to parse output template "+cmd)
		}
		if err := t.Execute(op.out, data); err != nil {
			return errors.Wrap(err, "failed to write the template")
		}
	}
	return nil
}

func NewOutPutFactory(out io.Writer) OutPutFactory {
	return OutPutFactory{out: out}
}

const (
	failedToOutPutInFormat = "failed to output %s in format %s"
)
