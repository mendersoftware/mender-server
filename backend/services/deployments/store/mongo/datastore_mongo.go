// Copyright 2024 Northern.tech AS
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
	"crypto/tls"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

const (
	DatabaseName                   = "deployment_service"
	CollectionLimits               = "limits"
	CollectionImages               = "images"
	CollectionDeployments          = "deployments"
	CollectionDeviceDeploymentLogs = "devices.logs"
	CollectionDevices              = "devices"
	CollectionDevicesLastStatus    = "devices_last_status"
	CollectionStorageSettings      = "settings"
	CollectionUploadIntents        = "uploads"
	CollectionReleases             = "releases"
	CollectionUpdateTypes          = "update_types"
)

const DefaultDocumentLimit = 20
const maxCountDocuments = int64(10000)

// Internal status codes from
// https://github.com/mongodb/mongo/blob/4.4/src/mongo/base/error_codes.yml
const (
	errorCodeNamespaceNotFound = 26
	errorCodeIndexNotFound     = 27
)

const (
	mongoOpSet = "$set"
)

var currentDbVersion map[string]*migrate.Version

var (
	// Indexes (version: 1.2.2)
	IndexUniqueNameAndDeviceTypeName          = "uniqueNameAndDeviceTypeIndex"
	IndexDeploymentArtifactName               = "deploymentArtifactNameIndex"
	IndexDeploymentDeviceStatusesName         = "deviceIdWithStatusByCreated"
	IndexDeploymentDeviceIdStatusName         = "devicesIdWithStatus"
	IndexDeploymentDeviceCreatedStatusName    = "devicesIdWithCreatedStatus"
	IndexDeploymentDeviceDeploymentIdName     = "devicesDeploymentId"
	IndexDeploymentStatusFinishedName         = "deploymentStatusFinished"
	IndexDeploymentStatusPendingName          = "deploymentStatusPending"
	IndexDeploymentCreatedName                = "deploymentCreated"
	IndexDeploymentDeviceStatusRebootingName  = "deploymentsDeviceStatusRebooting"
	IndexDeploymentDeviceStatusPendingName    = "deploymentsDeviceStatusPending"
	IndexDeploymentDeviceStatusInstallingName = "deploymentsDeviceStatusInstalling"
	IndexDeploymentDeviceStatusFinishedName   = "deploymentsFinished"

	// Indexes (version: 1.2.3)
	IndexArtifactNameDependsName = "artifactNameDepends"
	IndexNameAndDeviceTypeName   = "artifactNameAndDeviceTypeIndex"

	// Indexes (version: 1.2.4)
	IndexDeploymentStatus = "deploymentStatus"

	// Indexes 1.2.6
	IndexDeviceDeploymentStatusName = "deploymentid_status_deviceid"

	// Indexes 1.2.13
	IndexArtifactProvidesName = "artifact_provides"

	// Indexes 1.2.15
	IndexNameReleaseTags           = "release_tags"
	IndexNameReleaseUpdateTypes    = "release_update_types"
	IndexNameAggregatedUpdateTypes = "aggregated_release_update_types"
	IndexNameReleaseArtifactsCount = "release_artifacts_count"

	// Indexes 1.2.16
	IndexNameDeploymentConstructorChecksum = "deployment_deploymentconstructor_checksum"

	// Indexes 1.2.17
	IndexNameDeploymentName = "deployment_name"

	_false         = false
	_true          = true
	StorageIndexes = mongo.IndexModel{
		// NOTE: Keys should be bson.D as element
		//       order matters!
		Keys: bson.D{
			{Key: StorageKeyDeploymentName,
				Value: "text"},
			{Key: StorageKeyDeploymentArtifactName,
				Value: "text"},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentArtifactName,
		},
	}
	StatusIndexes = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentDeviceId,
				Value: 1},
			{Key: StorageKeyDeviceDeploymentStatus,
				Value: 1},
			{Key: StorageKeyDeploymentStatsCreated,
				Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentDeviceStatusesName,
		},
	}
	DeploymentStatusIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentStatus,
				Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentStatus,
		},
	}
	DeviceIDStatusIndexes = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
			{Key: StorageKeyDeviceDeploymentStatus, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentDeviceIdStatusName,
		},
	}
	DeviceIDCreatedStatusIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
			{Key: StorageKeyDeploymentStatsCreated, Value: 1},
			{Key: StorageKeyDeviceDeploymentStatus, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentDeviceCreatedStatusName,
		},
	}
	DeploymentIdIndexes = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentDeploymentID, Value: 1},
			{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentDeviceDeploymentIdName,
		},
	}
	DeviceDeploymentIdStatus = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentDeploymentID, Value: 1},
			{Key: StorageKeyDeviceDeploymentStatus, Value: 1},
			{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
		},
		Options: mopts.Index().
			SetName(IndexDeviceDeploymentStatusName),
	}
	DeploymentCreatedIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: "created", Value: -1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentCreatedName,
		},
	}
	DeploymentDeviceStatusFinishedIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: "finished", Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentDeviceStatusFinishedName,
		},
	}
	UniqueNameVersionIndex = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyImageName,
				Value: 1},
			{Key: StorageKeyImageDeviceTypes,
				Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexUniqueNameAndDeviceTypeName,
			Unique:     &_true,
		},
	}

	// 1.2.3
	IndexArtifactNameDepends = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyImageName,
				Value: 1},
			{Key: StorageKeyImageDependsIdx,
				Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexArtifactNameDependsName,
			Unique:     &_true,
		},
	}

	// Indexes 1.2.7
	IndexImageMetaDescription      = "image_meta_description"
	IndexImageMetaDescriptionModel = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyImageDescription, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexImageMetaDescription,
		},
	}

	IndexImageMetaArtifactDeviceTypeCompatible      = "image_meta_artifact_device_type_compatible"
	IndexImageMetaArtifactDeviceTypeCompatibleModel = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyImageDeviceTypes, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexImageMetaArtifactDeviceTypeCompatible,
		},
	}

	// Indexes 1.2.8
	IndexDeploymentsActiveCreated      = "active_created"
	IndexDeploymentsActiveCreatedModel = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeploymentCreated, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Name:       &IndexDeploymentsActiveCreated,
			PartialFilterExpression: bson.M{
				StorageKeyDeploymentActive: true,
			},
		},
	}

	// Index 1.2.9
	IndexDeviceDeploymentsActiveCreated      = "active_deviceid_created"
	IndexDeviceDeploymentsActiveCreatedModel = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentActive, Value: 1},
			{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
			{Key: StorageKeyDeviceDeploymentCreated, Value: 1},
		},
		Options: mopts.Index().
			SetName(IndexDeviceDeploymentsActiveCreated),
	}

	// Index 1.2.11
	IndexDeviceDeploymentsLogs      = "devices_logs"
	IndexDeviceDeploymentsLogsModel = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeviceDeploymentDeploymentID, Value: 1},
			{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
		},
		Options: mopts.Index().
			SetName(IndexDeviceDeploymentsLogs),
	}

	// 1.2.13
	IndexArtifactProvides = mongo.IndexModel{
		Keys: bson.D{
			{Key: model.StorageKeyImageProvidesIdxKey,
				Value: 1},
			{Key: model.StorageKeyImageProvidesIdxValue,
				Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_false,
			Sparse:     &_true,
			Name:       &IndexArtifactProvidesName,
		},
	}

	// 1.2.17
	IndexDeploymentName = mongo.IndexModel{
		Keys: bson.D{
			{Key: StorageKeyDeploymentName, Value: 1},
			{Key: StorageKeyDeploymentCreated, Value: 1},
		},
		Options: &mopts.IndexOptions{
			Background: &_true,
			Name:       &IndexNameDeploymentName,
		},
	}
)

// Errors
var (
	ErrImagesStorageInvalidID           = errors.New("Invalid id")
	ErrImagesStorageInvalidArtifactName = errors.New("Invalid artifact name")
	ErrImagesStorageInvalidName         = errors.New("Invalid name")
	ErrImagesStorageInvalidDeviceType   = errors.New("Invalid device type")
	ErrImagesStorageInvalidImage        = errors.New("Invalid image")

	ErrStorageInvalidDeviceDeployment = errors.New("Invalid device deployment")

	ErrDeploymentStorageInvalidDeployment = errors.New("Invalid deployment")
	ErrStorageInvalidID                   = errors.New("Invalid id")
	ErrStorageNotFound                    = errors.New("Not found")
	ErrDeploymentStorageInvalidQuery      = errors.New("Invalid query")
	ErrDeploymentStorageCannotExecQuery   = errors.New("Cannot execute query")
	ErrStorageInvalidInput                = errors.New("invalid input")

	ErrLimitNotFound      = errors.New("limit not found")
	ErrDevicesCountFailed = errors.New("failed to count devices")
	ErrConflictingDepends = errors.New(
		"an artifact with the same name and depends already exists",
	)
	ErrConflictingDeployment = errors.New(
		"an active deployment with the same parameter already exists",
	)
)

