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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusToString(t *testing.T) {
	assert.Equal(t, "pending", StatusToString(StatusPending))
	assert.Equal(t, "processing", StatusToString(StatusProcessing))
	assert.Equal(t, "done", StatusToString(StatusDone))
	assert.Equal(t, "pending", StatusToString(StatusPending))
	assert.Equal(t, "failed", StatusToString(StatusFailure))
	assert.Equal(t, "unknown", StatusToString(999999))
}

func TestValidateWithoutErrors(t *testing.T) {
	workflow := &Workflow{
		Name: "test",
		InputParameters: []string{
			"key",
		},
	}
	job := Job{
		InputParameters: []InputParameter{
			{
				Name:  "key",
				Value: "test",
			},
		},
	}
	err := job.Validate(workflow)
	assert.Nil(t, err)
}
func TestValidateWithErrors(t *testing.T) {
	workflow := &Workflow{
		Name: "test",
		InputParameters: []string{
			"key",
			"missing_key",
		},
	}
	job := Job{
		InputParameters: []InputParameter{
			{
				Name:  "key",
				Value: "test",
			},
		},
	}
	err := job.Validate(workflow)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Missing input parameters: [missing_key]")
}

func TestPrepareForJSONMarshalling(t *testing.T) {
	job := Job{
		Status: StatusPending,
		InputParameters: []InputParameter{
			{
				Name:  "key",
				Value: "test",
				Raw:   "test",
			},
		},
	}
	job.PrepareForJSONMarshalling()
	assert.Equal(t, "pending", job.StatusString)
	assert.Nil(t, job.InputParameters[0].Raw)
}
