// Copyright 2026 Northern.tech AS
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
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

type migration_1_1_1 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_1_1) Up(from migrate.Version) error {
	databaseName := mstore.DbFromContext(m.ctx, DbName)
	coll := m.ms.client.Database(databaseName).Collection(DbDevicesColl)
	indexView := coll.Indexes()

	cur, err := indexView.List(m.ctx)
	if err != nil {
		return err
	}

	var indices []bson.M
	err = cur.All(m.ctx, &indices)
	if err != nil {
		return err
	}

	for _, idx := range indices {
		if idx["name"] == DbDevIdentitiesName {
			return nil
		}
	}

	keys := bson.D{
		{
			Key:   DbDevIdentitiesName,
			Value: 1,
		},
	}
	_, err = indexView.CreateOne(
		m.ctx,
		mongo.IndexModel{
			Keys:    keys,
			Options: mopts.Index().SetName(DbDevIdentitiesName),
		},
	)

	return err
}

func (m *migration_1_1_1) Version() migrate.Version {
	return migrate.MakeVersion(1, 1, 1)
}
