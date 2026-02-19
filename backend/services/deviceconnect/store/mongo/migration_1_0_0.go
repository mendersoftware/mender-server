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

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

const (
	IndexNameLogsExpire = "RecordingExpire"
)

type migration_1_0_0 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_0_0) Up(from migrate.Version) error {
	ctx := context.Background()
	indexModels := []mongo.IndexModel{
		{
			// Index for expiring old events
			Keys: bson.D{{Key: "expire_ts", Value: 1}},
			Options: mopts.Index().
				SetExpireAfterSeconds(0).
				SetName(IndexNameLogsExpire),
		},
	}
	collLogs := m.client.Database(m.db).Collection(RecordingsCollectionName)
	idxView := collLogs.Indexes()

	_, err := idxView.CreateMany(ctx, indexModels)
	return err
}

func (m *migration_1_0_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 0, 0)
}
