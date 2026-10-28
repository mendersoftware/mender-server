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

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func TestGetLastDeviceDeploymentStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SaveLastDeviceDeploymentStatus in short mode.")
	}

	deviceId1 := bson.NewObjectID().String()
	tenantId := uuid.New().String()
	now := time.Now()
	pastNow := now.Add(time.Hour)
	testCases := map[string]struct {
		deviceDeployments []model.DeviceDeployment
		tenantId          string
	}{
		"last status added": {
			deviceDeployments: []model.DeviceDeployment{
				{
					Created:      &now,
					Finished:     &now,
					Status:       model.DeviceDeploymentStatusAborted,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusNoArtifact,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusFailure,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
			},
			tenantId: tenantId,
		},
		"deployment successful status stored": {
			deviceDeployments: []model.DeviceDeployment{
				{
					Created:      &now,
					Finished:     &now,
					Status:       model.DeviceDeploymentStatusAborted,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusNoArtifact,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusSuccess,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
			},
			tenantId: tenantId,
		},
		"multiple failed deployments status stored": {
			deviceDeployments: []model.DeviceDeployment{
				{
					Created:      &now,
					Finished:     &now,
					Status:       model.DeviceDeploymentStatusAborted,
					DeviceId:     bson.NewObjectID().String(),
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusNoArtifact,
					DeviceId:     bson.NewObjectID().String(),
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusFailure,
					DeviceId:     bson.NewObjectID().String(),
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
			},
			tenantId: tenantId,
		},
		"empty tenant": {
			deviceDeployments: []model.DeviceDeployment{
				{
					Created:      &now,
					Finished:     &now,
					Status:       model.DeviceDeploymentStatusAborted,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusNoArtifact,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
				{
					Created:      &pastNow,
					Finished:     &pastNow,
					Status:       model.DeviceDeploymentStatusFailure,
					DeviceId:     deviceId1,
					DeploymentId: bson.NewObjectID().String(),
					Id:           bson.NewObjectID().String(),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			if tc.tenantId != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tenantId})
			}
			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)
			db.Wipe()
			ids := make([]string, len(tc.deviceDeployments))
			for i := range tc.deviceDeployments {
				ids[i] = tc.deviceDeployments[i].DeviceId
			}
			deployments, e := ds.GetLastDeviceDeploymentStatus(ctx, ids)
			assert.NoError(t, e)
			assert.Equal(t, len(deployments), 0)
			for i := range tc.deviceDeployments {
				err := ds.InsertDeviceDeployment(ctx, &tc.deviceDeployments[i], false)
				assert.NoError(t, err)
			}
			if tc.deviceDeployments[0].DeviceId != tc.deviceDeployments[1].DeviceId &&
				tc.deviceDeployments[0].DeviceId != tc.deviceDeployments[2].DeviceId &&
				tc.deviceDeployments[1].DeviceId != tc.deviceDeployments[2].DeviceId {
				for _, d := range tc.deviceDeployments {
					deployments, e = ds.GetLastDeviceDeploymentStatus(ctx, []string{d.DeviceId})
					assert.NoError(t, e)
					assert.Equal(t, len(deployments), 1)
					assert.Equal(t, deployments[0].DeviceId, d.DeviceId)
					assert.Equal(t, deployments[0].DeploymentId, d.DeploymentId)
				}
				deployments, e = ds.GetLastDeviceDeploymentStatus(ctx, ids)
				assert.NoError(t, e)
				for i := range deployments {
					found := false
					for j := range tc.deviceDeployments {
						if deployments[i].DeviceId == tc.deviceDeployments[j].DeviceId {
							found = true
							break
						}
					}
					assert.True(t, found)
				}
			} else {
				deployments, e = ds.GetLastDeviceDeploymentStatus(ctx, ids)
				assert.NoError(t, e)
				assert.Equal(t, len(deployments), 1)
				assert.Equal(t, deployments[0].DeviceId, tc.deviceDeployments[0].DeviceId)
			}
		})
	}
}
