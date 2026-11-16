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

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func parseTime(t *testing.T, value string) *time.Time {
	tm, err := time.Parse(time.RFC3339, value)
	if assert.NoError(t, err) == false {
		t.Fatalf("failed to parse time %s", value)
	}

	return &tm
}

func TestSaveDeviceDeploymentLog(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestSaveDeviceDeploymentLog in short mode.")
	}

	const (
		deviceID     = "b500c9e9-e298-4e53-ae9b-9d5188cce714"
		deploymentID = "0fb22abf-dc50-44b5-9a70-9b11de9fb395"
	)

	messages := []model.LogMessage{
		{
			Level:     "notice",
			Message:   "foo",
			Timestamp: parseTime(t, "2006-01-02T15:04:05-07:00"),
		},
		{
			Level:     "notice",
			Message:   "bar",
			Timestamp: parseTime(t, "2006-01-02T15:05:05-07:00"),
		},
	}
	testCases := []struct {
		InputDeviceDeploymentLog model.DeploymentLog
		InputTenant              string
		OutputError              error
	}{
		{
			InputDeviceDeploymentLog: model.DeploymentLog{
				Messages: []model.LogMessage{},
			},
			OutputError: errors.New("messages: cannot be blank."),
		},
		{
			InputDeviceDeploymentLog: model.DeploymentLog{
				Messages: messages,
			},
			OutputError: nil,
		},
		{
			InputDeviceDeploymentLog: model.DeploymentLog{
				Messages: messages,
			},
			// same as previous case but with tenant
			InputTenant: "acme",
			OutputError: nil,
		},
	}

	for _, testCase := range testCases {

		t.Logf("testing case %v %v",
			testCase.InputDeviceDeploymentLog, testCase.OutputError)

		// Make sure we start test with empty database
		db.Wipe()

		client := db.Client()
		store := NewDataStoreMongoWithClient(client)

		ctx := context.Background()
		if testCase.InputTenant != "" {
			ctx = identity.WithContext(ctx, &identity.Identity{
				Tenant: testCase.InputTenant,
			})
		} else {
			ctx = context.Background()
		}
		require.NoError(t, store.Migrate(ctx, DbVersion, true))
		require.NoError(t, store.InsertDeviceDeployment(ctx, &model.DeviceDeployment{
			DeviceId:     deviceID,
			DeploymentId: deploymentID,
		}, false))

		err := store.SaveDeviceDeploymentLog(ctx,
			deviceID,
			deploymentID,
			testCase.InputDeviceDeploymentLog)

		if testCase.OutputError != nil {
			assert.EqualError(t, err, testCase.OutputError.Error())
		} else {
			assert.NoError(t, err)

			// no errors, so we should be able to find the log in DB
			dlog, err := store.GetDeviceDeploymentLog(ctx, deviceID, deploymentID)

			assert.NoError(t, err)

			// message timestamp is a pointer, so we cannot use assert.EqualValues()
			// or reflect.DeepEqual() as both will choke on *time.Time pointing to
			// different, but value-equal instances, just compare if length is ok for now
			assert.Len(t, dlog.Messages, len(testCase.InputDeviceDeploymentLog.Messages))

			if testCase.InputTenant != "" {
				// logs were saved to tenant's DB, double check
				// that they are not found in default DB
				log, err := store.GetDeviceDeploymentLog(context.Background(), deviceID, deploymentID)
				assert.NoError(t, err)
				assert.Nil(t, log)
			}

		}
	}
}

func TestGetDeviceDeploymentLog(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestGetDeviceDeploymentLog in short mode.")
	}

	messages := []model.LogMessage{
		{
			Level:     "notice",
			Message:   "foo",
			Timestamp: parseTime(t, "2006-01-02T15:04:05-07:00"),
		},
		{
			Level:     "notice",
			Message:   "bar",
			Timestamp: parseTime(t, "2006-01-02T15:05:05-07:00"),
		},
	}

	logs := []model.DeviceDeployment{
		{
			Id:           "43c3c7aa-50c7-49e8-a32c-45f58dcfe19f",
			DeviceId:     "123",
			DeploymentId: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			Log: &model.DeploymentLog{
				Messages: messages,
			},
		},
		{
			Id:           "43c3c7aa-50c7-49e8-a32c-45f58dcfe190",
			DeviceId:     "234",
			DeploymentId: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
			Log: &model.DeploymentLog{
				Messages: messages,
			},
		},
		{
			Id:           "43c3c7aa-50c7-49e8-a32c-45f58dcfe191",
			DeviceId:     "345",
			DeploymentId: "30b3e62c-9ec2-4312-a7fa-cff24cc7397c",
			Log: &model.DeploymentLog{
				Messages: messages,
			},
		},
	}

	testCases := []struct {
		InputDeviceID      string
		InputDeploymentID  string
		InputDeploymentLog *model.DeploymentLog
		InputTenant        string
		OutputError        error
	}{
		{
			InputDeviceID:      "123",
			InputDeploymentID:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputDeploymentLog: logs[0].Log,
		},
		{
			InputDeviceID:      "234",
			InputDeploymentID:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
			InputDeploymentLog: logs[1].Log,
		},
		{
			InputDeviceID:      "345",
			InputDeploymentID:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397c",
			InputDeploymentLog: logs[2].Log,
		},
		// not found
		{
			InputDeviceID:      "456",
			InputDeploymentID:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397c",
			InputDeploymentLog: nil,
		},
		{
			// same as first case, but accessing tenant's DB
			InputDeviceID:      "123",
			InputDeploymentID:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputTenant:        "acme",
			InputDeploymentLog: nil,
		},
	}

	// Make sure we start test with empty database
	db.Wipe()

	client := db.Client()

	store := NewDataStoreMongoWithClient(client)

	for _, dl := range logs {
		// save all messages to default DB
		require.NoError(t, store.InsertDeviceDeployment(context.Background(), &dl, false))
	}

	for _, testCase := range testCases {

		t.Logf("testing case %v %v",
			testCase.InputDeploymentLog, testCase.OutputError)

		ctx := context.Background()
		if testCase.InputTenant != "" {
			ctx = identity.WithContext(ctx, &identity.Identity{
				Tenant: testCase.InputTenant,
			})
		} else {
			ctx = context.Background()
		}

		dlog, err := store.GetDeviceDeploymentLog(ctx,
			testCase.InputDeviceID, testCase.InputDeploymentID)
		if testCase.OutputError != nil {
			assert.EqualError(t, err, testCase.OutputError.Error())
		} else {
			assert.NoError(t, err)

			if testCase.InputDeploymentLog == nil {
				assert.Nil(t, dlog)
			} else {
				// message timestamp is a pointer, so we cannot use assert.EqualValues()
				// or reflect.DeepEqual() as both will choke on *time.Time pointing to
				// different, but value-equal instances
				assert.Len(t, dlog.Messages, len(testCase.InputDeploymentLog.Messages))
				for i, m := range testCase.InputDeploymentLog.Messages {
					assert.True(t, m.Timestamp.Equal(*dlog.Messages[i].Timestamp))
					assert.Equal(t, m.Level, dlog.Messages[i].Level)
					assert.Equal(t, m.Message, dlog.Messages[i].Message)
				}
			}
		}
	}
	db.Wipe()
}
