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
	"testing"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func TestMigration_1_2_1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_1 in short mode.")
	}
	ctx := context.Background()

	testCases := map[string]struct {
		// ST or MT naming convention
		db    string
		dbVer string

		// non-empty deployments also means the index is already present
		// must be nil for MT databases - the index name issue implies no deployments were created ever
		deployments []*model.Deployment

		err error
	}{
		"ST, no index, 0.0.0": {
			db:          "deployments_service",
			dbVer:       "",
			deployments: nil,
		},
		"ST, no index, 0.0.1": {
			db:          "deployments_service",
			dbVer:       "0.0.1",
			deployments: nil,
		},
		"ST, with index, 0.0.1": {
			db:    "deployments_service",
			dbVer: "0.0.1",
			deployments: []*model.Deployment{
				makeDeployment(t, "one", "artifact1"),
				makeDeployment(t, "two", "artifact2"),
			},
		},
		"MT, no index, 0.0.0": {
			db:          "deployments_service-59afdb71c704db002a86ad95",
			dbVer:       "",
			deployments: nil,
		},
		"MT, no index, 0.0.1": {
			db:          "deployments_service-59afdb71c704db002a86ad95",
			dbVer:       "0.0.1",
			deployments: nil,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		db.Wipe()
		c := db.Client()

		// setup
		// setup existing migrations
		if tc.dbVer != "" {
			ver, err := migrate.NewVersion(tc.dbVer)
			assert.NoError(t, err)
			migrate.UpdateMigrationInfo(ctx, *ver, c, tc.db)
		}

		// setup existing deployments
		for _, d := range tc.deployments {
			err := insertDeployment(d, c, tc.db)
			assert.NoError(t, err)
		}

		migrations := []migrate.Migration{
			&migration_1_2_1{
				client: c,
				db:     tc.db,
			},
		}

		m := migrate.SimpleMigrator{
			Client:      c,
			Db:          tc.db,
			Automigrate: true,
		}

		// verify for DBs with data - just in case verify the old/long
		// index was created
		if tc.deployments != nil {
			coll := c.Database(tc.db).
				Collection(CollectionDeployments)
			iw := coll.Indexes()
			hasOld, err := hasIndex(
				ctx, IndexDeploymentArtifactName_0_0_0, iw)
			assert.NoError(t, err)
			assert.True(t, hasOld)
		}

		err := m.Apply(ctx, migrate.MakeVersion(1, 2, 1), migrations)
		assert.NoError(t, err)

		// verify old index dropped
		coll := c.Database(tc.db).Collection(CollectionDeployments)
		indexes := coll.Indexes()
		hasOld, err := hasIndex(
			ctx, IndexDeploymentArtifactName_0_0_0, indexes)
		assert.NoError(t, err)
		assert.False(t, hasOld)

		// verify new index present - only if deployment inserted
		if tc.deployments != nil {
			hasNew, err := hasIndex(
				ctx, IndexDeploymentArtifactName, indexes)
			assert.NoError(t, err)
			assert.True(t, hasNew)
		}
	}
}

// makeDeployments creates a bare-bones deployment struct
func makeDeployment(t *testing.T, name, artifactName string) *model.Deployment {
	d, err := model.NewDeploymentFromConstructor(
		&model.DeploymentConstructor{
			Name:         name,
			ArtifactName: artifactName})

	assert.NoError(t, err)
	return d
}

// insertDeployment mimics the 0.0.1 method of inserting deployments (now deleted)
// creates the deployment name + artifact name index with a 'long' name
func insertDeployment(
	d *model.Deployment, client *mongo.Client, db string) error {
	ctx := context.Background()
	collection := client.Database(db).Collection(CollectionDeployments)
	indexes := collection.Indexes()

	indexOptions := mopts.Index()
	indexOptions.SetBackground(false)
	indexOptions.SetName(IndexDeploymentArtifactName_0_0_0)
	indexModel := mongo.IndexModel{
		Keys:    StorageIndexes.Keys,
		Options: indexOptions,
	}
	_, err := indexes.CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	_, err = collection.InsertOne(ctx, *d)
	return err
}

// find indexes by name - new and old indexes have the same key so it's not useful
// Name is what causes the issue with long index name, so use that
func hasIndex(
	ctx context.Context, name string, iw mongo.IndexView) (bool, error) {
	var index bson.M
	cursor, err := iw.List(ctx)
	if err != nil {
		return false, err
	}
	for cursor.Next(ctx) {
		if err := cursor.Decode(&index); err != nil {
			// Should not occurr
			return false, err
		}
		if index["name"] == name {
			return true, nil
		}
	}

	return false, nil
}
