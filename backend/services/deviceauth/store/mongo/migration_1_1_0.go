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
	"time"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

type migration_1_1_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

type device_0_1_0 struct {
	Id          string    `bson:"_id,omitempty"`
	TenantToken string    `bson:"tenant_token,omitempty"`
	PubKey      string    `bson:"pubkey,omitempty"`
	IdData      string    `bson:"id_data,omitempty"`
	Status      string    `bson:"status,omitempty"`
	CreatedTs   time.Time `bson:"created_ts,omitempty"`
	UpdatedTs   time.Time `bson:"updated_ts,omitempty"`
}

type token_0_1_0 struct {
	Id    string `bson:"_id,omitempty"`
	DevId string `bson:"dev_id,omitempty"`
	Token string `bson:"token,omitempty"`
}

func (m *migration_1_1_0) Up(from migrate.Version) error {
	devColl := m.ms.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbDevicesColl)
	asColl := m.ms.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbAuthSetColl)
	tColl := m.ms.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).Collection(DbTokensColl)

	if err := m.ensureIndexes(m.ctx); err != nil {
		return errors.Wrap(err, "database indexing failed")
	}

	cursor, err := devColl.Find(m.ctx, bson.M{})
	if err != nil {
		return err
	}

	var olddev device_0_1_0

	for cursor.Next(m.ctx) {
		if err = cursor.Decode(&olddev); err != nil {
			continue
		}
		// first prepare an auth set

		// reuse device ID as auth set ID
		asetId := olddev.Id

		aset := model.AuthSet{
			Id:        asetId,
			IdData:    olddev.IdData,
			PubKey:    olddev.PubKey,
			DeviceId:  olddev.Id,
			Status:    olddev.Status,
			Timestamp: &olddev.UpdatedTs,
		}

		if _, err := asColl.InsertOne(m.ctx, aset); err != nil {
			return errors.Wrapf(err, "failed to insert auth set for device %v",
				olddev.Id)
		}

		// update tokens

		filter := token_0_1_0{
			DevId: olddev.Id,
		}

		update := bson.M{
			"$set": bson.M{
				// see model.Token for field naming
				"auth_id": asetId,
			},
		}

		if _, err = tColl.UpdateMany(m.ctx, filter, update); err != nil {
			return errors.Wrapf(err, "failed to update tokens of device %v", olddev.Id)
		}
	}

	if err := cursor.Close(m.ctx); err != nil {
		return errors.Wrap(err, "failed to close DB iterator")
	}

	return nil
}

func (m *migration_1_1_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 1, 0)
}

func (m *migration_1_1_0) ensureIndexes(ctx context.Context) error {
	_false := false
	_true := true

	devIdDataUniqueIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: model.DevKeyIdData, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &indexDevices_IdentityData,
			Unique:     &_true,
		},
	}

	authSetUniqueIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: model.AuthSetKeyDeviceId, Value: 1},
			{Key: model.AuthSetKeyIdData, Value: 1},
			{Key: model.AuthSetKeyPubKey, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &indexAuthSet_DeviceId_IdentityData_PubKey,
			Unique:     &_true,
		},
	}

	cDevs := m.ms.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
	devIndexes := cDevs.Indexes()
	_, err := devIndexes.CreateOne(ctx, devIdDataUniqueIndex)
	if err != nil {
		return err
	}

	cAuthSets := m.ms.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbAuthSetColl)
	authSetIndexes := cAuthSets.Indexes()
	_, err = authSetIndexes.CreateOne(ctx, authSetUniqueIndex)

	return err
}