// Database keys
const (
	// Need to be kept in sync with structure filed names
	StorageKeyId       = "_id"
	StorageKeyTenantId = "tenant_id"

	StorageKeyImageProvides    = "meta_artifact.provides"
	StorageKeyImageProvidesIdx = "meta_artifact.provides_idx"
	StorageKeyImageDepends     = "meta_artifact.depends"
	StorageKeyImageDependsIdx  = "meta_artifact.depends_idx"
	StorageKeyImageSize        = "size"
	StorageKeyImageDeviceTypes = "meta_artifact.device_types_compatible"
	StorageKeyImageName        = "meta_artifact.name"
	StorageKeyUpdateType       = "meta_artifact.updates.typeinfo.type"
	StorageKeyImageDescription = "meta.description"
	StorageKeyImageModified    = "modified"

	// releases
	StorageKeyReleaseName                      = "_id"
	StorageKeyReleaseModified                  = "modified"
	StorageKeyReleaseTags                      = "tags"
	StorageKeyReleaseNotes                     = "notes"
	StorageKeyReleaseArtifacts                 = "artifacts"
	StorageKeyReleaseArtifactsCount            = "artifacts_count"
	StorageKeyReleaseArtifactsIndexDescription = StorageKeyReleaseArtifacts + ".$." +
		StorageKeyImageDescription
	StorageKeyReleaseArtifactsDescription = StorageKeyReleaseArtifacts + "." +
		StorageKeyImageDescription
	StorageKeyReleaseArtifactsDeviceTypes = StorageKeyReleaseArtifacts + "." +
		StorageKeyImageDeviceTypes
	StorageKeyReleaseArtifactsUpdateTypes = StorageKeyReleaseArtifacts + "." +
		StorageKeyUpdateType
	StorageKeyReleaseArtifactsIndexModified = StorageKeyReleaseArtifacts + ".$." +
		StorageKeyImageModified
	StorageKeyReleaseArtifactsId = StorageKeyReleaseArtifacts + "." +
		StorageKeyId
	StorageKeyReleaseImageDependsIdx = StorageKeyReleaseArtifacts + "." +
		StorageKeyImageDependsIdx
	StorageKeyReleaseImageProvidesIdx = StorageKeyReleaseArtifacts + "." +
		StorageKeyImageProvidesIdx

	StorageKeyDeviceDeploymentLogMessages = "messages"

	StorageKeyDeviceDeploymentAssignedImage   = "image"
	StorageKeyDeviceDeploymentAssignedImageId = StorageKeyDeviceDeploymentAssignedImage +
		"." + StorageKeyId

	StorageKeyDeviceDeploymentActive         = "active"
	StorageKeyDeviceDeploymentCreated        = "created"
	StorageKeyDeviceDeploymentDeviceId       = "deviceid"
	StorageKeyDeviceDeploymentStatus         = "status"
	StorageKeyDeviceDeploymentStarted        = "started"
	StorageKeyDeviceDeploymentSubState       = "substate"
	StorageKeyDeviceDeploymentDeploymentID   = "deploymentid"
	StorageKeyDeviceDeploymentFinished       = "finished"
	StorageKeyDeviceDeploymentIsLogAvailable = "log"
	StorageKeyDeviceDeploymentArtifact       = "image"
	StorageKeyDeviceDeploymentRequest        = "request"
	StorageKeyDeviceDeploymentDeleted        = "deleted"

	StorageKeyDeploymentName                = "deploymentconstructor.name"
	StorageKeyDeploymentArtifactName        = "deploymentconstructor.artifactname"
	StorageKeyDeploymentConstructorChecksum = "deploymentconstructor_checksum"
	StorageKeyDeploymentStats               = "stats"
	StorageKeyDeploymentActive              = "active"
	StorageKeyDeploymentStatus              = "status"
	StorageKeyDeploymentCreated             = "created"
	StorageKeyDeploymentDeviceList          = "device_list"
	StorageKeyDeploymentStatsCreated        = "created"
	StorageKeyDeploymentFinished            = "finished"
	StorageKeyDeploymentArtifacts           = "artifacts"
	StorageKeyDeploymentDeviceCount         = "device_count"
	StorageKeyDeploymentMaxDevices          = "max_devices"
	StorageKeyDeploymentType                = "type"
	StorageKeyDeploymentTotalSize           = "statistics.total_size"

	StorageKeyStorageSettingsDefaultID      = "settings"
	StorageKeyStorageSettingsBucket         = "bucket"
	StorageKeyStorageSettingsRegion         = "region"
	StorageKeyStorageSettingsKey            = "key"
	StorageKeyStorageSettingsSecret         = "secret"
	StorageKeyStorageSettingsURI            = "uri"
	StorageKeyStorageSettingsExternalURI    = "external_uri"
	StorageKeyStorageSettingsToken          = "token"
	StorageKeyStorageSettingsForcePathStyle = "force_path_style"
	StorageKeyStorageSettingsUseAccelerate  = "use_accelerate"

	StorageKeyStorageReleaseUpdateTypes = "update_types"

	ArtifactDependsDeviceType = "device_type"
)

type DataStoreMongo struct {
	client *mongo.Client
}

func NewDataStoreMongoWithClient(client *mongo.Client) *DataStoreMongo {
	return &DataStoreMongo{
		client: client,
	}
}

func NewMongoClient(ctx context.Context, c config.Reader) (*mongo.Client, error) {

	clientOptions := mopts.Client()
	mongoURL := c.GetString(dconfig.SettingMongo)
	if !strings.Contains(mongoURL, "://") {
		return nil, errors.Errorf("Invalid mongoURL %q: missing schema.",
			mongoURL)
	}
	clientOptions.ApplyURI(mongoURL)

	username := c.GetString(dconfig.SettingDbUsername)
	if username != "" {
		credentials := mopts.Credential{
			Username: c.GetString(dconfig.SettingDbUsername),
		}
		password := c.GetString(dconfig.SettingDbPassword)
		if password != "" {
			credentials.Password = password
			credentials.PasswordSet = true
		}
		clientOptions.SetAuth(credentials)
	}

	if c.GetBool(dconfig.SettingDbSSL) {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = c.GetBool(dconfig.SettingDbSSLSkipVerify)
		clientOptions.SetTLSConfig(tlsConfig)
	}

	// Set 10s timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to mongo server")
	}

	// Validate connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "Error reaching mongo server")
	}

	return client, nil
}

func (db *DataStoreMongo) Ping(ctx context.Context) error {
	res := db.client.Database(DbName).RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
}

func (db *DataStoreMongo) setCurrentDbVersion(
	ctx context.Context,
) error {
	versions, err := migrate.GetMigrationInfo(
		ctx, db.client, mstore.DbFromContext(ctx, DatabaseName))
	if err != nil {
		return errors.Wrap(err, "failed to list applied migrations")
	}
	var current migrate.Version
	if len(versions) > 0 {
		// sort applied migrations wrt. version
		sort.Slice(versions, func(i int, j int) bool {
			return migrate.VersionIsLess(versions[i].Version, versions[j].Version)
		})
		current = versions[len(versions)-1].Version
	}
	if currentDbVersion == nil {
		currentDbVersion = map[string]*migrate.Version{}
	}
	currentDbVersion[mstore.DbFromContext(ctx, DatabaseName)] = &current
	return nil
}

func (db *DataStoreMongo) getCurrentDbVersion(
	ctx context.Context,
) (*migrate.Version, error) {
	if currentDbVersion == nil ||
		currentDbVersion[mstore.DbFromContext(ctx, DatabaseName)] == nil {
		if err := db.setCurrentDbVersion(ctx); err != nil {
			return nil, err
		}
	}
	return currentDbVersion[mstore.DbFromContext(ctx, DatabaseName)], nil
}

func (db *DataStoreMongo) GetReleases(
	ctx context.Context,
	filt *model.ReleaseOrImageFilter,
) ([]model.Release, int, error) {
	current, err := db.getCurrentDbVersion(ctx)
	if err != nil {
		return []model.Release{}, 0, err
	} else if current == nil {
		return []model.Release{}, 0, errors.New("couldn't get current database version")
	}
	if migrate.VersionIsLess(*current, migrate.Version{
		Major: 1, Minor: 2, Patch: 15,
	}) {
		return db.getReleases_1_2_14(ctx, filt)
	} else {
		return db.getReleases_1_2_15(ctx, filt)
	}
}

func (db *DataStoreMongo) getReleases_1_2_14(
	ctx context.Context,
	filt *model.ReleaseOrImageFilter,
) ([]model.Release, int, error) {
	l := log.FromContext(ctx)
	l.Infof("get releases method version 1.2.14")
	var pipe []bson.D

	pipe = []bson.D{}
	if filt != nil && filt.Name != "" {
		pipe = append(pipe, bson.D{
			{Key: "$match", Value: bson.M{
				StorageKeyImageName: bson.M{
					"$regex": primitive.Regex{
						Pattern: ".*" + regexp.QuoteMeta(filt.Name) + ".*",
						Options: "i",
					},
				},
			}},
		})
	}

	pipe = append(pipe, bson.D{
		// Remove (possibly expensive) sub-documents from pipeline
		{
			Key: "$project",
			Value: bson.M{
				StorageKeyImageDependsIdx:  0,
				StorageKeyImageProvidesIdx: 0,
			},
		},
	})

	pipe = append(pipe, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$" + StorageKeyImageName},
			{Key: "name", Value: bson.M{"$first": "$" + StorageKeyImageName}},
			{Key: "artifacts", Value: bson.M{"$push": "$$ROOT"}},
			{Key: "modified", Value: bson.M{"$max": "$modified"}},
		}},
	})

	if filt != nil && filt.Description != "" {
		pipe = append(pipe, bson.D{
			{Key: "$match", Value: bson.M{
				"artifacts." + StorageKeyImageDescription: bson.M{
					"$regex": primitive.Regex{
						Pattern: ".*" + regexp.QuoteMeta(filt.Description) + ".*",
						Options: "i",
					},
				},
			}},
		})
	}
	if filt != nil && filt.DeviceType != "" {
		pipe = append(pipe, bson.D{
			{Key: "$match", Value: bson.M{
				"artifacts." + StorageKeyImageDeviceTypes: bson.M{
					"$regex": primitive.Regex{
						Pattern: ".*" + regexp.QuoteMeta(filt.DeviceType) + ".*",
						Options: "i",
					},
				},
			}},
		})
	}

	sortField, sortOrder := getReleaseSortFieldAndOrder(filt)
	if sortField == "" {
		sortField = "name"
	}
	if sortOrder == 0 {
		sortOrder = 1
	}

	page := 1
	perPage := math.MaxInt64
	if filt != nil && filt.Page > 0 && filt.PerPage > 0 {
		page = filt.Page
		perPage = filt.PerPage
	}
	pipe = append(pipe,
		bson.D{{Key: "$facet", Value: bson.D{
			{Key: "results", Value: []bson.D{
				{
					{Key: "$sort", Value: bson.D{
						{Key: sortField, Value: sortOrder},
						{Key: "_id", Value: 1},
					}},
				},
				{{Key: "$skip", Value: int64((page - 1) * perPage)}},
				{{Key: "$limit", Value: int64(perPage)}},
			}},
			{Key: "count", Value: []bson.D{
				{{Key: "$count", Value: "count"}},
			}},
		}}},
	)

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	cursor, err := collImg.Aggregate(ctx, pipe)
	if err != nil {
		return []model.Release{}, 0, err
	}
	defer cursor.Close(ctx)

	result := struct {
		Results []model.Release       `bson:"results"`
		Count   []struct{ Count int } `bson:"count"`
	}{}
	if !cursor.Next(ctx) {
		return []model.Release{}, 0, nil
	} else if err = cursor.Decode(&result); err != nil {
		return []model.Release{}, 0, err
	} else if len(result.Count) == 0 {
		return []model.Release{}, 0, err
	}
	return result.Results, result.Count[0].Count, nil
}

