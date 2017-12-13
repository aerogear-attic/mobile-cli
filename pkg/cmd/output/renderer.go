package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/pkg/errors"
)

type Renderer struct {
	out io.Writer
}

func (r *Renderer) Render(cmd, outputType string, data interface{}) error {
	if strings.ToLower(outputType) == "json" {
		encoder := json.NewEncoder(r.out)
		return encoder.Encode(data)
	}

	if render, ok := renderers[cmd+outputType]; ok {
		if err := render(r.out, data); err != nil {
			return err
		}
		return nil
	}

	return errors.New("no renderer registed for " + cmd + outputType)
}

func (r *Renderer) AddRenderer(name, outType string, renderer func(out io.Writer, data interface{}) error) {
	renderers[name+outType] = renderer
}

var renderers = map[string]func(out io.Writer, data interface{}) error{}

func NewRenderer(out io.Writer) *Renderer {
	return &Renderer{out: out}
}

const (
	FailedToOutPutInFormat = "failed to output %s in format %s"
)
