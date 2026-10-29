// Copyright 2026 Northern.tech AS
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
	"fmt"
	"slices"
	"strings"

	"github.com/mendersoftware/mender-server/pkg/log"
	pkgmongo "github.com/mendersoftware/mender-server/pkg/mongo/v2"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	"github.com/mendersoftware/mender-server/services/deployments/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var _ pkgmongo.MigrationMajor = migration_2_0_0{}

type migration_2_0_0 struct{}

const legacyDbNamePrefix = "deployment_service-"

type legacyReleaseExt struct {
	model.Release `bson:"inline"`
	// The legacy release used the name in the primary index
	LegacyName string `bson:"_id,omitempty"`
	TenantID   string `bson:"tenant_id"`
}

func (migration_2_0_0) migrateReleaseDocuments(ctx context.Context, dst, src *mongo.Collection) error {
	databaseName := src.Database().Name()
	_, tenantID, found := strings.Cut(databaseName, legacyDbNamePrefix)
	if !found {
		return nil
	}

	cur, err := src.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to copy release documents: %w", err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var release legacyReleaseExt
		err := cur.Decode(&release)
		if err != nil {
			return fmt.Errorf("failed to decode release document: %w", err)
		}
		release.Name = release.LegacyName
		release.LegacyName = ""
		release.TenantID = tenantID
		// Initialize with best possible value from data.
		release.Created = slices.MinFunc(release.Artifacts, func(a, b model.Image) int {
			return int(min(a.Modified.Unix(), b.Modified.Unix()))
		}).Modified

		_, err = dst.InsertOne(ctx, release)
		if err != nil {
			return fmt.Errorf("failed to copy release document: %w", err)
		}
	}
	return nil
}

func (m migration_2_0_0) Up(ctx context.Context, db *mongo.Database, from migrate.Version) error {
	logger := log.FromContext(ctx)
	var indexes = map[string][]mongo.IndexModel{
		CollectionDeployments: {
			{
				Keys: bson.D{ // Base sorting
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeploymentCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeploymentName, Value: 1},
					{Key: StorageKeyDeploymentCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeploymentStats + ".$**", Value: 1},
					{Key: StorageKeyDeploymentCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeploymentActive, Value: 1},
					{Key: StorageKeyDeploymentCreated, Value: -1},
				},
			},
			{
				// NOTE: Can potentially replace the previous index as active
				// deployment has `finished == null`
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeploymentFinished, Value: 1},
					{Key: StorageKeyDeploymentCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeploymentStatus, Value: 1},
					{Key: StorageKeyDeploymentCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeploymentConstructorChecksum, Value: 1},
				},
				Options: options.Index().
					SetUnique(true).
					SetPartialFilterExpression(bson.M{
						StorageKeyDeploymentFinished: nil,
					}),
			},
		},
		CollectionDevices: {
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeviceDeploymentDeploymentID, Value: 1},
					{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
					{Key: StorageKeyDeviceDeploymentCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
					{Key: StorageKeyDeviceDeploymentStatus, Value: -1},
					{Key: StorageKeyDeviceDeploymentCreated, Value: -1},
				},
			},
			{
				// For getting the last status
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
					{Key: StorageKeyDeviceDeploymentFinished, Value: -1},
				},
			},
		},
		CollectionReleases: {
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyReleaseArtifactsId, Value: 1},
					// WARN: (enterprise) manifest model needs a face-lift 🗿
				},
				Options: options.Index().
					SetUnique(true).
					SetPartialFilterExpression(bson.D{
						{Key: StorageKeyReleaseArtifactsId, Value: bson.D{
							{Key: "$exists", Value: true},
						}},
					}),
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyReleaseName, Value: 1},
				},
				Options: options.Index().
					SetUnique(true),
			},
			{
				Keys: bson.D{ // Base sorting by newest
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyReleaseCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyReleaseTags, Value: 1},
					{Key: StorageKeyReleaseCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyReleaseArtifactsUpdateTypes, Value: 1},
					{Key: StorageKeyReleaseCreated, Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: StorageKeyTenantId, Value: 1},
					{Key: StorageKeyReleaseArtifactsCount, Value: 1},
					{Key: StorageKeyReleaseCreated, Value: -1},
				},
			},
		},
		// CollectionUploadIntents only use primary index (artifact ID)
		// CollectionUploadLimits only use primary index (tenant ID)
		// CollectionUploadStorageSettings only use primary index (tenant ID)
	}
	for _, coll := range []string{
		CollectionDeployments,
		CollectionDevices,
		CollectionLimits,
		CollectionReleases,
		CollectionStorageSettings,
		CollectionUploadIntents,
	} {
		err := db.CreateCollection(ctx, coll)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		if indexModels, ok := indexes[coll]; ok {
			_, err = db.Collection(coll).
				Indexes().
				CreateMany(ctx, indexModels)
			if err != nil {
				return fmt.Errorf("failed to create indexes [%s.%s]: %w", db.Name(), coll, err)
			}
		}
	}

	// Migrate existing documents
	dbClient := db.Client()
	dbNames, err := dbClient.ListDatabaseNames(ctx,
		bson.M{"name": bson.M{"$regex": "^" + legacyDbNamePrefix}},
		options.ListDatabases().SetNameOnly(true))
	if err != nil {
		return fmt.Errorf("failed to list existing databases: %w", err)
	}
	collReleases := db.Collection(CollectionReleases)
	for _, dbName := range dbNames {
		logger.Debugf("migrating database documents from %q", dbName)
		legacyDB := dbClient.Database(dbName,
			options.Database().SetReadPreference(readpref.Nearest()))
		err := m.migrateReleaseDocuments(ctx, collReleases, legacyDB.Collection(CollectionReleases))
		if err != nil {
			return fmt.Errorf("[%s] failed to migrate releases: %w", dbName, err)
		}
	}
	return nil
}

func (migration_2_0_0) MinimumVersion() migrate.Version {
	return migrate.Version{Major: 1, Minor: 2, Patch: 17}
}

func (migration_2_0_0) Version() migrate.Version {
	return migrate.Version{Major: 2, Minor: 0, Patch: 0}
}
