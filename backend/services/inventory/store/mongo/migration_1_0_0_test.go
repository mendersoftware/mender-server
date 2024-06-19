// Copyright 2021 Northern.tech AS
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
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

func TestMigration_1_0_0(t *testing.T) {
	testTimestamp := time.Now()
	cases := map[string]struct {
		inDevs  []interface{}
		outDevs []model.Device
		tenant  string
	}{
		"single dev": {
			inDevs: []interface{}{
				legacyDevice{
					ID: model.DeviceID("1"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					Group:     "foobar",
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "group",
						Value: "foobar",
						Scope: model.AttrScopeSystem,
					}},
					Group: "foobar",
				},
			},
		},
		"one ungrouped and one grouped device": {
			inDevs: []interface{}{
				legacyDevice{
					ID: model.DeviceID("1"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					Group:     "foobar",
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
				legacyDevice{
					ID: model.DeviceID("2"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "group",
						Value: "foobar",
						Scope: model.AttrScopeSystem,
					}},
					Group: "foobar",
				}, {
					ID: model.DeviceID("2"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}},
				},
			},
		},
		"multiple devs, with tenant": {
			tenant: "foobarbaz",
			inDevs: []interface{}{
				legacyDevice{
					ID: model.DeviceID("1"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					Group:     "foobar",
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
				legacyDevice{
					ID: model.DeviceID("2"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
				legacyDevice{
					ID:        model.DeviceID("3"),
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "group",
						Value: "foobar",
						Scope: model.AttrScopeSystem,
					}},
					Group: "foobar",
				}, {
					ID: model.DeviceID("2"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}},
				}, {
					ID: model.DeviceID("3"),
					Attributes: model.DeviceAttributes{{
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}},
				},
			},
		},
	}
	for n, tc := range cases {
		t.Run(fmt.Sprintf("tc %s", n), func(t *testing.T) {
			ctx := context.Background()

			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			// setup
			db.Wipe()
			s := db.Client()
			ds := NewDataStoreMongoWithSession(s).(*DataStoreMongo)

			migrations := []migrate.Migration{
				&migration_0_2_0{
					ms:  ds,
					ctx: ctx,
				},
				&migration_1_0_0{
					ms:  ds,
					ctx: ctx,
				},
			}
			migrator := &migrate.SimpleMigrator{
				Client:      s,
				Db:          mstore.DbFromContext(ctx, DbName),
				Automigrate: true,
			}

			c := s.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
			_, err := c.InsertMany(ctx, tc.inDevs)
			assert.NoError(t, err)

			err = migrator.Apply(ctx, migrate.MakeVersion(1, 0, 0), migrations)
			assert.NoError(t, err)

			var dbdevs []*model.Device
			devsColl := s.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
			cursor, err := devsColl.Find(ctx, bson.M{})
			cursor.All(ctx, &dbdevs)

			assert.NoError(t, err)
			assert.Len(t, dbdevs, len(tc.outDevs))

			sort.Slice(dbdevs, func(i, j int) bool { return dbdevs[i].ID < dbdevs[j].ID })
			for i := range dbdevs {
				compareDevsWithoutTimestamps(t, &tc.outDevs[i], dbdevs[i])
			}
		})
	}
}

