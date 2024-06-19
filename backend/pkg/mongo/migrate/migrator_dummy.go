// Copyright 2023 Northern.tech AS
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
package migrate

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/log"
)

// MigratorDummy does not actually apply migrations, just inserts the
// target version into the db to mark the initial/current state.
type DummyMigrator struct {
	Client      *mongo.Client
	Db          string
	Automigrate bool
}

// Apply makes MigratorDummy implement the Migrator interface.
func (m *DummyMigrator) Apply(ctx context.Context, target Version, migrations []Migration) error {
	l := log.FromContext(ctx).F(log.Ctx{"db": m.Db})

	applied, err := GetMigrationInfo(ctx, m.Client, m.Db)
	if err != nil {
		return err
	}

	if len(applied) > 1 {
		return errors.New("dummy migrator cannot apply migrations, more than 1 already applied")
	}

	last := Version{}
	if len(applied) == 1 {
		last = applied[0].Version
	}

	if !m.Automigrate {
		if VersionIsLess(last, target) {
			return fmt.Errorf(ErrNeedsMigration+": %s has version %s, needs version %s", m.Db, last.String(), target.String())
		} else {
			return nil
		}
	}

	if VersionIsLess(last, target) {
		l.Infof("applying migration from version %s to %s", last, target)
		return UpdateMigrationInfo(ctx, target, m.Client, m.Db)
	} else {
		l.Infof("migration to version %s skipped", target)
	}

	return nil
}
