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
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

type migration_1_2_5 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_2_5) Up(from migrate.Version) error {
	ctx := context.Background()
	storage := NewDataStoreMongoWithClient(m.client)
	coll := m.client.Database(m.db).Collection(CollectionDevices)
	// Drop IndexDeploymentDeviceStatusesName and
	// IndexDeploymentDeviceDeploymentIdName if they exist.
	_, err := coll.Indexes().DropOne(ctx, IndexDeploymentDeviceStatusesName)
	if err != nil {
		if except, ok := err.(mongo.CommandError); ok {
			if except.Code != errorCodeNamespaceNotFound &&
				except.Code != errorCodeIndexNotFound {
				return err
			}
			// continue
		} else {
			return err
		}
	}
	_, err = coll.Indexes().DropOne(ctx, IndexDeploymentDeviceDeploymentIdName)
	if err != nil {
		if except, ok := err.(mongo.CommandError); ok {
			if except.Code != errorCodeNamespaceNotFound &&
				except.Code != errorCodeIndexNotFound {
				return err
			}
			// continue
		} else {
			return err
		}
	}
	return storage.EnsureIndexes(m.db, CollectionDevices,
		DeviceIDCreatedStatusIndex,
		DeploymentIdIndexes,
	)
}

func (m *migration_1_2_5) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 5)
}
