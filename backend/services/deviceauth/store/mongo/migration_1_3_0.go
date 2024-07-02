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
	"crypto/rsa"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

type migration_1_3_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_3_0) Up(from migrate.Version) error {
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
		newKey, err := normalizeKey(set.PubKey)

		if err != nil {
			return errors.Wrapf(
				err,
				"failed to normalize key of auth set %v: %v",
				set.Id,
				set.PubKey,
			)
		}

		update := bson.M{
			"$set": model.AuthSetUpdate{
				PubKey: newKey,
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

	return nil
}

func (m *migration_1_3_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 3, 0)
}

func normalizeKey(key string) (string, error) {
	parsed, err := utils.ParsePubKey(key)
	if err != nil {
		return "", err
	}

	keyStruct, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("cannot decode public key")
	}

	serialized, err := utils.SerializePubKey(keyStruct)
	if err != nil {
		return "", err
	}

	return serialized, nil
}
