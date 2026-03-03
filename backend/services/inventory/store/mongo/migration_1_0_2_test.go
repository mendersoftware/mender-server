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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

func TestMigration_1_0_2(t *testing.T) {
	testTimestamp := time.Now()
	cases := map[string]struct {
		inDevs  []interface{}
		outDevs []model.Device
		tenant  string
	}{
		"no revision": {
			inDevs: []interface{}{
				legacyDevice{
					ID:        model.DeviceID("1"),
					UpdatedTs: testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID:       model.DeviceID("1"),
					Revision: 0,
				},
			},
		},
		"existing revision": {
			inDevs: []interface{}{
				model.Device{
					ID:        model.DeviceID("1"),
					Revision:  1,
					UpdatedTs: &testTimestamp,
					CreatedTs: testTimestamp,
				},
			},
			outDevs: []model.Device{
				{
					ID:       model.DeviceID("1"),
					Revision: 1,
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
				&migration_1_0_1{
					ms:  ds,
					ctx: ctx,
				},
				&migration_1_0_2{
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

			err = migrator.Apply(ctx, migrate.MakeVersion(1, 0, 2), migrations)
			assert.NoError(t, err)

			var dbdevs []*model.Device
			devsColl := s.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
			cursor, err := devsColl.Find(ctx, bson.M{})
			cursor.All(ctx, &dbdevs)

			assert.NoError(t, err)
			assert.Len(t, dbdevs, len(tc.outDevs))

			for i, outDev := range tc.outDevs {
				assert.Equal(t, outDev.Revision, dbdevs[i].Revision)
			}
		})
	}
}
