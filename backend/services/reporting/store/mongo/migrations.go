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

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/pkg/errors"
)

const (
	// DbVersion is the current schema version
	DbVersion = "1.0.0"

	// DbName is the database name
	DbName = "reporting"
)

// Migrate applies all migrations under the given context. That is, if ctx
// has an associated identity.Identity set, it will ONLY migrate the single
// tenant's db. If ctx does not have an identity, all reporting databases
// will be migrated to the given version.
func (db *MongoStore) Migrate(ctx context.Context, version string, automigrate bool) error {
	ver, err := migrate.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "failed to parse service version")
	}
	l := log.FromContext(ctx)
	l.Infof("Migrating database: %s", db.config.DbName)

	m := migrate.SimpleMigrator{
		Client:      db.client,
		Db:          db.config.DbName,
		Automigrate: automigrate,
	}
	migrations := []migrate.Migration{
		&migration_1_0_0{
			client: db.client,
			db:     db.config.DbName,
		},
	}
	err = m.Apply(ctx, *ver, migrations)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}
	return nil
}

func (db *MongoStore) MigrateLatest(ctx context.Context) error {
	return db.Migrate(ctx, DbVersion, true)
}
