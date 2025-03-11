// Copyright 2025 Northern.tech AS
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
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

func TestMigration_1_2_0(t *testing.T) {
	ctx := context.Background()
	client := db.Client()
	m := &migration_1_2_0{
		client: client,
		db:     DbName,
	}
	from := migrate.MakeVersion(0, 0, 0)

	err := m.Up(from)
	require.NoError(t, err)
	specs, err := client.Database(DbName).
		Collection(CollNameIntegrations).
		Indexes().
		ListSpecifications(ctx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	var foundIndex bool
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		if spec.Name == IndexNameOneIntegration {
			foundIndex = true
			assert.True(
				t,
				spec.Unique != nil && *spec.Unique,
				"index must allow only one integration per tenant")
			var keys bson.M
			_ = bson.Unmarshal(spec.KeysDocument, &keys)
			assert.Equal(t,
				keys,
				bson.M{KeyTenantID: int32(1)},
				"unexpected index keys")
			break
		}
	}
	assert.True(t, foundIndex, "Failed to find index created by migration 1.2.0")
}
