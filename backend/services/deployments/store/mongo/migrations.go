// Copyright 2024 Northern.tech AS
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

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctx_store "github.com/mendersoftware/mender-server/pkg/store"
)

const (
	DbVersion        = "1.2.17"
	DbMinimumVersion = "1.2.17"
	DbName           = "deployment_service"
)

func Migrate(ctx context.Context,
	version string,
	client *mongo.Client,
	automigrate bool) error {

	l := log.FromContext(ctx)

	dbs, err := migrate.GetTenantDbs(ctx, client, ctx_store.IsTenantDb(DbName))
	if err != nil {
		return errors.Wrap(err, "failed go retrieve tenant DBs")
	}

	if len(dbs) == 0 {
		dbs = []string{DbName}
	}

	if automigrate {
		l.Infof("automigrate is ON, will apply migrations")
	} else {
		l.Infof("automigrate is OFF, will check db version compatibility")
	}

	for _, d := range dbs {
		err := MigrateSingle(ctx, d, version, client, automigrate)
		if err != nil {
			return err
		}
	}

	return nil
}

func MigrateSingle(ctx context.Context,
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
		&migration_1_2_1{
			client: client,
			db:     db,
		},
		&migration_1_2_2{
			client: client,
			db:     db,
		},
		&migration_1_2_3{
			client: client,
			db:     db,
		},
		&migration_1_2_4{
			client: client,
			db:     db,
		},
		&migration_1_2_5{
			client: client,
			db:     db,
		},
		&migration_1_2_6{
			client: client,
			db:     db,
		},
		&migration_1_2_7{
			client: client,
			db:     db,
		},
		&migration_1_2_9{
			client: client,
			db:     db,
		},
		&migration_1_2_10{
			client: client,
			db:     db,
		},
		&migration_1_2_11{
			client: client,
			db:     db,
		},
		&migration_1_2_13{
			client: client,
			db:     db,
		},
		&migration_1_2_14{
			client: client,
			db:     db,
		},
		&migration_1_2_15{
			client: client,
			db:     db,
		},
		&migration_1_2_16{
			client: client,
			db:     db,
		},
		&migration_1_2_17{
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
