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

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/mendersoftware/mender-server/services/useradm/model"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration_2_0_1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestMigration_2_0_1 in short mode")
	}
	testCases := []struct {
		Name string
	}{
		{
			Name: "Successful migration",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db.Wipe()
			ctx := context.Background()
			client := db.Client()
			ds, err := NewDataStoreMongoWithClient(client)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			migrations := []migrate.Migration{
				&migration_2_0_1{
					ds:     ds,
					ctx:    ctx,
					dbName: DbName,
				},
			}

			coll := client.Database(DbName).Collection(DbSettingsColl)

			userID := uuid.NewString()
			tenantID := uuid.NewString()
			coll.InsertOne(ctx, bson.M{
				"tenant_id": tenantID,
				"key":       "value",
				userID: bson.M{
					"user-key": "user-value",
					"onboarding": bson.M{
						"nested-key": "nested-value",
					},
				},
			})

			m := migrate.SimpleMigrator{
				Client:      client,
				Db:          DbName,
				Automigrate: true,
			}
			err = m.Apply(ctx, migrate.MakeVersion(2, 0, 1), migrations)
			assert.NoError(t, err)

			coll = client.Database(DbName).Collection(DbUserSettingsColl)

			settings := &model.Settings{}
			err = coll.FindOne(ctx, bson.M{
				"tenant_id": tenantID,
				"user_id":   userID,
			}).Decode(settings)
			assert.NoError(t, err)
		})
	}
}
