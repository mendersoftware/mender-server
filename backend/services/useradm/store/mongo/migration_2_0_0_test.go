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
	"sort"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore_v1 "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/useradm/model"
)

func TestMigration_2_0_0(t *testing.T) {
	//now := time.Now().UTC().Round(time.Second).Truncate(0)

	cases := map[string]struct {
		data map[string]struct {
			inUsers  []model.User
			outUsers []model.User
			outErr   error
		}
	}{
		"ok": {
			data: map[string]struct {
				inUsers  []model.User
				outUsers []model.User
				outErr   error
			}{
				"foo": {
					inUsers: []model.User{
						{
							ID:    "foo",
							Email: "foo@bar.com",
						},
					},
					outUsers: []model.User{
						{
							ID:    "foo",
							Email: "foo@bar.com",
						},
					},
				},
				"bar": {
					inUsers: []model.User{
						{
							ID:    "bar",
							Email: "bar@bar.com",
						},
					},
					outUsers: []model.User{
						{
							ID:    "bar",
							Email: "bar@bar.com",
						},
					},
				},
			},
		},
		"duplicated email address": {
			data: map[string]struct {
				inUsers  []model.User
				outUsers []model.User
				outErr   error
			}{
				"bar": {
					inUsers: []model.User{
						{
							ID:    "bar",
							Email: "foo@bar.com",
						},
					},
					outUsers: []model.User{
						{
							ID:    "bar",
							Email: "foo@bar.com",
						},
					},
				},
				"foo": {
					inUsers: []model.User{
						{
							ID:    "foo",
							Email: "foo@bar.com",
						},
					},
					outUsers: []model.User{},
					outErr:   errors.New("failed to apply migration from 0.0.0 to 2.0.0: bulk write exception: write errors: [E11000 duplicate key error collection: useradm.users index: email_1 dup key: { email: \"foo@bar.com\" }]"),
				},
			},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			ctx := context.Background()
			db.Wipe()
			c := db.Client()
			ds, _ := NewDataStoreMongoWithClient(c)

			dbName := DbName
			// run the 2.0.0 migration for the default database
			migrations := []migrate.Migration{
				&migration_2_0_0{
					ds:     ds,
					dbName: dbName,
					ctx:    ctx,
				},
			}
			migrator := &migrate.SimpleMigrator{
				Client:      c,
				Db:          dbName,
				Automigrate: true,
			}

			err := migrator.Apply(ctx, migrate.MakeVersion(2, 0, 0), migrations)
			assert.NoError(t, err)

			// create the documents in the tenant-specific databases
			for tenant, d := range tc.data {
				ctx := ctx
				if tenant != "" {
					ctx = identity.WithContext(ctx, &identity.Identity{
						Tenant: tenant,
					})
				}

				dbName := mstore_v1.DbNameForTenant(tenant, DbName)
				// insert users
				if len(d.inUsers) > 0 {
					coll := c.Database(dbName).Collection(DbUsersColl)
					docs := make([]interface{}, len(d.inUsers))
					for i, v := range d.inUsers {
						docs[i] = v
					}
					_, err := coll.InsertMany(ctx, docs)
					assert.NoError(t, err)
				}
			}

			// in the test, migration order is important
			keys := make([]string, 0, len(tc.data))
			for k := range tc.data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, tenant := range keys {
				d := tc.data[tenant]
				ctx := ctx
				if tenant != "" {
					ctx = identity.WithContext(ctx, &identity.Identity{
						Tenant: tenant,
					})
				}
				dbName := mstore_v1.DbNameForTenant(tenant, DbName)

				migrations := []migrate.Migration{
					&migration_2_0_0{
						ds:     ds,
						dbName: dbName,
						ctx:    ctx,
					},
				}
				migrator := &migrate.SimpleMigrator{
					Client:      c,
					Db:          dbName,
					Automigrate: true,
				}

				err := migrator.Apply(ctx, migrate.MakeVersion(2, 0, 0), migrations)
				if d.outErr != nil {
					assert.EqualError(t, err, d.outErr.Error())
				} else {
					assert.NoError(t, err)
				}
			}
			for tenant, d := range tc.data {
				ctx := ctx
				if tenant != "" {
					ctx = identity.WithContext(ctx, &identity.Identity{
						Tenant: tenant,
					})
				}

				foundUsers, err := ds.GetUsers(ctx, model.UserFilter{})
				assert.NoError(t, err)
				assert.Equal(t, d.outUsers, foundUsers)
			}
		})
	}

}
