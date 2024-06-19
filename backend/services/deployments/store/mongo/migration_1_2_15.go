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
	"time"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type migration_1_2_15 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_2_15) Up(from migrate.Version) error {
	err := m.createCollectionReleases()
	if err == nil {
		err = m.indexReleaseTags()
	}
	if err == nil {
		err = m.indexReleaseUpdateType()
	}
	if err == nil {
		err = m.indexUpdateTypes()
	}
	if err == nil {
		err = m.indexReleaseArtifactsCount()
	}
	return err
}

func (m *migration_1_2_15) createCollectionReleases() error {
	ctx := context.Background()
	c := m.client.Database(m.db).Collection(CollectionImages)
	cr := m.client.Database(m.db).Collection(CollectionReleases)

	cursor, err := c.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var a model.Image
		err = cursor.Decode(&a)
		if err != nil {
			break
		}

		opt := &mopts.UpdateOptions{}
		upsert := true
		opt.Upsert = &upsert
		update := bson.M{
			"$set": bson.M{
				StorageKeyReleaseName:     a.ArtifactMeta.Name,
				StorageKeyReleaseModified: time.Now(),
			},
			"$push": bson.M{StorageKeyReleaseArtifacts: a},
		}

		res, err := cr.UpdateOne(
			ctx,
			bson.M{
				"$and": []bson.M{
					{
						StorageKeyReleaseName: a.ArtifactMeta.Name,
					},
					{
						StorageKeyReleaseArtifactsId: bson.M{
							"$ne": a.Id,
						},
					},
				}},
			update,
			opt,
		)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			return err
		}

		if res.ModifiedCount != 1 && res.UpsertedCount != 1 {
			return errors.New(fmt.Sprintf("failed to update release %s", a.ArtifactMeta.Name))
		}
	}
	if err = cursor.Err(); err != nil {
		return errors.WithMessage(err, "failed to decode artifact")
	}
	return nil
}

func (m *migration_1_2_15) indexReleaseArtifactsCount() error {
	ctx := context.Background()
	idxReleases := m.client.
		Database(m.db).
		Collection(CollectionReleases).
		Indexes()

	_, err := idxReleases.CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{
				Key:   StorageKeyReleaseArtifactsCount,
				Value: 1,
			},
		},
		Options: mopts.Index().SetName(IndexNameReleaseArtifactsCount),
	})
	if err != nil {
		return fmt.Errorf("mongo(1.2.15): failed to create index: %w", err)
	}

	_, err = m.client.
		Database(m.db).
		Collection(CollectionReleases).
		UpdateMany(
			ctx,
			bson.M{},
			[]bson.M{{
				"$set": bson.M{
					"artifacts_count": bson.M{
						"$size": bson.M{
							"$ifNull": bson.A{
								"$artifacts",
								bson.A{},
							}},
					},
				},
			}},
		)
	if err != nil {
		return fmt.Errorf("mongo(1.2.15): failed to update adrtifact counts: %w", err)
	}

	return nil
}

func (m *migration_1_2_15) indexReleaseTags() error {
	ctx := context.Background()
	idxReleases := m.client.
		Database(m.db).
		Collection(CollectionReleases).
		Indexes()

	_, err := idxReleases.CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{
			Key:   StorageKeyReleaseTags,
			Value: 1,
		}, {
			// Sort by modified date by default when querying by tags
			Key:   StorageKeyReleaseModified,
			Value: -1,
		}},
		Options: mopts.Index().
			SetName(IndexNameReleaseTags),
	})
	if err != nil {
		return fmt.Errorf("mongo(1.2.15): failed to create index: %w", err)
	}
	return nil
}

func (m *migration_1_2_15) indexReleaseUpdateType() error {
	ctx := context.Background()
	idxReleases := m.client.
		Database(m.db).
		Collection(CollectionReleases).
		Indexes()

	_, err := idxReleases.CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{
				Key:   StorageKeyReleaseArtifactsUpdateTypes,
				Value: 1,
			},
		},
		Options: mopts.Index().
			SetName(IndexNameReleaseUpdateTypes).
			SetSparse(true),
	})
	if err != nil {
		return fmt.Errorf("mongo(1.2.15): failed to create index: %w", err)
	}
	return nil
}

func (m *migration_1_2_15) indexUpdateTypes() error {
	ctx := context.Background()
	idxReleases := m.client.
		Database(m.db).
		Collection(CollectionUpdateTypes).
		Indexes()

	_, err := idxReleases.CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{
				Key:   StorageKeyTenantId,
				Value: 1,
			},
		},
		Options: mopts.Index().SetName(IndexNameAggregatedUpdateTypes).SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("mongo(1.2.15): failed to create index: %w", err)
	}
	return nil
}

func (m *migration_1_2_15) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 15)
}
