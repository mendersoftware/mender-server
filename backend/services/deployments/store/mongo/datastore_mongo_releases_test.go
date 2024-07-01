// Copyright 2024 Northern.tech AS
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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/pkg/identity"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
)

func TestGetReleases_1_2_14(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetReleases_1_2_14 in short mode.")
	}
	db.Wipe()

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
	}

	// Setup test context
	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())
	for _, img := range inputImgs {
		err := ds.InsertImage(ctx, img)
		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t,
				"error setting up image collection for testing")
		}

		// Convert Depends["device_type"] to bson.A for the sake of
		// simplifying test case definitions.
		img.ArtifactMeta.Depends = make(map[string]interface{})
		img.ArtifactMeta.Depends["device_type"] = make(bson.A,
			len(img.ArtifactMeta.DeviceTypesCompatible),
		)
		for i, devType := range img.ArtifactMeta.DeviceTypesCompatible {
			img.ArtifactMeta.Depends["device_type"].(bson.A)[i] = devType
		}
	}

	testCases := map[string]struct {
		releaseFilt *model.ReleaseOrImageFilter

		releases []model.Release
		err      error
	}{
		"ok, all": {
			releases: []model.Release{
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
			},
		},
		"ok, description partial": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Description: "description",
			},
			releases: []model.Release{
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
			},
		},
		"ok, description exact": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Description: "extended description",
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
				},
			},
		},
		"ok, sort by modified asc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "modified:asc",
			},
			releases: []model.Release{
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
			},
		},
		"ok, sort by modified desc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "modified:desc",
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
				},
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
				},
			},
		},
		"ok, with sort and pagination": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort:    "name:desc",
				Page:    2,
				PerPage: 1,
			},
			releases: []model.Release{
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
				},
			},
		},
		"ok, by name": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Name: "App2 v0.1",
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
				},
			},
		},
		"ok, not found": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Name: "App3 v1.0",
			},
			releases: []model.Release{},
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {
			releases, count, err := ds.getReleases_1_2_14(ctx, tc.releaseFilt)

			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			//ignore modification timestamp
			for i := range releases {
				releases[i].Modified = nil
			}
			assert.Equal(t, tc.releases, releases)
			assert.GreaterOrEqual(t, count, len(tc.releases))
		})
	}
}

