package integration

import (
	"regexp"
)

func CleanStringByRegex(input string, regexes []*regexp.Regexp) string {
	for _, regex := range regexes {
		input = regex.ReplaceAllString(input, "")
	}
	return input
}
