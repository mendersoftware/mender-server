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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func TestArtifactUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestArtifactUpdate in short mode.")
	}

	//image dataset - common for all cases
	img := &model.Image{
		Id: "a3719bc6-62af-4d65-b781-effa992048ba",
		ImageMeta: &model.ImageMeta{
			Description: "description",
		},

		ArtifactMeta: &model.ArtifactMeta{
			Name:                  "app1-v1.0",
			DeviceTypesCompatible: []string{"foo", "bar"},
			Updates:               []model.Update{},
		},
	}

	//setup db - common for all cases
	ctx := context.Background()
	db.Wipe()
	client := db.Client()

	store := NewDataStoreMongoWithClient(client)
	store.AddArtifactToRelease(ctx, *img)

	img.ImageMeta.Description = "updated description"
	done, err := store.Update(ctx, img)
	assert.NoError(t, err)
	assert.True(t, done)

	release, err := store.GetRelease(ctx, img.Name)
	require.NoError(t, err)
	require.Len(t, release.Artifacts, 1)
	imgFromDB := release.Artifacts[0]
	assert.NoError(t, err)
	assert.Equal(t, img.ImageMeta.Description, imgFromDB.ImageMeta.Description)
}

func makeTestImages(t *testing.T) (*DataStoreMongo, []*model.Image) {
	inputImgs := []*model.Image{
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d80",
			ImageMeta: &model.ImageMeta{
				Description: "description",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App1 v1.0",
				DeviceTypesCompatible: []string{"foo"},
				Updates:               []model.Update{},
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
			Modified: timePtr("2010-09-22T22:01:00+00:00"),
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
			Modified: timePtr("2010-09-22T22:02:00+00:00"),
		},
		{
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d83",
			ImageMeta: &model.ImageMeta{
				Description: `description*`,
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "App1 v1.0",
				DeviceTypesCompatible: []string{"bork"},
				Updates:               []model.Update{},
			},
			Modified: timePtr("2010-09-22T22:03:00+00:00"),
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
			Modified: timePtr("2010-09-22T22:04:00+00:00"),
		}, {
			Id: "6d4f6e27-c3bb-438c-ad9c-d9de30e59d85",
			ImageMeta: &model.ImageMeta{
				Description: "foo",
			},

			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "Not an app",
				DeviceTypesCompatible: []string{"bork1"},
				Updates:               []model.Update{},
			},
			Modified: timePtr("2010-09-22T22:05:00+00:00"),
		},
	}

	// setup test context
	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())
	for _, img := range inputImgs {
		require.NoError(t, ds.AddArtifactToRelease(ctx, *img))

		// Convert Depends["device_type"] to bson.A for the sake of
		// simplifying test case definitions.
		img.ArtifactMeta.Depends = make(map[string][]string)
		img.ArtifactMeta.Depends["device_type"] = img.ArtifactMeta.DeviceTypesCompatible
	}
	return ds, inputImgs
}

func TestListImages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestListImages in short mode.")
	}

	// Make sure we start test with empty database
	db.Wipe()
	ctx := context.Background()
	ds, inputImgs := makeTestImages(t)

	testCases := map[string]struct {
		filter *model.ReleaseOrImageFilter

		images      []*model.Image
		imagesCount int
		err         error
	}{
		"ok, all": {
			filter: &model.ReleaseOrImageFilter{
				Name: "App",
			},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[2],
				inputImgs[3],
				inputImgs[1],
				inputImgs[4],
				inputImgs[5],
			},
			imagesCount: 6,
		},
		"ok, description": {
			filter: &model.ReleaseOrImageFilter{
				Description: "description",
			},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[2],
				inputImgs[3],
				inputImgs[1],
				inputImgs[4],
			},
			imagesCount: 5,
		},
		"ok, description extended": {
			filter: &model.ReleaseOrImageFilter{
				Description: "extended description",
			},
			images: []*model.Image{
				inputImgs[4],
			},
			imagesCount: 1,
		},
		"ok, sort by modified asc": {
			filter: &model.ReleaseOrImageFilter{
				Sort: "modified:asc",
			},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[1],
				inputImgs[2],
				inputImgs[3],
				inputImgs[4],
				inputImgs[5],
			},
			imagesCount: 6,
		},
		"ok, sort by modified desc": {
			filter: &model.ReleaseOrImageFilter{
				Sort: "modified:desc",
			},
			images: []*model.Image{
				inputImgs[5],
				inputImgs[4],
				inputImgs[3],
				inputImgs[2],
				inputImgs[1],
				inputImgs[0],
			},
			imagesCount: 6,
		},
		"ok, by device type": {
			filter: &model.ReleaseOrImageFilter{
				DeviceType: "bork",
			},
			images: []*model.Image{
				inputImgs[3],
				inputImgs[5],
			},
			imagesCount: 2,
		},
		"ok, by name": {
			filter: &model.ReleaseOrImageFilter{
				Name: "App2 v0.1",
			},
			images: []*model.Image{
				inputImgs[1],
				inputImgs[4],
			},
			imagesCount: 2,
		},
		"ok, by name desc, page 1": {
			filter: &model.ReleaseOrImageFilter{
				Name:    "App2 v0.1",
				Sort:    "name:desc",
				Page:    1,
				PerPage: 1,
			},
			images: []*model.Image{
				inputImgs[1],
			},
			imagesCount: 2,
		},
		"ok, by name desc, page 2": {
			filter: &model.ReleaseOrImageFilter{
				Name:    "App2 v0.1",
				Sort:    "name:desc",
				Page:    2,
				PerPage: 1,
			},
			images: []*model.Image{
				inputImgs[4],
			},
			imagesCount: 2,
		},
		"ok, not found": {
			filter: &model.ReleaseOrImageFilter{
				Name: "App3 v1.0",
			},
			images:      nil,
			imagesCount: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			images, count, err := ds.ListImages(ctx, tc.filter)

			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			if !assert.Equal(t, tc.images, images) {
				t.Log(*tc.images[0])
				t.Log(*images[0])
			}
			assert.Equal(t, tc.imagesCount, count)
		})
	}
}

