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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func TestMigration_1_2_15_createCollectionReleases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_15_createCollectionReleases in short mode.")
	}

	db.Wipe()
	c := db.Client()

	ctx := context.TODO()

	//store := NewDataStoreMongoWithClient(c)
	database := c.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)
	collRel := database.Collection(CollectionReleases)

	inputImages := []interface{}{
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70f6",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "foo",
				DeviceTypesCompatible: []string{"foo"},
				Provides:              map[string]string{"rootfs-image.checksum": "bar"},
			},
		},
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70f7",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "foo",
				DeviceTypesCompatible: []string{"foo"},
				Provides:              map[string]string{"rootfs-image.checksum": "bar"},
			},
		},
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70f8",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "bar",
				DeviceTypesCompatible: []string{"foo"},
				Provides:              map[string]string{"rootfs-image.checksum": "bar"},
			},
		},
	}

	outputReleases := []model.Release{
		{
			Name: "foo",
			Artifacts: []model.Image{
				{
					Id: "0cb87b3d-4f08-420b-b004-4347c07f70f6",
					ArtifactMeta: &model.ArtifactMeta{
						Name:                  "foo",
						DeviceTypesCompatible: []string{"foo"},
						Provides:              map[string]string{"rootfs-image.checksum": "bar"},
						Depends:               map[string]interface{}{"device_type": bson.A{"foo"}},
					},
				},
				{
					Id: "0cb87b3d-4f08-420b-b004-4347c07f70f7",
					ArtifactMeta: &model.ArtifactMeta{
						Name:                  "foo",
						DeviceTypesCompatible: []string{"foo"},
						Provides:              map[string]string{"rootfs-image.checksum": "bar"},
						Depends:               map[string]interface{}{"device_type": bson.A{"foo"}},
					},
				},
			},
			ArtifactsCount: 2,
		},
		{
			Name: "bar",
			Artifacts: []model.Image{
				{
					Id: "0cb87b3d-4f08-420b-b004-4347c07f70f8",
					ArtifactMeta: &model.ArtifactMeta{
						Name:                  "bar",
						DeviceTypesCompatible: []string{"foo"},
						Provides:              map[string]string{"rootfs-image.checksum": "bar"},
						Depends:               map[string]interface{}{"device_type": bson.A{"foo"}},
					},
				},
			},
			ArtifactsCount: 1,
		},
	}
	// insert images
	_, err := collImg.InsertMany(ctx, inputImages)
	assert.NoError(t, err)

	// get releases
	// there should be no documents in the result
	releases := []model.Release{}
	cursor, err := collRel.Find(ctx, bson.M{})
	assert.NoError(t, err)
	err = cursor.All(ctx, &releases)
	assert.NoError(t, err)
	assert.Len(t, releases, 0)

	// apply migration (1.2.15)
	mnew := &migration_1_2_15{
		client: c,
		db:     DbName,
	}
	err = mnew.Up(migrate.MakeVersion(1, 2, 15))
	assert.NoError(t, err)

	// get release
	// this time the releas should be in the result
	cursor, err = collRel.Find(ctx, bson.M{})
	assert.NoError(t, err)
	err = cursor.All(ctx, &releases)
	assert.NoError(t, err)
	assert.Len(t, releases, 2)
	// ignore modification timestamp
	for i := range releases {
		releases[i].Modified = nil
	}
	assert.Equal(t, outputReleases, releases)
}

func TestMigration_1_2_15_indexReleaseTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_15_indexReleaseTags in short mode.")
	}

	db.Wipe()
	c := db.Client()

	ctx := context.TODO()

	//store := NewDataStoreMongoWithClient(c)
	database := c.Database(mstore.DbFromContext(ctx, DatabaseName))
	collRel := database.Collection(CollectionReleases)

	// apply migration (1.2.15)
	mnew := &migration_1_2_15{
		client: c,
		db:     DbName,
	}
	err := mnew.Up(migrate.MakeVersion(1, 2, 15))
	assert.NoError(t, err)

	indices := collRel.Indexes()
	exists, err := hasIndex(ctx, IndexNameReleaseTags, indices)
	assert.NoError(t, err)
	assert.True(t, exists, "index "+IndexNameReleaseTags+" must exist in 1.2.15")
}

func TestMigration_1_2_15_indexReleaseUpdateType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_15_indexReleaseUpdateType in short mode.")
	}

	db.Wipe()
	c := db.Client()

	ctx := context.TODO()

	//store := NewDataStoreMongoWithClient(c)
	database := c.Database(mstore.DbFromContext(ctx, DatabaseName))
	collRel := database.Collection(CollectionReleases)

	// apply migration (1.2.15)
	mnew := &migration_1_2_15{
		client: c,
		db:     DbName,
	}
	err := mnew.Up(migrate.MakeVersion(1, 2, 15))
	assert.NoError(t, err)

	indices := collRel.Indexes()
	exists, err := hasIndex(ctx, IndexNameReleaseUpdateTypes, indices)
	assert.NoError(t, err)
	assert.True(t, exists, "index "+IndexNameReleaseUpdateTypes+" must exist in 1.2.15")
}

