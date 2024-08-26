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
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
)

type migration_1_11_0 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

// Up creates an index on status and id in the devices collection
func (m *migration_1_11_0) Up(from migrate.Version) error {
	const pubkey = "pubkey"
	cDevs := m.ds.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbDevicesColl)
	_, err := cDevs.UpdateMany(m.ctx, bson.M{
		pubkey: bson.M{"$exists": 1},
	}, bson.M{
		"$unset": bson.M{pubkey: 1},
	})

	return err
}

func (m *migration_1_11_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 11, 0)
}
