// Copyright 2023 Northern.tech AS
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

package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	inventory_mocks "github.com/mendersoftware/mender-server/services/deployments/client/inventory/mocks"
	workflows_mocks "github.com/mendersoftware/mender-server/services/deployments/client/workflows/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	fs_mocks "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store/mocks"
)

func TestGetDeviceDeploymentLastStatus(t *testing.T) {
	t.Parallel()

	deviceId := uuid.New().String()
	lotsOfDevicesIds := make([]string, MaxDeviceArrayLength+1)
	for i := 0; i < MaxDeviceArrayLength+1; i++ {
		lotsOfDevicesIds[i] = uuid.New().String()
	}
	testCases := []struct {
		Name string

		DeviceIds     []string
		Statuses      []model.DeviceDeploymentLastStatus
		DbError       error
		ExpectedError error
	}{
		{
			Name: "ok",
			Statuses: []model.DeviceDeploymentLastStatus{
				{
					DeviceId:               deviceId,
					DeploymentId:           uuid.New().String(),
					DeviceDeploymentId:     uuid.New().String(),
					DeviceDeploymentStatus: model.DeviceDeploymentStatusNoArtifact,
					TenantId:               uuid.New().String(),
				},
			},
			DeviceIds: []string{deviceId},
		},
		{
			Name: "error no device ids",
			Statuses: []model.DeviceDeploymentLastStatus{
				{
					DeviceId:               uuid.New().String(),
					DeploymentId:           uuid.New().String(),
					DeviceDeploymentId:     uuid.New().String(),
					DeviceDeploymentStatus: model.DeviceDeploymentStatusNoArtifact,
					TenantId:               uuid.New().String(),
				},
			},
			ExpectedError: ErrNoIdsGiven,
		},
		{
			Name: "error no device ids",
			Statuses: []model.DeviceDeploymentLastStatus{
				{
					DeviceId:               uuid.New().String(),
					DeploymentId:           uuid.New().String(),
					DeviceDeploymentId:     uuid.New().String(),
					DeviceDeploymentStatus: model.DeviceDeploymentStatusNoArtifact,
					TenantId:               uuid.New().String(),
				},
			},
			DeviceIds:     lotsOfDevicesIds,
			ExpectedError: ErrArrayTooBig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.TODO()
			mDStore := &mocks.DataStore{}
			mFStore := &fs_mocks.ObjectStorage{}
			mWorkflows := &workflows_mocks.Client{}
			mInventory := &inventory_mocks.Client{}
			dep := &Deployments{
				db:              mDStore,
				objectStorage:   mFStore,
				workflowsClient: mWorkflows,
				inventoryClient: mInventory,
			}
			mDStore.On(
				"GetLastDeviceDeploymentStatus",
				ctx,
				mock.AnythingOfType("[]string"),
			).
				Return(
					tc.Statuses,
					tc.DbError,
				)
			statuses, err := dep.GetDeviceDeploymentLastStatus(ctx, tc.DeviceIds)
			if tc.ExpectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.Statuses), len(statuses.DeviceDeploymentLastStatuses))
			} else {
				assert.Error(t, err)
				assert.EqualError(t, tc.ExpectedError, err.Error())
			}
		})
	}
}
