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

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
)

var (
	ErrExpirationTooLong = errors.New("token expiration time to high")
	ErrExpirationVoid    = errors.New("token expiration time cannot be 0")
	ErrExpNotPresent     = errors.New("token expiration not present in context")
)

type migration_1_0_0 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_0_0) Up(from migrate.Version) error {
	dbName := ctxstore.DbFromContext(m.ctx, DbName)
	db := m.ds.client.Database(dbName)
	collTokens := db.Collection(DbTokensColl)

	// Drop existing tokens - makes users logged out.
	err := collTokens.Drop(m.ctx)
	if err != nil {
		return err
	}
	// Recreate token collection (and unique _id index)
	collTokens = db.Collection(DbTokensColl)

	idxOpts := mopts.Index()
	// Expire document at exp timestamp.
	idxOpts.SetExpireAfterSeconds(0)
	idxOpts.SetName("TokenExpiration")
	idxModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "exp.time", Value: 1}},
		Options: idxOpts,
	}
	idxView := collTokens.Indexes()
	_, err = idxView.CreateOne(m.ctx, idxModel)
	return err
}

func (m *migration_1_0_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 0, 0)
}
