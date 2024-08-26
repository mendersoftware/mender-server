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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
)

func TestMigration_1_3_2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestMigration_3_1_2 in short mode")
	}
	testCases := []struct {
		Name string

		Tenant string
	}{
		{
			Name: "Successful migration",
		},
		{
			Name:   "Successful migration, MT expire token",
			Tenant: primitive.NewObjectID().Hex(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db.Wipe()
			ctx := context.Background()
			if testCase.Tenant != "" {
				ctx = identity.WithContext(
					ctx,
					&identity.Identity{
						Tenant: testCase.Tenant,
					})
			}
			client := db.Client()
			ds, err := NewDataStoreMongoWithClient(client)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			dbName := ctxstore.DbFromContext(ctx, DbName)

			migrations := []migrate.Migration{
				&migration_1_3_2{
					ds:  ds,
					ctx: ctx,
				},
			}
			m := migrate.SimpleMigrator{
				Client:      client,
				Db:          dbName,
				Automigrate: true,
			}
			err = m.Apply(ctx, migrate.MakeVersion(1, 3, 2), migrations)
			assert.NoError(t, err)

			c := client.Database(dbName).Collection(DbTokensColl)
			tokensCollectionIndexes, err := c.Indexes().List(ctx)
			assert.NoError(t, err)
			count := 0
			var index bson.M
			for tokensCollectionIndexes.Next(ctx) {
				err := tokensCollectionIndexes.Decode(&index)
				assert.NoError(t, err)
				// do not count the default index on id
				if index["name"] == "_id_" {
					continue
				}
				assert.Contains(
					t,
					[]string{DbUniqueTokenNameIndexName, DbTokenSubjectIndexName},
					index["name"],
				)
				count = count + 1
			}
			assert.Equal(t, 2, count)
		})
	}
}
