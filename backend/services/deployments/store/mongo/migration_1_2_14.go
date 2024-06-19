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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type migration_1_2_14 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_2_14) Up(from migrate.Version) error {
	if m.db != DatabaseName {
		return nil
	}
	ctx := context.Background()
	idx := mongo.IndexModel{
		Keys: bson.D{{
			Key: "status", Value: 1,
		}, {
			Key: "expire", Value: 1,
		}},
		Options: options.Index().
			SetName("UploadExpire"),
	}

	_, err := m.client.Database(m.db).
		Collection(CollectionUploadIntents).
		Indexes().
		CreateOne(ctx, idx)
	return err
}

func (m *migration_1_2_14) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 14)
}
