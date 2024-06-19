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

package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestValidateDeploymentLog(t *testing.T) {
	tref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05-07:00")
	assert.NoError(t, err)

	tcs := []struct {
		input DeploymentLog
		err   error
	}{
		{
			input: DeploymentLog{
				DeviceID:     "1234",
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
				// messages should be picked up
				Messages: []LogMessage{
					{
						Level:     "notice",
						Message:   "foo",
						Timestamp: &tref,
					},
				},
			},
		},
		{
			input: DeploymentLog{
				// device ID and messages are to be skipped when parsing/marshaling JSON
				DeviceID:     "1234",
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
				// messages should be picked up
				Messages: []LogMessage{
					{
						Level:     "",
						Message:   "foo",
						Timestamp: &tref,
					},
				},
			},
			err: errors.New("messages: (0: (level: cannot be blank.).)."),
		},
		{
			input: DeploymentLog{
				// device ID and messages are to be skipped when parsing/marshaling JSON
				DeviceID:     "1234",
				DeploymentID: "asdasdad1231",
				// messages should be picked up
				Messages: []LogMessage{
					{
						Level:     "notice",
						Message:   "foo",
						Timestamp: &tref,
					},
				},
			},
			err: errors.New("DeploymentID: must be a valid UUID."),
		},
	}

	for _, tc := range tcs {
		err := tc.input.Validate()
		if tc.err != nil {
			assert.Error(t, err)
			assert.EqualError(t, err, tc.err.Error())
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestUnmarshalDeploymentLog(t *testing.T) {

	t.Parallel()

	tref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05-07:00")
	assert.NoError(t, err)

	tcs := []struct {
		input    string
		err      error
		expected interface{}
	}{
		{
			input: `{ "messages": []}`,
			err:   errors.Wrapf(ErrInvalidDeploymentLog, "no messages"),
		},
		{
			input: `{ }`,
			err:   errors.Wrapf(ErrInvalidDeploymentLog, "no messages"),
		},
		{
			input: `{ "dev_id": "007",  "deployment_id": "001", "messages": [{
"timestamp": "2006-01-02T15:04:05-07:00", "level": "notice", "message": "foo"
}]}`,
			expected: &DeploymentLog{
				// device ID and messages are to be skipped when parsing/marshaling JSON
				DeviceID:     "",
				DeploymentID: "",
				// messages should be picked up
				Messages: []LogMessage{
					{
						Level:     "notice",
						Message:   "foo",
						Timestamp: &tref,
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Logf("testing: %v %v", tc.input, tc.err)
		var dl DeploymentLog
		err := json.Unmarshal([]byte(tc.input), &dl)

		if tc.err != nil {
			assert.Error(t, err)
			assert.EqualError(t, err, tc.err.Error())
		} else {
			assert.NoError(t, err)
			assert.EqualValues(t, tc.expected, &dl)
		}
	}
}

func TestUnmarshalLogMessage(t *testing.T) {

	t.Parallel()

	tref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05-07:00")
	assert.NoError(t, err)

	tcs := []struct {
		input    string
		err      error
		expected interface{}
	}{
		{
			input: `{ "message": "foo", "level": "notice"}`,
			err:   errors.New("timestamp: cannot be blank."),
		},
		{
			input: `{ "level": "notice", "timestamp": "2006-01-02T15:04:05-07:00"}`,
			err:   errors.New("message: cannot be blank."),
		},
		{
			input: `{ "message": "foo", "timestamp": "2006-01-02T15:04:05-07:00"}`,
			err:   errors.New("level: cannot be blank."),
		},
		{
			input: `{ "message": "foo", "level": "notice", "timestamp": "2006-01-02T15:04:05-07:00"}`,
			expected: &LogMessage{
				Level:     "notice",
				Message:   "foo",
				Timestamp: &tref,
			},
		},
	}

	for _, tc := range tcs {
		t.Logf("testing: %v %v", tc.input, tc.err)
		var lm LogMessage
		err := json.Unmarshal([]byte(tc.input), &lm)

		if tc.err != nil {
			assert.Error(t, err)
			assert.EqualError(t, err, tc.err.Error())
		} else {
			assert.NoError(t, err)
			assert.EqualValues(t, tc.expected, &lm)
		}
	}

}
