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

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

type migration_1_10_0 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

// Up creates an index on status and id in the devices collection
func (m *migration_1_10_0) Up(from migrate.Version) error {
	_false := false

	// create device index on status and id
	devStatusIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: model.DevKeyStatus, Value: 1},
			{Key: model.DevKeyId, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &indexDevices_Status,
			Unique:     &_false,
		},
	}

	cDevs := m.ds.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbDevicesColl)
	devIndexes := cDevs.Indexes()
	_, err := devIndexes.CreateOne(m.ctx, devStatusIndex)
	if err != nil {
		return errors.Wrap(err, "failed to create unique index for status on devices")
	}

	return nil
}

func (m *migration_1_10_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 10, 0)
}
