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

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

const (
	DbVersion        = "2.0.0"
	DbMinimumVersion = "1.2.17"
	DefaultDbName    = "deployments"
)

func (mgo *DataStoreMongo) Migrate(
	ctx context.Context,
	version string,
	automigrate bool,
) error {

	l := log.FromContext(ctx)

	if automigrate {
		l.Infof("automigrate is ON, will apply migrations")
	} else {
		l.Infof("automigrate is OFF, will check db version compatibility")
	}

	l.Infof("migrating %s", mgo.database.Name())

	ver, err := migrate.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "failed to parse service version")
	}

	m := migrate.SimpleMigrator{
		// FIXME: SimplerMigrator that swaps Client and Db with *mongo.Database
		Client:      mgo.database.Client(),
		Db:          mgo.database.Name(),
		Automigrate: automigrate,
	}

	// FIXME: Insert BIG migration initializing collections and transforming existing data
	//      - Can purge existing migrations from codebase but require last version is either nil or DbMinimumVersion
	migrations := []migrate.Migration{}

	err = m.Apply(ctx, *ver, migrations)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}

	return nil
}
