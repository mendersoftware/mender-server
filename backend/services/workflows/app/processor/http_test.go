// Copyright 2021 Northern.tech AS
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
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store/mock"
)

func TestProcessJobHTTP(t *testing.T) {
	const responseBody = "HTTP/Response"
	testCases := []struct {
		Name string

		Workflow        *model.Workflow
		JobRequest      *model.TaskResultHTTPRequest
		InputParameters model.InputParameters
		Error           interface{}
	}{{
		Name: "Plain GET Task",
		Workflow: &model.Workflow{
			Name: "test",
			Tasks: []model.Task{
				{
					Name: "task_1",
					Type: model.TaskTypeHTTP,
					HTTP: &model.HTTPTask{
						URI:    "http://localhost",
						Method: http.MethodGet,
						Headers: map[string]string{
							"X-Header": "Value",
						},
					},
				},
			},
		},
		JobRequest: &model.TaskResultHTTPRequest{
			URI:     "http://localhost",
			Method:  http.MethodGet,
			Headers: []string{"X-Header: Value"},
		},
		Error: nil,
	}, {
		Name: "POST JSON",
		Workflow: &model.Workflow{
			Name: "test",
			Tasks: []model.Task{
				{
					Name: "task_1",
					Type: model.TaskTypeHTTP,
					HTTP: &model.HTTPTask{
						URI:    "http://localhost",
						Method: http.MethodPost,
						JSON: map[string]interface{}{
							"foo": "${workflow.input.foo}",
						},
					},
				},
			},
		},
		JobRequest: &model.TaskResultHTTPRequest{
			URI:    "http://localhost",
			Method: http.MethodPost,
			Body:   `{"foo":"bar"}`,
		},
		InputParameters: model.InputParameters{
			{Name: "foo", Value: "bar", Raw: "bar"},
		},
		Error: nil,
	}, {
		Name: "POST form data",
		Workflow: &model.Workflow{
			Name: "test",
			Tasks: []model.Task{
				{
					Name: "task_1",
					Type: model.TaskTypeHTTP,
					HTTP: &model.HTTPTask{
						URI:    "http://localhost",
						Method: http.MethodPost,
						FormData: map[string]string{
							"key": "value to be encoded",
						},
					},
				},
			},
		},
		JobRequest: &model.TaskResultHTTPRequest{
			URI:    "http://localhost",
			Method: http.MethodPost,
			Body:   "key=value+to+be+encoded",
		},
		Error: nil,
	}}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			dataStore := mock.NewDataStore()
			defer dataStore.AssertExpectations(t)

			job := &model.Job{
				WorkflowName:    testCase.Workflow.Name,
				Status:          model.StatusPending,
				InputParameters: testCase.InputParameters,
			}
			dataStore.On("GetWorkflowByName",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				testCase.Workflow.Name,
				mocklib.AnythingOfType("string"),
			).Return(testCase.Workflow, nil)

			dataStore.On("UpsertJob",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
			).Return(job, nil)

			dataStore.On("UpdateJobStatus",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
				model.StatusDone,
			).Return(nil)

			if testCase.Error == nil {
				dataStore.On("UpdateJobAddResult",
					mocklib.MatchedBy(
						func(_ context.Context) bool {
							return true
						}),
					job,
					mocklib.MatchedBy(
						func(taskResult *model.TaskResult) bool {
							t.Log(taskResult.HTTPRequest)
							assert.True(t, taskResult.Success)
							assert.Equal(t,
								testCase.Workflow.Tasks[0].HTTP.URI,
								taskResult.HTTPRequest.URI,
							)
							assert.Equal(t,
								testCase.Workflow.Tasks[0].HTTP.Method,
								taskResult.HTTPRequest.Method,
							)
							assert.Equal(t, testCase.JobRequest.Headers, taskResult.HTTPRequest.Headers)
							assert.Equal(t, testCase.JobRequest.Body, taskResult.HTTPRequest.Body)
							assert.Equal(t, http.StatusOK, taskResult.HTTPResponse.StatusCode)
							assert.Equal(t, responseBody, taskResult.HTTPResponse.Body)

							return true
						}),
				).Return(nil)
			}

			makeHTTPRequestOriginal := makeHTTPRequest
			makeHTTPRequest = func(req *http.Request, timeout time.Duration) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
				}
				return resp, nil
			}
			jp := mockJobProcessor(job, dataStore, nil, nil)
			err := jp.ProcessJob(ctx)
			makeHTTPRequest = makeHTTPRequestOriginal

			switch e := testCase.Error.(type) {
			case bool:
				if e {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			case error:
				assert.EqualError(t, err, e.Error())
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessJobHTTPValidStatusCode(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeHTTP,
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: http.MethodGet,
					Headers: map[string]string{
						"X-Header": "Value",
					},
					StatusCodes: []int{
						http.StatusOK,
					},
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	requestBody := "BODY"

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	dataStore.On("UpsertJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusDone,
	).Return(nil)

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.True(t, taskResult.Success)
				assert.Equal(t, workflow.Tasks[0].HTTP.URI, taskResult.HTTPRequest.URI)
				assert.Equal(t, workflow.Tasks[0].HTTP.Method, taskResult.HTTPRequest.Method)
				assert.Equal(t, []string{
					"X-Header: Value",
				}, taskResult.HTTPRequest.Headers)
				assert.Equal(t, http.StatusOK, taskResult.HTTPResponse.StatusCode)
				assert.Equal(t, requestBody, taskResult.HTTPResponse.Body)

				return true
			}),
	).Return(nil)

	makeHTTPRequestOriginal := makeHTTPRequest
	makeHTTPRequest = func(req *http.Request, timeout time.Duration) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(requestBody)),
		}
		return resp, nil
	}
	jp := mockJobProcessor(job, dataStore, nil, nil)
	err := jp.ProcessJob(ctx)
	makeHTTPRequest = makeHTTPRequestOriginal

	assert.Nil(t, err)
}

