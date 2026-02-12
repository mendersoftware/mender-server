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

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"

	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

const (
	DbVersion = "2.0.3"
	DbName    = "useradm"
)

// MigrateTenant migrates a single tenant database.
func (db *DataStoreMongo) MigrateTenant(ctx context.Context, version string, tenant string) error {
	ver, err := migrate.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "failed to parse service version")
	}

	tenantCtx := identity.WithContext(ctx, &identity.Identity{
		Tenant: tenant,
	})

	m := migrate.SimpleMigrator{
		Client:      db.client,
		Db:          mstore.DbFromContext(tenantCtx, DbName),
		Automigrate: db.automigrate,
	}
	migrations := []migrate.Migration{
		&migration_1_0_0{
			ds:  db,
			ctx: tenantCtx,
		},
		&migration_1_1_4{
			ds:  db,
			ctx: tenantCtx,
		},
		&migration_1_3_2{
			ds:  db,
			ctx: tenantCtx,
		},
		&migration_2_0_0{
			ds:     db,
			dbName: mstore.DbFromContext(tenantCtx, DbName),
			ctx:    tenantCtx,
		},
		&migration_2_0_1{
			ds:     db,
			dbName: mstore.DbFromContext(tenantCtx, DbName),
		},
		&migration_2_0_3{
			ds:     db,
			dbName: mstore.DbFromContext(tenantCtx, DbName),
		},
	}

	err = m.Apply(tenantCtx, *ver, migrations)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}
	return nil
}

func (db *DataStoreMongo) Migrate(ctx context.Context, version string) error {
	l := log.FromContext(ctx)

	dbs := []string{DbName}

	if db.multitenant {
		l.Infof("running migrations in multitenant mode")

		tdbs, err := migrate.GetTenantDbs(ctx, db.client, mstore.IsTenantDb(DbName))
		if err != nil {
			return errors.Wrap(err, "failed go retrieve tenant DBs")
		}
		dbs = append(tdbs, DbName)
	} else {
		l.Infof("running migrations in single tenant mode")
	}

	if db.automigrate {
		l.Infof("automigrate is ON, will apply migrations")
	} else {
		l.Infof("automigrate is OFF, will check db version compatibility")
	}

	for _, d := range dbs {
		l.Infof("migrating %s", d)

		// if not in multi tenant, then tenant will be "" and identity
		// will be the same as default
		tenant := mstore.TenantFromDbName(d, DbName)

		if err := db.MigrateTenant(ctx, version, tenant); err != nil {
			return err
		}
	}

	return nil
}
