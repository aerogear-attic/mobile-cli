package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// ServiceParams for creating integration and binding
type ServiceParams struct {
	AdditionalProperties bool                              `json:"additionalProperties"`
	Properties           map[string]map[string]interface{} `json:"properties"`
	Required             []string                          `json:"required"`
	Type                 string                            `json:"type"`
}

func parseParams(keyVals []string) (map[string]string, error) {
	params := map[string]string{}
	for _, p := range keyVals {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			return nil, NewIncorrectParameterFormat("key value pairs are needed failed to find one: " + p)
		}
		params[strings.TrimSpace(kv[0])] = kv[1]
	}
	return params, nil
}

func isRequired(params ServiceParams, key string) bool {
	for _, r := range params.Required {
		if r == key {
			return true
		}
	}
	return false
}

// GetParams - Gets the service parameters (i.e. for provision/bind service) from the params
//             flag or as a user input
func GetParams(flagParams []string, params *ServiceParams) (*ServiceParams, error) {
	parsedParams, err := parseParams(flagParams)
	if err != nil {
		return params, errors.WithStack(err)
	}

	if len(parsedParams) > 0 {
		for k, v := range params.Properties {
			defaultVal := v["default"]
			if pVal, ok := parsedParams[k]; !ok && isRequired(*params, k) || isRequired(*params, k) && pVal == "" {
				if defaultVal != nil {
					//use default
					v["value"] = defaultVal
					continue
				}
				return params, errors.New(fmt.Sprintf("missing required parameter %s", k))
			}
			v["value"] = parsedParams[k]
			params.Properties[k] = v
		}
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for k, v := range params.Properties {
			validInput := false
			val := ""
			for validInput == false {
				questionFormat := "Set value for %s [default value: %s, required: %v]"
				if v["default"] != nil {
					fmt.Println(fmt.Sprintf(questionFormat, k, v["default"], isRequired(*params, k)))
				} else {
					fmt.Println(fmt.Sprintf(questionFormat, k, "<no default value>", isRequired(*params, k)))
				}
				scanner.Scan()

				val = strings.TrimSpace(scanner.Text())

				if len(val) > 0 {
					validInput = true
				}
				if validInput == false && val == "" && v["default"] != nil {
					val = v["default"].(string)
					validInput = true
				}
				if validInput == false && val == "" && !isRequired(*params, k) {
					validInput = true
				}
				if validInput == false {
					fmt.Println("Invalid option for required field.")
				}
			}
			v["value"] = val
			params.Properties[k] = v
			fmt.Println(fmt.Sprintf("Value for %s set to: %s", k, val))
		}
	}
	return params, nil
}
