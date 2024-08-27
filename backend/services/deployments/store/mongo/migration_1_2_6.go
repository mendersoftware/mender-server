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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type migration_1_2_6 struct {
	client *mongo.Client
	db     string
}

// Up replaces all devicedeployment documents status field with an enumerated type
func (m *migration_1_2_6) Up(from migrate.Version) error {
	ctx := context.Background()
	coll := m.client.Database(m.db).
		Collection(CollectionDevices)

	for _, status := range []model.DeviceDeploymentStatus{
		model.DeviceDeploymentStatusFailure,
		model.DeviceDeploymentStatusPauseBeforeInstall,
		model.DeviceDeploymentStatusPauseBeforeCommit,
		model.DeviceDeploymentStatusPauseBeforeReboot,
		model.DeviceDeploymentStatusDownloading,
		model.DeviceDeploymentStatusInstalling,
		model.DeviceDeploymentStatusRebooting,
		model.DeviceDeploymentStatusPending,
		model.DeviceDeploymentStatusSuccess,
		model.DeviceDeploymentStatusAborted,
		model.DeviceDeploymentStatusNoArtifact,
		model.DeviceDeploymentStatusAlreadyInst,
		model.DeviceDeploymentStatusDecommissioned,
	} {
		oldStatus := status.String()
		_, err := coll.UpdateMany(ctx,
			bson.D{{Key: StorageKeyDeviceDeploymentStatus, Value: oldStatus}},
			bson.D{{Key: "$set", Value: bson.D{{
				Key: StorageKeyDeviceDeploymentStatus, Value: status,
			}}}},
		)
		if err != nil {
			return err
		}
	}
	// TODO: Undo DeploymentIdIndexes ?

	iw := coll.Indexes()
	_, err := iw.CreateOne(ctx, DeviceDeploymentIdStatus)
	return err
}

func (m *migration_1_2_6) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 6)
}
