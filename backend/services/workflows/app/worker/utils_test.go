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

package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/services/workflows/app/processor"
	"github.com/mendersoftware/mender-server/services/workflows/model"
)

func TestProcessJobString(t *testing.T) {
	workflow := &model.Workflow{
		Name: "test",
		InputParameters: []string{
			"key",
		},
	}
	job := &model.Job{
		InputParameters: []model.InputParameter{
			{
				Name:  "key",
				Value: "test+test",
			},
		},
	}

	ps := processor.NewJobStringProcessor(workflow, job)
	res := ps.ProcessJobString("_${workflow.input.key}_")
	assert.Equal(t, "_test+test_", res)

	res = ps.ProcessJobString("_${workflow.input.another_key|default}_")
	assert.Equal(t, "_default_", res)

	res = ps.ProcessJobString("_${encoding=url;workflow.input.key}_")
	assert.Equal(t, "_test%2Btest_", res)

	res = ps.ProcessJobString("_${encoding=plain;workflow.input.key}_")
	assert.Equal(t, "_test+test_", res)
}

func TestProcessJobStringEnvVariable(t *testing.T) {
	workflow := &model.Workflow{
		Name: "test",
	}
	job := &model.Job{}

	ps := processor.NewJobStringProcessor(workflow, job)
	res := ps.ProcessJobString("_${env.PWD}_")
	pwd := os.Getenv("PWD")
	expected := fmt.Sprintf("_%s_", pwd)
	assert.Equal(t, expected, res)

	res = ps.ProcessJobString("_${env.ENV_VARIABLE_WHICH_DOES_NOT_EXIST|default}_")
	expected = "_default_"
	assert.Equal(t, expected, res)
}

func TestProcessJobStringJSONOutputFromPreviousResult(t *testing.T) {
	var tests = []struct {
		taskResult    model.TaskResult
		expression    string
		expectedValue string
	}{
		{
			taskResult: model.TaskResult{
				Name:    "task_1",
				Type:    model.TaskTypeHTTP,
				Success: true,
				HTTPResponse: &model.TaskResultHTTPResponse{
					StatusCode: http.StatusOK,
					Body:       "{\"key\": \"value\"}",
				},
			},
			expression:    "_${task_1.json.key}_",
			expectedValue: "_value_",
		},
		{
			taskResult: model.TaskResult{
				Name:    "task_1",
				Type:    model.TaskTypeHTTP,
				Success: true,
				HTTPResponse: &model.TaskResultHTTPResponse{
					StatusCode: http.StatusOK,
					Body:       "{\"key\": {\"subkey\": \"value\"}}",
				},
			},
			expression:    "_${task_1.json.key.subkey}_",
			expectedValue: "_value_",
		},
		{
			taskResult: model.TaskResult{
				Name:    "task_1",
				Type:    model.TaskTypeHTTP,
				Success: true,
				HTTPResponse: &model.TaskResultHTTPResponse{
					StatusCode: http.StatusOK,
					Body:       "{\"key\": {\"subkey\": 1}}",
				},
			},
			expression:    "_${task_1.json.key.subkey}_",
			expectedValue: "_1_",
		},
		{
			taskResult: model.TaskResult{
				Name:    "task_1",
				Type:    model.TaskTypeHTTP,
				Success: true,
				HTTPResponse: &model.TaskResultHTTPResponse{
					StatusCode: http.StatusOK,
					Body:       "{\"key\": {\"subkey\": [\"value\", \"value2\"]}}",
				},
			},
			expression:    "_${task_1.json.key.subkey.[1]}_",
			expectedValue: "_value2_",
		},
		{
			taskResult: model.TaskResult{
				Name:    "task_1",
				Type:    model.TaskTypeCLI,
				Success: true,
				CLI: &model.TaskResultCLI{
					ExitCode: 0,
					Output:   "dummy",
				},
			},
			expression:    "_${task_1.json.key}_",
			expectedValue: "__",
		},
		{
			taskResult: model.TaskResult{
				Name:    "task_1",
				Type:    model.TaskTypeCLI,
				Success: true,
				CLI: &model.TaskResultCLI{
					ExitCode: 0,
					Output:   "dummy",
				},
			},
			expression:    "_${task_1.json.key|default}_",
			expectedValue: "_default_",
		},
	}

	for _, test := range tests {
		workflow := &model.Workflow{
			Name: "test",
		}
		job := &model.Job{
			Results: []model.TaskResult{test.taskResult},
		}

		ps := processor.NewJobStringProcessor(workflow, job)
		res := ps.ProcessJobString(test.expression)
		assert.Equal(t, test.expectedValue, res)
	}
}

