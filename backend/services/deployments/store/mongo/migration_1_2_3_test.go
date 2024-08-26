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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func TestMigration_1_2_3_DeviceTypeNameIndexReplaced(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_3_DeviceTypeNameIndexReplace in short mode.")
	}

	//verify that, for 'old' artifacts, inserted before the migration:
	// - uniqueness of device_types_compatible and name will be preserved, even though the index
	//   was dropped (and values rewritten to 'depends')
	// - old artifacts won't prevent new ones from being inserted only based on device type + name
	//   (extra depends = no overlap)
	inputImages := []*model.Image{
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70f6",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "release1",
				DeviceTypesCompatible: []string{"arm6"},
			},
		},
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70f7",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "release1",
				DeviceTypesCompatible: []string{"arm7"},
			},
		},
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70f8",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "release2",
				DeviceTypesCompatible: []string{"arm8", "arm9"},
			},
		},
	}

	testCases := map[string]struct {
		img          *model.Image
		idxConflicts map[string]interface{}
	}{
		"conflict 1": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70f9",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm6"},
				},
			},
			idxConflicts: map[string]interface{}{
				"meta_artifact.name": "release1",
				"meta_artifact.depends_idx": map[interface{}]interface{}{
					"device_type": "arm6",
				},
			},
		},
		"conflict 2": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70f9",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm7"},
				},
			},
			idxConflicts: map[string]interface{}{
				"meta_artifact.name": "release1",
				"meta_artifact.depends_idx": map[interface{}]interface{}{
					"device_type": "arm7",
				},
			},
		},
		"conflict 3": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70f9",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm6", "arm7"},
				},
			},
			idxConflicts: map[string]interface{}{
				"meta_artifact.name": "release1",
				"meta_artifact.depends_idx": map[interface{}]interface{}{
					"device_type": "arm6",
				},
			},
		},
		"no conflict 1: different release": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70f9",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release2",
					DeviceTypesCompatible: []string{"arm6", "arm7"},
				},
			},
		},
		"no conflict 2: artifact has extra depends = no overlap": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70f9",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm6", "arm7"},
					Depends: bson.M{
						"checksum": "1",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		db.Wipe()
		c := db.Client()

		ctx := context.TODO()

		store := NewDataStoreMongoWithClient(c)

		// bring db to version just-before new (1.2.2)
		migrations := []migrate.Migration{
			&migration_1_2_1{
				client: c,
				db:     DbName,
			},
			&migration_1_2_2{
				client: c,
				db:     DbName,
			},
		}

		m := migrate.SimpleMigrator{
			Client:      c,
			Db:          DbName,
			Automigrate: true,
		}

		err := m.Apply(ctx, migrate.MakeVersion(1, 2, 2), migrations)
		assert.NoError(t, err)

		// insert input images
		for _, i := range inputImages {
			err = store.InsertImage(ctx, i)
			assert.NoError(t, err)
		}

		// bring db to latest version (1.2.3)
		mnew := &migration_1_2_3{
			client: c,
			db:     DbName,
		}

		err = mnew.Up(migrate.MakeVersion(1, 2, 3))
		assert.NoError(t, err)

		// try insert image under test
		err = store.InsertImage(ctx, tc.img)
		if tc.idxConflicts != nil {
			if assert.NotNil(t, err) {
				assert.IsType(t, &model.ConflictError{}, err)
			}
		} else {
			assert.NoError(t, err)
		}

		all, _, _ := store.ListImages(ctx, nil)
		for _, a := range all {
			assert.NotNil(t, a.ArtifactMeta.Depends)
		}
	}

}