func TestMaintenance_1_0_0(t *testing.T) {
	testTimestamp := time.Now()
	cases := map[string]struct {
		inDevs  []interface{}
		outDevs []model.Device
		tenant  string
	}{
		"single dev": {
			inDevs: []interface{}{
				legacyDevice{
					ID: model.DeviceID("1"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					Group:     "foobar",
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "group",
						Value: "foobar",
						Scope: model.AttrScopeSystem,
					}},
					Group: "foobar",
				},
			},
		},
		"one ungrouped and one grouped device": {
			inDevs: []interface{}{
				legacyDevice{
					ID: model.DeviceID("1"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					Group:     "foobar",
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
				legacyDevice{
					ID: model.DeviceID("2"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "group",
						Value: "foobar",
						Scope: model.AttrScopeSystem,
					}},
					Group: "foobar",
				}, {
					ID: model.DeviceID("2"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}},
				},
			},
		},
		"multiple devs, with tenant": {
			tenant: "foobarbaz",
			inDevs: []interface{}{
				legacyDevice{
					ID: model.DeviceID("1"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					Group:     "foobar",
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
				legacyDevice{
					ID: model.DeviceID("2"),
					Attributes: map[string]model.DeviceAttribute{
						"foo": {
							Name:        "foo",
							Value:       "val3",
							Description: strPtr("desc"),
						},
					},
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
				legacyDevice{
					ID:        model.DeviceID("3"),
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "group",
						Value: "foobar",
						Scope: model.AttrScopeSystem,
					}},
					Group: "foobar",
				}, {
					ID: model.DeviceID("2"),
					Attributes: model.DeviceAttributes{{
						Name:        "foo",
						Value:       "val3",
						Description: strPtr("desc"),
						Scope:       model.AttrScopeInventory,
					}, {
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}},
				}, {
					ID: model.DeviceID("3"),
					Attributes: model.DeviceAttributes{{
						Name:  "created_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}, {
						Name:  "updated_ts",
						Value: testTimestamp,
						Scope: model.AttrScopeSystem,
					}},
				},
			},
		},
	}
	for n, tc := range cases {
		t.Run(fmt.Sprintf("tc %s", n), func(t *testing.T) {
			var err error
			ctx := context.Background()

			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			// setup
			db.Wipe()
			s := db.Client()
			ds := NewDataStoreMongoWithSession(s).(*DataStoreMongo)

			migrations := []migrate.Migration{
				&migration_0_2_0{
					ms:  ds,
					ctx: ctx,
				},
			}
			migrator := &migrate.SimpleMigrator{
				Client:      s,
				Db:          mstore.DbFromContext(ctx, DbName),
				Automigrate: true,
			}

			c := s.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
			_, err = c.InsertMany(ctx, tc.inDevs)
			assert.NoError(t, err)

			err = migrator.Apply(ctx, migrate.MakeVersion(0, 2, 0), migrations)
			assert.NoError(t, err)

			// Do maintenance migration
			if tc.tenant != "" {
				err = ds.Maintenance(ctx, "1.0.0", tc.tenant)
			} else {
				err = ds.Maintenance(ctx, "1.0.0")
			}

			dbDevs := make([]model.Device, len(tc.inDevs))
			dbDoc := make([]bson.M, len(tc.inDevs))
			database := s.Database(mstore.DbFromContext(ctx, DbName))
			devsColl := database.Collection(DbDevicesColl)
			mgrInfoColl := database.Collection("migration_info")

			res := mgrInfoColl.FindOne(ctx, bson.M{"maintenance": bson.M{"$exists": true}})
			if !assert.NoError(t, res.Err()) {
				t.FailNow()
			}
			cursor, err := devsColl.Find(ctx, bson.M{})
			assert.NoError(t, err)
			for i := 0; cursor.Next(ctx); i++ {
				err = cursor.Decode(&dbDevs[i])
				assert.NoError(t, err)
				err = cursor.Decode(&dbDoc[i])
				assert.NoError(t, err)
			}
			// Check that backwards compatibility is preserved
			for i, dev := range dbDevs {
				for _, attr := range dev.Attributes {
					switch attr.Name {
					case "group":
						_, ok := dbDoc[i]["group"]
						assert.True(t, ok,
							"Backward-compatible group field not present",
						)
					case "updated_ts":
						_, ok := dbDoc[i]["updated_ts"]
						assert.True(t, ok,
							"Backward-compatible updated_ts field not present",
						)

					case "created_ts":
						_, ok := dbDoc[i]["created_ts"]
						assert.True(t, ok,
							"Backward-compatible created_ts field not present",
						)
					}
				}
			}

			assert.NoError(t, err)
			assert.Len(t, dbDevs, len(tc.outDevs))

			sort.Slice(dbDevs, func(i, j int) bool { return dbDevs[i].ID < dbDevs[j].ID })
			for i := range dbDevs {
				compareDevsWithoutTimestamps(t, &tc.outDevs[i], &dbDevs[i])
			}
			// Finally check that old fields are removed on migration
			ds.automigrate = true
			if tc.tenant != "" {
				err = ds.MigrateTenant(ctx, "1.0.0", tc.tenant)
			} else {
				err = ds.Migrate(ctx, "1.0.0")
			}
			assert.NoError(t, err)
			cursor, err = devsColl.Find(ctx, bson.M{})
			assert.NoError(t, err)
			// Reset document arrays
			dbDevs = make([]model.Device, len(tc.inDevs))
			dbDoc = make([]bson.M, len(tc.inDevs))
			for i := 0; cursor.Next(ctx); i++ {
				err = cursor.Decode(&dbDevs[i])
				assert.NoError(t, err)
				err = cursor.Decode(&dbDoc[i])
				assert.NoError(t, err)
			}
			// Check that backwards compatibility is preserved
			for i, dev := range dbDevs {
				for _, attr := range dev.Attributes {
					switch attr.Name {
					case "group":
						_, ok := dbDoc[i]["group"]
						assert.False(t, ok,
							"Backward-compatible group field present after migration",
						)
					case "updated_ts":
						_, ok := dbDoc[i]["updated_ts"]
						assert.False(t, ok,
							"Backward-compatible updated_ts field present after migration",
						)

					case "created_ts":
						_, ok := dbDoc[i]["created_ts"]
						assert.False(t, ok,
							"Backward-compatible created_ts field present after migration",
						)
					}
				}
			}

			assert.NoError(t, err)
			assert.Len(t, dbDevs, len(tc.outDevs))

			sort.Slice(dbDevs, func(i, j int) bool { return dbDevs[i].ID < dbDevs[j].ID })
			for i := range dbDevs {
				compareDevsWithoutTimestamps(t, &tc.outDevs[i], &dbDevs[i])
			}
		})
	}
}
