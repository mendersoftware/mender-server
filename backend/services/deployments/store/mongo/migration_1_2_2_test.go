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
)

func TestMigration_1_2_2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_2 in short mode.")
	}
	ctx := context.Background()

	testCases := map[string]struct {
		// ST or MT naming convention
		db    string
		dbVer string

		err error
	}{
		"ST, no index, 0.0.0": {
			db:    "deployments_service",
			dbVer: "",
		},
		"MT, no index, 0.0.0": {
			db:    "deployments_service-59afdb71c704db002a86ad95",
			dbVer: "",
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
			migrate.UpdateMigrationInfo(db.CTX(), *ver, c, tc.db)
		}

		migrations := []migrate.Migration{
			&migration_1_2_1{
				client: c,
				db:     tc.db,
			},
			&migration_1_2_2{
				client: c,
				db:     tc.db,
			},
		}

		m := migrate.SimpleMigrator{
			Client:      c,
			Db:          tc.db,
			Automigrate: true,
		}

		err := m.Apply(ctx, migrate.MakeVersion(1, 2, 2), migrations)
		assert.NoError(t, err)

		devicesCollectionIndicesNames := []string{
			IndexDeploymentDeviceStatusesName,
			IndexDeploymentDeviceIdStatusName,
			IndexDeploymentDeviceDeploymentIdName,
		}
		// verify new indices present
		collection := c.Database(tc.db).Collection(CollectionDevices)
		indexes := collection.Indexes()
		cursor, _ := indexes.List(ctx)
		for cursor.Next(ctx) {
			var tmp map[string]interface{}
			cursor.Decode(&tmp)
			t.Log(tmp)
		}

		for _, indexName := range devicesCollectionIndicesNames {
			hasNew, err := hasIndex(ctx, indexName, indexes)
			assert.NoError(t, err)
			assert.True(t, hasNew)
		}

		deploymentsCollectionIndicesNames := []string{
			IndexDeploymentCreatedName,
			IndexDeploymentDeviceStatusFinishedName,
		}
		// verify new deployment indices present
		collection = c.Database(tc.db).Collection(CollectionDeployments)
		indexes = collection.Indexes()
		for _, indexName := range deploymentsCollectionIndicesNames {
			hasNew, err := hasIndex(ctx, indexName, indexes)
			assert.NoError(t, err)
			t.Log("Checking index: " + indexName)
			assert.True(t, hasNew)
		}
	}
}