func TestGetReleases_1_2_15(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetReleases_1_2_15 in short mode.")
	}
	db.Wipe()

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

	releaseNameToTags := make(map[string]model.Tags, 8)
	releaseNameToTags["App4 v2.0"] = model.Tags{
		"demo",
	}
	releaseNameToTags["App1 v1.0"] = model.Tags{
		"production",
	}
	releaseNameToTags["App2 v0.1"] = model.Tags{
		"root-fs",
	}
	// Setup test context
	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())
	for _, img := range inputImgs {
		err := ds.InsertImage(ctx, img)
		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t,
				"error setting up image collection for testing")
		}
		err = ds.UpdateReleaseArtifacts(ctx, img, nil, img.ArtifactMeta.Name)
		assert.NoError(t, err)

		// Convert Depends["device_type"] to bson.A for the sake of
		// simplifying test case definitions.
		img.ArtifactMeta.Depends = make(map[string]interface{})
		img.ArtifactMeta.Depends["device_type"] = make(bson.A,
			len(img.ArtifactMeta.DeviceTypesCompatible),
		)
		for i, devType := range img.ArtifactMeta.DeviceTypesCompatible {
			img.ArtifactMeta.Depends["device_type"].(bson.A)[i] = devType
		}
		time.Sleep(time.Millisecond * 10)
	}

	testCases := map[string]struct {
		releaseFilt *model.ReleaseOrImageFilter

		releases []model.Release
		err      error
	}{
		"ok, all": {
			releases: []model.Release{
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
				},
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
				},
			},
		},
		"ok, description partial": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Description: "description",
			},
			releases: []model.Release{
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
				},
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
				},
			},
		},
		"ok, description exact": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Description: "extended description",
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
			},
		},
		"ok, tag": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Tags: []string{"root-fs"},
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
					Tags:           releaseNameToTags["App2 v0.1"],
				},
			},
		},
		"ok, tags": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Tags: []string{"production", "demo"},
			},
			releases: []model.Release{
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
					Tags:           releaseNameToTags["App1 v1.0"],
				},
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
					Tags:           releaseNameToTags["App4 v2.0"],
				},
			},
		},
		"ok, sort by modified asc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "modified:asc",
			},
			releases: []model.Release{
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
				},
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
				},
			},
		},
		"ok, sort by modified desc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "modified:desc",
			},
			releases: []model.Release{
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
				},
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
				},
			},
		},
		"ok, sort by artifacts count asc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "artifacts_count:asc",
			},
			releases: []model.Release{
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
				},
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
				},
			},
		},
		"ok, sort by artifacts count desc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "artifacts_count:desc",
			},
			releases: []model.Release{
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
				},
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
				},
			},
		},
		"ok, sort by tags asc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "tags:asc",
			},
			releases: []model.Release{
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
					Tags:           releaseNameToTags["App4 v2.0"],
				},
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
					Tags:           releaseNameToTags["App1 v1.0"],
				},
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
					Tags:           releaseNameToTags["App2 v0.1"],
				},
			},
		},
		"ok, sort by tags desc": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort: "tags:desc",
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
					Tags:           releaseNameToTags["App2 v0.1"],
				},
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
					Tags:           releaseNameToTags["App1 v1.0"],
				},
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
					Tags:           releaseNameToTags["App4 v2.0"],
				},
			},
		},
		"ok, device type": {
			releaseFilt: &model.ReleaseOrImageFilter{
				DeviceType: "bork",
			},
			releases: []model.Release{
				{
					Name: "App1 v1.0",
					Artifacts: []model.Image{
						*inputImgs[0],
						*inputImgs[2],
						*inputImgs[3],
					},
					ArtifactsCount: 3,
				},
			},
		},
		"ok, with sort and pagination": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Sort:    "name:desc",
				Page:    2,
				PerPage: 1,
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
			},
		},
		"ok, by name": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Name: "App2 v0.1",
			},
			releases: []model.Release{
				{
					Name: "App2 v0.1",
					Artifacts: []model.Image{
						*inputImgs[1],
						*inputImgs[4],
					},
					ArtifactsCount: 2,
				},
			},
		},
		"ok, by update_type": {
			releaseFilt: &model.ReleaseOrImageFilter{
				UpdateType: artifactType,
			},
			releases: []model.Release{
				{
					Name: "App4 v2.0",
					Artifacts: []model.Image{
						*inputImgs[5],
					},
					ArtifactsCount: 1,
				},
			},
		},
		"ok, not found": {
			releaseFilt: &model.ReleaseOrImageFilter{
				Name: "App3 v1.0",
			},
			releases: []model.Release{},
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {
			for _, r := range tc.releases {
				ds.ReplaceReleaseTags(ctx, r.Name, r.Tags)
			}
			releases, count, err := ds.getReleases_1_2_15(ctx, tc.releaseFilt)

			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			//ignore modification timestamp
			for i := range releases {
				releases[i].Modified = nil
			}
			assert.Equal(t, tc.releases, releases)
			assert.GreaterOrEqual(t, count, len(tc.releases))

			countsByName := make(map[string]int, len(inputImgs))
			for _, img := range inputImgs {
				if v, ok := countsByName[img.ArtifactMeta.Name]; ok {
					countsByName[img.ArtifactMeta.Name] = v + 1
				} else {
					countsByName[img.ArtifactMeta.Name] = 1
				}
			}
			for _, img := range inputImgs {
				err = ds.UpdateReleaseArtifacts(ctx, nil, img, img.ArtifactMeta.Name)
				countsByName[img.ArtifactMeta.Name] = countsByName[img.ArtifactMeta.Name] - 1
				if countsByName[img.ArtifactMeta.Name] > 0 {
					_, count, _ = ds.getReleases_1_2_15(ctx, &model.ReleaseOrImageFilter{
						Name: img.ArtifactMeta.Name,
					})
					assert.True(t, count > 0, "before the removal of the last artifact the release should still exist")
				}
				assert.NoError(t, err)
			}
			_, count, _ = ds.getReleases_1_2_15(ctx, tc.releaseFilt)
			assert.Equal(t, count, 0, "after the removal of the last artifact the release should vanish")
			// we have to re-add the images for the rest of the tests
			for _, img := range inputImgs {
				err = ds.UpdateReleaseArtifacts(ctx, img, nil, img.ArtifactMeta.Name)
				assert.NoError(t, err)

				// Convert Depends["device_type"] to bson.A for the sake of
				// simplifying test case definitions.
				img.ArtifactMeta.Depends = make(map[string]interface{})
				img.ArtifactMeta.Depends["device_type"] = make(bson.A,
					len(img.ArtifactMeta.DeviceTypesCompatible),
				)
				for i, devType := range img.ArtifactMeta.DeviceTypesCompatible {
					img.ArtifactMeta.Depends["device_type"].(bson.A)[i] = devType
				}
				time.Sleep(time.Millisecond * 10)
			}
		})
	}
}