func TestListV2Images(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestListImages in short mode.")
	}

	// Make sure we start test with empty database
	db.Wipe()
	ctx := context.Background()
	ds, inputImgs := makeTestImages(t)

	testCases := map[string]struct {
		filter *model.ImageFilter

		images      []*model.Image
		imagesCount int
		err         error
	}{
		"ok, all": {
			filter: &model.ImageFilter{},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[2],
				inputImgs[3],
				inputImgs[1],
				inputImgs[4],
				inputImgs[5],
			},
			imagesCount: 6,
		},
		"ok, description": {
			filter: &model.ImageFilter{
				Description: "extended description",
			},
			images: []*model.Image{
				inputImgs[4],
			},
			imagesCount: 1,
		},
		"ok, by device type": {
			filter: &model.ImageFilter{
				DeviceType: "bork",
			},
			images: []*model.Image{
				inputImgs[3],
			},
			imagesCount: 1,
		},
		"ok, by exact name": {
			filter: &model.ImageFilter{
				ExactNames: []string{"App1 v1.0"},
			},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[2],
				inputImgs[3],
			},
			imagesCount: 3,
		},
		"ok, by multiple exact names": {
			filter: &model.ImageFilter{
				ExactNames: []string{"App1 v1.0", "App2 v0.1"},
			},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[2],
				inputImgs[3],
				inputImgs[1],
				inputImgs[4],
			},
			imagesCount: 5,
		},
		"ok, by name prefix": {
			filter: &model.ImageFilter{
				NamePrefixes: []string{"App1"},
			},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[2],
				inputImgs[3],
			},
			imagesCount: 3,
		},
		"ok, by prefix description": {
			filter: &model.ImageFilter{
				Description: "des*",
			},
			images: []*model.Image{
				inputImgs[0],
				inputImgs[2],
				inputImgs[3],
				inputImgs[1],
			},
			imagesCount: 4,
		},
		"ok, by prefix decvice type ": {
			filter: &model.ImageFilter{
				DeviceType: "bork*",
			},
			images: []*model.Image{
				inputImgs[3],
				inputImgs[5],
			},
			imagesCount: 2,
		},
		"ok, by name, page 2": {
			filter: &model.ImageFilter{
				ExactNames: []string{"App2 v0.1"},
				Page:       2,
				PerPage:    1,
			},
			images: []*model.Image{
				inputImgs[4],
			},
			imagesCount: 1,
		},
		"ok, not found": {
			filter: &model.ImageFilter{
				ExactNames: []string{"App3 v1.0"},
			},
			images:      nil,
			imagesCount: 0,
		},
		"ok, by description with escape character": {
			filter: &model.ImageFilter{
				Description: `description\*`,
			},
			images: []*model.Image{
				inputImgs[3],
			},
			imagesCount: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			images, err := ds.ListImagesV2(ctx, tc.filter)

			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.images, images)
			assert.Equal(t, tc.imagesCount, len(images))
		})
	}
}
