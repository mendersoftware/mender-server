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

type migration1_0_0 struct {
	client *mongo.Client
	db     string
}

// Up creates the jobs capped collection
func (m *migration1_0_0) Up(from migrate.Version) error {
	ctx := context.Background()
	database := m.client.Database(m.db)
	collJobs := database.Collection(JobsCollectionName)
	idxJobs := collJobs.Indexes()
	indexOptions := mopts.Index()
	indexOptions.SetName("status")
	statusIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: 1}},
		Options: indexOptions,
	}
	if _, err := idxJobs.CreateOne(ctx, statusIndex); err != nil {
		return err
	}

	indexOptions.SetName("workflow_name")
	nameIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "workflow_name", Value: 1}},
		Options: indexOptions,
	}
	_, err := idxJobs.CreateOne(ctx, nameIndex)

	return err
}

func (m *migration1_0_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 0, 0)
}
