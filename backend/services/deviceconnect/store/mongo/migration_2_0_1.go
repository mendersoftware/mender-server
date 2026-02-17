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
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mongostore "github.com/mendersoftware/mender-server/pkg/mongo"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstorev1 "github.com/mendersoftware/mender-server/pkg/store"
)

const (
	findBatchSize = 255
)

type migration_2_0_1 struct {
	client *mongo.Client
	db     string
}

func (m *migration_2_0_1) Up(from migrate.Version) error {
	ctx := context.Background()
	client := m.client

	collections := map[string]struct {
		Indexes []mongo.IndexModel
	}{
		DevicesCollectionName: {},
		SessionsCollectionName: {
			Indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: mongostore.FieldTenantID, Value: 1},
						{Key: dbFieldDeviceID, Value: 1},
					},
					Options: mopts.Index().
						SetName(mongostore.FieldTenantID + "_" + dbFieldDeviceID),
				},
			},
		},
		RecordingsCollectionName: {
			Indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: mongostore.FieldTenantID, Value: 1},
						{Key: dbFieldSessionID, Value: 1},
					},
					Options: mopts.Index().
						SetName(mongostore.FieldTenantID + "_" + dbFieldSessionID),
				},
				{
					// Index for expiring old events
					Keys: bson.D{{Key: "expire_ts", Value: 1}},
					Options: mopts.Index().
						SetExpireAfterSeconds(0).
						SetName(IndexNameLogsExpire),
				},
			},
		},
		ControlCollectionName: {
			Indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: mongostore.FieldTenantID, Value: 1},
						{Key: dbFieldSessionID, Value: 1},
					},
					Options: mopts.Index().
						SetName(mongostore.FieldTenantID + "_" + dbFieldSessionID),
				},
			},
		},
	}

	tenantID := mstorev1.TenantFromDbName(m.db, DbName)
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantID,
	})
	writes := make([]mongo.WriteModel, 0, findBatchSize)

	for collection, idxes := range collections {
		writes = writes[:0]
		findOptions := mopts.Find().
			SetBatchSize(findBatchSize).
			SetSort(bson.D{{Key: dbFieldID, Value: 1}})
		collOut := client.Database(DbName).Collection(collection)
		if m.db == DbName {
			if len(idxes.Indexes) > 0 {
				_, err := collOut.Indexes().CreateMany(ctx, collections[collection].Indexes)
				if err != nil {
					return err
				}
			}
			_, err := collOut.UpdateMany(ctx, bson.D{
				{Key: mongostore.FieldTenantID, Value: bson.D{
					{Key: "$exists", Value: false},
				}},
			}, bson.D{{Key: "$set", Value: bson.D{
				{Key: mongostore.FieldTenantID, Value: ""},
			}}},
			)
			if err != nil {
				return err
			}
			continue
		}

		coll := client.Database(m.db).Collection(collection)
		// get all the documents in the collection
		cur, err := coll.Find(ctx, bson.D{}, findOptions)
		if err != nil {
			return err
		}
		defer cur.Close(ctx)

		// migrate the documents
		for cur.Next(ctx) {
			id := cur.Current.Lookup(dbFieldID)
			var item bson.D
			if err = cur.Decode(&item); err != nil {
				return err
			}
			writes = append(writes, mongo.
				NewReplaceOneModel().
				SetFilter(bson.D{{Key: dbFieldID, Value: id}}).
				SetUpsert(true).
				SetReplacement(mongostore.WithTenantID(ctx, item)))
			if len(writes) == findBatchSize {
				_, err = collOut.BulkWrite(ctx, writes)
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

	return nil
}

func (m *migration_2_0_1) Version() migrate.Version {
	return migrate.MakeVersion(2, 0, 1)
}
