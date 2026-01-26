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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/executor"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store/mock"
)

var (
	executorMock = executor.New([]string{
		"/usr/bin/echo",
		"/usr/bin/bash",
	})
)

func TestProcessJobCLI(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
				CLI: &model.CLITask{
					Command: []string{
						"/usr/bin/echo",
						"TEST",
					},
				},
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
				assert.Equal(t, workflow.Tasks[0].CLI.Command, taskResult.CLI.Command)
				assert.Equal(t, "TEST\n", taskResult.CLI.Output)
				assert.Equal(t, "", taskResult.CLI.Error)
				assert.Equal(t, 0, taskResult.CLI.ExitCode)

				return true
			}),
	).Return(nil)

	jp := mockJobProcessor(job, dataStore, nil, executorMock)
	err := jp.ProcessJob(ctx)

	assert.Nil(t, err)
}

func TestProcessJobCLIWrongExitCode(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
				CLI: &model.CLITask{
					Command: []string{
						"/usr/bin/bash",
						"-c",
						"exit 10",
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
				assert.Equal(t, workflow.Tasks[0].CLI.Command, taskResult.CLI.Command)
				assert.Equal(t, "", taskResult.CLI.Output)
				assert.Equal(t, "", taskResult.CLI.Error)
				assert.Equal(t, 10, taskResult.CLI.ExitCode)

				return true
			}),
	).Return(nil)

	jp := mockJobProcessor(job, dataStore, nil, executorMock)
	err := jp.ProcessJob(ctx)
	assert.Nil(t, err)
}

func TestProcessJobCLTimeOut(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
				CLI: &model.CLITask{
					Command: []string{
						"/usr/bin/bash",
						"-c",
						"sleep 10",
					},
					ExecutionTimeOut: 2,
				},
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

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.False(t, taskResult.Success)
				assert.Equal(t, workflow.Tasks[0].CLI.Command, taskResult.CLI.Command)
				assert.Equal(t, "", taskResult.CLI.Output)
				assert.Equal(t, "", taskResult.CLI.Error)
				assert.Equal(t, -1, taskResult.CLI.ExitCode)

				return true
			}),
	).Return(nil)

	jp := mockJobProcessor(job, dataStore, nil, executorMock)
	err := jp.ProcessJob(ctx)
	assert.Nil(t, err)
}

func TestProcessJobCLIFailedIncompatibleDefinition(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
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

	jp := mockJobProcessor(job, dataStore, nil, executorMock)
	err := jp.ProcessJob(ctx)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Error: Task definition incompatible with specified type (cli)")
}

func TestProcessJobCLINotAllowedCommand(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
				CLI: &model.CLITask{
					Command: []string{
						"/usr/bin/python3",
						"-c",
						`'print("Get Hacked!")'`,
					},
				},
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

	jp := mockJobProcessor(job, dataStore, nil, executorMock)
	err := jp.ProcessJob(ctx)

	assert.ErrorIs(t, err, executor.ErrCommandNotAllowed)
}
