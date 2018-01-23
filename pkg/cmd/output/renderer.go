// Copyright Red Hat, Inc., and individual contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
