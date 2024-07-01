// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/services/deployments/client/workflows"
	workflows_mocks "github.com/mendersoftware/mender-server/services/deployments/client/workflows/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	h "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
)

func TestPropagateReporting(t *testing.T) {
	var active *bool
	cases := map[string]struct {
		workflowsMock *workflows_mocks.Client
		storeMock     *mocks.DataStore

		cmdTenant string
		cmdDryRun bool
	}{
		"ok, default db, no tenant": {
			storeMock: func() *mocks.DataStore {
				ds := new(mocks.DataStore)

				ds.On("GetTenantDbs").
					Return([]string{""}, nil)
				ds.On("GetDeviceDeployments",
					h.ContextMatcher(),
					0,
					deviceDeploymentsBatchSize,
					"",
					active,
					true,
				).Return(
					[]model.DeviceDeployment{
						{
							Id:           "foo",
							DeviceId:     "bar",
							DeploymentId: "baz",
						},
						{
							Id:           "foo1",
							DeviceId:     "bar1",
							DeploymentId: "baz1",
						},
					},
					nil,
				)

				return ds
			}(),
			workflowsMock: func() *workflows_mocks.Client {
				wf := new(workflows_mocks.Client)
				wf.On(
					"StartReindexReportingDeploymentBatch",
					h.ContextMatcher(),
					[]workflows.DeviceDeploymentShortInfo{
						{
							ID:           "foo",
							DeviceID:     "bar",
							DeploymentID: "baz",
						},
						{
							ID:           "foo1",
							DeviceID:     "bar1",
							DeploymentID: "baz1",
						},
					},
				).Return(nil)
				return wf
			}(),
		},
		"ok, default db, dry-run": {
			cmdDryRun: true,
			storeMock: func() *mocks.DataStore {
				ds := new(mocks.DataStore)

				ds.On("GetTenantDbs").
					Return([]string{""}, nil)
				ds.On("GetDeviceDeployments",
					h.ContextMatcher(),
					0,
					deviceDeploymentsBatchSize,
					"",
					active,
					true,
				).Return(
					[]model.DeviceDeployment{
						{
							Id:           "foo",
							DeviceId:     "bar",
							DeploymentId: "baz",
						},
						{
							Id:           "foo1",
							DeviceId:     "bar1",
							DeploymentId: "baz1",
						},
					},
					nil,
				)

				return ds
			}(),
			workflowsMock: func() *workflows_mocks.Client {
				wf := new(workflows_mocks.Client)
				return wf
			}(),
		},
	}

	for k := range cases {
		tc := cases[k]
		t.Run(fmt.Sprintf("tc %s", k), func(t *testing.T) {
			defer tc.workflowsMock.AssertExpectations(t)
			defer tc.storeMock.AssertExpectations(t)
			err := propagateReporting(tc.storeMock, tc.workflowsMock, tc.cmdTenant, time.Microsecond, tc.cmdDryRun)
			assert.NoError(t, err)
		})
	}
}
