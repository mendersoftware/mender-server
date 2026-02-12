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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	mongostore "github.com/mendersoftware/mender-server/pkg/mongo/v2"
)

type index struct {
	Keys        map[string]int `bson:"key"`
	Name        string         `bson:"name"`
	ExpireAfter *int           `bson:"expireAfterSeconds"`
}

func TestMigration_2_0_1(t *testing.T) {
	db := db.Client().Database(DbName)

	collDevs := db.Collection(DevicesCollectionName)
	ctx := context.Background()

	_, err := collDevs.InsertMany(ctx, func() []interface{} {
		ret := make([]interface{}, findBatchSize+1)
		for i := range ret {
			ret[i] = bson.D{}
		}
		return ret
	}())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = db.Client().
		Database(DbName+"-123456789012345678901234").
		Collection(DevicesCollectionName).
		InsertMany(ctx, func() []interface{} {
			ret := make([]interface{}, findBatchSize+1)
			for i := range ret {
				ret[i] = bson.D{}
			}
			return ret
		}())
	require.NoError(t, err)

	// NOTE: We're using 2.0.1 since it's the same migration, thus verifying
	//       that the migration is reentrant.
	err = Migrate(ctx, DbName, "2.0.1", db.Client(), true)
	require.NoError(t, err)

	collNames, err := db.ListCollectionNames(ctx, bson.D{})
	for _, collName := range collNames {
		idxes, err := db.Collection(collName).
			Indexes().
			ListSpecifications(ctx)
		require.NoError(t, err)
		expectedLen := 1
		switch collName {
		case RecordingsCollectionName:
			expectedLen = 3
		case SessionsCollectionName, ControlCollectionName:
			expectedLen = 2
		}
		require.Len(t, idxes, expectedLen)
		for _, idx := range idxes {
			if idx.Name == "_id_" {
				// Skip default index
				continue
			}
			switch idx.Name {
			case IndexNameLogsExpire:
				var keys bson.D
				assert.True(t, idx.ExpireAfterSeconds != nil, "Expected index to have TTL")
				err := bson.Unmarshal(idx.KeysDocument, &keys)
				require.NoError(t, err)
				assert.Equal(t, bson.D{
					{Key: "expire_ts", Value: int32(1)},
				}, keys)

			case mongostore.FieldTenantID + "_" + dbFieldSessionID:
				var keys bson.D
				err := bson.Unmarshal(idx.KeysDocument, &keys)
				require.NoError(t, err)
				assert.Equal(t, bson.D{
					{Key: mongostore.FieldTenantID, Value: int32(1)},
					{Key: dbFieldSessionID, Value: int32(1)},
				}, keys)

			case mongostore.FieldTenantID + "_" + dbFieldDeviceID:
				var keys bson.D
				err := bson.Unmarshal(idx.KeysDocument, &keys)
				require.NoError(t, err)
				assert.Equal(t, bson.D{
					{Key: mongostore.FieldTenantID, Value: int32(1)},
					{Key: dbFieldDeviceID, Value: int32(1)},
				}, keys)

			default:
				assert.Failf(t, "index test failed", "Index name \"%s\" not recognized", idx.Name)
			}
		}
	}

	actual, err := collDevs.CountDocuments(ctx, bson.D{
		{Key: mongostore.FieldTenantID, Value: bson.D{{Key: "$exists", Value: false}}},
	})
	require.NoError(t, err)
	assert.Zerof(t, actual, "%d documents are not indexed by tenant_id", actual)
}
