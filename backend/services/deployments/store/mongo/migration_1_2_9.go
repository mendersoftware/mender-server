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

package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type migration_1_2_9 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_2_9) Up(from migrate.Version) (err error) {
	ctx := context.Background()
	storage := NewDataStoreMongoWithClient(m.client)
	database := storage.client.Database(m.db)

	collDev := database.Collection(CollectionDevices)

	bulkReqs := []mongo.WriteModel{
		mongo.NewUpdateManyModel().
			SetFilter(bson.D{{
				Key: StorageKeyDeviceDeploymentStatus, Value: bson.D{{
					Key: "$lt", Value: model.DeviceDeploymentStatusActiveLow,
				}},
			}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{{
				Key: StorageKeyDeviceDeploymentActive, Value: false,
			}}}}),
		mongo.NewUpdateManyModel().
			SetFilter(bson.D{{
				Key: StorageKeyDeviceDeploymentStatus, Value: bson.D{
					{Key: "$gte", Value: model.DeviceDeploymentStatusActiveLow},
					{Key: "$lte", Value: model.DeviceDeploymentStatusActiveHigh},
				},
			}}).SetUpdate(bson.D{{Key: "$set", Value: bson.D{{
			Key: StorageKeyDeviceDeploymentActive, Value: true,
		}}}}),
		mongo.NewUpdateManyModel().
			SetFilter(bson.D{{
				Key: StorageKeyDeviceDeploymentStatus, Value: bson.D{{
					Key: "$gt", Value: model.DeviceDeploymentStatusActiveHigh,
				}},
			}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{{
				Key: StorageKeyDeviceDeploymentActive, Value: false,
			}}}}),
	}
	_, err = collDev.BulkWrite(ctx, bulkReqs)
	if err == nil {
		err = storage.EnsureIndexes(m.db, CollectionDevices,
			IndexDeviceDeploymentsActiveCreatedModel,
		)
	}
	return err
}

func (m *migration_1_2_9) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 9)
}