func TestReplaceReleaseTags(t *testing.T) {
	ctx := context.Background()
	client := db.Client()
	db.Wipe()

	type testCase struct {
		Name string

		context.Context

		Init        func(t *testing.T, self *testCase)
		ReleaseName string

		Tags model.Tags

		assert.ErrorAssertionFunc
	}

	testCases := []testCase{{
		Name: "ok",

		Context: context.Background(),

		Init: func(t *testing.T, self *testCase) {
			t.Helper()
			_, err := client.Database(DbName).
				Collection(CollectionReleases).
				InsertMany(ctx, []interface{}{model.Release{
					Name: self.ReleaseName,
					Tags: model.Tags{"bar", "foo"},
				}, model.Release{
					Name: "v100.2.3",
					Tags: model.Tags{"bar", "baz"},
				}})
			if err != nil {
				t.Errorf("failed to initialize dataset: %s", err)
				t.FailNow()
			}
		},
		Tags: func() model.Tags {
			newTags := make(model.Tags, model.TagsMaxUnique/2)
			for i := range newTags {
				newTags[i] = model.Tag("field" + strconv.Itoa(i))
			}
			return newTags
		}(),
		ReleaseName: "v1.0",
	}, {
		Name: "clear tags",

		Context: identity.WithContext(context.Background(),
			&identity.Identity{
				Tenant: "111111111111111111111111",
			},
		),

		Init: func(t *testing.T, self *testCase) {
			t.Helper()
			_, err := client.Database(ctxstore.DbFromContext(self, DbName)).
				Collection(CollectionReleases).
				InsertMany(self,
					[]interface{}{model.Release{
						Name: self.ReleaseName,
						Tags: func() model.Tags {
							newTags := make(model.Tags, model.TagsMaxUnique)
							for i := range newTags {
								newTags[i] = model.Tag(
									"field" + strconv.Itoa(i),
								)
							}
							return newTags
						}(),
					}, model.Release{
						Name: "v1.2.3-beta",
						Tags: model.Tags{"bar", "foo"},
					}})
			if err != nil {
				t.Errorf("failed to initialize dataset: %s", err)
				t.FailNow()
			}
		},
		Tags:        model.Tags{},
		ReleaseName: "v1.0",
	}, {
		Name: "error/too many tags",

		Context: identity.WithContext(context.Background(),
			&identity.Identity{
				Tenant: "222222222222222222222222",
			},
		),

		Init: func(t *testing.T, self *testCase) {
			t.Helper()
			_, err := client.Database(ctxstore.DbFromContext(self, DbName)).
				Collection(CollectionReleases).
				InsertMany(self,
					[]interface{}{model.Release{
						Name: self.ReleaseName,
						Tags: func() model.Tags {
							newTags := make(model.Tags, model.TagsMaxUnique)
							for i := range newTags {
								newTags[i] = model.Tag(
									"field" + strconv.Itoa(i),
								)
							}
							return newTags
						}(),
					}, model.Release{
						Name: "v1.2.3-beta",
						Tags: model.Tags{"bar", "foo"},
					}})
			if err != nil {
				t.Errorf("failed to initialize dataset: %s", err)
				t.FailNow()
			}
		},
		Tags:        model.Tags{"oneFieldTooMany"},
		ReleaseName: "v1.0",
		ErrorAssertionFunc: func(t assert.TestingT, err error, vargs ...interface{}) bool {
			return assert.ErrorIs(t, err, model.ErrTooManyUniqueTags)
		},
	}, {
		Name: "error/too many tags in input",

		Context: identity.WithContext(context.Background(),
			&identity.Identity{
				Tenant: "333333333333333333333333",
			},
		),

		Init: func(t *testing.T, self *testCase) {
			t.Helper()
			_, err := client.Database(ctxstore.DbFromContext(self, DbName)).
				Collection(CollectionReleases).
				InsertMany(self,
					[]interface{}{model.Release{
						Name: self.ReleaseName,
						Tags: model.Tags{},
					}, model.Release{
						Name: "v1.2.3-beta",
						Tags: model.Tags{"bar", "foo"},
					}})
			if err != nil {
				t.Errorf("failed to initialize dataset: %s", err)
				t.FailNow()
			}
		},
		Tags: func() model.Tags {
			newTags := make(model.Tags, model.TagsMaxUnique+1)
			for i := range newTags {
				newTags[i] = model.Tag(
					"field" + strconv.Itoa(i),
				)
			}
			return newTags
		}(),
		ReleaseName: "v1.0",
		ErrorAssertionFunc: func(t assert.TestingT, err error, vargs ...interface{}) bool {
			return assert.ErrorIs(t, err, model.ErrTooManyUniqueTags)
		},
	}, {
		Name: "error/no documents",

		Context: context.Background(),
		Init:    func(t *testing.T, self *testCase) {},

		Tags:        model.Tags{},
		ReleaseName: "not_found",
		ErrorAssertionFunc: func(t assert.TestingT, err error, vargs ...interface{}) bool {
			return assert.ErrorIs(t, err, store.ErrNotFound)
		},
	}, {
		Name: "error/aggregate/decode error",

		Context: identity.WithContext(ctx, &identity.Identity{
			Tenant: "deadbeefdeadbeefdeadbeef",
		}),
		Init: func(t *testing.T, self *testCase) {
			t.Helper()
			_, err := client.Database(ctxstore.DbFromContext(self, DbName)).
				Collection(CollectionReleases).
				InsertMany(self,
					[]interface{}{model.Release{
						Name: self.ReleaseName,
						Tags: model.Tags{},
					}, map[string]interface{}{
						StorageKeyReleaseName: "v1.2.3-beta",
						StorageKeyReleaseTags: []map[string]interface{}{{
							"key":   123,
							"value": true,
						}},
					}})
			if err != nil {
				t.Errorf("failed to initialize dataset: %s", err)
				t.FailNow()
			}
		},

		Tags:        model.Tags{"bar", "foo"},
		ReleaseName: "v1.0",
	}, {
		Name: "error/aggregate context cancelled",

		Context: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		Init: func(t *testing.T, self *testCase) {},

		Tags:        model.Tags{"oneMore", "tag"},
		ReleaseName: "v1.0",
		ErrorAssertionFunc: func(t assert.TestingT, err error, vargs ...interface{}) bool {
			return assert.ErrorIs(t, err, context.Canceled)
		},
	}, {
		Name: "error/update context cancelled",

		Context: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		Init: func(t *testing.T, self *testCase) {},

		Tags:        model.Tags{},
		ReleaseName: "v1.0",
		ErrorAssertionFunc: func(t assert.TestingT, err error, vargs ...interface{}) bool {
			return assert.ErrorIs(t, err, context.Canceled)
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var tenantID string
			if id := identity.FromContext(tc.Context); id != nil {
				tenantID = id.Tenant
			}
			err := MigrateSingle(ctx,
				ctxstore.DbNameForTenant(tenantID, DbName),
				DbVersion,
				client,
				true)
			if err != nil {
				panic(err)
			}
			tc.Init(t, &tc)

			ds := NewDataStoreMongoWithClient(client)
			err = ds.ReplaceReleaseTags(tc.Context, tc.ReleaseName, tc.Tags)
			if tc.ErrorAssertionFunc == nil {
				if assert.NoError(t, err) {
					var release model.Release
					err := client.Database(
						ctxstore.DbNameForTenant(tenantID, DbName)).
						Collection(CollectionReleases).
						FindOne(ctx, bson.D{}).
						Decode(&release)
					if assert.NoError(t, err, "failed to decode updated release") {
						assert.EqualValues(t, tc.Tags, release.Tags)
					}
				}
			} else {
				tc.ErrorAssertionFunc(t, err)
			}
		})
	}
}