func TestProcessJobHTTPWrongStatusCode(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeHTTP,
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: http.MethodGet,
					Headers: map[string]string{
						"X-Header": "Value",
					},
					StatusCodes: []int{
						http.StatusOK,
						http.StatusCreated,
					},
				},
			},
			{
				Name: "task_2",
				Type: model.TaskTypeHTTP,
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: http.MethodGet,
					Headers: map[string]string{
						"X-Header": "Value",
					},
					StatusCodes: []int{
						http.StatusOK,
						http.StatusCreated,
					},
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	requestBody := "BODY"

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	dataStore.On("UpsertJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusFailure,
	).Return(nil)

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.False(t, taskResult.Success)
				assert.Equal(t, workflow.Tasks[0].HTTP.URI, taskResult.HTTPRequest.URI)
				assert.Equal(t, workflow.Tasks[0].HTTP.Method, taskResult.HTTPRequest.Method)
				assert.Equal(t, []string{
					"X-Header: Value",
				}, taskResult.HTTPRequest.Headers)
				assert.Equal(t, http.StatusBadRequest, taskResult.HTTPResponse.StatusCode)
				assert.Equal(t, requestBody, taskResult.HTTPResponse.Body)

				return true
			}),
	).Return(nil)

	makeHTTPRequestOriginal := makeHTTPRequest
	makeHTTPRequest = func(req *http.Request, timeout time.Duration) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(strings.NewReader(requestBody)),
		}
		return resp, nil
	}
	jp := mockJobProcessor(job, dataStore, nil, nil)
	err := jp.ProcessJob(ctx)
	makeHTTPRequest = makeHTTPRequestOriginal

	assert.Nil(t, err)
}

func TestProcessJobHTTPFailedIncompatibleDefinition(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeHTTP,
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	dataStore.On("UpsertJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusFailure,
	).Return(nil)

	jp := mockJobProcessor(job, dataStore, nil, nil)
	err := jp.ProcessJob(ctx)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Error: Task definition incompatible with specified type (http)")
}
