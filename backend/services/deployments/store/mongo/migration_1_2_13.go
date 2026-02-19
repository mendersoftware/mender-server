// Copyright 2023 Northern.tech AS
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
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type migration_1_2_13 struct {
	client *mongo.Client
	db     string
}

// Up intrduces special representation of artifact provides and index them.
func (m *migration_1_2_13) Up(from migrate.Version) error {
	ctx := context.Background()
	c := m.client.Database(m.db).Collection(CollectionImages)

	query := bson.M{
		StorageKeyImageProvides:    bson.M{"$exists": true},
		StorageKeyImageProvidesIdx: bson.M{"$exists": false},
	}
	cursor, err := c.Find(ctx, query)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	var a model.Image
	for cursor.Next(ctx) {
		err = cursor.Decode(&a)
		if err != nil {
			break
		}

		providesIdx := model.ProvidesIdx(a.ArtifactMeta.Provides)

		up := bson.M{
			"$set": bson.M{
				StorageKeyImageProvidesIdx: providesIdx,
			},
		}

		res, err := c.UpdateOne(ctx, bson.M{"_id": a.Id}, up)
		if err != nil {
			return err
		}

		if res.MatchedCount != 1 {
			return errors.New(fmt.Sprintf("failed to update artifact %s: not found", a.Id))
		}
	}
	if err = cursor.Err(); err != nil {
		return errors.WithMessage(err, "failed to decode artifact")
	}
	// create index for artifact provides
	storage := NewDataStoreMongoWithClient(m.client)
	return storage.EnsureIndexes(m.db,
		CollectionImages,
		IndexArtifactProvides)
}

func (m *migration_1_2_13) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 13)
}
