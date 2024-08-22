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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

func TestMigration_1_1_0(t *testing.T) {
	cases := map[string]struct {
		tenant string
	}{
		"ok, single tenant": {},
		"ok, multi tenant": {
			tenant: "tenant",
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
				&migration_1_1_0{
					ms:  ds,
					ctx: ctx,
				},
			}
			migrator := &migrate.SimpleMigrator{
				Client:      s,
				Db:          mstore.DbFromContext(ctx, DbName),
				Automigrate: true,
			}

			err := migrator.Apply(ctx, migrate.MakeVersion(1, 1, 0), migrations)
			assert.NoError(t, err)

			devsColl := s.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
			indexView := devsColl.Indexes()
			cur, err := indexView.List(ctx)
			assert.NoError(t, err)

			var idxs []bson.M
			err = cur.All(context.TODO(), &idxs)
			assert.NoError(t, err)

			found := false
			for _, idx := range idxs {
				if idx["name"] == DbDevAttributesText {
					found = true
					break
				}
			}
			assert.True(t, found)
		})
	}
}
