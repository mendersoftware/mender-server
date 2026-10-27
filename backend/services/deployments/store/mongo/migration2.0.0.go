// Copyright 2026 Northern.tech AS
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

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var _ pkgmongo.MigrationMajor = migration_2_0_0{}

type migration_2_0_0 struct{}

func (m migration_2_0_0) Up(ctx context.Context, db *mongo.Database, from migrate.Version) error {
	return nil
}

func (migration_2_0_0) MinimumVersion() migrate.Version {
	return migrate.Version{Major: 1, Minor: 2, Patch: 17}
}

func (migration_2_0_0) Version() migrate.Version {
	return migrate.Version{Major: 2, Minor: 0, Patch: 0}
}
