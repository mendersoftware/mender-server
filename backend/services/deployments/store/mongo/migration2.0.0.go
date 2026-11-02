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
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate/utils"
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

type legacyDeployment struct {
	Deployment `bson:",inline"`
	Stats      *model.Stats `bson:"stats,omitempty"`
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
		CollectionUploadIntents: {
			{Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "expire", Value: 1},
			}},
		},
		// CollectionSettings only use primary index (tenant ID)
	}
	for _, coll := range []string{
		CollectionDeployments,
		CollectionDevices,
		CollectionReleases,
		CollectionSettings,
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
	registry := newRegistry()
	for _, dbName := range dbNames {
		logger.Infof("migrating database documents from %q", dbName)
		tenantID, found := strings.CutPrefix(dbName, legacyDbNamePrefix)
		if !found {
			continue
		} else if tenantID == "" {
			// Only update types are stored in "non-tenant" db.
			// The collection is now obsolete.
			continue
		}
		legacyDB := dbClient.Database(dbName,
			options.Database().
				SetReadPreference(readpref.Nearest()).
				SetRegistry(registry))

		err = utils.CopyCollection(ctx,
			db.Collection(CollectionReleases),
			legacyDB.Collection(CollectionReleases),
			func(release legacyReleaseExt) legacyReleaseExt {
				release.Name = release.LegacyName
				release.LegacyName = ""
				release.TenantID = tenantID
				// Initialize with best possible value from data.
				release.Created = slices.MinFunc(release.Artifacts, func(a, b model.Image) int {
					return int(min(a.Modified.Unix(), b.Modified.Unix()))
				}).Modified
				return release
			},
		)
		if err != nil {
			return fmt.Errorf("[%s] failed to migrate Releases: %w", dbName, err)
		}

		err = utils.CopyCollection(ctx,
			db.Collection(CollectionDeployments),
			legacyDB.Collection(CollectionDeployments),
			func(doc legacyDeployment) Deployment {
				doc.TenantID = tenantID
				if doc.Stats != nil {
					doc.Statistics.Status = *doc.Stats
				}
				return doc.Deployment
			})
		if err != nil {
			return fmt.Errorf("[%s] failed to migrate Deployments: %w", dbName, err)
		}

		err = utils.CopyCollection(ctx,
			db.Collection(CollectionDevices),
			legacyDB.Collection(CollectionDevices),
			func(doc DeviceDeployment) DeviceDeployment {
				doc.TenantID = tenantID
				return doc
			},
			utils.WithFindOptions(options.Find().
				SetProjection(bson.M{
					"log": 0, // Changed from bool to object with logs
				}),
			),
		)
		if err != nil {
			return fmt.Errorf("[%s] failed to migrate DeviceDeployments: %w", dbName, err)
		}
		type DeploymentLogExt struct {
			model.DeploymentLog `bson:",inline"`
			DeviceID            string `bson:"deviceid"`
			DeploymentID        string `bson:"deploymentid"`
		}
		err = utils.CopyUpsertCollection(ctx,
			db.Collection(CollectionDevices),
			legacyDB.Collection(legacyCollectionDeviceDeploymentLogs),
			func(doc DeploymentLogExt) any {
				return bson.D{{
					Key:   StorageKeyDeviceDeploymentDeploymentID,
					Value: doc.DeploymentID,
				}, {
					Key:   StorageKeyDeviceDeploymentDeviceId,
					Value: doc.DeviceID,
				}, {
					Key:   StorageKeyTenantId,
					Value: tenantID,
				}}
			}, func(doc DeploymentLogExt) any {
				return bson.D{{Key: "$set", Value: bson.D{
					{Key: "deployment_log", Value: doc.DeploymentLog},
				}}}
			},
		)
		if err != nil {
			return fmt.Errorf("[%s] failed to migrate DeviceDeployments: %w", dbName, err)
		}
		err = utils.CopyCollection(ctx,
			db.Collection(CollectionSettings),
			legacyDB.Collection(CollectionSettings),
			func(doc model.StorageSettings) Settings {
				return Settings{
					Storage:  &doc,
					TenantID: tenantID,
				}
			})
		if err != nil {
			return fmt.Errorf("[%s] failed to migrate storage settings: %w", dbName, err)
		}
		err = utils.CopyUpsertCollection(ctx,
			db.Collection(CollectionSettings),
			legacyDB.Collection(legacyCollectionLimits),
			func(doc model.Limit) any {
				return bson.D{
					{Key: "_id", Value: tenantID},
				}
			}, func(doc model.Limit) any {
				return bson.D{
					{Key: "$push", Value: bson.D{
						{Key: "limits", Value: doc},
					}},
				}
			},
			utils.WithFindOptions(options.Find().
				SetProjection(bson.M{
					"name": "$_id",
				}),
			),
		)
		if err != nil {
			return fmt.Errorf("[%s] failed to migrate limits: %w", dbName, err)
		}
		// NOTE: (enterprise) delta configuration merges here

		// UploadIntents does not require document migration.
		// Only caches state for ad-hoc uploads.
	}
	return nil
}

func (migration_2_0_0) MinimumVersion() migrate.Version {
	return migrate.Version{Major: 1, Minor: 2, Patch: 17}
}

func (migration_2_0_0) Version() migrate.Version {
	return migrate.Version{Major: 2, Minor: 0, Patch: 0}
}
