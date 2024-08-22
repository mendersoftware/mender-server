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

package store

import (
	"errors"
	"testing"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/stretchr/testify/assert"
)

func str2ptr(s string) *string {
	return &s
}

func TestListQueryValidate(t *testing.T) {
	testCases := map[string]struct {
		query *ListQuery
		err   error
	}{
		"limit": {
			query: &ListQuery{
				Limit: -1,
			},
			err: errors.New("limit: must be a positive integer"),
		},
		"deployment ID": {
			query: &ListQuery{
				Limit:        1,
				DeploymentID: "",
			},
			err: errors.New("deployment_id: cannot be blank"),
		},
		"status": {
			query: &ListQuery{
				Limit:        1,
				DeploymentID: "dummy",
				Status:       str2ptr("dummy"),
			},
			err: errors.New("status: must be a valid value"),
		},
		"status, pause": {
			query: &ListQuery{
				Limit:        1,
				DeploymentID: "dummy",
				Status:       str2ptr(model.DeviceDeploymentStatusPauseStr),
			},
		},
		"status, active": {
			query: &ListQuery{
				Limit:        1,
				DeploymentID: "dummy",
				Status:       str2ptr(model.DeviceDeploymentStatusActiveStr),
			},
		},
		"status, pending": {
			query: &ListQuery{
				Limit:        1,
				DeploymentID: "dummy",
				Status:       str2ptr(model.DeviceDeploymentStatusPendingStr),
			},
		},
		"status, finished": {
			query: &ListQuery{
				Limit:        1,
				DeploymentID: "dummy",
				Status:       str2ptr(model.DeviceDeploymentStatusFinishedStr),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.query.Validate()
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
