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

	"github.com/mendersoftware/mender-server/services/deployments/model"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

type migration_1_2_10 struct {
	client *mongo.Client
	db     string
}

const IndexDeploymentsActiveCreatedV2 = "active_created:v2"

func (m *migration_1_2_10) Up(from migrate.Version) (err error) {

	ctx := context.Background()
	// MEN-5695: This index from the parent collection is not reliably
	//           created in the parent collection because of the
	//           renameCollection operation (removed).
	db := m.client.Database(m.db)
	_, err = db.Collection(CollectionDevices).
		Indexes().
		CreateOne(ctx, IndexDeviceDeploymentsActiveCreatedModel)
	if err != nil {
		return errors.Wrapf(err, "failed to create index '%s'",
			IndexDeviceDeploymentsActiveCreated)
	}

	// Ensure that we can index the 'active' field.
	// NOTE: Requires that no client prior to 1.2.8 writes to the
	//       database.
	collDpl := db.Collection(CollectionDeployments)
	_, err = collDpl.UpdateMany(ctx, bson.D{{
		Key: StorageKeyDeploymentStatus,
		Value: bson.D{{
			Key: "$ne", Value: model.DeploymentStatusFinished,
		}},
	}}, bson.D{{
		Key: "$set", Value: bson.D{{
			Key:   StorageKeyDeploymentActive,
			Value: true,
		}},
	}})
	if err != nil {
		return errors.Wrap(err, "failed to update database schema")
	}

	// MEN-5695: The index from migration 1.2.8 (removed) is not
	//           utilized for the majority of queries targetted at
	//           it because of the missing key on the "active"
	//           field. We make it sparse since we don't care about
	//           old 'inactive' deployments.
	idxDpl := collDpl.Indexes()
	_, err = idxDpl.CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeploymentActive, Value: 1},
			{Key: StorageKeyDeploymentCreated, Value: 1},
		},
		Options: mopts.Index().
			SetName(IndexDeploymentsActiveCreatedV2).
			SetSparse(true),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create index '%s'",
			IndexDeploymentsActiveCreatedV2)
	}
	// MEN-5695: This index is no longer needed.
	err = idxDpl.DropOne(ctx, IndexDeploymentsActiveCreated)
	var srvErr mongo.ServerError
	if errors.As(err, &srvErr) {
		if srvErr.HasErrorCode(errorCodeIndexNotFound) {
			err = nil
		}
	}

	return err
}

func (m *migration_1_2_10) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 10)
}