func TestMigration_1_2_3_OverlappingDepends(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_3_OverlappingDepends in short mode.")
	}

	// verify that, for new v3 artifacts, 'depends' uniqueness is
	// detected correctly, i.e. by checking overlaps via the
	// exploded depends index
	inputImages := []*model.Image{
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70f9",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "release1",
				DeviceTypesCompatible: []string{"arm6"},
				Depends: bson.M{
					"checksum": "1",
				},
			},
		},
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70fa",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "release1",
				DeviceTypesCompatible: []string{"arm7"},
				Depends: bson.M{
					"checksum": "1",
				},
			},
		},
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70fb",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "release1",
				DeviceTypesCompatible: []string{"arm6"},
				Depends: bson.M{
					"checksum": "2",
				},
			},
		},
		&model.Image{
			Id: "0cb87b3d-4f08-420b-b004-4347c07f70fc",
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  "release1",
				DeviceTypesCompatible: []string{"arm8", "arm9"},
				Depends: bson.M{
					"checksum": "3",
					"foo":      []interface{}{"foo1", "foo2"},
				},
			},
		},
	}

	testCases := map[string]struct {
		img          *model.Image
		idxConflicts map[string]interface{}
	}{
		"conflict 1": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70ff",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm6", "arm7"},
					Depends: map[string]interface{}{
						"checksum": "2",
					},
				},
			},
			idxConflicts: map[string]interface{}{
				"meta_artifact.name": "release1",
				"meta_artifact.depends_idx": map[interface{}]interface{}{
					"device_type": "arm6",
					"checksum":    "2",
				},
			},
		},
		"conflict 2": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70ff",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm6", "arm8"},
					Depends: map[string]interface{}{
						"checksum": "1",
					},
				},
			},
			idxConflicts: map[string]interface{}{
				"meta_artifact.name": "release1",
				"meta_artifact.depends_idx": map[interface{}]interface{}{
					"device_type": "arm6",
					"checksum":    "1",
				},
			},
		},
		"conflict 3": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70ff",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm6"},
					Depends: map[string]interface{}{
						"checksum": "1",
					},
				},
			},
			idxConflicts: map[string]interface{}{
				"meta_artifact.name": "release1",
				"meta_artifact.depends_idx": map[interface{}]interface{}{
					"device_type": "arm6",
					"checksum":    "1",
				},
			},
		},
		"conflict 4": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70ff",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm8"},
					Depends: map[string]interface{}{
						"checksum": "3",
						"foo":      "foo1",
					},
				},
			},
			idxConflicts: map[string]interface{}{
				"meta_artifact.name": "release1",
				"meta_artifact.depends_idx": map[interface{}]interface{}{
					"device_type": "arm8",
					"checksum":    "3",
					"foo":         "foo1",
				},
			},
		},
		"no conflict: overlap + extra param": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70ff",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release1",
					DeviceTypesCompatible: []string{"arm8"},
					Depends: map[string]interface{}{
						"checksum": "3",
						"foo":      "foo1",
						"bar":      "bar1",
					},
				},
			},
		},
		"no conflict: overlap but different release": {
			img: &model.Image{
				Id: "0cb87b3d-4f08-420b-b004-4347c07f70ff",
				ArtifactMeta: &model.ArtifactMeta{
					Name:                  "release2",
					DeviceTypesCompatible: []string{"arm6"},
					Depends: map[string]interface{}{
						"checksum": "1",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		db.Wipe()
		c := db.Client()

		ctx := context.TODO()

		store := NewDataStoreMongoWithClient(c)

		// bring db to latest version (1.2.3)
		migrations := []migrate.Migration{
			&migration_1_2_1{
				client: c,
				db:     DbName,
			},
			&migration_1_2_2{
				client: c,
				db:     DbName,
			},
			&migration_1_2_3{
				client: c,
				db:     DbName,
			},
		}

		m := migrate.SimpleMigrator{
			Client:      c,
			Db:          DbName,
			Automigrate: true,
		}

		err := m.Apply(ctx, migrate.MakeVersion(1, 2, 3), migrations)
		assert.NoError(t, err)

		// insert input images
		for _, i := range inputImages {
			err = store.InsertImage(ctx, i)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
		}

		// try insert image under test
		err = store.InsertImage(ctx, tc.img)
		if tc.idxConflicts != nil {
			if assert.NotNil(t, err) {
				// The best we can do is check the error type.
				// The driver has a tendency to return invalid
				// in the error message.
				assert.IsType(t, &model.ConflictError{}, err)
			}
		} else {
			assert.NoError(t, err)
		}
	}
}