func TestMigration_1_2_15_indexUpdateTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_15_indexUpdateTypes in short mode.")
	}

	db.Wipe()
	c := db.Client()

	ctx := context.TODO()

	//store := NewDataStoreMongoWithClient(c)
	database := c.Database(mstore.DbFromContext(ctx, DatabaseName))
	collRel := database.Collection(CollectionUpdateTypes)

	// apply migration (1.2.15)
	mnew := &migration_1_2_15{
		client: c,
		db:     DbName,
	}
	err := mnew.Up(migrate.MakeVersion(1, 2, 18))
	assert.NoError(t, err)

	indices := collRel.Indexes()
	exists, err := hasIndex(ctx, IndexNameAggregatedUpdateTypes, indices)
	assert.NoError(t, err)
	assert.True(t, exists, "index "+IndexNameAggregatedUpdateTypes+" must exist in 1.2.18")
}

func TestMigration_1_2_15_indexReleaseArtifactsCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_15_indexReleaseArtifactsCount in short mode.")
	}

	db.Wipe()
	c := db.Client()

	ctx := context.TODO()

	//store := NewDataStoreMongoWithClient(c)
	database := c.Database(mstore.DbFromContext(ctx, DatabaseName))
	collRel := database.Collection(CollectionReleases)

	artifactType := "app"
	inputImgs := []*model.Image{
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d80",
			ImageMeta: &model.ImageMeta{
				Description: "description",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App1 v1.0",
				DeviceTypesCompatible: []string{"foo"},
			},
			Modified: timePtr("2010-09-22T22:00:00+00:00"),
		},
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d81",
			ImageMeta: &model.ImageMeta{
				Description: "description",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App2 v0.1",
				DeviceTypesCompatible: []string{"foo"},
				Updates:               []model.Update{},
			},
			Modified: timePtr("2010-09-22T23:02:00+00:00"),
		},
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d82",
			ImageMeta: &model.ImageMeta{
				Description: "description",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App1 v1.0",
				DeviceTypesCompatible: []string{"bar, baz"},
				Updates:               []model.Update{},
			},
			Modified: timePtr("2010-09-22T22:00:01+00:00"),
		},
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d83",
			ImageMeta: &model.ImageMeta{
				Description: "description",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App1 v1.0",
				DeviceTypesCompatible: []string{"bork"},
				Updates:               []model.Update{},
			},
			Modified: timePtr("2010-09-22T22:00:04+00:00"),
		},
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d84",
			ImageMeta: &model.ImageMeta{
				Description: "extended description",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App2 v0.1",
				DeviceTypesCompatible: []string{"bar", "baz"},
				Updates:               []model.Update{},
			},
			Modified: timePtr("2010-09-22T23:00:00+00:00"),
		},
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d85",
			ImageMeta: &model.ImageMeta{
				Description: "description2",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App4 v2.0",
				DeviceTypesCompatible: []string{"foo2"},
				Updates: []model.Update{
					{
						TypeInfo: model.ArtifactUpdateTypeInfo{
							Type: &artifactType,
						},
					},
				},
			},
			Modified: timePtr("2023-09-22T22:00:00+00:00"),
		},
	}
	releases := []model.Release{
		{
			Name: "App1 v1.0",
			Artifacts: []model.Image{
				*inputImgs[0],
				*inputImgs[2],
				*inputImgs[3],
			},
		},
		{
			Name: "App2 v0.1",
			Artifacts: []model.Image{
				*inputImgs[1],
				*inputImgs[4],
			},
		},
		{
			Name: "App4 v2.0",
			Artifacts: []model.Image{
				*inputImgs[5],
			},
		},
	}
	items := make([]interface{}, len(releases))
	for i, _ := range releases {
		items[i] = releases[i]
	}
	r, err := collRel.InsertMany(ctx, items)
	assert.NotNil(t, r)
	assert.NoError(t, err)

	// apply migration (1.2.15)
	mnew := &migration_1_2_15{
		client: c,
		db:     DbName,
	}
	err = mnew.Up(migrate.MakeVersion(1, 2, 15))
	assert.NoError(t, err)

	indices := collRel.Indexes()
	exists, err := hasIndex(ctx, IndexNameReleaseArtifactsCount, indices)
	assert.NoError(t, err)
	assert.True(t, exists, "index "+IndexNameAggregatedUpdateTypes+" must exist in 1.2.15")

	cursor, err := collRel.Find(ctx, bson.M{})
	var releases010219 []model.Release
	err = cursor.All(ctx, &releases010219)
	assert.NoError(t, err)
	for _, r := range releases010219 {
		assert.Equal(t, len(r.Artifacts), r.ArtifactsCount)
	}
}
