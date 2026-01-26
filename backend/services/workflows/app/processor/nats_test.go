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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"

	mock_nats "github.com/mendersoftware/mender-server/services/workflows/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store/mock"
)

func TestProcessJobNATS(t *testing.T) {
	testCases := map[string]struct {
		err error
	}{
		"ok": {},
		"ko": {
			err: errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			dataStore := mock.NewDataStore()
			defer dataStore.AssertExpectations(t)

			nats := &mock_nats.Client{}
			defer nats.AssertExpectations(t)

			workflow := &model.Workflow{
				Name: "test",
				Tasks: []model.Task{
					{
						Name: "task_1",
						Type: model.TaskTypeNATS,
						NATS: &model.NATSTask{
							Subject: "test",
							Data: map[string]interface{}{
								"key": "${workflow.input.key}",
							},
						},
					},
				},
			}

			nats.On("StreamName").Return("STREAM")

			nats.On("JetStreamPublish",
				"STREAM."+workflow.Tasks[0].NATS.Subject,
				[]byte(`{"key":"value"}`),
			).Return(tc.err)

			job := &model.Job{
				WorkflowName: workflow.Name,
				Status:       model.StatusPending,
				InputParameters: model.InputParameters{
					{
						Name:  "key",
						Value: "value",
						Raw:   "value",
					},
				},
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

			status := model.StatusDone
			if tc.err != nil {
				status = model.StatusFailure
			}
			dataStore.On("UpdateJobStatus",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
				status,
			).Return(nil)

			dataStore.On("UpdateJobAddResult",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
				mocklib.MatchedBy(
					func(taskResult *model.TaskResult) bool {
						if tc.err == nil {
							assert.True(t, taskResult.Success)
							assert.Equal(t, "", taskResult.NATS.Error)
						} else {
							assert.False(t, taskResult.Success)
							assert.Equal(t, tc.err.Error(), taskResult.NATS.Error)
						}
						return true
					}),
			).Return(nil)

			jp := mockJobProcessor(job, dataStore, nats, nil)
			err := jp.ProcessJob(ctx)
			assert.Nil(t, err)
		})
	}
}
