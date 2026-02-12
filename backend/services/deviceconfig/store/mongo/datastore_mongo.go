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
	"crypto/tls"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/codec"
	mstore "github.com/mendersoftware/mender-server/pkg/store/v2"

	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
)

const (
	// CollDevices refers to the collection name for device configurations
	CollDevices = "devices"
	// fields
	fieldID           = "_id"
	fieldConfigured   = "configured"
	fieldReported     = "reported"
	fieldUpdatedTs    = "updated_ts"
	fieldReportedTs   = "reported_ts"
	fieldDeploymentID = "deployment_id"

	KeyTenantID = "tenant_id"
)

type MongoStoreConfig struct {
	// MongoURL holds the URL to the MongoDB server.
	MongoURL *url.URL
	// TLSConfig holds optional tls configuration options for connecting
	// to the MongoDB server.
	TLSConfig *tls.Config
	// Username holds the user id credential for authenticating with the
	// MongoDB server.
	Username string
	// Password holds the password credential for authenticating with the
	// MongoDB server.
	Password string

	// DbName contains the name of the deviceconfig database.
	DbName string
}

// newClient returns a mongo client
func newClient(ctx context.Context, config MongoStoreConfig) (*mongo.Client, error) {

	clientOptions := mopts.Client()
	if config.MongoURL == nil {
		return nil, errors.New("mongo: missing URL")
	}
	clientOptions.ApplyURI(config.MongoURL.String()).
		SetRegistry(codec.NewRegistry())

	if config.Username != "" {
		credentials := mopts.Credential{
			Username: config.Username,
		}
		if config.Password != "" {
			credentials.Password = config.Password
			credentials.PasswordSet = true
		}
		clientOptions.SetAuth(credentials)
	}

	if config.TLSConfig != nil {
		clientOptions.SetTLSConfig(config.TLSConfig)
	}

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "mongo: failed to connect with server")
	}

	// Validate connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "mongo: error reaching mongo server")
	}

	return client, nil
}

// MongoStore is the data storage service
type MongoStore struct {
	// client holds the reference to the client used to communicate with the
	// mongodb server.
	client *mongo.Client

	config MongoStoreConfig
}

// SetupDataStore returns the mongo data store and optionally runs migrations
func NewMongoStore(ctx context.Context, config MongoStoreConfig) (*MongoStore, error) {
	dbClient, err := newClient(ctx, config)
	if err != nil {
		return nil, err
	}
	return &MongoStore{
		client: dbClient,
		config: config,
	}, nil
}

func (db *MongoStore) Database(ctx context.Context,
	opt ...mopts.Lister[mopts.DatabaseOptions]) *mongo.Database {
	return db.client.Database(db.config.DbName, opt...)
}

// Ping verifies the connection to the database
func (db *MongoStore) Ping(ctx context.Context) error {
	return db.client.Ping(ctx, nil)
}

// Close disconnects the client
func (db *MongoStore) Close(ctx context.Context) error {
	err := db.client.Disconnect(ctx)
	return err
}

//nolint:unused
func (db *MongoStore) DropDatabase(ctx context.Context) error {
	err := db.client.
		Database(db.config.DbName).
		Drop(ctx)
	return err
}

func (db *MongoStore) InsertDevice(ctx context.Context, dev model.Device) error {
	if err := dev.Validate(); err != nil {
		return err
	}
	collDevs := db.Database(ctx).Collection(CollDevices)

	_, err := collDevs.InsertOne(ctx, mstore.WithTenantID(ctx, dev))
	if mongo.IsDuplicateKeyError(err) {
		return store.ErrDeviceAlreadyExists
	}

	return errors.Wrap(err, "mongo: failed to store device configuration")
}

func (db *MongoStore) ReplaceConfiguration(ctx context.Context, dev model.Device) error {
	if err := dev.Validate(); err != nil {
		return err
	}
	collDevs := db.Database(ctx).Collection(CollDevices)

	fltr := bson.D{{
		Key:   fieldID,
		Value: dev.ID,
	}}

	update := bson.M{
		"$set": bson.D{
			{
				Key:   fieldConfigured,
				Value: dev.ConfiguredAttributes,
			},
			{
				Key:   fieldUpdatedTs,
				Value: time.Now().UTC(),
			},
		},
	}
	mopts.UpdateOne()
	_, err := collDevs.UpdateOne(ctx,
		mstore.WithTenantID(ctx, fltr),
		update,
		mopts.UpdateOne().SetUpsert(true))

	return errors.Wrap(err, "mongo: failed to store device configuration")
}

