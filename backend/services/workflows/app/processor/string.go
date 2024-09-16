// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package processor

import (
	"html"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/thedevsaddam/gojsonq"

	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/utils"
)

const (
	workflowEnvVariable   = "env."
	workflowInputVariable = "workflow.input."
	regexVariable         = `\$\{(?P<options>(?:(?:[a-zA-Z]+)=(?:[a-zA-Z0-9]+);)*)` +
		`(?P<name>[^;\}\|]+)(?:\|(?P<default>[^\}]+))?}`
	regexOutputVariable = `(.*)\.json\.(.*)`

	encodingFlag = "encoding"
	encodingURL  = "url"
	encodingHTML = "html"
)

var (
	reExpression       = regexp.MustCompile(regexVariable)
	reExpressionOutput = regexp.MustCompile(regexOutputVariable)

	reMatchIndexOptions = reExpression.SubexpIndex("options")
	reMatchIndexName    = reExpression.SubexpIndex("name")
	reMatchIndexDefault = reExpression.SubexpIndex("default")
)

type Encoding int64

const (
	EncodingPlain Encoding = iota
	EncodingURL
	EncodingHTML
)

func ParseString(s string) Encoding {
	var enc Encoding
	switch s {
	case encodingURL:
		enc = EncodingURL
	case encodingHTML:
		enc = EncodingHTML
	default:
		enc = EncodingPlain
	}
	return enc
}

func (enc Encoding) Apply(s string) string {
	switch enc {
	case EncodingURL:
		return url.QueryEscape(s)
	case EncodingHTML:
		return html.EscapeString(s)
	}
	return s
}

type JobStringProcessor struct {
	job *model.Job
}

type Options struct {
	Encoding Encoding
}

func NewJobStringProcessor(
	job *model.Job,
) *JobStringProcessor {
	return &JobStringProcessor{
		job: job,
	}
}

func processOptionString(expression string) (opts Options) {
	const (
		flagTokenCount = 2
		lValueIndex    = 0
		rValueIndex    = 1
	)
	for _, flagToken := range strings.Split(expression, ";") {
		flagValueTokens := strings.SplitN(flagToken, "=", 2)
		if len(flagValueTokens) < flagTokenCount {
			continue
		}
		if flagValueTokens[lValueIndex] == encodingFlag {
			switch flagValueTokens[rValueIndex] {
			case encodingURL:
				opts.Encoding = EncodingURL
			case encodingHTML:
				opts.Encoding = EncodingHTML
			}
		}
	}
	return
}

func (j *JobStringProcessor) ProcessJobString(data string) string {
	matches := reExpression.FindAllStringSubmatch(data, -1)

	// search for ${...} expressions in the data string
SubMatchLoop:
	for _, submatch := range matches {
		// content of the ${...} expression, without the brackets
		varName := submatch[reMatchIndexName]
		value := submatch[reMatchIndexDefault]
		options := processOptionString(submatch[reMatchIndexOptions])
		// now it is possible to override the encoding with flags: ${encoding=plain;identifier}
		// if encoding is supplied via flags, it takes precedence, we return the match
		// without the flags, otherwise fail back to original match and encoding
		if strings.HasPrefix(varName, workflowInputVariable) &&
			len(varName) > len(workflowInputVariable) {
			// Replace ${workflow.input.KEY} with the KEY input variable
			paramName := varName[len(workflowInputVariable):]
			for _, param := range j.job.InputParameters {
				if param.Name == paramName {
					value = param.Value
					break
				}
			}
		} else if strings.HasPrefix(varName, workflowEnvVariable) &&
			len(varName) > len(workflowEnvVariable) {
			// Replace ${env.KEY} with the KEY environment variable
			envName := varName[len(workflowEnvVariable):]
			if envValue := os.Getenv(envName); envValue != "" {
				value = envValue
			}
		} else if output := reExpressionOutput.FindStringSubmatch(varName); len(output) > 0 {
			// Replace ${TASK_NAME.json.JSONPATH} with the value of the JSONPATH expression from the
			// JSON output of the previous task with name TASK_NAME. If the output is not a valid
			// JSON or the JSONPATH does not resolve to a value, replace with empty string
			for _, result := range j.job.Results {
				if result.Name == output[1] {
					varKey := output[2]
					var output string
					if result.Type == model.TaskTypeHTTP {
						output = result.HTTPResponse.Body
					} else if result.Type == model.TaskTypeCLI {
						output = result.CLI.Output
					} else {
						continue
					}
					varValue := gojsonq.New().FromString(output).Find(varKey)
					if varValue == nil {
						varValue = ""
					}
					varValueString, err := utils.ConvertAnythingToString(varValue)
					if err == nil {
						if varValueString != "" {
							value = varValueString
						}
					} else {
						continue SubMatchLoop
					}
					break
				}
			}
		}
		value = options.Encoding.Apply(value)
		data = strings.ReplaceAll(data, submatch[0], value)
	}

	return data
}
