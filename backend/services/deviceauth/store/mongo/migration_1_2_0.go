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

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

type migration_1_2_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_2_0) Up(from migrate.Version) error {
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
		newIdData, err := utils.JsonSort(set.IdData)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to sort id data of auth set  %v: %v",
				set.Id,
				set.IdData,
			)
		}

		update := bson.M{
			"$set": model.AuthSetUpdate{
				IdData: newIdData,
			},
		}

		_, err = asColl.UpdateOne(m.ctx, bson.M{"_id": set.Id}, update)
		if err != nil {
			return errors.Wrapf(err, "failed to update auth set  %v", set.Id)
		}

	}

	if err := cursor.Close(m.ctx); err != nil {
		return errors.Wrap(err, "failed to close DB iterator")
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
		newIdData, err := utils.JsonSort(dev.IdData)

		if err != nil {
			return errors.Wrapf(err, "failed to sort id data of device  %v: %v", dev.Id, set.IdData)
		}

		update := bson.M{
			"$set": model.DeviceUpdate{
				IdData: newIdData,
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

func (m *migration_1_2_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 0)
}