func (db *MongoStore) ReplaceReportedConfiguration(ctx context.Context, dev model.Device) error {
	if err := dev.Validate(); err != nil {
		return err
	}
	collDevs := db.Database(ctx).Collection(CollDevices)

	fltr := bson.D{{
		Key:   fieldID,
		Value: dev.ID,
	}}

	update := bson.M{
		"$set": bson.D{
			{
				Key:   fieldReported,
				Value: dev.ReportedAttributes,
			},
			{
				Key:   fieldReportedTs,
				Value: time.Now().UTC(),
			},
		},
	}

	_, err := collDevs.UpdateOne(ctx,
		mstore.WithTenantID(ctx, fltr),
		update,
		mopts.UpdateOne().SetUpsert(true))
	return errors.Wrap(err, "mongo: failed to store device reported configuration")
}

func (db *MongoStore) UpdateConfiguration(
	ctx context.Context,
	devID string,
	attrs model.Attributes,
) error {
	if len(attrs) == 0 {
		return nil
	} else if err := attrs.Validate(); err != nil {
		return err
	}
	var tenantID string
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	collDevs := db.Database(ctx).Collection(CollDevices)
	attrKeys := make([]string, len(attrs))
	for i, attr := range attrs {
		attrKeys[i] = attr.Key
	}

	fltr := bson.D{{
		Key:   fieldID,
		Value: devID,
	}, {
		Key:   mstore.FieldTenantID,
		Value: tenantID,
	}, {
		Key: fieldConfigured,
		Value: bson.D{{
			Key: "$exists", Value: true,
		}},
	}}
	bwm := []mongo.WriteModel{
		mongo.NewUpdateOneModel().
			SetFilter(fltr).
			SetUpdate(bson.D{{
				Key: "$pull",
				Value: bson.D{{
					Key: fieldConfigured,
					Value: bson.D{{Key: "key", Value: bson.D{{
						Key:   "$in",
						Value: attrKeys,
					}}}},
				}},
			}}),
		mongo.NewUpdateOneModel().
			SetUpsert(true).
			SetFilter(fltr[:2]).
			SetUpdate(bson.D{{
				Key: "$set", Value: bson.D{{
					Key:   fieldUpdatedTs,
					Value: time.Now().UTC(),
				}},
			}, {
				Key: "$push",
				Value: bson.D{{
					Key: fieldConfigured,
					Value: bson.D{{
						Key:   "$each",
						Value: attrs,
					}, {
						// Enforce validation constraint
						Key:   "$slice",
						Value: model.AttributesMaxLength,
					}},
				}},
			}}),
	}
	_, err := collDevs.BulkWrite(ctx,
		bwm,
		mopts.BulkWrite().
			SetOrdered(true),
	)

	return errors.Wrap(err, "mongo: failed to update configuration")
}

func (db *MongoStore) SetDeploymentID(ctx context.Context, devID string,
	deploymentID uuid.UUID) error {
	collDevs := db.Database(ctx).Collection(CollDevices)

	fltr := bson.D{{
		Key:   fieldID,
		Value: devID,
	}}

	update := bson.M{
		"$set": bson.D{
			{
				Key:   fieldDeploymentID,
				Value: deploymentID,
			},
		},
	}

	res, err := collDevs.UpdateOne(ctx, mstore.WithTenantID(ctx, fltr), update, mopts.UpdateOne())
	if err != nil {
		return errors.Wrap(err, "mongo: failed to set the deployment ID")
	} else if res.MatchedCount == 0 {
		return errors.Wrap(store.ErrDeviceNoExist, "mongo")
	}
	return nil
}

func (db *MongoStore) DeleteDevice(ctx context.Context, devID string) error {
	collDevs := db.Database(ctx).Collection(CollDevices)

	fltr := bson.D{{
		Key:   fieldID,
		Value: devID,
	}}
	res, err := collDevs.DeleteOne(ctx, mstore.WithTenantID(ctx, fltr))

	if res != nil && res.DeletedCount == 0 {
		return errors.Wrap(store.ErrDeviceNoExist, "mongo")
	}
	return errors.Wrap(err, "mongo: failed to delete device configuration")
}

func (db *MongoStore) GetDevice(ctx context.Context, devID string) (model.Device, error) {
	collDevs := db.Database(ctx).Collection(CollDevices)

	fltr := bson.D{{
		Key:   fieldID,
		Value: devID,
	}}
	res := collDevs.FindOne(ctx, mstore.WithTenantID(ctx, fltr))

	var device model.Device

	err := res.Decode(&device)
	if err != nil {
		return device, errors.Wrap(store.ErrDeviceNoExist, "mongo")
	}

	return device, nil
}

func (db *MongoStore) DeleteTenant(ctx context.Context, tenant_id string) error {
	database := db.Database(ctx)
	collectionNames, err := database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}
	for _, collName := range collectionNames {
		collection := database.Collection(collName)
		_, e := collection.DeleteMany(ctx, bson.M{KeyTenantID: tenant_id})
		if e != nil {
			return e
		}
	}
	return nil
}
