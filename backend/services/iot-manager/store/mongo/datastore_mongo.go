// Copyright 2025 Northern.tech AS
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
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	mongostore "github.com/mendersoftware/mender-server/pkg/mongo/v2"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/codec"

	dconfig "github.com/mendersoftware/mender-server/services/iot-manager/config"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	"github.com/mendersoftware/mender-server/services/iot-manager/store"
)

const (
	CollNameDevices      = "devices"
	CollNameIntegrations = "integrations"

	KeyID             = "_id"
	KeyIntegrationIDs = "integration_ids"
	KeyProvider       = "provider"
	KeyTenantID       = "tenant_id"
	KeyCredentials    = "credentials"
	KeyStatus         = "status"
	KeyIntegrationID  = "integration_id"

	ConnectTimeoutSeconds = 10
	defaultAutomigrate    = false
)

var (
	ErrFailedToGetIntegrations = errors.New("failed to get integrations")
	ErrFailedToGetDevice       = errors.New("failed to get device")
	ErrFailedToGetSettings     = errors.New("failed to get settings")
)

type Config struct {
	Automigrate *bool
	DbName      *string
}

func NewConfig() *Config {
	conf := new(Config)
	return conf.
		SetAutomigrate(defaultAutomigrate).
		SetDbName(DbName)
}

func (c *Config) SetAutomigrate(migrate bool) *Config {
	c.Automigrate = &migrate
	return c
}

func (c *Config) SetDbName(name string) *Config {
	c.DbName = &name
	return c
}

func mergeConfig(configs ...*Config) *Config {
	config := NewConfig()
	for _, c := range configs {
		if c == nil {
			continue
		}
		if c.Automigrate != nil {
			config.SetAutomigrate(*c.Automigrate)
		}
		if c.DbName != nil {
			config.DbName = c.DbName
		}
	}
	return config
}

// SetupDataStore returns the mongo data store and optionally runs migrations
func SetupDataStore(conf *Config) (*DataStoreMongo, error) {
	conf = mergeConfig(conf)
	ctx := context.Background()
	dbClient, err := NewClient(ctx, config.Config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect to db: %v", err))
	}
	dataStore := NewDataStoreWithClient(dbClient, conf)

	if conf.Automigrate != nil && *conf.Automigrate {
		err := dataStore.Migrate(ctx)
		if err != nil {
			return nil, err
		}
	}

	return dataStore, nil
}

func (ds *DataStoreMongo) Migrate(ctx context.Context) error {
	return Migrate(ctx, *ds.DbName, DbVersion, ds.client, *ds.Automigrate)
}

// NewClient returns a mongo client
func NewClient(ctx context.Context, c config.Reader) (*mongo.Client, error) {

	mongoURL := c.GetString(dconfig.SettingMongo)
	if !strings.Contains(mongoURL, "://") {
		return nil, errors.Errorf("Invalid mongoURL %q: missing schema.",
			mongoURL)
	}
	clientOptions := mongostore.BaseClientOptions(mongoURL)
	clientOptions.SetRegistry(codec.NewRegistry())

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
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ConnectTimeoutSeconds*time.Second)
		defer cancel()
	}
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to mongo server")
	}

	// Validate connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "Error reaching mongo server")
	}

	return client, nil
}

// DataStoreMongo is the data storage service
type DataStoreMongo struct {
	// client holds the reference to the client used to communicate with the
	// mongodb server.
	client *mongo.Client

	*Config
}

var _ store.DataStore = &DataStoreMongo{}

// NewDataStoreWithClient initializes a DataStore object
func NewDataStoreWithClient(client *mongo.Client, conf ...*Config) *DataStoreMongo {
	return &DataStoreMongo{
		client: client,
		Config: mergeConfig(conf...),
	}
}

// Ping verifies the connection to the database
func (db *DataStoreMongo) Ping(ctx context.Context) error {
	res := db.client.Database(*db.DbName).RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
}

func (db *DataStoreMongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := db.client.Disconnect(ctx)
	return err
}

func (db *DataStoreMongo) Collection(
	name string,
	opts ...mopts.Lister[mopts.CollectionOptions],
) *mongo.Collection {
	return db.client.Database(*db.DbName).Collection(name, opts...)
}

func (db *DataStoreMongo) ListCollectionNames(
	ctx context.Context,
) ([]string, error) {
	return db.client.Database(*db.DbName).ListCollectionNames(ctx, bson.D{})
}

