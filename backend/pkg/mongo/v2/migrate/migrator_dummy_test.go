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
package migrate_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/v2/bson"

	. "github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

func TestDummyMigratorApply(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDummyMigratorApply in short mode.")
	}

	testCases := map[string]struct {
		Automigrate     bool
		InputMigrations []*MigrationEntry
		InputVersion    Version

		OutputVersion Version
		OutputError   error
	}{
		"ok - empty state, automigrate": {
			Automigrate:     true,
			InputMigrations: []*MigrationEntry{},
			InputVersion:    Version{Major: 1, Minor: 0, Patch: 0},

			OutputVersion: Version{Major: 1, Minor: 0, Patch: 0},
		},

		"ok - already has version, automigrate": {
			Automigrate: true,
			InputMigrations: []*MigrationEntry{
				&MigrationEntry{
					Version:   Version{Major: 1, Minor: 0, Patch: 0},
					Timestamp: time.Now(),
				},
			},
			InputVersion:  Version{Major: 1, Minor: 0, Patch: 0},
			OutputVersion: Version{Major: 1, Minor: 0, Patch: 0},
		},

		"ok - empty state, no automigrate": {
			Automigrate:     false,
			InputMigrations: []*MigrationEntry{},
			InputVersion:    Version{Major: 1, Minor: 0, Patch: 0},

			OutputVersion: Version{Major: 0, Minor: 0, Patch: 0},
			OutputError:   errors.New("db needs migration: test has version 0.0.0, needs version 1.0.0"),
		},

		"ok - already has version, no automigrate": {
			Automigrate: false,
			InputMigrations: []*MigrationEntry{
				&MigrationEntry{
					Version:   Version{Major: 1, Minor: 0, Patch: 0},
					Timestamp: time.Now(),
				},
			},
			InputVersion:  Version{Major: 1, Minor: 0, Patch: 0},
			OutputVersion: Version{Major: 1, Minor: 0, Patch: 0},
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		//setup
		db.Wipe()
		client := db.Client()
		for _, m := range tc.InputMigrations {
			_, err := client.Database("test").
				Collection(DbMigrationsColl).
				InsertOne(db.CTX(), m)
			assert.NoError(t, err)
		}

		//test
		m := &DummyMigrator{Client: client, Db: "test", Automigrate: tc.Automigrate}
		m.Apply(context.Background(), tc.InputVersion, nil)

		//verify
		var out []MigrationEntry
		cursor, _ := client.Database("test").
			Collection(DbMigrationsColl).
			Find(db.CTX(), bson.M{})

		count := 0
		for cursor.Next(db.CTX()) {
			var res MigrationEntry
			count++
			elem := &bson.D{}
			_ = cursor.Decode(elem)
			bsonBytes, _ := bson.Marshal(elem)
			bson.Unmarshal(bsonBytes, &res)
			out = append(out, res)
		}

		if tc.Automigrate {
			assert.Len(t, out, 1)
			assert.Equal(t, tc.OutputVersion, out[0].Version)
		} else {
			assert.Len(t, out, len(tc.InputMigrations))
		}
	}
}
