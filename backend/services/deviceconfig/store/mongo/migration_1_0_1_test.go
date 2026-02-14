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

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store/v2"
)

type index struct {
	Keys        map[string]int `bson:"key"`
	Name        string         `bson:"name"`
	ExpireAfter *int           `bson:"expireAfterSeconds"`
}

func TestMigration_1_0_1(t *testing.T) {

	collDevs := client.Database(DbName).
		Collection(CollDevices)
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

	m := &migration_1_0_1{
		client: client,
		db:     DbName,
	}
	from := migrate.MakeVersion(0, 0, 0)

	err = m.Up(from)
	require.NoError(t, err)

	iv := client.Database(DbName).
		Collection(CollDevices).
		Indexes()
	cur, err := iv.List(ctx)
	require.NoError(t, err)

	var idxes []index
	err = cur.All(ctx, &idxes)
	require.NoError(t, err)
	require.Len(t, idxes, 2)
	for _, idx := range idxes {
		if _, ok := idx.Keys["_id"]; ok && len(idx.Keys) == 1 {
			// Skip default index
			continue
		}
		switch idx.Name {
		case mstore.FieldTenantID + "_" + fieldID:
			assert.Equal(t, map[string]int{
				fieldID:     1,
				KeyTenantID: 1,
			}, idx.Keys)

		default:
			assert.Failf(t, "index test failed", "Index name \"%s\" not recognized", idx.Name)
		}
	}
	assert.Equal(t, "1.0.1", m.Version().String())
	actual, err := collDevs.CountDocuments(ctx, bson.D{
		{Key: mstore.FieldTenantID, Value: bson.D{{Key: "$exists", Value: false}}},
	})
	require.NoError(t, err)
	assert.Zerof(t, actual, "%d documents are not indexed by tenant_id", actual)
}