func TestProcessJobJSON(t *testing.T) {
	var tests = map[string]struct {
		json       interface{}
		result     interface{}
		resultJSON string
	}{
		"string": {
			json:       "_${workflow.input.key}_",
			result:     "_test_",
			resultJSON: `"_test_"`,
		},
		"number": {
			json:       "${workflow.input.num}",
			result:     1,
			resultJSON: "1",
		},
		"map": {
			json: map[string]interface{}{
				"key":         "_${workflow.input.key}_",
				"other-key":   "other-value",
				"numeric-key": 1,
			},
			result: map[string]interface{}{
				"key":         "_test_",
				"other-key":   "other-value",
				"numeric-key": 1,
			},
			resultJSON: `{"key":"_test_","numeric-key":1,"other-key":"other-value"}`,
		},
		"nested map": {
			json: map[string]interface{}{
				"parent": map[string]interface{}{
					"key": "_${workflow.input.key}_",
				},
				"other-key":   "other-value",
				"numeric-key": 1,
			},
			result: map[string]interface{}{
				"parent": map[string]interface{}{
					"key": "_test_",
				},
				"other-key":   "other-value",
				"numeric-key": 1,
			},
			resultJSON: `{"numeric-key":1,"other-key":"other-value","parent":{"key":"_test_"}}`,
		},
		"list": {
			json: []interface{}{
				map[string]interface{}{
					"key": "_${workflow.input.key}_",
				},
				map[string]interface{}{
					"other-key": "other-value",
				},
			},
			result: []interface{}{
				map[string]interface{}{
					"key": "_test_",
				},
				map[string]interface{}{
					"other-key": "other-value",
				},
			},
			resultJSON: `[{"key":"_test_"},{"other-key":"other-value"}]`,
		},
		"bson.D": {
			json: bson.D{
				{Key: "key", Value: "_${workflow.input.key}_"},
				{Key: "other-key", Value: "other-value"},
			},
			result: map[string]interface{}{
				"key":       "_test_",
				"other-key": "other-value",
			},
			resultJSON: `{"key":"_test_","other-key":"other-value"}`,
		},
		"list of bson.D": {
			json: []bson.D{
				{
					{Key: "key", Value: "_${workflow.input.key}_"},
					{Key: "other-key", Value: "other-value"},
				},
			},
			result: []interface{}{
				map[string]interface{}{
					"key":       "_test_",
					"other-key": "other-value",
				},
			},
			resultJSON: `[{"key":"_test_","other-key":"other-value"}]`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			workflow := &model.Workflow{
				Name: "test",
				InputParameters: []string{
					"key",
					"num",
				},
			}
			job := &model.Job{
				InputParameters: []model.InputParameter{
					{
						Name:  "key",
						Value: "test",
					},
					{
						Name:  "num",
						Value: "1",
						Raw:   1,
					},
				},
			}

			ps := processor.NewJobStringProcessor(workflow, job)
			jp := processor.NewJobProcessor(job)

			res := jp.ProcessJSON(test.json, ps)
			assert.Equal(t, test.result, res)

			jsonString, err := json.Marshal(res)
			assert.NoError(t, err)
			assert.Equal(t, test.resultJSON, string(jsonString))
		})
	}
}
