// Copyright 2021 Northern.tech AS
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

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"

	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

type migration_1_0_2 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_0_2) Up(from migrate.Version) error {
	l := log.FromContext(m.ctx)

	databaseName := mstore.DbFromContext(m.ctx, DbName)
	coll := m.ms.client.Database(databaseName).Collection(DbDevicesColl)
	filter := bson.M{DbDevRevision: bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{DbDevRevision: 0}}
	resp, err := coll.UpdateMany(m.ctx, filter, update)
	if err != nil {
		return err
	}

	l.Infof("Set revision to 0 for %d devices", resp.ModifiedCount)

	return nil
}

func (m *migration_1_0_2) Version() migrate.Version {
	return migrate.MakeVersion(1, 0, 2)
}
