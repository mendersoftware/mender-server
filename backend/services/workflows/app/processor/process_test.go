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
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mocklib "github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/executor"
	"github.com/mendersoftware/mender-server/pkg/log"

	natsmock "github.com/mendersoftware/mender-server/services/workflows/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store"
	storemock "github.com/mendersoftware/mender-server/services/workflows/store/mock"
)

func mockCommandExecutor(ctx context.Context, command []string) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, command[0], command[1:]...), nil

}

func mockJobProcessor(job *model.Job, store store.DataStore, client *natsmock.Client,
	exec executor.BinaryExecutor) *JobProcessor {
	return &JobProcessor{
		store:    store,
		client:   client,
		executor: exec,
		job:      job,
	}
}

func TestProcessJobFailedWorkflowDoesNotExist(t *testing.T) {
	ctx := context.Background()
	dataStore := storemock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	job := &model.Job{
		ID:           "job",
		WorkflowName: "does_not_exist",
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		ctx,
		job.WorkflowName,
		job.WorkflowVersion,
	).Return(nil, errors.New("workflow not found"))

	dataStore.On("UpdateJobStatus",
		ctx,
		job,
		model.StatusFailure,
	).Return(nil)
	jp := mockJobProcessor(job, dataStore, nil, nil)
	err := jp.ProcessJob(ctx)
	assert.Nil(t, err)
}

func TestProcessJobFailedUpsert(t *testing.T) {
	ctx := context.Background()
	dataStore := storemock.NewDataStore()
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
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusDone,
	}

	dataStore.On("GetWorkflowByName",
		ctx,
		job.WorkflowName,
		job.WorkflowVersion,
	).Return(workflow, nil)

	dataStore.On("UpsertJob",
		ctx,
		job,
	).Return(nil, errors.New("failed"))

	jp := mockJobProcessor(job, dataStore, nil, nil)
	err := jp.ProcessJob(ctx)
	assert.EqualError(t, err, "insert of the job failed: failed")
}

func TestProcessTaskSkipped(t *testing.T) {
	testCases := map[string]struct {
		workflow *model.Workflow
		job      *model.Job
		task     *model.Task
		skipped  bool
	}{
		"skipped, missing parameter": {
			workflow: &model.Workflow{
				Name: "test",
				InputParameters: []string{
					"request_id",
				},
			},
			job: &model.Job{
				InputParameters: []model.InputParameter{
					{
						Name:  "request_id",
						Value: "",
					},
				},
			},
			task: &model.Task{
				Name: "task_1",
				Type: model.TaskTypeHTTP,
				Requires: []string{
					"${workflow.input.request_id}",
				},
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: http.MethodGet,
					Headers: map[string]string{
						"X-Header": "Value",
					},
				},
			},
			skipped: true,
		},
		"executed, requires parameter": {
			workflow: &model.Workflow{
				Name: "test",
			},
			job: &model.Job{
				InputParameters: []model.InputParameter{
					{
						Name:  "request_id",
						Value: "value",
					},
				},
			},
			task: &model.Task{
				Name: "task_1",
				Type: model.TaskTypeHTTP,
				Requires: []string{
					"${workflow.input.request_id}",
				},
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: http.MethodGet,
					Headers: map[string]string{
						"X-Header": "Value",
					},
				},
			},
			skipped: false,
		},
		"executed, requires parameter but empty": {
			workflow: &model.Workflow{
				Name: "test",
			},
			job: &model.Job{
				InputParameters: []model.InputParameter{
					{
						Name:  "request_id",
						Value: "",
					},
				},
			},
			task: &model.Task{
				Name: "task_1",
				Type: model.TaskTypeHTTP,
				Requires: []string{
					"${workflow.input.request_id}",
				},
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: http.MethodGet,
					Headers: map[string]string{
						"X-Header": "Value",
					},
				},
			},
			skipped: true,
		},
		"executed, requires env": {
			workflow: &model.Workflow{
				Name: "test",
			},
			job: &model.Job{},
			task: &model.Task{
				Name: "task_1",
				Type: model.TaskTypeHTTP,
				Requires: []string{
					"${env.PWD}",
				},
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: http.MethodGet,
					Headers: map[string]string{
						"X-Header": "Value",
					},
				},
			},
			skipped: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			makeHTTPRequestOriginal := makeHTTPRequest
			makeHTTPRequest = func(req *http.Request, timeout time.Duration) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			}

			ctx := context.Background()
			l := log.FromContext(ctx)
			jp := mockJobProcessor(tc.job, nil, nil, nil)
			result, err := jp.processTask(*tc.task, tc.workflow, l)
			assert.NoError(t, err)
			assert.Equal(t, tc.skipped, result.Skipped)

			makeHTTPRequest = makeHTTPRequestOriginal
		})
	}
}

func TestProcessTaskRetries(t *testing.T) {
	testCases := map[string]struct {
		workflow *model.Workflow
		job      *model.Job
	}{
		"retries": {
			workflow: &model.Workflow{
				Name: "test",
				Tasks: []model.Task{
					{
						Name:    "task_1",
						Type:    model.TaskTypeHTTP,
						Retries: 1,
						HTTP: &model.HTTPTask{
							URI:    "http://localhost",
							Method: http.MethodGet,
						},
					},
				},
			},
			job: &model.Job{
				WorkflowName: "test",
			},
		},
		"retries with delay": {
			workflow: &model.Workflow{
				Name: "test",
				Tasks: []model.Task{
					{
						Name:              "task_1",
						Type:              model.TaskTypeHTTP,
						Retries:           1,
						RetryDelaySeconds: 1,
						HTTP: &model.HTTPTask{
							URI:    "http://localhost",
							Method: http.MethodGet,
						},
					},
				},
			},
			job: &model.Job{
				WorkflowName: "test",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			makeHTTPRequestOriginal := makeHTTPRequest
			firstCallHappened := false
			makeHTTPRequest = func(req *http.Request, timeout time.Duration) (*http.Response, error) {
				status := http.StatusOK
				if !firstCallHappened {
					firstCallHappened = true
					status = http.StatusBadGateway
				}
				return &http.Response{
					StatusCode: status,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			}

			ctx := context.Background()
			dataStore := storemock.NewDataStore()
			defer dataStore.AssertExpectations(t)

			dataStore.On("GetWorkflowByName",
				ctx,
				tc.job.WorkflowName,
				tc.job.WorkflowVersion,
			).Return(tc.workflow, nil)

			dataStore.On("UpsertJob",
				ctx,
				tc.job,
			).Return(tc.job, nil)

			dataStore.On("UpdateJobStatus",
				ctx,
				tc.job,
				model.StatusDone,
			).Return(nil)

			dataStore.On("UpdateJobAddResult",
				ctx,
				tc.job,
				mock.AnythingOfType("*model.TaskResult"),
			).Return(nil)
			jp := mockJobProcessor(tc.job, dataStore, nil, nil)
			err := jp.ProcessJob(ctx)
			assert.NoError(t, err)
			assert.Equal(t, true, firstCallHappened)

			makeHTTPRequest = makeHTTPRequestOriginal
		})
	}
}

func TestProcessJobUnrecognizedTaskType(t *testing.T) {
	ctx := context.Background()
	dataStore := storemock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: "dummy",
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
	assert.EqualError(t, err, "Unrecognized task type: dummy")
}
