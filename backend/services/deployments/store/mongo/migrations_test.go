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
	mgopts "go.mongodb.org/mongo-driver/mongo/options"
)

func TestMigration_1_2_15(t *testing.T) {
	ctx := context.Background()
	client := db.Client()

	t.Run("ok", func(t *testing.T) {
		migration := &migration_1_2_15{
			client: client,
			db:     "deployments_test",
		}
		err := migration.Up(migrate.MakeVersion(1, 2, 15))
		assert.NoError(t, err)
		assert.Equal(t, migrate.MakeVersion(1, 2, 15), migration.Version())
	})

	t.Run("error/indexAlreadyExist", func(t *testing.T) {
		collBadReleases := client.Database("deployments_test_bad").
			Collection(CollectionReleases)
		idxView := collBadReleases.Indexes()

		// Create a bad index with the same name
		_, _ = idxView.CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{
				Key: "bad", Value: 1,
			}},
			Options: mgopts.Index().
				SetName(IndexNameReleaseTags),
		})

		migration_fail := &migration_1_2_15{
			client: client,
			db:     "deployments_test_bad",
		}
		err := migration_fail.Up(migrate.MakeVersion(1, 2, 15))
		var srvErr mongo.ServerError
		assert.ErrorAs(t, err, &srvErr)
	})
}
