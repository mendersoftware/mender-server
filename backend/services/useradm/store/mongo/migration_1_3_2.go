// Copyright 2022 Northern.tech AS
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
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

type migration_1_3_2 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

func (m *migration_1_3_2) Up(from migrate.Version) error {
	c := m.ds.client.Database(mstore.DbFromContext(m.ctx, DbName)).
		Collection(DbTokensColl)

	idxTokens := c.Indexes()

	// create new index with predictable name
	tokenNameIndexOptions := mopts.Index()
	tokenNameIndexOptions.SetUnique(true)
	tokenNameIndexOptions.SetPartialFilterExpression(
		bson.M{
			DbTokenName: bson.M{"$exists": true},
		})
	tokenNameIndexOptions.SetName(DbUniqueTokenNameIndexName)

	uniqueTokenNameIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: DbTokenName, Value: 1},
			{Key: DbTokenSubject, Value: 1},
		},
		Options: tokenNameIndexOptions,
	}

	tokenSubjectIndexOptions := mopts.Index()
	tokenSubjectIndexOptions.SetName(DbTokenSubjectIndexName)

	tokenSubjectIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: DbTokenSubject, Value: 1}},
		Options: tokenSubjectIndexOptions,
	}

	_, err := idxTokens.CreateMany(
		m.ctx, []mongo.IndexModel{uniqueTokenNameIndex, tokenSubjectIndex})

	return err
}

func (m *migration_1_3_2) Version() migrate.Version {
	return migrate.MakeVersion(1, 3, 2)
}
