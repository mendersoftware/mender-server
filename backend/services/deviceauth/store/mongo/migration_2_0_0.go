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
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mongostore "github.com/mendersoftware/mender-server/pkg/mongo/v2"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstorev1 "github.com/mendersoftware/mender-server/pkg/store"
)

const (
	findBatchSize = 255
)

type migration_2_0_0 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

var DbDevicesCollectionIndices = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: mongostore.FieldTenantID, Value: 1},
			{Key: dbFieldIDDataSha, Value: 1},
		},
		//nolint:staticcheck // SA1019
		Options: mopts.Index().
			SetName(strings.Join([]string{
				mongostore.FieldTenantID,
				dbFieldIDDataSha,
			}, "_")).
			SetUnique(true),
	},
	{
		Keys: bson.D{
			{Key: mongostore.FieldTenantID, Value: 1},
			{Key: dbFieldStatus, Value: 1},
			{Key: dbFieldID, Value: 1},
		},
		//nolint:staticcheck // SA1019
		Options: mopts.Index().
			SetName(strings.Join([]string{
				mongostore.FieldTenantID,
				dbFieldStatus,
				dbFieldID,
			}, "_")),
	},
}

var DbAuthSetsCollectionIndices = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: mongostore.FieldTenantID, Value: 1},
			{Key: dbFieldIDDataSha, Value: 1},
			{Key: dbFieldPubKey, Value: 1},
		},
		Options: mopts.Index().
			SetName(strings.Join([]string{
				mongostore.FieldTenantID,
				dbFieldIDDataSha,
				dbFieldPubKey,
			}, "_")).
			SetUnique(true),
	},
	{
		Keys: bson.D{
			{Key: mongostore.FieldTenantID, Value: 1},
			{Key: dbFieldDeviceID, Value: 1},
		},
		Options: mopts.Index().
			SetName(strings.Join([]string{
				mongostore.FieldTenantID,
				dbFieldDeviceID,
			}, "_")),
	},
}

var DbLimitsCollectionIndices = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: mongostore.FieldTenantID, Value: 1},
			{Key: dbFieldName, Value: 1},
		},
		Options: mopts.Index().
			SetName(strings.Join([]string{
				mongostore.FieldTenantID,
				dbFieldName,
			},
				"_",
			),
			).
			SetUnique(true),
	},
}

var DbTokensCollectionIndices = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: dbFieldTenantClaim, Value: 1},
			{Key: dbFieldSubject, Value: 1},
		},
		Options: mopts.Index().
			SetName(
				strings.Join(
					[]string{
						strings.ReplaceAll(dbFieldTenantClaim, ".", "_"),
						dbFieldSubject,
					},
					"_",
				),
			),
	},
	{
		Keys: bson.D{
			{Key: dbFieldExpTime, Value: 1},
		},
		Options: mopts.Index().
			SetName(
				strings.Join(
					[]string{
						strings.ReplaceAll(dbFieldExpTime, ".", "_"),
					},
					"_",
				),
			).
			SetExpireAfterSeconds(0),
	},
}

// Up creates an index on status and id in the devices collection
// nolint: gocyclo
func (m *migration_2_0_0) Up(from migrate.Version) error {
	ctx := context.Background()
	client := m.ds.client

	collections := map[string]struct {
		Indexes []mongo.IndexModel
	}{
		DbAuthSetColl: {
			Indexes: DbAuthSetsCollectionIndices,
		},
		DbDevicesColl: {
			Indexes: DbDevicesCollectionIndices,
		},
		DbLimitsColl: {
			Indexes: DbLimitsCollectionIndices,
		},
		DbTokensColl: {
			Indexes: DbTokensCollectionIndices,
		},
	}

	databaseName := mstorev1.DbFromContext(m.ctx, DbName)
	tenantID := mstorev1.TenantFromDbName(databaseName, DbName)
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
		if databaseName == DbName {
			indices := collOut.Indexes()
			_ = indices.DropAll(ctx)

			if len(idxes.Indexes) > 0 {
				_, err := collOut.Indexes().CreateMany(ctx, collections[collection].Indexes)
				if err != nil {
					return err
				}
			}
			if collection == DbTokensColl {
				continue
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

		if collection == DbTokensColl {
			continue
		}
		coll := client.Database(databaseName).Collection(collection)
		// get all the documents in the collection
		cur, err := coll.Find(ctx, bson.D{}, findOptions)
		if err != nil {
			return err
		}
		defer cur.Close(ctx)

		// migrate the documents
		if collection == DbLimitsColl {
			for cur.Next(ctx) {
				id, ok := cur.Current.Lookup(dbFieldID).StringValueOK()
				if !ok {
					return fmt.Errorf(
						"found suspicious limits document: %s",
						cur.Current.String())
				}
				var item bson.M
				if err = cur.Decode(&item); err != nil {
					return err
				}
				item[dbFieldName] = id
				item[dbFieldTenantID] = tenantID
				item[dbFieldID] = bson.NewObjectID()
				writes = append(writes, mongo.
					NewReplaceOneModel().
					SetFilter(bson.D{
						{Key: dbFieldTenantID, Value: tenantID},
						{Key: dbFieldName, Value: id},
					}).
					SetUpsert(true).
					SetReplacement(item))
				if len(writes) == findBatchSize {
					_, err = collOut.BulkWrite(ctx, writes)
					if err != nil {
						return err
					}
					writes = writes[:0]
				}
			}
		} else {
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

func (m *migration_2_0_0) Version() migrate.Version {
	return migrate.MakeVersion(2, 0, 0)
}
