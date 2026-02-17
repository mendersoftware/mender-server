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
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/v2/bson"

	. "github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate/mocks"
)

func TestSimpleMigratorApply(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDummyMigratorApply in short mode.")
	}

	makeMigration := func(v Version, from Version, err error) Migration {
		m := &mocks.Migration{}
		m.On("Up", from).Return(err)
		m.On("Version").Return(v)
		return m
	}

	testCases := map[string]struct {
		Automigrate     bool
		InputMigrations []MigrationEntry
		InputVersion    Version

		Migrators []Migration

		OutputVersion Version
		OutputError   error
	}{
		"ok - empty state": {
			Automigrate:     true,
			InputMigrations: nil,
			InputVersion:    MakeVersion(1, 0, 0),

			OutputVersion: MakeVersion(1, 0, 0),
		},

		"ok - already has version": {
			Automigrate: true,
			InputMigrations: []MigrationEntry{
				{
					Version:   MakeVersion(1, 0, 0),
					Timestamp: time.Now(),
				},
			},
			InputVersion:  MakeVersion(1, 0, 0),
			OutputVersion: MakeVersion(1, 0, 0),
		},
		"ok - already has version, no automigrate": {
			Automigrate: false,
			InputMigrations: []MigrationEntry{
				{
					Version:   MakeVersion(1, 0, 0),
					Timestamp: time.Now(),
				},
			},
			InputVersion:  MakeVersion(1, 0, 0),
			OutputVersion: MakeVersion(1, 0, 0),
		},
		"ok - add default target version": {
			Automigrate: true,
			InputMigrations: []MigrationEntry{
				{
					Version:   MakeVersion(1, 0, 0),
					Timestamp: time.Now(),
				},
			},
			InputVersion:  MakeVersion(1, 1, 0),
			OutputVersion: MakeVersion(1, 1, 0),
		},
		"ok - add default target version, no automigrate": {
			Automigrate: false,
			InputMigrations: []MigrationEntry{
				{
					Version:   MakeVersion(1, 0, 0),
					Timestamp: time.Now(),
				},
			},
			InputVersion:  MakeVersion(1, 1, 0),
			OutputVersion: MakeVersion(1, 0, 0),
			OutputError:   errors.New("db needs migration: test has version 1.0.0, needs version 1.1.0"),
		},
		"ok - ran migrations": {
			Automigrate: true,
			InputMigrations: []MigrationEntry{
				{
					Version:   MakeVersion(1, 0, 0),
					Timestamp: time.Now(),
				},
				{
					Version:   MakeVersion(1, 0, 1),
					Timestamp: time.Now(),
				},
			},
			InputVersion:  MakeVersion(1, 1, 0),
			OutputVersion: MakeVersion(1, 1, 0),

			Migrators: []Migration{
				makeMigration(MakeVersion(1, 0, 1), MakeVersion(1, 0, 0), nil),
				makeMigration(MakeVersion(1, 1, 0), MakeVersion(1, 0, 1), nil),
			},
		},
		"ok - ran migrations, no automigrate": {
			Automigrate: false,
			InputMigrations: []MigrationEntry{
				{
					Version:   MakeVersion(1, 0, 0),
					Timestamp: time.Now(),
				},
				{
					Version:   MakeVersion(1, 0, 1),
					Timestamp: time.Now(),
				},
			},
			InputVersion:  MakeVersion(1, 1, 0),
			OutputVersion: MakeVersion(1, 0, 1),

			Migrators: []Migration{
				makeMigration(MakeVersion(1, 0, 1), MakeVersion(1, 0, 0), nil),
				makeMigration(MakeVersion(1, 1, 0), MakeVersion(1, 0, 1), nil),
			},
			OutputError: errors.New("db needs migration: test has version 1.0.1, needs version 1.1.0"),
		},
		"ok - migration to lower": {
			Automigrate:     true,
			InputMigrations: nil,
			InputVersion:    MakeVersion(0, 1, 0),
			OutputVersion:   MakeVersion(0, 1, 0),

			Migrators: []Migration{
				makeMigration(MakeVersion(1, 0, 1), MakeVersion(0, 0, 0), nil),
				makeMigration(MakeVersion(1, 1, 0), MakeVersion(1, 0, 1), nil),
			},
		},
		"err - failed migration": {
			Automigrate: true,
			InputMigrations: []MigrationEntry{
				{
					Version:   MakeVersion(1, 0, 0),
					Timestamp: time.Now(),
				},
			},
			InputVersion: MakeVersion(1, 1, 0),
			// migration 1.0.3 fails, thus the output should remain at 1.0.2
			OutputVersion: MakeVersion(1, 0, 2),

			Migrators: []Migration{
				makeMigration(MakeVersion(1, 0, 1), MakeVersion(1, 0, 0), nil),
				makeMigration(MakeVersion(1, 0, 3), MakeVersion(1, 0, 2), errors.New("failed")),
				makeMigration(MakeVersion(1, 0, 2), MakeVersion(1, 0, 1), nil),
			},

			OutputError: errors.New("failed to apply migration from 1.0.2 to 1.0.3: failed"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {

			//setup
			db.Wipe()
			client := db.Client()
			for i := range tc.InputMigrations {
				_, err := client.Database("test").
					Collection(DbMigrationsColl).
					InsertOne(db.CTX(), tc.InputMigrations[i])
				assert.NoError(t, err)
			}

			//test
			m := &SimpleMigrator{Client: client, Db: "test", Automigrate: tc.Automigrate}
			err := m.Apply(context.Background(), tc.InputVersion, tc.Migrators)
			if tc.OutputError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				assert.NoError(t, err)
			}

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
				err = cursor.Decode(elem)
				bsonBytes, _ := bson.Marshal(elem)
				bson.Unmarshal(bsonBytes, &res)
				out = append(out, res)
			}

			// sort applied migrations
			sort.Slice(out, func(i int, j int) bool {
				return VersionIsLess(out[i].Version, out[j].Version)
			})
			// applied migration should be last
			assert.Equal(t, tc.OutputVersion, out[len(out)-1].Version)
		})
	}
}

func TestErrNeedsMigration(t *testing.T) {
	err := errors.New("db needs migration: mydbname has version 1.0.0, needs version 1.1.0")

	assert.True(t, IsErrNeedsMigration(err))
}