func (db *DataStoreMongo) getIntegrations(
	ctx context.Context,
	fltr model.IntegrationFilter,
) (*mongo.Cursor, error) {
	var (
		err error
	)

	collIntegrations := db.Collection(CollNameIntegrations)
	findOpts := mopts.Find().
		SetSort(bson.D{{
			Key:   KeyTenantID,
			Value: 1,
		}, {
			Key:   KeyProvider,
			Value: 1,
		}, {
			Key:   KeyID,
			Value: 1,
		}}).SetSkip(fltr.Skip)
	if fltr.Limit > 0 {
		findOpts.SetLimit(fltr.Limit)
	}

	fltrDoc := make(bson.D, 0, 3)
	id := identity.FromContext(ctx)
	if id != nil {
		fltrDoc = append(fltrDoc, bson.E{Key: KeyTenantID, Value: id.Tenant})
	}
	if fltr.Provider != model.ProviderEmpty {
		fltrDoc = append(fltrDoc, bson.E{Key: KeyProvider, Value: fltr.Provider})
	}
	if fltr.IDs != nil {
		switch len(fltr.IDs) {
		case 0:
			// Won't match anything, let's save the request
			return nil, nil
		case 1:
			fltrDoc = append(fltrDoc, bson.E{Key: KeyID, Value: fltr.IDs[0]})

		default:
			fltrDoc = append(fltrDoc, bson.E{Key: KeyID, Value: bson.D{{
				Key: "$in", Value: fltr.IDs,
			}}})
		}
	}

	cur, err := collIntegrations.Find(ctx,
		fltrDoc,
		findOpts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error executing integrations collection request")
	}
	return cur, nil
}

func (db *DataStoreMongo) GetIntegrationsIter(
	ctx context.Context,
	fltr model.IntegrationFilter,
) iter.Seq2[*model.Integration, error] {
	cur, err := db.getIntegrations(ctx, fltr)
	if err != nil {
		return func(yield func(*model.Integration, error) bool) {
			yield(nil, err)
		}
	}
	return func(yield func(*model.Integration, error) bool) {
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var integration model.Integration
			err := cur.Decode(&integration)
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(&integration, nil) {
				return
			}
		}
		if err := cur.Err(); err != nil {
			yield(nil, err)
		}
	}

}

func (db *DataStoreMongo) GetIntegrations(
	ctx context.Context,
	fltr model.IntegrationFilter,
) ([]model.Integration, error) {
	cur, err := db.getIntegrations(ctx, fltr)
	if err != nil {
		return nil, err
	}
	results := []model.Integration{}
	if cur != nil {
		if err = cur.All(ctx, &results); err != nil {
			return nil, errors.Wrap(err, "error retrieving integrations collection results")
		}
	}
	return results, nil
}

func (db *DataStoreMongo) GetIntegrationById(
	ctx context.Context,
	integrationId uuid.UUID,
) (*model.Integration, error) {
	var integration = new(model.Integration)

	collIntegrations := db.Collection(CollNameIntegrations)
	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil {
		tenantId = id.Tenant
	}

	if err := collIntegrations.FindOne(ctx,
		bson.M{KeyTenantID: tenantId},
	).Decode(&integration); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, store.ErrObjectNotFound
		default:
			return nil, errors.Wrap(err, ErrFailedToGetIntegrations.Error())
		}
	}
	return integration, nil
}

func (db *DataStoreMongo) CreateIntegration(
	ctx context.Context,
	integration model.Integration,
) (*model.Integration, error) {
	collIntegrations := db.Collection(CollNameIntegrations)
	integration.ID = uuid.New()
	_, err := collIntegrations.
		InsertOne(ctx, mongostore.WithTenantID(ctx, integration))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, store.ErrObjectExists
		}
		return nil, errors.Wrapf(err, "failed to store integration %v", integration)
	}

	return &integration, err
}

func (db *DataStoreMongo) SetIntegrationCredentials(
	ctx context.Context,
	integrationId uuid.UUID,
	credentials model.Credentials,
) error {
	collIntegrations := db.client.Database(*db.DbName).Collection(CollNameIntegrations)

	fltr := bson.D{{
		Key:   KeyID,
		Value: integrationId,
	}}

	update := bson.M{
		"$set": bson.D{
			{
				Key:   KeyCredentials,
				Value: credentials,
			},
		},
	}

	result, err := collIntegrations.UpdateOne(ctx,
		mongostore.WithTenantID(ctx, fltr),
		update,
	)
	if result.MatchedCount == 0 {
		return store.ErrObjectNotFound
	}

	return errors.Wrap(err, "mongo: failed to set integration credentials")
}

func (db *DataStoreMongo) RemoveIntegration(ctx context.Context, integrationId uuid.UUID) error {
	collIntegrations := db.client.Database(*db.DbName).Collection(CollNameIntegrations)
	fltr := bson.D{{
		Key:   KeyID,
		Value: integrationId,
	}}
	res, err := collIntegrations.DeleteOne(ctx, mongostore.WithTenantID(ctx, fltr))
	if err != nil {
		return err
	} else if res.DeletedCount == 0 {
		return store.ErrObjectNotFound
	}
	return nil
}

