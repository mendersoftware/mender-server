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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

func TestMigrate(t *testing.T) {
	testCases := []struct {
		Name string

		CTX         context.Context
		Version     string
		Automigrate bool

		Error error
	}{{
		Name: "ok",

		CTX:         context.Background(),
		Version:     DbVersion,
		Automigrate: true,
	}, {
		Name: "error, context canceled applying migrations",
		CTX: func() context.Context {
			ctx := identity.WithContext(context.TODO(),
				&identity.Identity{
					Tenant: "001122334455667788990011",
				})
			ctx, cancel := context.WithCancel(ctx)
			cancel()
			return ctx
		}(),
		Version: DbVersion,
		Error: errors.Errorf(
			"failed to apply migrations: failed to list applied " +
				"migrations: db: failed to get migration " +
				"info: .*context canceled",
		),
	}, {
		Name: "error, bad version",

		CTX:   context.Background(),
		Error: errors.New("failed to parse service version: failed to parse Version: EOF"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := GetTestDataStore(t)
			t.Log(ds.config.DbName)
			defer func() {
				ds.DropDatabase(tc.CTX)
				if identity.FromContext(tc.CTX) != nil {
					ds.DropDatabase(context.Background())
				}
			}()
			err := ds.Migrate(tc.CTX, tc.Version, tc.Automigrate)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				migrationInfo, err := migrate.GetMigrationInfo(
					context.Background(), ds.client, ds.config.DbName,
				)
				assert.NoError(t, err)
				// We don't have any migrations (yet).
				expectedVersion, err := migrate.NewVersion(DbVersion)
				assert.NoError(t, err)
				assert.Equal(t, []migrate.MigrationEntry{{
					Version:   *expectedVersion,
					Timestamp: migrationInfo[0].Timestamp,
				}}, migrationInfo)
			}
		})
	}
}