func TestUpdateRelease(t *testing.T) {
	ctx := context.Background()
	client := db.Client()
	db.Wipe()

	longReleaseNotes := make([]byte, model.NotesLengthMaximumCharacters+1)
	for i := range longReleaseNotes {
		longReleaseNotes[i] = '1'
	}

	type testCase struct {
		Name string

		context.Context

		Init        func(t *testing.T, self *testCase)
		ReleaseName string

		Release       model.ReleasePatch
		ReleaseUpdate model.ReleasePatch

		assert.ErrorAssertionFunc
	}

	testCases := []testCase{
		{
			Name: "ok",

			Context: context.Background(),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				_, err := client.Database(DbName).
					Collection(CollectionReleases).
					InsertMany(ctx, []interface{}{model.Release{
						Name: self.ReleaseName,
						Tags: model.Tags{"bar", "foo"},
					}, model.Release{
						Name: "v100.2.3",
						Tags: model.Tags{"bar", "baz"},
					}})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			Release: model.ReleasePatch{
				Notes: "New release 2023",
			},
			ReleaseUpdate: model.ReleasePatch{
				Notes: "Brand New release 2023",
			},
			ReleaseName: "v1.0",
		},
		{
			Name: "ok same update",

			Context: context.Background(),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				_, err := client.Database(DbName).
					Collection(CollectionReleases).
					InsertMany(ctx, []interface{}{model.Release{
						Name: self.ReleaseName,
						Tags: model.Tags{"bar", "foo"},
					}, model.Release{
						Name: "v100.2.4",
						Tags: model.Tags{"bar", "baz"},
					}})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			Release: model.ReleasePatch{
				Notes: "New release 2023",
			},
			ReleaseUpdate: model.ReleasePatch{
				Notes: "New release 2023",
			},
			ReleaseName: "v1.1",
		},
		{
			Name: "error/notes too long",

			Context: identity.WithContext(context.Background(),
				&identity.Identity{
					Tenant: "222222222222222222222222",
				},
			),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				_, err := client.Database(ctxstore.DbFromContext(self, DbName)).
					Collection(CollectionReleases).
					InsertMany(self,
						[]interface{}{model.Release{
							Name: self.ReleaseName,
							Tags: func() model.Tags {
								newTags := make(model.Tags, model.TagsMaxUnique)
								for i := range newTags {
									newTags[i] = model.Tag(
										"field" + strconv.Itoa(i),
									)
								}
								return newTags
							}(),
						}, model.Release{
							Name: "v1.2.3-beta",
							Tags: model.Tags{"bar", "foo"},
						}})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			Release: model.ReleasePatch{
				Notes: model.Notes(longReleaseNotes),
			},
			ReleaseName: "v1.0",
			ErrorAssertionFunc: func(t assert.TestingT, err error, vargs ...interface{}) bool {
				return assert.ErrorIs(t, err, model.ErrReleaseNotesTooLong)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var tenantID string
			if id := identity.FromContext(tc.Context); id != nil {
				tenantID = id.Tenant
			}
			err := MigrateSingle(ctx,
				ctxstore.DbNameForTenant(tenantID, DbName),
				DbVersion,
				client,
				true)
			if err != nil {
				panic(err)
			}
			tc.Init(t, &tc)

			ds := NewDataStoreMongoWithClient(client)
			err = ds.UpdateRelease(tc.Context, tc.ReleaseName, tc.Release)
			if tc.ErrorAssertionFunc == nil {
				if assert.NoError(t, err) {
					var release model.Release
					err = client.Database(
						ctxstore.DbNameForTenant(tenantID, DbName)).
						Collection(CollectionReleases).
						FindOne(ctx, bson.D{
							{StorageKeyReleaseName, tc.ReleaseName},
						}).
						Decode(&release)
					if assert.NoError(t, err, "failed to decode updated release") {
						assert.Equal(t, tc.Release.Notes, release.Notes)
					}
					err = ds.UpdateRelease(tc.Context, tc.ReleaseName, tc.ReleaseUpdate)
					err = client.Database(
						ctxstore.DbNameForTenant(tenantID, DbName)).
						Collection(CollectionReleases).
						FindOne(ctx, bson.D{
							{StorageKeyReleaseName, tc.ReleaseName},
						}).
						Decode(&release)
					if assert.NoError(t, err, "failed to decode updated release") {
						assert.Equal(t, tc.ReleaseUpdate.Notes, release.Notes)
					}
				}
			} else {
				tc.ErrorAssertionFunc(t, err)
			}
		})
	}
}

func TestDeleteReleasesByNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeleteReleasesByNames in short mode.")
	}

	testCases := map[string]struct {
		inputReleases  []interface{}
		names          []string
		outputReleases []model.Release
	}{
		"ok": {
			inputReleases: []interface{}{
				&model.Release{
					Name: "foo",
				},
				&model.Release{
					Name: "bar",
				},
				&model.Release{
					Name: "baz",
				},
			},
			names: []string{"foo", "bar"},
			outputReleases: []model.Release{
				{
					Name: "baz",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collReleases := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionReleases)

			if tc.inputReleases != nil {
				_, err := collReleases.InsertMany(
					ctx, tc.inputReleases)
				assert.NoError(t, err)
			}

			err := ds.DeleteReleasesByNames(ctx, tc.names)
			assert.NoError(t, err)
			cur, err := collReleases.Find(ctx, bson.M{})
			assert.NoError(t, err)
			var releases []model.Release
			err = cur.All(ctx, &releases)
			assert.NoError(t, err)
			assert.Equal(t, tc.outputReleases, releases)
		})
	}
}
