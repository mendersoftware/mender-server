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
	"io"
	"testing"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration1dot2dot14(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestMigration1dot2dot14 in short mode")
		return
	}

	mgoClient := db.Client()
	ctx := context.Background()
	logger := log.NewEmpty()
	logger.Logger.Out = io.Discard
	ctx = log.WithContext(ctx, logger)

	t.Run("ok", func(t *testing.T) {
		db.Wipe()
		migration := &migration_1_2_14{
			client: mgoClient,
			db:     DatabaseName,
		}
		migrator := migrate.SimpleMigrator{
			Client:      mgoClient,
			Db:          DatabaseName,
			Automigrate: true,
		}
		err := migrator.Apply(ctx, migration.Version(), []migrate.Migration{
			migration,
		})
		assert.NoError(t, err)

		database := mgoClient.Database(DatabaseName)
		var migrationInfo struct {
			Version migrate.Version `bson:"version"`
		}
		err = database.Collection("migration_info").
			FindOne(ctx, bson.D{}).
			Decode(&migrationInfo)
		if assert.NoError(t, err) {
			assert.Equal(t, migration.Version(), migrationInfo.Version)
		}
		cur, err := database.Collection(CollectionUploadIntents).
			Indexes().
			List(ctx)
		if !assert.NoError(t, err) {
			return
		}
		var index struct {
			Name string `bson:"name"`
			Key  bson.D `bson:"key"`
		}
		for cur.Next(ctx) {
			err = cur.Decode(&index)
			if !assert.NoError(t, err) {
				break
			} else if index.Name == "_id_" {
				continue
			}
			assert.Equal(t, "UploadExpire", index.Name)
			assert.Equal(t, bson.D{{
				Key: "status", Value: int32(1),
			}, {
				Key: "expire", Value: int32(1),
			}}, index.Key)
		}
	})

	t.Run("noop/wrong database name", func(t *testing.T) {
		db.Wipe()
		const databaseName = DatabaseName + "-123456789012345678901234"
		migration := &migration_1_2_14{
			client: mgoClient,
			db:     databaseName,
		}
		migrator := migrate.SimpleMigrator{
			Client:      mgoClient,
			Db:          databaseName,
			Automigrate: true,
		}
		err := migrator.Apply(ctx, migration.Version(), []migrate.Migration{
			migration,
		})
		assert.NoError(t, err)

		database := mgoClient.Database(databaseName)
		var migrationInfo struct {
			Version migrate.Version `bson:"version"`
		}
		err = database.Collection("migration_info").
			FindOne(ctx, bson.D{}).
			Decode(&migrationInfo)
		if assert.NoError(t, err) {
			assert.Equal(t, migration.Version(), migrationInfo.Version)
		}
		names, err := database.ListCollectionNames(ctx, bson.D{{
			Key: "name", Value: CollectionUploadIntents,
		}})
		if !assert.NoError(t, err) {
			return
		}
		assert.Emptyf(t, names,
			"collection %q should not exist in database %q",
			CollectionUploadIntents, databaseName)
	})
}
