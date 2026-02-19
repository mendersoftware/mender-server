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

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

const (
	IndexNameIntegrationsGet = KeyTenantID + "_" + KeyProvider + "_" + KeyID
	IndexNameDevices         = KeyTenantID + "_" + KeyIntegrationIDs + "_" + KeyID
)

type migration_1_0_0 struct {
	client *mongo.Client
	db     string
}

// Up creates indexes for fetching device and integration documents
func (m *migration_1_0_0) Up(from migrate.Version) error {
	ctx := context.Background()
	itgModels := []mongo.IndexModel{{
		Keys: bson.D{
			// $match ($eq)
			{Key: KeyTenantID, Value: 1},
			// $sort (and/or $match)
			{Key: KeyProvider, Value: 1},
			// $sort
			{Key: KeyID, Value: 1},
		},
		Options: mopts.Index().
			SetName(IndexNameIntegrationsGet),
	}}
	devicesModels := []mongo.IndexModel{{
		Keys: bson.D{
			// $match ($eq)
			{Key: KeyTenantID, Value: 1},
			// $match ($eq)
			{Key: KeyIntegrationIDs, Value: 1},
			// $match or $sort
			{Key: KeyID, Value: 1},
		},
		Options: mopts.Index().
			SetName(IndexNameDevices),
	}}
	idxItg := m.client.
		Database(m.db).
		Collection(CollNameIntegrations).
		Indexes()

	_, err := idxItg.CreateMany(ctx, itgModels)
	if err != nil {
		return err
	}
	idxDevices := m.client.Database(m.db).
		Collection(CollNameDevices).
		Indexes()

	_, err = idxDevices.CreateMany(ctx, devicesModels)
	return err
}

func (m *migration_1_0_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 0, 0)
}