// DoDevicesExistByIntegrationID checks if there is at least one device connected
// with given integration ID
func (db *DataStoreMongo) DoDevicesExistByIntegrationID(
	ctx context.Context,
	integrationID uuid.UUID,
) (bool, error) {
	var (
		err error
	)
	collDevices := db.client.Database(*db.DbName).Collection(CollNameDevices)

	fltr := bson.D{
		{
			Key: KeyIntegrationIDs, Value: integrationID,
		},
	}
	if err = collDevices.FindOne(ctx, mongostore.WithTenantID(ctx, fltr)).Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (db *DataStoreMongo) GetDeviceByIntegrationID(
	ctx context.Context,
	deviceID string,
	integrationID uuid.UUID,
) (*model.Device, error) {
	var device *model.Device

	collDevices := db.Collection(CollNameDevices)
	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil {
		tenantId = id.Tenant
	}

	filter := bson.D{{
		Key: KeyTenantID, Value: tenantId,
	}, {
		Key: KeyID, Value: deviceID,
	}, {
		Key: KeyIntegrationIDs, Value: integrationID,
	}}
	if err := collDevices.FindOne(ctx,
		filter,
	).Decode(&device); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, store.ErrObjectNotFound
		default:
			return nil, errors.Wrap(err, ErrFailedToGetDevice.Error())
		}
	}
	return device, nil
}

func (db *DataStoreMongo) GetDevice(
	ctx context.Context,
	deviceID string,
) (*model.Device, error) {
	var (
		tenantID string
		result   *model.Device = new(model.Device)
	)
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	filter := bson.D{{
		Key: KeyID, Value: deviceID,
	}, {
		Key: KeyTenantID, Value: tenantID,
	}}
	collDevices := db.Collection(CollNameDevices)

	err := collDevices.FindOne(ctx, filter).
		Decode(result)
	if err == mongo.ErrNoDocuments {
		return nil, store.ErrObjectNotFound
	}
	return result, err
}

func (db *DataStoreMongo) DeleteDevice(ctx context.Context, deviceID string) error {
	var tenantID string
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	collDevices := db.Collection(CollNameDevices)

	filter := bson.D{{
		Key: KeyID, Value: deviceID,
	}, {
		Key: KeyTenantID, Value: tenantID,
	}}

	res, err := collDevices.DeleteOne(ctx, filter)
	if err != nil {
		return err
	} else if res.DeletedCount == 0 {
		return store.ErrObjectNotFound
	}
	return nil
}

func (db *DataStoreMongo) RemoveDevicesFromIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
) (int64, error) {
	var tenantID string
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	filter := bson.D{{
		Key: KeyTenantID, Value: tenantID,
	}, {
		Key: KeyIntegrationIDs, Value: integrationID,
	}}
	update := bson.D{{
		Key: "$pull", Value: bson.D{{
			Key: KeyIntegrationIDs, Value: integrationID,
		}},
	}}

	collDevices := db.Collection(CollNameDevices)

	res, err := collDevices.UpdateMany(ctx, filter, update)
	if res != nil {
		return res.ModifiedCount, err
	}
	return 0, errors.Wrap(err, "mongo: failed to remove device from integration")
}

func (db *DataStoreMongo) UpsertDeviceIntegrations(
	ctx context.Context,
	deviceID string,
	integrationIDs []uuid.UUID,
) (*model.Device, error) {
	var (
		tenantID string
		result   = new(model.Device)
	)
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	if integrationIDs == nil {
		integrationIDs = []uuid.UUID{}
	}
	filter := bson.D{{
		Key: KeyID, Value: deviceID,
	}, {
		Key: KeyTenantID, Value: tenantID,
	}}
	update := bson.D{{
		Key: "$addToSet", Value: bson.D{{
			Key: KeyIntegrationIDs, Value: bson.D{{
				Key: "$each", Value: integrationIDs,
			}},
		}},
	}}
	updateOpts := mopts.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(mopts.After)
	collDevices := db.Collection(CollNameDevices)
	err := collDevices.FindOneAndUpdate(ctx, filter, update, updateOpts).
		Decode(result)
	return result, err
}

func (db *DataStoreMongo) GetAllDevices(ctx context.Context) (store.Iterator, error) {
	collDevs := db.Collection(CollNameDevices)

	return collDevs.Find(ctx,
		bson.D{},
		mopts.Find().
			SetSort(bson.D{{Key: KeyTenantID, Value: 1}}),
	)

}

func (db *DataStoreMongo) DeleteTenantData(
	ctx context.Context,
) error {
	id := identity.FromContext(ctx)
	if id == nil {
		return errors.New("identity is empty")
	}
	if len(id.Tenant) < 1 {
		return errors.New("tenant id is empty")
	}

	collectionNames, err := db.ListCollectionNames(ctx)
	if err != nil {
		return err
	}
	for _, collName := range collectionNames {
		collection := db.Collection(collName)
		_, e := collection.DeleteMany(ctx, bson.M{KeyTenantID: id.Tenant})
		if e != nil {
			return e
		}
	}
	return nil
}
