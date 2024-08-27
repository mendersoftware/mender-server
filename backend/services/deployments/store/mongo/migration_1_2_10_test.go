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
	"encoding/base32"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

const (
	mongoMaxDatabaseNameLen = 63

	mongoIndexKeySpecsConflict = 86
)

func Test_migration_1_2_10_Up(t *testing.T) {
	tests := []struct {
		Name string

		Setup func(db *mongo.Database)

		ErrorCode int
	}{{
		Name: "ok",

		Setup: func(db *mongo.Database) {
			err := MigrateSingle(
				context.Background(),
				db.Name(),
				"1.2.9",
				db.Client(),
				true,
			)
			if err != nil {
				panic(err)
			}
		},
	}, {
		Name: "ok/index already exist",

		Setup: func(db *mongo.Database) {
			_, err := db.Collection(CollectionDevices).
				Indexes().
				CreateOne(context.Background(),
					IndexDeviceDeploymentsActiveCreatedModel)
			if err != nil {
				panic(err)
			}
		},
	}, {
		Name: "error/devices index name taken",

		Setup: func(db *mongo.Database) {
			_, err := db.Collection(CollectionDevices).
				Indexes().
				CreateOne(context.Background(),
					mongo.IndexModel{
						Keys: bson.D{{Key: "bogus", Value: 1}},
						Options: options.Index().
							SetName(IndexDeviceDeploymentsActiveCreated),
					},
				)
			if err != nil {
				panic(err)
			}
		},
		ErrorCode: mongoIndexKeySpecsConflict,
	}, {
		Name: "error/deployments index name taken",

		Setup: func(db *mongo.Database) {
			_, err := db.Collection(CollectionDeployments).
				Indexes().
				CreateOne(context.Background(),
					mongo.IndexModel{
						Keys: bson.D{{Key: "bogus", Value: 1}},
						Options: options.Index().
							SetName(IndexDeploymentsActiveCreatedV2),
					},
				)
			if err != nil {
				panic(err)
			}
		},

		ErrorCode: mongoIndexKeySpecsConflict,
	}}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			dbName := strings.ToLower(strings.Trim(base32.StdEncoding.
				EncodeToString([]byte(t.Name())), "="))
			if len(dbName) > mongoMaxDatabaseNameLen {
				dbName = dbName[len(dbName)-mongoMaxDatabaseNameLen:]
			}
			client := db.Client()
			db := client.Database(dbName)
			if tc.Setup != nil {
				tc.Setup(client.Database(dbName))
			}
			m := &migration_1_2_10{
				client: client,
				db:     dbName,
			}
			err := m.Up(migrate.Version{})
			if tc.ErrorCode != 0 {
				var srvErr mongo.ServerError
				if assert.ErrorAs(t, err, &srvErr) {
					assert.True(t, srvErr.HasErrorCode(tc.ErrorCode))
				}
			} else if assert.NoError(t, err) {
				_, err = db.Collection(CollectionDevices).
					Indexes().
					ListSpecifications(context.Background())
				if err != nil {
					panic(err)
				}
				_, err = db.Collection(CollectionDeployments).
					Indexes().
					ListSpecifications(context.Background())
				if err != nil {
					panic(err)
				}
			}
		})
	}
}