func (db *DataStoreMongo) getReleases_1_2_15(
	ctx context.Context,
	filt *model.ReleaseOrImageFilter,
) ([]model.Release, int, error) {
	l := log.FromContext(ctx)
	l.Infof("get releases method version 1.2.15")

	sortField, sortOrder := getReleaseSortFieldAndOrder(filt)
	if sortField == "" {
		sortField = "_id"
	} else if sortField == "name" {
		sortField = StorageKeyReleaseName
	}
	if sortOrder == 0 {
		sortOrder = 1
	}

	page := 1
	perPage := DefaultDocumentLimit
	if filt != nil {
		if filt.Page > 0 {
			page = filt.Page
		}
		if filt.PerPage > 0 {
			perPage = filt.PerPage
		}
	}

	opts := &mopts.FindOptions{}
	opts.SetSort(bson.D{{Key: sortField, Value: sortOrder}})
	opts.SetSkip(int64((page - 1) * perPage))
	opts.SetLimit(int64(perPage))
	projection := bson.M{
		StorageKeyReleaseImageDependsIdx:  0,
		StorageKeyReleaseImageProvidesIdx: 0,
	}
	opts.SetProjection(projection)

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collReleases := database.Collection(CollectionReleases)

	filter := bson.M{}
	if filt != nil {
		if filt.Name != "" {
			filter[StorageKeyReleaseName] = bson.M{"$regex": primitive.Regex{
				Pattern: regexp.QuoteMeta(filt.Name) + ".*",
				Options: "i",
			}}
		}
		if len(filt.Tags) > 0 {
			filter[StorageKeyReleaseTags] = bson.M{"$in": filt.Tags}
		}
		if filt.Description != "" {
			filter[StorageKeyReleaseArtifactsDescription] = bson.M{"$regex": primitive.Regex{
				Pattern: ".*" + regexp.QuoteMeta(filt.Description) + ".*",
				Options: "i",
			}}
		}
		if filt.DeviceType != "" {
			filter[StorageKeyReleaseArtifactsDeviceTypes] = filt.DeviceType
		}
		if filt.UpdateType != "" {
			filter[StorageKeyReleaseArtifactsUpdateTypes] = filt.UpdateType
		}
	}
	releases := []model.Release{}
	cursor, err := collReleases.Find(ctx, filter, opts)
	if err != nil {
		return []model.Release{}, 0, err
	}
	if err := cursor.All(ctx, &releases); err != nil {
		return []model.Release{}, 0, err
	}

	// TODO: can we return number of all documents in the collection
	// using EstimatedDocumentCount?
	count, err := collReleases.CountDocuments(ctx, filter)
	if err != nil {
		return []model.Release{}, 0, err
	}

	if count < 1 {
		return []model.Release{}, int(count), nil
	}
	return releases, int(count), nil
}

// limits
func (db *DataStoreMongo) GetLimit(ctx context.Context, name string) (*model.Limit, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collLim := database.Collection(CollectionLimits)

	limit := new(model.Limit)
	if err := collLim.FindOne(ctx, bson.M{"_id": name}).
		Decode(limit); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrLimitNotFound
		}
		return nil, err
	}

	return limit, nil
}

func (db *DataStoreMongo) ProvisionTenant(ctx context.Context, tenantId string) error {

	dbname := mstore.DbNameForTenant(tenantId, DbName)

	return MigrateSingle(ctx, dbname, DbVersion, db.client, true)
}

//images

