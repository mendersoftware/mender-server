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

type migration_1_9_0 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

// Up removes device and authset indexes which include raw id data
// and creates device index on id data sha256
func (m *migration_1_9_0) Up(from migrate.Version) error {
	_false := false
	_true := true

	// create device index on id data sha
	devIdDataSha256UniqueIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: model.DevKeyIdDataSha256, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &indexDevices_IdentityDataSha256,
			Unique:     &_true,
		},
	}

	cDevs := m.ds.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbDevicesColl)
	devIndexes := cDevs.Indexes()
	_, err := devIndexes.CreateOne(m.ctx, devIdDataSha256UniqueIndex)
	if err != nil {
		return errors.Wrap(err, "failed to create unique index containing IdDataSha256 on devices")
	}

	// drop device index on raw identity data, if exists
	_, err = devIndexes.DropOne(m.ctx, indexDevices_IdentityData)
	if err != nil && !isIndexNotFound(err) {
		return errors.Wrap(err, "failed to drop index devices:IdentityData")
	}

	// drop authset index on raw identity data, if exists
	cAuthsets := m.ds.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).
		Collection(DbAuthSetColl)
	asetIndexes := cAuthsets.Indexes()
	_, err = asetIndexes.DropOne(m.ctx, indexAuthSet_DeviceId_IdentityData_PubKey)
	if err != nil && !isIndexNotFound(err) {
		return errors.Wrap(err, "failed to drop index auth_sets:DeviceId:IdData:PubKey")
	}

	return nil
}

func (m *migration_1_9_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 9, 0)
}

// ref: https://github.com/mongodb/mongo/blob/master/src/mongo/base/error_codes.yml
func isIndexNotFound(e error) bool {
	if mgoErr, ok := e.(mongo.CommandError); ok {
		if mgoErr.Code == 27 || // IndexNotFound - index does not exist
			mgoErr.Code == 26 { // NamespaceNotFound - collection does not exist
			return true
		}
	}
	return false
}
