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
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	systemDBs     = map[string]bool{"local": true, "admin": true, "config": true, "workflows": true}
	tenantDBRegex = regexp.MustCompile(`^(deployment_service|inventory)-[0-9a-f]{24}`)
)

// Client wraps a mongo client for test DB cleanup.
type Client struct {
	client *mongo.Client
}

// NewClient creates a new mongo client connected to the given address.
func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+addr))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

// Cleanup drops tenant-specific databases and clears all non-system,
// non-capped, non-migration collections. Mirrors the Python testutils
// infra/mongo.py cleanup() method.
func (c *Client) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbs, err := c.client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return err
	}

	for _, dbName := range dbs {
		if systemDBs[dbName] {
			continue
		}
		if tenantDBRegex.MatchString(dbName) {
			if err := c.client.Database(dbName).Drop(ctx); err != nil {
				return err
			}
			continue
		}
		db := c.client.Database(dbName)
		colls, err := db.ListCollections(ctx, bson.M{
			"name": bson.M{"$ne": "migration_info"},
			"$or": bson.A{
				bson.M{"options.capped": bson.M{"$exists": false}},
				bson.M{"options.capped": false},
			},
		})
		if err != nil {
			return err
		}
		for colls.Next(ctx) {
			var result struct {
				Name string `bson:"name"`
			}
			if err := colls.Decode(&result); err != nil {
				return err
			}
			if _, err := db.Collection(result.Name).DeleteMany(ctx, bson.M{}); err != nil {
				return err
			}
		}
		if err := colls.Err(); err != nil {
			return err
		}
	}
	return nil
}

// Close disconnects the mongo client.
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}
