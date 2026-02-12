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
package migrate

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	storev1 "github.com/mendersoftware/mender-server/pkg/store"
)

// this is a small internal data layer for the migration utils, may be shared by diff migrators
const (
	DbMigrationsColl = "migration_info"
)

type MigrationEntry struct {
	Version   Version   `bson:"version"`
	Timestamp time.Time `bson:"timestamp"`
}

// GetMigrationInfo retrieves a list of migrations applied to the db.
// On success the function returns the MigrationEntries present in the db
// sorted by the version in decending order.
func GetMigrationInfo(
	ctx context.Context,
	sess *mongo.Client,
	db string,
) ([]MigrationEntry, error) {
	c := sess.Database(db).Collection(DbMigrationsColl)
	findOpts := mopts.Find()
	findOpts.SetSort(bson.D{{
		Key: "version.major", Value: -1,
	}, {
		Key: "version.minor", Value: -1,
	}, {
		Key: "version.patch", Value: -1,
	}})

	cursor, err := c.Find(ctx, bson.M{
		"version": bson.M{"$exists": true},
	}, findOpts)
	if cursor == nil || err != nil {
		return nil, errors.Wrap(err, "db: failed to get migration info")
	}

	var infoArray []MigrationEntry
	err = cursor.All(ctx, &infoArray)
	return infoArray, err
}

// UpdateMigrationInfo inserts a migration entry in the migration info collection.
func UpdateMigrationInfo(
	ctx context.Context,
	version Version,
	sess *mongo.Client,
	db string,
) error {
	c := sess.Database(db).Collection(DbMigrationsColl)

	entry := MigrationEntry{
		Version:   version,
		Timestamp: time.Now(),
	}
	_, err := c.InsertOne(ctx, entry)
	if err != nil {
		return errors.Wrap(err, "db: failed to insert migration info")
	}
	// result.InsertedID (InsertOneResult)

	return nil
}

func GetTenantDbs(
	ctx context.Context,
	client *mongo.Client,
	matcher storev1.TenantDbMatchFunc,
) ([]string, error) {
	result, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	tenantDbs := make([]string, len(result))
	j := 0
	for _, db := range result {
		if matcher(db) {
			tenantDbs[j] = db
			j++
		}
	}
	tenantDbs = tenantDbs[:j]

	return tenantDbs, nil
}
