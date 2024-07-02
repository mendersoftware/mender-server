// Copyright 2023 Northern.tech AS
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
	"net/url"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/services/reporting/model"
)

const (
	collNameMapping   = "mapping"
	keyNameTenantID   = "tenant_id"
	indexNameTenantID = "tenant_id_ndx"
)

type MongoStoreConfig struct {
	// MongoURL holds the URL to the MongoDB server.
	MongoURL *url.URL
	// SSL Enables SSL for mongo connections
	SSL bool
	// SkipVerify controls whether a mongo client verifies the server's
	// certificate chain and host name.
	SSLSkipVerify bool
	// Username holds the user id credential for authenticating with the
	// MongoDB server.
	Username string
	// Password holds the password credential for authenticating with the
	// MongoDB server.
	Password string
	// DbName contains the name of the reporting database.
	DbName string
}

// newClient returns a mongo client
func newClient(ctx context.Context, config MongoStoreConfig) (*mongo.Client, error) {

	clientOptions := mopts.Client()
	if config.MongoURL == nil {
		return nil, errors.New("mongo: missing URL")
	}
	clientOptions.ApplyURI(config.MongoURL.String())

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

	if config.SSL {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = config.SSLSkipVerify
		clientOptions.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.Connect(ctx, clientOptions)
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

func (db *MongoStore) Database(ctx context.Context, opt ...*mopts.DatabaseOptions) *mongo.Database {
	return db.client.Database(db.config.DbName, opt...)
}

// Ping verifies the connection to the database
func (db *MongoStore) Ping(ctx context.Context) error {
	res := db.client.
		Database(db.config.DbName).
		RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
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

// GetMapping returns the mapping
func (db *MongoStore) GetMapping(ctx context.Context, tenantID string) (*model.Mapping, error) {
	query := bson.M{
		keyNameTenantID: tenantID,
	}
	res := db.client.
		Database(db.config.DbName).
		Collection(collNameMapping).
		FindOne(ctx, query)

	mapping := &model.Mapping{}
	err := res.Decode(mapping)
	if err == mongo.ErrNoDocuments {
		mapping = &model.Mapping{
			TenantID:  tenantID,
			Inventory: []string{},
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to update and get the mapping")
	}
	return mapping, nil
}

// UpdateAndGetMapping updates the mapping and returns it
func (db *MongoStore) UpdateAndGetMapping(ctx context.Context, tenantID string,
	inventory []string) (*model.Mapping, error) {
	inventoryLastField := fmt.Sprintf("inventory.%d", model.MaxMappingInventoryAttributes-1)
	query := bson.M{
		keyNameTenantID: tenantID,
		inventoryLastField: bson.M{
			"$exists": false,
		},
	}
	update := bson.M{
		"$addToSet": bson.M{
			"inventory": bson.M{
				"$each": inventory,
			},
		},
	}
	projection := bson.M{
		"tenant_id": 1,
		"inventory": bson.M{
			"$slice": model.MaxMappingInventoryAttributes,
		},
	}
	opts := mopts.FindOneAndUpdate().
		SetReturnDocument(mopts.After).
		SetUpsert(true).
		SetProjection(projection)
	mapping := &model.Mapping{}
	err := db.client.
		Database(db.config.DbName).
		Collection(collNameMapping).
		FindOneAndUpdate(ctx, query, update, opts).
		Decode(mapping)

	if mongo.IsDuplicateKeyError(err) {
		// Tenant attribute quota already full
		err = db.client.
			Database(db.config.DbName).
			Collection(collNameMapping).
			FindOne(ctx, bson.D{{Key: "tenant_id", Value: tenantID}},
				mopts.FindOne().SetProjection(projection),
			).
			Decode(mapping)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to update and get the mapping")
	}
	return mapping, nil
}
