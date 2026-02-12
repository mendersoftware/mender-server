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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

const (
	findBatchSize = 100
)

type migration_2_0_0 struct {
	ds     *DataStoreMongo
	dbName string
	ctx    context.Context
}

func (m *migration_2_0_0) Up(from migrate.Version) error {
	logger := log.FromContext(m.ctx)
	ctx := context.Background()

	collectionsIndexes := map[string]struct {
		Indexes []mongo.IndexModel
	}{
		DbUsersColl: {
			Indexes: []mongo.IndexModel{
				{
					Keys: bson.D{{Key: DbUserEmail, Value: 1}},
					Options: mopts.Index().
						SetUnique(true).
						SetName(DbUniqueEmailIndexName),
				},
			},
		},
		DbTokensColl: {
			Indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: mstore.FieldTenantID, Value: 1},
						{Key: DbTokenSubject, Value: 1},
						{Key: DbTokenName, Value: 1},
					},
					Options: mopts.Index().
						SetUnique(true).
						SetPartialFilterExpression(
							bson.M{
								DbTokenName: bson.M{"$exists": true},
							}).
						SetName(DbTenantUniqueTokenNameIndexName),
				},
				{
					Keys: bson.D{
						{Key: mstore.FieldTenantID, Value: 1},
						{Key: DbTokenSubject, Value: 1},
					},
					Options: mopts.Index().
						SetName(DbTenantTokenSubjectIndexName),
				},
			},
		},
		DbSettingsColl: {
			Indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: mstore.FieldTenantID, Value: 1},
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
	}

	tenantID := mstore.TenantFromDbName(m.dbName, DbName)
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantID,
	})

	// for each collection
	for collection := range collectionsIndexes {
		coll := m.ds.client.Database(m.dbName).Collection(collection)
		collOut := m.ds.client.Database(DbName).Collection(collection)
		writes := make([]mongo.WriteModel, 0, findBatchSize)

		if m.dbName == DbName {
			// if any documents already exist in "useradm" ds,
			// add empty "tenant_id": "" key-value pair
			tenantIdFilter := bson.D{
				{Key: mstore.FieldTenantID, Value: bson.D{{Key: "$exists", Value: false}}}}
			update := bson.M{"$set": bson.M{mstore.FieldTenantID: ""}}
			result, err := collOut.UpdateMany(ctx, tenantIdFilter, update)
			logger.Debugf("Modified documents in main useradm database count: %d",
				result.ModifiedCount)
			if err != nil {
				return err
			}
		} else {
			// get all the documents in the collection
			findOptions := mopts.Find().
				SetBatchSize(findBatchSize).
				SetSort(bson.D{{Key: "_id", Value: 1}})
			cur, err := coll.Find(ctx, bson.D{}, findOptions)
			if err != nil {
				return err
			}
			defer cur.Close(ctx)

			// migrate the documents
			for cur.Next(ctx) {
				item := bson.D{}
				err := cur.Decode(&item)
				if err != nil {
					return err
				}

				item = mstore.WithTenantID(ctx, item)
				writes = append(writes, mongo.NewInsertOneModel().SetDocument(item))

				if len(writes) == findBatchSize {
					_, err := collOut.BulkWrite(ctx, writes)
					if err != nil {
						return err
					}
					writes = writes[:0]
				}
			}
			if len(writes) > 0 {
				_, err := collOut.BulkWrite(ctx, writes)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *migration_2_0_0) Version() migrate.Version {
	return migrate.MakeVersion(2, 0, 0)
}
