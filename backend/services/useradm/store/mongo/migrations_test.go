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
	"testing"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/stretchr/testify/assert"
)

func TestMigration_2_0_3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestMigration_2_0_3 in short mode")
	}

	client := db.Client()
	ds, err := NewDataStoreMongoWithClient(client)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	migrator := migrate.SimpleMigrator{
		Client:      client,
		Db:          DbName,
		Automigrate: true,
	}

	migration := []migrate.Migration{
		&migration_2_0_3{
			ds:     ds,
			ctx:    ctx,
			dbName: DbName,
		},
	}
	err = migrator.Apply(ctx, migrate.MakeVersion(2, 0, 3), migration)
	if assert.NoError(t, err, "migration is broken :'(") {
		iw := client.Database(DbName).
			Collection(DbTokensColl).
			Indexes()
		specs, err := iw.ListSpecifications(ctx)
		if err != nil {
			panic(err)
		}

		var found bool
		for _, spec := range specs {
			if spec.Name == IndexNameTokenExpire {
				found = true
				if assert.NotNil(t, spec.ExpireAfterSeconds, "not a TTL index") {
					assert.Equal(t, int32(0), *spec.ExpireAfterSeconds)
					_, err := spec.KeysDocument.LookupErr(DbTokenExpireTime)
					assert.NoError(t, err, "index did not contain expected key")
				}
			}
		}
		assert.True(t, found, "could not find index")
	}
}
