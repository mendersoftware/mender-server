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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

const (
	// DbVersion is the current schema version
	DbVersion = "1.1.1"

	// DbName is the database name
	DbName = "iot_manager"
)

// Migrate applies migrations to the database
func Migrate(ctx context.Context,
	db string,
	version string,
	client *mongo.Client,
	automigrate bool) error {
	l := log.FromContext(ctx)

	l.Infof("migrating %s", db)

	ver, err := migrate.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "failed to parse service version")
	}

	m := migrate.SimpleMigrator{
		Client:      client,
		Db:          db,
		Automigrate: automigrate,
	}

	migrations := []migrate.Migration{
		&migration_1_0_0{
			client: client,
			db:     db,
		},
		&migration_1_1_0{
			client: client,
			db:     db,
		},
		&migration_1_1_1{
			client: client,
			db:     db,
		},
	}

	err = m.Apply(ctx, *ver, migrations)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}

	return nil
}
