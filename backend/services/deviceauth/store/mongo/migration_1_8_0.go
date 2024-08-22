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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"
)

type migration_1_8_0 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_8_0) Up(from migrate.Version) error {
	database := m.ds.client.Database(ctxstore.DbFromContext(m.ctx, DbName))
	collTokens := database.Collection(DbTokensColl)
	err := collTokens.Drop(m.ctx)
	if err != nil {
		return err
	}
	// Recreate tokens collection
	collTokens = database.Collection(DbTokensColl)
	idxView := collTokens.Indexes()
	idxOpts := mopts.Index()
	idxOpts.SetExpireAfterSeconds(0)
	idxOpts.SetName("TokenExpiration")
	idxModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "exp.time", Value: 1}},
		Options: idxOpts,
	}
	_, err = idxView.CreateOne(m.ctx, idxModel)
	return err
}

func (m *migration_1_8_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 8, 0)
}
