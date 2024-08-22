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
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	uto "github.com/mendersoftware/mender-server/services/deviceauth/utils/to"
)

type migration_1_5_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_5_0) Up(from migrate.Version) error {
	devColl := m.ms.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbDevicesColl)
	asColl := m.ms.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbAuthSetColl)

	cursor, err := asColl.Find(m.ctx, bson.M{})
	if err != nil {
		return err
	}

	var set model.AuthSet

	for cursor.Next(m.ctx) {
		if err = cursor.Decode(&set); err != nil {
			continue
		}

		idDataStruct, err := decode(set.IdData)

		if err != nil {
			return errors.Wrapf(
				err,
				"failed to parse id data of auth set %v: %v",
				set.Id,
				set.IdData,
			)
		}

		hash := sha256.New()
		hash.Write([]byte(set.IdData))

		update := bson.M{
			"$set": model.AuthSetUpdate{
				IdDataStruct: idDataStruct,
				IdDataSha256: hash.Sum(nil),
				Timestamp:    uto.TimePtr(time.Now()),
			},
		}

		_, err = asColl.UpdateOne(m.ctx, bson.M{"_id": set.Id}, update)
		if err != nil {
			return errors.Wrapf(err, "failed to update auth set %v", set.Id)
		}

	}

	if err := cursor.Close(m.ctx); err != nil {
		return errors.Wrap(err, "failed to close DB iterator")
	}

	_false := false
	_true := true
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: model.AuthSetKeyDeviceId, Value: 1},
			{Key: model.AuthSetKeyIdDataSha256, Value: 1},
			{Key: model.AuthSetKeyPubKey, Value: 1},
		},
		Options: &options.IndexOptions{
			Background: &_false,
			Name:       &indexAuthSet_DeviceId_IdentityDataSha256_PubKey,
			Unique:     &_true,
		},
	}

	asIndexes := asColl.Indexes()
	_, err = asIndexes.CreateOne(m.ctx, index)
	if err != nil {
		return errors.Wrap(err, "failed to create index on auth sets")
	}

	cursor, err = devColl.Find(m.ctx, bson.M{})
	if err != nil {
		return err
	}

	var dev model.Device

	for cursor.Next(m.ctx) {
		if err = cursor.Decode(&dev); err != nil {
			continue
		}

		idDataStruct, err := decode(dev.IdData)

		if err != nil {
			return errors.Wrapf(err, "failed to parse id data of device %v: %v", dev.Id, dev.IdData)
		}

		hash := sha256.New()
		hash.Write([]byte(dev.IdData))

		update := bson.M{
			"$set": model.DeviceUpdate{
				IdDataStruct: idDataStruct,
				IdDataSha256: hash.Sum(nil),
				UpdatedTs:    uto.TimePtr(time.Now().UTC()),
			},
		}

		_, err = devColl.UpdateOne(m.ctx, bson.M{"_id": dev.Id}, update)
		if err != nil {
			return errors.Wrapf(err, "failed to update device %v", dev.Id)
		}

	}

	if err := cursor.Close(m.ctx); err != nil {
		return errors.Wrap(err, "failed to close DB iterator")
	}

	return nil
}

func (m *migration_1_5_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 5, 0)
}

func decode(what string) (map[string]interface{}, error) {
	var dec map[string]interface{}

	err := json.Unmarshal([]byte(what), &dec)
	if err != nil {
		return nil, err
	}

	return dec, nil
}