// Exists checks if object with ID exists
func (db *DataStoreMongo) Exists(ctx context.Context, id string) (bool, error) {
	var result interface{}

	if len(id) == 0 {
		return false, ErrImagesStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	if err := collImg.FindOne(ctx, bson.M{"_id": id}).
		Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Update provided Image
// Return false if not found
func (db *DataStoreMongo) Update(ctx context.Context,
	image *model.Image) (bool, error) {

	if err := image.Validate(); err != nil {
		return false, err
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	// add special representation of artifact provides
	image.ArtifactMeta.ProvidesIdx = model.ProvidesIdx(image.ArtifactMeta.Provides)

	image.SetModified(time.Now())
	if res, err := collImg.ReplaceOne(
		ctx, bson.M{"_id": image.Id}, image,
	); err != nil {
		return false, err
	} else if res.MatchedCount == 0 {
		return false, nil
	}

	return true, nil
}

// ImageByNameAndDeviceType finds image with specified application name and target device type
func (db *DataStoreMongo) ImageByNameAndDeviceType(ctx context.Context,
	name, deviceType string) (*model.Image, error) {

	if len(name) == 0 {
		return nil, ErrImagesStorageInvalidArtifactName
	}

	if len(deviceType) == 0 {
		return nil, ErrImagesStorageInvalidDeviceType
	}

	// equal to device type & software version (application name + version)
	query := bson.M{
		StorageKeyImageName:        name,
		StorageKeyImageDeviceTypes: deviceType,
	}

	// If multiple entries matches, pick the smallest one.
	findOpts := mopts.FindOne()
	findOpts.SetSort(bson.D{{Key: StorageKeyImageSize, Value: 1}})

	dbName := mstore.DbFromContext(ctx, DatabaseName)
	database := db.client.Database(dbName)
	collImg := database.Collection(CollectionImages)

	// Both we lookup unique object, should be one or none.
	var image model.Image
	if err := collImg.FindOne(ctx, query, findOpts).
		Decode(&image); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &image, nil
}

// ImageByIdsAndDeviceType finds image with id from ids and target device type
func (db *DataStoreMongo) ImageByIdsAndDeviceType(ctx context.Context,
	ids []string, deviceType string) (*model.Image, error) {

	if len(deviceType) == 0 {
		return nil, ErrImagesStorageInvalidDeviceType
	}

	if len(ids) == 0 {
		return nil, ErrImagesStorageInvalidID
	}

	query := bson.D{
		{Key: StorageKeyId, Value: bson.M{"$in": ids}},
		{Key: StorageKeyImageDeviceTypes, Value: deviceType},
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	// If multiple entries matches, pick the smallest one
	findOpts := mopts.FindOne()
	findOpts.SetSort(bson.D{{Key: StorageKeyImageSize, Value: 1}})

	// Both we lookup unique object, should be one or none.
	var image model.Image
	if err := collImg.FindOne(ctx, query, findOpts).
		Decode(&image); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &image, nil
}

// ImagesByName finds images with specified artifact name
func (db *DataStoreMongo) ImagesByName(
	ctx context.Context, name string) ([]*model.Image, error) {

	var images []*model.Image

	if len(name) == 0 {
		return nil, ErrImagesStorageInvalidName
	}

	// equal to artifact name
	query := bson.M{
		StorageKeyImageName: name,
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)
	cursor, err := collImg.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	// Both we lookup unique object, should be one or none.
	if err = cursor.All(ctx, &images); err != nil {
		return nil, err
	}

	return images, nil
}

func newDependsConflictError(mgoErr mongo.WriteError) *model.ConflictError {
	var err error
	conflictErr := model.NewConflictError(ErrConflictingDepends)
	// Try to lookup the document that caused the index violation:
	if raw, ok := mgoErr.Raw.Lookup("keyValue").DocumentOK(); ok {
		if raw, ok = raw.Lookup(StorageKeyImageDependsIdx).DocumentOK(); ok {
			var conflicts map[string]interface{}
			err = bson.Unmarshal([]byte(raw), &conflicts)
			if err == nil {
				_ = conflictErr.WithMetadata(
					map[string]interface{}{
						"conflict": conflicts,
					},
				)
			}
		}
	}
	return conflictErr
}

// Insert persists object
func (db *DataStoreMongo) InsertImage(ctx context.Context, image *model.Image) error {

	if image == nil {
		return ErrImagesStorageInvalidImage
	}

	if err := image.Validate(); err != nil {
		return err
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	// add special representation of artifact provides
	image.ArtifactMeta.ProvidesIdx = model.ProvidesIdx(image.ArtifactMeta.Provides)

	_, err := collImg.InsertOne(ctx, image)
	if err != nil {
		var wExc mongo.WriteException
		if errors.As(err, &wExc) {
			for _, wErr := range wExc.WriteErrors {
				if !mongo.IsDuplicateKeyError(wErr) {
					continue
				}
				return newDependsConflictError(wErr)
			}
		}
		return err
	}

	return nil
}

func (db *DataStoreMongo) InsertUploadIntent(ctx context.Context, link *model.UploadLink) error {
	collUploads := db.client.
		Database(DatabaseName).
		Collection(CollectionUploadIntents)
	if idty := identity.FromContext(ctx); idty != nil {
		link.TenantID = idty.Tenant
	}
	_, err := collUploads.InsertOne(ctx, link)
	return err
}

func (db *DataStoreMongo) UpdateUploadIntentStatus(
	ctx context.Context,
	id string,
	from, to model.LinkStatus,
) error {
	collUploads := db.client.
		Database(DatabaseName).
		Collection(CollectionUploadIntents)
	q := bson.D{
		{Key: "_id", Value: id},
		{Key: "status", Value: from},
	}
	if idty := identity.FromContext(ctx); idty != nil {
		q = append(q, bson.E{
			Key:   StorageKeyTenantId,
			Value: idty.Tenant,
		})
	}
	update := bson.D{{
		Key: "updated_ts", Value: time.Now(),
	}}
	if from != to {
		update = append(update, bson.E{
			Key: "status", Value: to,
		})
	}
	res, err := collUploads.UpdateOne(ctx, q, bson.D{
		{Key: "$set", Value: update},
	})
	if err != nil {
		return err
	} else if res.MatchedCount == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (db *DataStoreMongo) FindUploadLinks(
	ctx context.Context,
	expiredAt time.Time,
) (store.Iterator[model.UploadLink], error) {
	collUploads := db.client.
		Database(DatabaseName).
		Collection(CollectionUploadIntents)

	q := bson.D{{
		Key: "status",
		Value: bson.D{{
			Key:   "$lt",
			Value: model.LinkStatusProcessedBit,
		}},
	}, {
		Key: "expire",
		Value: bson.D{{
			Key:   "$lt",
			Value: expiredAt,
		}},
	}}
	cur, err := collUploads.Find(ctx, q)
	return IteratorFromCursor[model.UploadLink](cur), err
}

// FindImageByID search storage for image with ID, returns nil if not found
func (db *DataStoreMongo) FindImageByID(ctx context.Context,
	id string) (*model.Image, error) {

	if len(id) == 0 {
		return nil, ErrImagesStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)
	projection := bson.M{
		StorageKeyImageDependsIdx:  0,
		StorageKeyImageProvidesIdx: 0,
	}
	findOptions := mopts.FindOne()
	findOptions.SetProjection(projection)

	var image model.Image
	if err := collImg.FindOne(ctx, bson.M{"_id": id}, findOptions).
		Decode(&image); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &image, nil
}

// IsArtifactUnique checks if there is no artifact with the same artifactName
// supporting one of the device types from deviceTypesCompatible list.
// Returns true, nil if artifact is unique;
// false, nil if artifact is not unique;
// false, error in case of error.
func (db *DataStoreMongo) IsArtifactUnique(ctx context.Context,
	artifactName string, deviceTypesCompatible []string) (bool, error) {

	if len(artifactName) == 0 {
		return false, ErrImagesStorageInvalidArtifactName
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	query := bson.M{
		"$and": []bson.M{
			{
				StorageKeyImageName: artifactName,
			},
			{
				StorageKeyImageDeviceTypes: bson.M{
					"$in": deviceTypesCompatible},
			},
		},
	}

	// do part of the job manually
	// if candidate images have any extra 'depends' - guaranteed non-overlap
	// otherwise it's a match
	cur, err := collImg.Find(ctx, query)
	if err != nil {
		return false, err
	}

	var images []model.Image
	err = cur.All(ctx, &images)
	if err != nil {
		return false, err
	}

	for _, i := range images {
		// the artifact already has same name and overlapping dev type
		// if there are no more depends than dev type - it's not unique
		if len(i.ArtifactMeta.Depends) == 1 {
			if _, ok := i.ArtifactMeta.Depends["device_type"]; ok {
				return false, nil
			}
		} else if len(i.ArtifactMeta.Depends) == 0 {
			return false, nil
		}
	}

	return true, nil
}

// Delete image specified by ID
// Noop on if not found.
func (db *DataStoreMongo) DeleteImage(ctx context.Context, id string) error {

	if len(id) == 0 {
		return ErrImagesStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	if res, err := collImg.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		if res.DeletedCount == 0 {
			return nil
		}
		return err
	}

	return nil
}

func getReleaseSortFieldAndOrder(filt *model.ReleaseOrImageFilter) (string, int) {
	if filt != nil && filt.Sort != "" {
		sortParts := strings.SplitN(filt.Sort, ":", 2)
		if len(sortParts) == 2 &&
			(sortParts[0] == "name" ||
				sortParts[0] == "modified" ||
				sortParts[0] == "artifacts_count" ||
				sortParts[0] == "tags") {
			sortField := sortParts[0]
			sortOrder := 1
			if sortParts[1] == model.SortDirectionDescending {
				sortOrder = -1
			}
			return sortField, sortOrder
		}
	}
	return "", 0
}

// ListImages lists all images
func (db *DataStoreMongo) ListImages(
	ctx context.Context,
	filt *model.ReleaseOrImageFilter,
) ([]*model.Image, int, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collImg := database.Collection(CollectionImages)

	filters := bson.M{}
	if filt != nil {
		if filt.Name != "" {
			filters[StorageKeyImageName] = bson.M{
				"$regex": primitive.Regex{
					Pattern: ".*" + regexp.QuoteMeta(filt.Name) + ".*",
					Options: "i",
				},
			}
		}
		if filt.Description != "" {
			filters[StorageKeyImageDescription] = bson.M{
				"$regex": primitive.Regex{
					Pattern: ".*" + regexp.QuoteMeta(filt.Description) + ".*",
					Options: "i",
				},
			}
		}
		if filt.DeviceType != "" {
			filters[StorageKeyImageDeviceTypes] = bson.M{
				"$regex": primitive.Regex{
					Pattern: ".*" + regexp.QuoteMeta(filt.DeviceType) + ".*",
					Options: "i",
				},
			}
		}

	}

	projection := bson.M{
		StorageKeyImageDependsIdx:  0,
		StorageKeyImageProvidesIdx: 0,
	}
	findOptions := &mopts.FindOptions{}
	findOptions.SetProjection(projection)
	if filt != nil && filt.Page > 0 && filt.PerPage > 0 {
		findOptions.SetSkip(int64((filt.Page - 1) * filt.PerPage))
		findOptions.SetLimit(int64(filt.PerPage))
	}

	sortField, sortOrder := getReleaseSortFieldAndOrder(filt)
	if sortField == "" || sortField == "name" {
		sortField = StorageKeyImageName
	}
	if sortOrder == 0 {
		sortOrder = 1
	}
	findOptions.SetSort(bson.D{
		{Key: sortField, Value: sortOrder},
		{Key: "_id", Value: sortOrder},
	})

	cursor, err := collImg.Find(ctx, filters, findOptions)
	if err != nil {
		return nil, 0, err
	}

	// NOTE: cursor.All closes the cursor before returning
	var images []*model.Image
	if err := cursor.All(ctx, &images); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	count, err := collImg.CountDocuments(ctx, filters)
	if err != nil {
		return nil, -1, ErrDevicesCountFailed
	}

	return images, int(count), nil
}

func (db *DataStoreMongo) DeleteImagesByNames(ctx context.Context, names []string) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionImages)
	query := bson.M{
		StorageKeyImageName: bson.M{
			"$in": names,
		},
	}
	_, err := collDevs.DeleteMany(ctx, query)
	return err
}

// device deployment log
func (db *DataStoreMongo) SaveDeviceDeploymentLog(ctx context.Context,
	log model.DeploymentLog) error {

	if err := log.Validate(); err != nil {
		return err
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collLogs := database.Collection(CollectionDeviceDeploymentLogs)

	query := bson.D{
		{Key: StorageKeyDeviceDeploymentDeviceId,
			Value: log.DeviceID},
		{Key: StorageKeyDeviceDeploymentDeploymentID,
			Value: log.DeploymentID},
	}

	// update log messages
	// if the deployment log is already present than messages will be overwritten
	update := bson.D{
		{Key: "$set", Value: bson.M{
			StorageKeyDeviceDeploymentLogMessages: log.Messages,
		}},
	}
	updateOptions := mopts.Update()
	updateOptions.SetUpsert(true)
	if _, err := collLogs.UpdateOne(
		ctx, query, update, updateOptions); err != nil {
		return err
	}

	return nil
}

func (db *DataStoreMongo) GetDeviceDeploymentLog(ctx context.Context,
	deviceID, deploymentID string) (*model.DeploymentLog, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collLogs := database.Collection(CollectionDeviceDeploymentLogs)

	query := bson.M{
		StorageKeyDeviceDeploymentDeviceId:     deviceID,
		StorageKeyDeviceDeploymentDeploymentID: deploymentID,
	}

	var depl model.DeploymentLog
	if err := collLogs.FindOne(ctx, query).Decode(&depl); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &depl, nil
}

// device deployments

// Insert persists device deployment object
func (db *DataStoreMongo) InsertDeviceDeployment(
	ctx context.Context,
	deviceDeployment *model.DeviceDeployment,
	incrementDeviceCount bool,
) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	c := database.Collection(CollectionDevices)

	if deviceDeployment.Status != model.DeviceDeploymentStatusPending {
		startedTime := time.Now().UTC()
		deviceDeployment.Started = &startedTime
	}

	if _, err := c.InsertOne(ctx, deviceDeployment); err != nil {
		return err
	}

	if incrementDeviceCount {
		err := db.IncrementDeploymentDeviceCount(ctx, deviceDeployment.DeploymentId, 1)
		if err != nil {
			return err
		}
	}

	return nil
}

// InsertMany stores multiple device deployment objects.
// TODO: Handle error cleanup, multi insert is not atomic, loop into two-phase commits
func (db *DataStoreMongo) InsertMany(ctx context.Context,
	deployments ...*model.DeviceDeployment) error {

	if len(deployments) == 0 {
		return nil
	}

	deviceCountIncrements := make(map[string]int)

	// Writing to another interface list addresses golang gatcha interface{} == []interface{}
	var list []interface{}
	for _, deployment := range deployments {

		if deployment == nil {
			return ErrStorageInvalidDeviceDeployment
		}

		if err := deployment.Validate(); err != nil {
			return errors.Wrap(err, "Validating device deployment")
		}

		list = append(list, deployment)
		if deployment.Status != model.DeviceDeploymentStatusPending {
			startedTime := time.Now().UTC()
			deployment.Started = &startedTime
		}
		deviceCountIncrements[deployment.DeploymentId]++
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	if _, err := collDevs.InsertMany(ctx, list); err != nil {
		return err
	}

	for deploymentID := range deviceCountIncrements {
		err := db.IncrementDeploymentDeviceCount(
			ctx,
			deploymentID,
			deviceCountIncrements[deploymentID],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindOldestActiveDeviceDeployment finds the oldest deployment that has not finished yet.
func (db *DataStoreMongo) FindOldestActiveDeviceDeployment(
	ctx context.Context,
	deviceID string,
) (*model.DeviceDeployment, error) {

	// Verify ID formatting
	if len(deviceID) == 0 {
		return nil, ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	// Device should know only about deployments that are not finished
	query := bson.D{
		{Key: StorageKeyDeviceDeploymentActive, Value: true},
		{Key: StorageKeyDeviceDeploymentDeviceId, Value: deviceID},
		{Key: StorageKeyDeviceDeploymentDeleted, Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}

	// Find the oldest one by sorting the creation timestamp
	// in ascending order.
	findOptions := mopts.FindOne()
	findOptions.SetSort(bson.D{{Key: "created", Value: 1}})

	// Select only the oldest one that have not been finished yet.
	deployment := new(model.DeviceDeployment)
	if err := collDevs.FindOne(ctx, query, findOptions).
		Decode(deployment); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return deployment, nil
}

// FindLatestInactiveDeviceDeployment finds the latest device deployment
// matching device id that has not finished yet.
func (db *DataStoreMongo) FindLatestInactiveDeviceDeployment(
	ctx context.Context,
	deviceID string,
) (*model.DeviceDeployment, error) {

	// Verify ID formatting
	if len(deviceID) == 0 {
		return nil, ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	query := bson.D{
		{Key: StorageKeyDeviceDeploymentActive, Value: false},
		{Key: StorageKeyDeviceDeploymentDeviceId, Value: deviceID},
		{Key: StorageKeyDeviceDeploymentDeleted, Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}

	// Find the latest one by sorting by the creation timestamp
	// in ascending order.
	findOptions := mopts.FindOne()
	findOptions.SetSort(bson.D{{Key: "created", Value: -1}})

	// Select only the latest one that have not been finished yet.
	var deployment *model.DeviceDeployment
	if err := collDevs.FindOne(ctx, query, findOptions).
		Decode(&deployment); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return deployment, nil
}

func (db *DataStoreMongo) UpdateDeviceDeploymentStatus(
	ctx context.Context,
	deviceID string,
	deploymentID string,
	ddState model.DeviceDeploymentState,
	currentStatus model.DeviceDeploymentStatus,
) (model.DeviceDeploymentStatus, error) {

	// Verify ID formatting
	if len(deviceID) == 0 ||
		len(deploymentID) == 0 {
		return model.DeviceDeploymentStatusNull, ErrStorageInvalidID
	}

	if err := ddState.Validate(); err != nil {
		return model.DeviceDeploymentStatusNull, ErrStorageInvalidInput
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	// Device should know only about deployments that are not finished
	query := bson.D{
		{Key: StorageKeyDeviceDeploymentDeviceId, Value: deviceID},
		{Key: StorageKeyDeviceDeploymentDeploymentID, Value: deploymentID},
		{Key: StorageKeyDeviceDeploymentDeleted, Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}

	// update status field
	set := bson.M{
		StorageKeyDeviceDeploymentStatus: ddState.Status,
		StorageKeyDeviceDeploymentActive: ddState.Status.Active(),
	}
	// and finish time if provided
	if ddState.FinishTime != nil {
		set[StorageKeyDeviceDeploymentFinished] = ddState.FinishTime
	}

	if len(ddState.SubState) > 0 {
		set[StorageKeyDeviceDeploymentSubState] = ddState.SubState
	}

	if currentStatus == model.DeviceDeploymentStatusPending &&
		ddState.Status != currentStatus {
		startedTime := time.Now().UTC()
		set[StorageKeyDeviceDeploymentStarted] = startedTime
	}

	update := bson.D{
		{Key: "$set", Value: set},
	}

	var old model.DeviceDeployment

	if err := collDevs.FindOneAndUpdate(ctx, query, update).
		Decode(&old); err != nil {
		if err == mongo.ErrNoDocuments {
			return model.DeviceDeploymentStatusNull, ErrStorageNotFound
		}
		return model.DeviceDeploymentStatusNull, err

	}

	return old.Status, nil
}

func (db *DataStoreMongo) UpdateDeviceDeploymentLogAvailability(ctx context.Context,
	deviceID string, deploymentID string, log bool) error {

	// Verify ID formatting
	if len(deviceID) == 0 ||
		len(deploymentID) == 0 {
		return ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	selector := bson.D{
		{Key: StorageKeyDeviceDeploymentDeviceId, Value: deviceID},
		{Key: StorageKeyDeviceDeploymentDeploymentID, Value: deploymentID},
		{Key: StorageKeyDeviceDeploymentDeleted, Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}

	update := bson.D{
		{Key: "$set", Value: bson.M{
			StorageKeyDeviceDeploymentIsLogAvailable: log}},
	}

	if res, err := collDevs.UpdateOne(ctx, selector, update); err != nil {
		return err
	} else if res.MatchedCount == 0 {
		return ErrStorageNotFound
	}

	return nil
}

// SaveDeviceDeploymentRequest saves device deployment request
// with the device deployment object
func (db *DataStoreMongo) SaveDeviceDeploymentRequest(
	ctx context.Context,
	ID string,
	request *model.DeploymentNextRequest,
) error {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	res, err := collDevs.UpdateOne(
		ctx,
		bson.D{{Key: StorageKeyId, Value: ID}},
		bson.D{{Key: "$set", Value: bson.M{StorageKeyDeviceDeploymentRequest: request}}},
	)
	if err != nil {
		return err
	} else if res.MatchedCount == 0 {
		return ErrStorageNotFound
	}
	return nil
}

// AssignArtifact assigns artifact to the device deployment
func (db *DataStoreMongo) AssignArtifact(
	ctx context.Context,
	deviceID string,
	deploymentID string,
	artifact *model.Image,
) error {

	// Verify ID formatting
	if len(deviceID) == 0 ||
		len(deploymentID) == 0 {
		return ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	selector := bson.D{
		{Key: StorageKeyDeviceDeploymentDeviceId, Value: deviceID},
		{Key: StorageKeyDeviceDeploymentDeploymentID, Value: deploymentID},
		{Key: StorageKeyDeviceDeploymentDeleted, Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}

	update := bson.D{
		{Key: "$set", Value: bson.M{
			StorageKeyDeviceDeploymentArtifact: artifact,
		}},
	}

	if res, err := collDevs.UpdateOne(ctx, selector, update); err != nil {
		return err
	} else if res.MatchedCount == 0 {
		return ErrStorageNotFound
	}

	return nil
}

func (db *DataStoreMongo) AggregateDeviceDeploymentByStatus(ctx context.Context,
	id string) (model.Stats, error) {

	if len(id) == 0 {
		return nil, ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	match := bson.D{
		{Key: "$match", Value: bson.M{
			StorageKeyDeviceDeploymentDeploymentID: id,
			StorageKeyDeviceDeploymentDeleted: bson.D{
				{Key: "$exists", Value: false},
			},
		}},
	}
	group := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id",
				Value: "$" + StorageKeyDeviceDeploymentStatus},
			{Key: "count",
				Value: bson.M{"$sum": 1}}},
		},
	}
	pipeline := []bson.D{
		match,
		group,
	}
	var results []struct {
		Status model.DeviceDeploymentStatus `bson:"_id"`
		Count  int
	}
	cursor, err := collDevs.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &results); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	raw := model.NewDeviceDeploymentStats()
	for _, res := range results {
		raw.Set(res.Status, res.Count)
	}
	return raw, nil
}

// GetDeviceStatusesForDeployment retrieve device deployment statuses for a given deployment.
func (db *DataStoreMongo) GetDeviceStatusesForDeployment(ctx context.Context,
	deploymentID string) ([]model.DeviceDeployment, error) {

	statuses := []model.DeviceDeployment{}
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	query := bson.M{
		StorageKeyDeviceDeploymentDeploymentID: deploymentID,
		StorageKeyDeviceDeploymentDeleted: bson.D{
			{Key: "$exists", Value: false},
		},
	}

	cursor, err := collDevs.Find(ctx, query)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &statuses); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return statuses, nil
}

func (db *DataStoreMongo) GetDevicesListForDeployment(ctx context.Context,
	q store.ListQuery) ([]model.DeviceDeployment, int, error) {

	statuses := []model.DeviceDeployment{}
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	query := bson.D{
		{Key: StorageKeyDeviceDeploymentDeploymentID, Value: q.DeploymentID},
		{Key: StorageKeyDeviceDeploymentDeleted, Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}
	if q.Status != nil {
		if *q.Status == model.DeviceDeploymentStatusPauseStr {
			query = append(query, bson.E{
				Key: "status", Value: bson.D{{
					Key:   "$gte",
					Value: model.DeviceDeploymentStatusPauseBeforeInstall,
				}, {
					Key:   "$lte",
					Value: model.DeviceDeploymentStatusPauseBeforeReboot,
				}},
			})
		} else if *q.Status == model.DeviceDeploymentStatusActiveStr {
			query = append(query, bson.E{
				Key: "status", Value: bson.D{{
					Key:   "$gte",
					Value: model.DeviceDeploymentStatusPauseBeforeInstall,
				}, {
					Key:   "$lte",
					Value: model.DeviceDeploymentStatusPending,
				}},
			})
		} else if *q.Status == model.DeviceDeploymentStatusFinishedStr {
			query = append(query, bson.E{
				Key: "status", Value: bson.D{{
					Key: "$in",
					Value: []model.DeviceDeploymentStatus{
						model.DeviceDeploymentStatusFailure,
						model.DeviceDeploymentStatusAborted,
						model.DeviceDeploymentStatusSuccess,
						model.DeviceDeploymentStatusNoArtifact,
						model.DeviceDeploymentStatusAlreadyInst,
						model.DeviceDeploymentStatusDecommissioned,
					},
				}},
			})
		} else {
			var status model.DeviceDeploymentStatus
			err := status.UnmarshalText([]byte(*q.Status))
			if err != nil {
				return nil, -1, errors.Wrap(err, "invalid status query")
			}
			query = append(query, bson.E{
				Key: "status", Value: status,
			})
		}
	}

	options := mopts.Find()
	sortFieldQuery := bson.D{
		{Key: StorageKeyDeviceDeploymentStatus, Value: 1},
		{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
	}
	options.SetSort(sortFieldQuery)
	if q.Skip > 0 {
		options.SetSkip(int64(q.Skip))
	}
	if q.Limit > 0 {
		options.SetLimit(int64(q.Limit))
	} else {
		options.SetLimit(DefaultDocumentLimit)
	}

	cursor, err := collDevs.Find(ctx, query, options)
	if err != nil {
		return nil, -1, err
	}

	if err = cursor.All(ctx, &statuses); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, -1, nil
		}
		return nil, -1, err
	}

	count, err := collDevs.CountDocuments(ctx, query)
	if err != nil {
		return nil, -1, ErrDevicesCountFailed
	}

	return statuses, int(count), nil
}

func (db *DataStoreMongo) GetDeviceDeploymentsForDevice(ctx context.Context,
	q store.ListQueryDeviceDeployments) ([]model.DeviceDeployment, int, error) {

	statuses := []model.DeviceDeployment{}
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	query := bson.D{}
	if q.DeviceID != "" {
		query = append(query, bson.E{
			Key:   StorageKeyDeviceDeploymentDeviceId,
			Value: q.DeviceID,
		})
	} else if len(q.IDs) > 0 {
		query = append(query, bson.E{
			Key: StorageKeyId,
			Value: bson.D{{
				Key:   "$in",
				Value: q.IDs,
			}},
		})
	}

	if q.Status != nil {
		if *q.Status == model.DeviceDeploymentStatusPauseStr {
			query = append(query, bson.E{
				Key: "status", Value: bson.D{{
					Key:   "$gte",
					Value: model.DeviceDeploymentStatusPauseBeforeInstall,
				}, {
					Key:   "$lte",
					Value: model.DeviceDeploymentStatusPauseBeforeReboot,
				}},
			})
		} else if *q.Status == model.DeviceDeploymentStatusActiveStr {
			query = append(query, bson.E{
				Key: "status", Value: bson.D{{
					Key:   "$gte",
					Value: model.DeviceDeploymentStatusPauseBeforeInstall,
				}, {
					Key:   "$lte",
					Value: model.DeviceDeploymentStatusPending,
				}},
			})
		} else if *q.Status == model.DeviceDeploymentStatusFinishedStr {
			query = append(query, bson.E{
				Key: "status", Value: bson.D{{
					Key: "$in",
					Value: []model.DeviceDeploymentStatus{
						model.DeviceDeploymentStatusFailure,
						model.DeviceDeploymentStatusAborted,
						model.DeviceDeploymentStatusSuccess,
						model.DeviceDeploymentStatusNoArtifact,
						model.DeviceDeploymentStatusAlreadyInst,
						model.DeviceDeploymentStatusDecommissioned,
					},
				}},
			})
		} else {
			var status model.DeviceDeploymentStatus
			err := status.UnmarshalText([]byte(*q.Status))
			if err != nil {
				return nil, -1, errors.Wrap(err, "invalid status query")
			}
			query = append(query, bson.E{
				Key: "status", Value: status,
			})
		}
	}

	options := mopts.Find()
	sortFieldQuery := bson.D{
		{Key: StorageKeyDeviceDeploymentCreated, Value: -1},
		{Key: StorageKeyDeviceDeploymentStatus, Value: -1},
	}
	options.SetSort(sortFieldQuery)
	if q.Skip > 0 {
		options.SetSkip(int64(q.Skip))
	}
	if q.Limit > 0 {
		options.SetLimit(int64(q.Limit))
	} else {
		options.SetLimit(DefaultDocumentLimit)
	}

	cursor, err := collDevs.Find(ctx, query, options)
	if err != nil {
		return nil, -1, err
	}

	if err = cursor.All(ctx, &statuses); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, 0, nil
		}
		return nil, -1, err
	}

	maxCount := maxCountDocuments
	countOptions := &mopts.CountOptions{
		Limit: &maxCount,
	}
	count, err := collDevs.CountDocuments(ctx, query, countOptions)
	if err != nil {
		return nil, -1, ErrDevicesCountFailed
	}

	return statuses, int(count), nil
}

// Returns true if deployment of ID `deploymentID` is assigned to device with ID
// `deviceID`, false otherwise. In case of errors returns false and an error
// that occurred
func (db *DataStoreMongo) HasDeploymentForDevice(ctx context.Context,
	deploymentID string, deviceID string) (bool, error) {

	var dep model.DeviceDeployment
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	query := bson.D{
		{Key: StorageKeyDeviceDeploymentDeploymentID, Value: deploymentID},
		{Key: StorageKeyDeviceDeploymentDeviceId, Value: deviceID},
		{Key: StorageKeyDeviceDeploymentDeleted, Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}

	if err := collDevs.FindOne(ctx, query).Decode(&dep); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func (db *DataStoreMongo) AbortDeviceDeployments(ctx context.Context,
	deploymentId string) error {

	if len(deploymentId) == 0 {
		return ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)
	selector := bson.M{
		StorageKeyDeviceDeploymentDeploymentID: deploymentId,
		StorageKeyDeviceDeploymentActive:       true,
		StorageKeyDeviceDeploymentDeleted: bson.D{
			{Key: "$exists", Value: false},
		},
	}

	update := bson.M{
		"$set": bson.M{
			StorageKeyDeviceDeploymentStatus: model.DeviceDeploymentStatusAborted,
			StorageKeyDeviceDeploymentActive: false,
		},
	}

	if _, err := collDevs.UpdateMany(ctx, selector, update); err != nil {
		return err
	}

	return nil
}

func (db *DataStoreMongo) DeleteDeviceDeploymentsHistory(ctx context.Context,
	deviceID string) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)
	selector := bson.M{
		StorageKeyDeviceDeploymentDeviceId: deviceID,
		StorageKeyDeviceDeploymentActive:   false,
		StorageKeyDeviceDeploymentDeleted: bson.M{
			"$exists": false,
		},
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			StorageKeyDeviceDeploymentDeleted: &now,
		},
	}

	if _, err := collDevs.UpdateMany(ctx, selector, update); err != nil {
		return err
	}

	database = db.client.Database(DatabaseName)
	collDevs = database.Collection(CollectionDevicesLastStatus)
	_, err := collDevs.DeleteMany(ctx, bson.M{StorageKeyDeviceDeploymentDeviceId: deviceID})

	return err
}

func (db *DataStoreMongo) DecommissionDeviceDeployments(ctx context.Context,
	deviceId string) error {

	if len(deviceId) == 0 {
		return ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)
	selector := bson.M{
		StorageKeyDeviceDeploymentDeviceId: deviceId,
		StorageKeyDeviceDeploymentActive:   true,
		StorageKeyDeviceDeploymentDeleted: bson.D{
			{Key: "$exists", Value: false},
		},
	}

	update := bson.M{
		"$set": bson.M{
			StorageKeyDeviceDeploymentStatus: model.DeviceDeploymentStatusDecommissioned,
			StorageKeyDeviceDeploymentActive: false,
		},
	}

	if _, err := collDevs.UpdateMany(ctx, selector, update); err != nil {
		return err
	}

	return nil
}

func (db *DataStoreMongo) GetDeviceDeployment(ctx context.Context, deploymentID string,
	deviceID string, includeDeleted bool) (*model.DeviceDeployment, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	filter := bson.M{
		StorageKeyDeviceDeploymentDeploymentID: deploymentID,
		StorageKeyDeviceDeploymentDeviceId:     deviceID,
	}
	if !includeDeleted {
		filter[StorageKeyDeviceDeploymentDeleted] = bson.D{
			{Key: "$exists", Value: false},
		}
	}

	opts := &mopts.FindOneOptions{}
	opts.SetSort(bson.D{{Key: "created", Value: -1}})

	var dd model.DeviceDeployment
	if err := collDevs.FindOne(ctx, filter, opts).Decode(&dd); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}

	return &dd, nil
}

func (db *DataStoreMongo) GetDeviceDeployments(
	ctx context.Context,
	skip int,
	limit int,
	deviceID string,
	active *bool,
	includeDeleted bool,
) ([]model.DeviceDeployment, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	filter := bson.M{}
	if !includeDeleted {
		filter[StorageKeyDeviceDeploymentDeleted] = bson.D{
			{Key: "$exists", Value: false},
		}
	}
	if deviceID != "" {
		filter[StorageKeyDeviceDeploymentDeviceId] = deviceID
	}
	if active != nil {
		filter[StorageKeyDeviceDeploymentActive] = *active
	}

	opts := &mopts.FindOptions{}
	opts.SetSort(bson.D{{Key: "created", Value: -1}})
	if skip > 0 {
		opts.SetSkip(int64(skip))
	}
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	var deviceDeployments []model.DeviceDeployment
	cursor, err := collDevs.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &deviceDeployments); err != nil {
		return nil, err
	}

	return deviceDeployments, nil
}

// deployments

func (db *DataStoreMongo) EnsureIndexes(dbName string, collName string,
	indexes ...mongo.IndexModel) error {
	ctx := context.Background()
	dataBase := db.client.Database(dbName)

	coll := dataBase.Collection(collName)
	idxView := coll.Indexes()
	_, err := idxView.CreateMany(ctx, indexes)
	return err
}

// return true if required indexing was set up
func (db *DataStoreMongo) hasIndexing(ctx context.Context, client *mongo.Client) bool {

	var idx bson.M
	database := client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)
	idxView := collDpl.Indexes()

	cursor, err := idxView.List(ctx)
	if err != nil {
		// check failed, assume indexing is not there
		return false
	}

	has := map[string]bool{}
	for cursor.Next(ctx) {
		if err = cursor.Decode(&idx); err != nil {
			continue
		}
		if _, ok := idx["weights"]; ok {
			// text index
			for k := range idx["weights"].(bson.M) {
				has[k] = true
			}
		} else {
			for i := range idx["key"].(bson.M) {
				has[i] = true
			}

		}
	}
	if err != nil {
		return false
	}

	for _, key := range StorageIndexes.Keys.(bson.D) {
		_, ok := has[key.Key]
		if !ok {
			return false
		}
	}

	return true
}

// Insert persists object
func (db *DataStoreMongo) InsertDeployment(
	ctx context.Context,
	deployment *model.Deployment,
) error {

	if deployment == nil {
		return ErrDeploymentStorageInvalidDeployment
	}

	if err := deployment.Validate(); err != nil {
		return err
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	if _, err := collDpl.InsertOne(ctx, deployment); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrConflictingDeployment
		}
		return err
	}
	return nil
}

// Delete removed entry by ID
// Noop on ID not found
func (db *DataStoreMongo) DeleteDeployment(ctx context.Context, id string) error {

	if len(id) == 0 {
		return ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	if _, err := collDpl.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return err
	}

	return nil
}

func (db *DataStoreMongo) FindDeploymentByID(
	ctx context.Context,
	id string,
) (*model.Deployment, error) {

	if len(id) == 0 {
		return nil, ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	deployment := new(model.Deployment)
	if err := collDpl.FindOne(ctx, bson.M{"_id": id}).
		Decode(deployment); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return deployment, nil
}

func (db *DataStoreMongo) FindDeploymentStatsByIDs(
	ctx context.Context,
	ids ...string,
) (deploymentStats []*model.DeploymentStats, err error) {

	if len(ids) == 0 {
		return nil, errors.New("no IDs passed into the function. At least one is required")
	}

	for _, id := range ids {
		if len(id) == 0 {
			return nil, ErrStorageInvalidID
		}
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	query := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	statsProjection := &mopts.FindOptions{
		Projection: bson.M{"stats": 1},
	}

	results, err := collDpl.Find(
		ctx,
		query,
		statsProjection,
	)
	if err != nil {
		return nil, err
	}

	for results.Next(context.Background()) {
		depl := new(model.DeploymentStats)
		if err = results.Decode(&depl); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, nil
			}
			return nil, err
		}
		deploymentStats = append(deploymentStats, depl)
	}

	return deploymentStats, nil
}

func (db *DataStoreMongo) FindUnfinishedByID(ctx context.Context,
	id string) (*model.Deployment, error) {

	if len(id) == 0 {
		return nil, ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	var deployment *model.Deployment
	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: StorageKeyDeploymentFinished, Value: nil},
	}
	if err := collDpl.FindOne(ctx, filter).
		Decode(&deployment); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return deployment, nil
}

func (db *DataStoreMongo) IncrementDeploymentDeviceCount(
	ctx context.Context,
	deploymentID string,
	increment int,
) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collection := database.Collection(CollectionDeployments)

	filter := bson.M{
		"_id": deploymentID,
		StorageKeyDeploymentDeviceCount: bson.M{
			"$ne": nil,
		},
	}

	update := bson.M{
		"$inc": bson.M{
			StorageKeyDeploymentDeviceCount: increment,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (db *DataStoreMongo) SetDeploymentDeviceCount(
	ctx context.Context,
	deploymentID string,
	count int,
) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collection := database.Collection(CollectionDeployments)

	filter := bson.M{
		"_id": deploymentID,
		StorageKeyDeploymentDeviceCount: bson.M{
			"$eq": nil,
		},
	}

	update := bson.M{
		"$set": bson.M{
			StorageKeyDeploymentDeviceCount: count,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (db *DataStoreMongo) DeviceCountByDeployment(ctx context.Context,
	id string) (int, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevices)

	filter := bson.M{
		StorageKeyDeviceDeploymentDeploymentID: id,
		StorageKeyDeviceDeploymentDeleted: bson.D{
			{Key: "$exists", Value: false},
		},
	}

	deviceCount, err := collDevs.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(deviceCount), nil
}

func (db *DataStoreMongo) UpdateStats(ctx context.Context,
	id string, stats model.Stats) error {

	if len(id) == 0 {
		return ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	deployment, err := model.NewDeployment()
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	deployment.Stats = stats
	var update bson.M
	if deployment.IsFinished() {
		now := time.Now()

		update = bson.M{
			"$set": bson.M{
				StorageKeyDeploymentStats:    stats,
				StorageKeyDeploymentFinished: &now,
			},
		}
	} else {
		update = bson.M{
			"$set": bson.M{
				StorageKeyDeploymentStats: stats,
			},
		}
	}

	res, err := collDpl.UpdateOne(ctx, bson.M{"_id": id}, update)
	if res != nil && res.MatchedCount == 0 {
		return ErrStorageInvalidID
	}
	return err
}

func (db *DataStoreMongo) UpdateStatsInc(ctx context.Context, id string,
	stateFrom, stateTo model.DeviceDeploymentStatus) (model.Stats, error) {

	if len(id) == 0 {
		return nil, ErrStorageInvalidID
	}

	if _, err := stateTo.MarshalText(); err != nil {
		return nil, ErrStorageInvalidInput
	}

	// does not need any extra operations
	// following query won't handle this case well and increase the state_to value
	if stateFrom == stateTo {
		return nil, nil
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	var update bson.M

	if stateFrom == model.DeviceDeploymentStatusNull {
		// note dot notation on embedded document
		update = bson.M{
			"$inc": bson.M{
				"stats." + stateTo.String(): 1,
			},
		}
	} else {
		// note dot notation on embedded document
		update = bson.M{
			"$inc": bson.M{
				"stats." + stateFrom.String(): -1,
				"stats." + stateTo.String():   1,
			},
		}
	}

	var res struct {
		Stats model.Stats `bson:"stats"`
	}
	err := collDpl.FindOneAndUpdate(ctx,
		bson.M{StorageKeyId: id},
		update,
		mopts.FindOneAndUpdate().
			SetReturnDocument(mopts.After).
			SetProjection(bson.M{
				StorageKeyDeploymentStats: 1,
			}),
	).Decode(&res)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrStorageInvalidID
	}

	return res.Stats, err
}

func (db *DataStoreMongo) IncrementDeploymentTotalSize(
	ctx context.Context,
	deploymentID string,
	increment int64,
) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collection := database.Collection(CollectionDeployments)

	filter := bson.M{
		"_id": deploymentID,
	}

	update := bson.M{
		"$inc": bson.M{
			StorageKeyDeploymentTotalSize: increment,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (db *DataStoreMongo) FindDeployments(ctx context.Context,
	match model.Query) ([]*model.Deployment, int64, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	query, err := db.buildDeploymentsQuery(ctx, match)
	if err != nil {
		return nil, 0, err
	}

	options := db.findOptions(match)

	var deployments []*model.Deployment
	cursor, err := collDpl.Find(ctx, query, options)
	if err != nil {
		return nil, 0, err
	}
	if err := cursor.All(ctx, &deployments); err != nil {
		return nil, 0, err
	}
	// Count documents if we didn't find all already.
	count := int64(0)
	if !match.DisableCount {
		count = int64(len(deployments))
		if count >= int64(match.Limit) {
			count, err = collDpl.CountDocuments(ctx, query)
			if err != nil {
				return nil, 0, err
			}
		} else {
			// Don't forget to add the skipped documents
			count += int64(match.Skip)
		}
	}

	return deployments, count, nil
}

func (db *DataStoreMongo) buildDeploymentsQuery(
	ctx context.Context,
	match model.Query,
) (bson.M, error) {
	andq := []bson.M{}

	// filter by IDs
	if match.IDs != nil {
		tq := bson.M{
			"_id": bson.M{
				"$in": match.IDs,
			},
		}
		andq = append(andq, tq)
	}

	// filter by Names
	if match.Names != nil {
		tq := bson.M{
			StorageKeyDeploymentName: bson.M{
				"$in": match.Names,
			},
		}
		andq = append(andq, tq)
	}

	// build deployment by name part of the query
	if match.SearchText != "" {
		// we must have indexing for text search
		if !db.hasIndexing(ctx, db.client) {
			return nil, ErrDeploymentStorageCannotExecQuery
		}

		tq := bson.M{
			"$text": bson.M{
				"$search": "\"" + match.SearchText + "\"",
			},
		}

		andq = append(andq, tq)
	}

	// build deployment by status part of the query
	if match.Status != model.StatusQueryAny {
		var status model.DeploymentStatus
		if match.Status == model.StatusQueryPending {
			status = model.DeploymentStatusPending
		} else if match.Status == model.StatusQueryInProgress {
			status = model.DeploymentStatusInProgress
		} else {
			status = model.DeploymentStatusFinished
		}
		stq := bson.M{StorageKeyDeploymentStatus: status}
		andq = append(andq, stq)
	}

	// build deployment by type part of the query
	if match.Type != "" {
		if match.Type == model.DeploymentTypeConfiguration {
			andq = append(andq, bson.M{StorageKeyDeploymentType: match.Type})
		} else if match.Type == model.DeploymentTypeSoftware {
			andq = append(andq, bson.M{
				"$or": []bson.M{
					{StorageKeyDeploymentType: match.Type},
					{StorageKeyDeploymentType: ""},
				},
			})
		}
	}

	query := bson.M{}
	if len(andq) != 0 {
		// use search criteria if any
		query = bson.M{
			"$and": andq,
		}
	}

	if match.CreatedAfter != nil && match.CreatedBefore != nil {
		query["created"] = bson.M{
			"$gte": match.CreatedAfter,
			"$lte": match.CreatedBefore,
		}
	} else if match.CreatedAfter != nil {
		query["created"] = bson.M{
			"$gte": match.CreatedAfter,
		}
	} else if match.CreatedBefore != nil {
		query["created"] = bson.M{
			"$lte": match.CreatedBefore,
		}
	}

	return query, nil
}

func (db *DataStoreMongo) findOptions(match model.Query) *mopts.FindOptions {
	options := &mopts.FindOptions{}
	if match.Sort == model.SortDirectionAscending {
		options.SetSort(bson.D{{Key: "created", Value: 1}})
	} else {
		options.SetSort(bson.D{{Key: "created", Value: -1}})
	}
	if match.Skip > 0 {
		options.SetSkip(int64(match.Skip))
	}
	if match.Limit > 0 {
		options.SetLimit(int64(match.Limit))
	} else {
		options.SetLimit(DefaultDocumentLimit)
	}
	return options
}

// FindNewerActiveDeployments finds active deployments which were created
// after createdAfter
// Deprecated: No longer in use
func (db *DataStoreMongo) FindNewerActiveDeployments(ctx context.Context,
	createdAfter *time.Time, skip, limit int) ([]*model.Deployment, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	c := database.Collection(CollectionDeployments)

	queryFilters := make([]bson.M, 0)
	queryFilters = append(queryFilters, bson.M{StorageKeyDeploymentActive: true})
	queryFilters = append(queryFilters,
		bson.M{StorageKeyDeploymentCreated: bson.M{"$gt": createdAfter}})
	findQuery := bson.M{}
	findQuery["$and"] = queryFilters

	findOptions := &mopts.FindOptions{}
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	findOptions.SetSort(bson.D{{Key: StorageKeyDeploymentCreated, Value: 1}})
	cursor, err := c.Find(ctx, findQuery, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployments")
	}
	defer cursor.Close(ctx)

	var deployments []*model.Deployment

	if err = cursor.All(ctx, &deployments); err != nil {
		return nil, errors.Wrap(err, "failed to get deployments")
	}

	return deployments, nil
}

// FindNewerActiveDeployment finds active deployments which were created
// after createdAfter where deviceID is part of the device list.
func (db *DataStoreMongo) FindNewerActiveDeployment(ctx context.Context,
	createdAfter *time.Time, deviceID string) (*model.Deployment, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	c := database.Collection(CollectionDeployments)

	findQuery := bson.D{
		{Key: StorageKeyDeploymentActive, Value: true},
		{Key: StorageKeyDeploymentCreated, Value: bson.M{"$gt": createdAfter}},
		{Key: StorageKeyDeploymentDeviceList, Value: deviceID},
	}
	findOptions := mopts.FindOne().
		SetSort(bson.D{{Key: StorageKeyDeploymentCreated, Value: 1}}).
		SetProjection(bson.M{
			// Discard information we don't need
			StorageKeyDeploymentConstructorChecksum: 0,
			StorageKeyDeploymentDeviceList:          0,
		})

	var deployment = new(model.Deployment)
	err := c.FindOne(ctx, findQuery, findOptions).
		Decode(deployment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get deployments")
	}

	return deployment, nil
}

// SetDeploymentStatus simply sets the status field
// optionally sets 'finished time' if deployment is indeed finished
func (db *DataStoreMongo) SetDeploymentStatus(
	ctx context.Context,
	id string,
	status model.DeploymentStatus,
	now time.Time,
) error {
	if len(id) == 0 {
		return ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	var update bson.M
	if status == model.DeploymentStatusFinished {
		update = bson.M{
			"$set": bson.M{
				StorageKeyDeploymentActive:   false,
				StorageKeyDeploymentStatus:   status,
				StorageKeyDeploymentFinished: &now,
			},
		}
	} else {
		update = bson.M{
			"$set": bson.M{
				StorageKeyDeploymentActive: true,
				StorageKeyDeploymentStatus: status,
			},
		}
	}

	res, err := collDpl.UpdateOne(ctx, bson.M{"_id": id}, update)

	if res != nil && res.MatchedCount == 0 {
		return ErrStorageInvalidID
	}

	return err
}

// ExistUnfinishedByArtifactId checks if there is an active deployment that uses
// given artifact
func (db *DataStoreMongo) ExistUnfinishedByArtifactId(ctx context.Context,
	id string) (bool, error) {

	if len(id) == 0 {
		return false, ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	var tmp interface{}
	query := bson.D{
		{Key: StorageKeyDeploymentFinished, Value: nil},
		{Key: StorageKeyDeploymentArtifacts, Value: id},
	}
	if err := collDpl.FindOne(ctx, query).Decode(&tmp); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ExistUnfinishedByArtifactName checks if there is an active deployment that uses
// given artifact
func (db *DataStoreMongo) ExistUnfinishedByArtifactName(ctx context.Context,
	artifactName string) (bool, error) {

	if len(artifactName) == 0 {
		return false, ErrImagesStorageInvalidArtifactName
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	var tmp interface{}
	query := bson.D{
		{Key: StorageKeyDeploymentFinished, Value: nil},
		{Key: StorageKeyDeploymentArtifactName, Value: artifactName},
	}

	projection := bson.M{
		"_id": 1,
	}
	findOptions := mopts.FindOne()
	findOptions.SetProjection(projection)

	if err := collDpl.FindOne(ctx, query, findOptions).Decode(&tmp); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ExistByArtifactId check if there is any deployment that uses give artifact
func (db *DataStoreMongo) ExistByArtifactId(ctx context.Context,
	id string) (bool, error) {

	if len(id) == 0 {
		return false, ErrStorageInvalidID
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	var tmp interface{}
	query := bson.D{
		{Key: StorageKeyDeploymentArtifacts, Value: id},
	}
	if err := collDpl.FindOne(ctx, query).Decode(&tmp); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Per-tenant storage settings
func (db *DataStoreMongo) GetStorageSettings(ctx context.Context) (*model.StorageSettings, error) {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collection := database.Collection(CollectionStorageSettings)

	settings := new(model.StorageSettings)
	// supposed that it's only one document in the collection
	query := bson.M{
		"_id": StorageKeyStorageSettingsDefaultID,
	}
	if err := collection.FindOne(ctx, query).Decode(settings); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return settings, nil
}

func (db *DataStoreMongo) SetStorageSettings(
	ctx context.Context,
	storageSettings *model.StorageSettings,
) error {
	var err error
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collection := database.Collection(CollectionStorageSettings)

	filter := bson.M{
		"_id": StorageKeyStorageSettingsDefaultID,
	}
	if storageSettings != nil {
		replaceOptions := mopts.Replace()
		replaceOptions.SetUpsert(true)
		_, err = collection.ReplaceOne(ctx, filter, storageSettings, replaceOptions)
	} else {
		_, err = collection.DeleteOne(ctx, filter)
	}

	return err
}

func (db *DataStoreMongo) UpdateDeploymentsWithArtifactName(
	ctx context.Context,
	artifactName string,
	artifactIDs []string,
) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	query := bson.D{
		{Key: StorageKeyDeploymentFinished, Value: nil},
		{Key: StorageKeyDeploymentArtifactName, Value: artifactName},
	}
	update := bson.M{
		"$set": bson.M{
			StorageKeyDeploymentArtifacts: artifactIDs,
		},
	}

	_, err := collDpl.UpdateMany(ctx, query, update)
	return err
}

func (db *DataStoreMongo) GetDeploymentIDsByArtifactNames(
	ctx context.Context,
	artifactNames []string,
) ([]string, error) {

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDpl := database.Collection(CollectionDeployments)

	query := bson.M{
		StorageKeyDeploymentArtifactName: bson.M{
			"$in": artifactNames,
		},
	}

	projection := bson.M{
		"_id": 1,
	}
	findOptions := mopts.Find()
	findOptions.SetProjection(projection)

	cursor, err := collDpl.Find(ctx, query, findOptions)
	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(ctx)

	var deployments []*model.Deployment
	if err = cursor.All(ctx, &deployments); err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
		}
		return []string{}, err
	}

	ids := make([]string, len(deployments))
	for i, d := range deployments {
		ids[i] = d.Id
	}

	return ids, nil
}

func (db *DataStoreMongo) GetTenantDbs() ([]string, error) {
	return migrate.GetTenantDbs(context.Background(), db.client, mstore.IsTenantDb(DbName))
}
