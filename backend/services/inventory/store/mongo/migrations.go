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

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/inventory/store"
)

// WithAutomigrate enables automatic migration and returns a new datastore based
// on current one
func (db *DataStoreMongo) WithAutomigrate() store.DataStore {
	return &DataStoreMongo{
		client:      db.client,
		automigrate: true,
	}
}

func (db *DataStoreMongo) MigrateTenant(
	ctx context.Context,
	version string,
	tenantId string,
) error {
	l := log.FromContext(ctx)

	database := mstore.DbNameForTenant(tenantId, DbName)

	l.Infof("migrating %s", database)

	m := migrate.SimpleMigrator{
		Client:      db.client,
		Db:          database,
		Automigrate: db.automigrate,
	}

	ver, err := migrate.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "failed to parse service version")
	}

	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantId,
	})

	migrations := []migrate.Migration{
		&migration_0_2_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_0_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_0_1{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_0_2{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_1_0{
			ms:  db,
			ctx: ctx,
		},
	}

	err = m.Apply(ctx, *ver, migrations)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}

	return nil
}

func (db *DataStoreMongo) Migrate(ctx context.Context, version string) error {
	l := log.FromContext(ctx)

	dbs, err := migrate.GetTenantDbs(ctx, db.client, mstore.IsTenantDb(DbName))
	if err != nil {
		return errors.Wrap(err, "failed go retrieve tenant DBs")
	}

	if len(dbs) == 0 {
		dbs = []string{DbName}
	}

	if db.automigrate {
		l.Infof("automigrate is ON, will apply migrations")
	} else {
		l.Infof("automigrate is OFF, will check db version compatibility")
	}

	for _, d := range dbs {
		l.Infof("migrating %s", d)

		tenantId := mstore.TenantFromDbName(d, DbName)

		if err := db.MigrateTenant(ctx, version, tenantId); err != nil {
			return err
		}
	}
	return nil
}

func (db *DataStoreMongo) upgradeTenant(
	ctx context.Context,
	version string,
) error {
	var migration migrate.Migration
	_, err := migrate.NewVersion(version)
	if err != nil {
		return err
	}

	switch version {
	case "1.0.0":
		migration = &migration_1_0_0{
			ms:          db,
			ctx:         ctx,
			maintenance: true,
		}
	default:
		return errors.Errorf(
			"migration version %s does not provide "+
				"a maintenance interface",
			version,
		)
	}
	tenantDB := mstore.DbFromContext(ctx, DbName)

	migrationInfo, err := migrate.GetMigrationInfo(
		ctx, db.client, tenantDB,
	)
	if err != nil {
		return errors.Wrap(err, "failed to fetch migration info")
	}
	migrationVersion := migrate.Version{Major: 0, Minor: 0, Patch: 0}
	if len(migrationInfo) > 0 {
		migrationVersion = migrationInfo[0].Version
	}
	err = migration.Up(migrationVersion)
	if err != nil {
		return err
	}
	return nil

}

func (db *DataStoreMongo) Maintenance(
	ctx context.Context,
	version string,
	tenantIDs ...string,
) error {
	l := log.FromContext(ctx)

	if len(tenantIDs) > 0 {
		for _, tid := range tenantIDs {
			tenantCTX := identity.WithContext(ctx,
				&identity.Identity{
					Tenant: tid,
				},
			)
			err := db.upgradeTenant(tenantCTX, version)
			if err != nil {
				return err
			}
		}
	} else {
		dbs, err := migrate.GetTenantDbs(
			ctx, db.client, mstore.IsTenantDb(DbName),
		)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve tenant DBs")
		}
		if len(dbs) == 0 {
			dbs = []string{DbName}
		}
		for _, d := range dbs {
			tenantID := mstore.TenantFromDbName(d, DbName)
			l.Infof("Updating DB: %s", d)
			tenantCTX := identity.WithContext(
				ctx,
				&identity.Identity{
					Tenant: tenantID,
				},
			)
			err := db.upgradeTenant(tenantCTX, version)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
