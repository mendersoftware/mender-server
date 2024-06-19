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
	"strconv"
	"testing"
	"time"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// capped version of the device deployment document
type DeviceDeployment1_2_8 struct {
	ID             string                       `bson:"_id"`
	DeviceID       string                       `bson:"deviceid"`
	Status         model.DeviceDeploymentStatus `bson:"status"`
	Created        time.Time                    `bson:"created"`
	ExpectedActive bool                         `bson:"-"`
}

func TestMigration_1_2_9(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_9 in short mode.")
	}
	ctx := context.Background()

	now := time.Now()
	dataSet := []interface{}{
		DeviceDeployment1_2_8{
			ID:       "0",
			DeviceID: "1",
			Status:   model.DeviceDeploymentStatusAborted,
			Created:  now.Add(-time.Hour),
		},
		DeviceDeployment1_2_8{
			ID:             "1",
			DeviceID:       "1",
			Status:         model.DeviceDeploymentStatusPending,
			Created:        now.Add(-time.Hour / 2),
			ExpectedActive: true,
		},
		DeviceDeployment1_2_8{
			ID:             "2",
			DeviceID:       "1",
			Status:         model.DeviceDeploymentStatusDownloading,
			Created:        now.Add(-time.Hour / 4),
			ExpectedActive: true,
		},
		DeviceDeployment1_2_8{
			ID:       "3",
			DeviceID: "1",
			Status:   model.DeviceDeploymentStatusSuccess,
			Created:  now.Add(-time.Minute),
		},
		DeviceDeployment1_2_8{
			ID:       "4",
			DeviceID: "1",
			Status:   model.DeviceDeploymentStatusDecommissioned,
			Created:  now.Add(-time.Minute),
		},
	}

	// TODO: Test schema migration!

	testCases := map[string]struct {
		// ST or MT naming convention
		db    string
		dbVer string

		err error
	}{
		"ST, from 1.2.8": {
			db:    "deployments_service",
			dbVer: "1.2.8",
		},
		"MT, from 1.2.8": {
			db:    "deployments_service-59afdb71c704db002a86ad95",
			dbVer: "1.2.8",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()
			c := db.Client()

			// setup existing migrations
			if tc.dbVer != "" {
				ver, err := migrate.NewVersion(tc.dbVer)
				assert.NoError(t, err)
				migrate.UpdateMigrationInfo(db.CTX(), *ver, c, tc.db)
			}

			collDevs := c.Database(tc.db).
				Collection(CollectionDevices)
			collDevs.InsertMany(ctx, dataSet)

			migrations := []migrate.Migration{
				&migration_1_2_9{
					client: c,
					db:     tc.db,
				},
			}

			m := migrate.SimpleMigrator{
				Client:      c,
				Db:          tc.db,
				Automigrate: true,
			}

			err := m.Apply(ctx, migrate.MakeVersion(1, 2, 9), migrations)
			assert.NoError(t, err)

			cur, err := collDevs.Find(ctx, bson.D{})
			if err != nil {
				panic(err)
			}
			for cur.Next(ctx) {
				var item struct {
					ID     string `bson:"id"`
					Active *bool  `bson:"active"`
				}
				err := cur.Decode(&item)
				if err != nil {
					panic(err)
				}
				if assert.NotNilf(t,
					item.Active,
					"'active' field is not set for document with _id: '%s'",
					item.ID) {
					continue
				}
				i, err := strconv.Atoi(item.ID)
				if err != nil {
					panic(err)
				} else if i >= len(dataSet) || i < 0 {
					panic("document index out of bounds")
				}
				assert.Equal(t,
					dataSet[i].(DeviceDeployment1_2_8).ExpectedActive,
					*item.Active,
				)

			}

			// verify new indexes are present
			cursor, err := collDevs.
				Indexes().
				List(ctx)
			if !assert.NoError(t, err) {
				return
			}

			expectedIndexes := map[string]bson.D{
				IndexDeviceDeploymentsActiveCreated: {
					{Key: StorageKeyDeviceDeploymentActive, Value: int32(1)},
					{Key: StorageKeyDeviceDeploymentDeviceId, Value: int32(1)},
					{Key: StorageKeyDeviceDeploymentCreated, Value: int32(1)},
				},
			}

			for cursor.Next(ctx) {
				var idx struct {
					Name string `bson:"name"`
					Key  bson.D `bson:"key"`
				}
				err = cursor.Decode(&idx)
				if err != nil {
					panic(err)
				}
				if idx.Name == "_id_" {
					// Skip the default index
					continue
				}

				if assert.Containsf(t,
					expectedIndexes, idx.Name,
					"Found an unexpected index '%s'", idx.Name,
				) {
					assert.EqualValues(t,
						expectedIndexes[idx.Name], idx.Key,
						"Index keys did not match expectations",
					)
				}
			}
		})
	}
}
