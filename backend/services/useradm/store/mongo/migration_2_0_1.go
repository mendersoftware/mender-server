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

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

type migration_2_0_1 struct {
	ds     *DataStoreMongo
	dbName string
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func (m *migration_2_0_1) Up(from migrate.Version) error {
	ctx := context.Background()

	collectionsIndexes := map[string]struct {
		Indexes []mongo.IndexModel
	}{
		DbUserSettingsColl: {
			Indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: mstore.FieldTenantID, Value: 1},
						{Key: DbSettingsUserID, Value: 1},
					},
					Options: mopts.Index().
						SetUnique(true).
						SetName(DbSettingsTenantIndexName),
				},
			},
		},
	}

	// for each collection in main useradm database
	if m.dbName == DbName {
		for collection, indexModel := range collectionsIndexes {
			coll := m.ds.client.Database(m.dbName).Collection(collection)
			// drop all the existing indexes, ignoring the errors
			_, _ = coll.Indexes().DropAll(ctx)

			// create the new indexes
			if len(indexModel.Indexes) != 0 {
				_, err := coll.Indexes().CreateMany(ctx, indexModel.Indexes)
				if err != nil {
					return err
				}
			}
		}

		// migrate user settings to the dedicated collection
		coll := m.ds.client.Database(m.dbName).Collection(DbSettingsColl)
		opts := &mopts.FindOptions{}
		opts.SetSort(bson.D{{Key: "_id", Value: 1}})
		cur, err := coll.Find(ctx, bson.M{}, opts)
		if err != nil {
			return err
		}

		usersColl := m.ds.client.Database(m.dbName).Collection(DbUserSettingsColl)

		defer cur.Close(ctx)

		// migrate the documents
		for cur.Next(ctx) {
			item := map[string]interface{}{}
			err := cur.Decode(&item)
			if err != nil {
				return err
			}
			for key, value := range item {
				if isValidUUID(key) {
					valueMap, ok := value.(map[string]interface{})
					if ok {
						valueMap[DbSettingsUserID] = key
						valueMap[mstore.FieldTenantID] = item[mstore.FieldTenantID]
						_, err = usersColl.InsertOne(ctx, valueMap)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (m *migration_2_0_1) Version() migrate.Version {
	return migrate.MakeVersion(2, 0, 1)
}
