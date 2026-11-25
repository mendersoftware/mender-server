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
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

type testMigration struct {
	version migrate.Version
	calls   *[]string
	err     error
}

func (m testMigration) Version() migrate.Version { return m.version }

func (m testMigration) Up(ctx context.Context, db *mongo.Database, from migrate.Version) error {
	if m.calls != nil {
		*m.calls = append(*m.calls, from.String()+"->"+m.version.String())
	}
	return m.err
}

type testMigrationMinor struct{ testMigration }

func (testMigrationMinor) Down(ctx context.Context, db *mongo.Database) error { return nil }

type testMigrationMajor struct {
	testMigration
	minVersion migrate.Version
}

func (m testMigrationMajor) MinimumVersion() migrate.Version { return m.minVersion }

func TestApplyMigrations(t *testing.T) {
	mongoURL := os.Getenv("TEST_MONGO_URL")
	if testing.Short() {
		t.Skip()
	} else if mongoURL == "" {
		t.Skip("env: TEST_MONGO_URL not set")
	}
	t.Parallel()

	c, err := mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Fatalf("failed to setup mongodb client: %s", err)
	}
	// Database names are limited to 63 characters.
	dbName := func(t *testing.T) string {
		name := strings.NewReplacer("/", "", " ", "_").Replace(t.Name())
		return name[len("TestApplyMigrations"):][:min(len(name)-len("TestApplyMigrations"), 63)]
	}
	v := migrate.MakeVersion

	// chain returns a typical migration list for major version 2.
	chain := func(calls *[]string) []Migration {
		return []Migration{
			testMigrationMinor{testMigration{v(2, 1, 0), calls, nil}},
			testMigrationMajor{testMigration{v(2, 0, 0), calls, nil}, v(1, 2, 17)},
			testMigration{v(2, 1, 1), calls, nil},
			testMigrationMajor{testMigration{v(1, 0, 0), calls, nil}, v(0, 0, 0)},
		}
	}

	testCases := []struct {
		name       string
		dbVersions []migrate.Version
		migrations func(calls *[]string) []Migration
		opts       []MigrationOption

		calls   []string
		version migrate.Version
		err     error
	}{{
		name:       "empty database starts from latest major",
		migrations: chain,
		calls:      []string{"0.0.0->2.0.0", "2.0.0->2.1.0", "2.1.0->2.1.1"},
		version:    v(2, 1, 1),
	}, {
		name:       "upgrade from previous major",
		dbVersions: []migrate.Version{v(1, 2, 16), v(1, 2, 17)},
		migrations: chain,
		calls:      []string{"1.2.17->2.0.0", "2.0.0->2.1.0", "2.1.0->2.1.1"},
		version:    v(2, 1, 1),
	}, {
		name:       "pending patch only",
		dbVersions: []migrate.Version{v(2, 0, 0), v(2, 1, 0)},
		migrations: chain,
		calls:      []string{"2.1.0->2.1.1"},
		version:    v(2, 1, 1),
	}, {
		name:       "up to date",
		dbVersions: []migrate.Version{v(2, 1, 1)},
		migrations: chain,
		version:    v(2, 1, 1),
	}, {
		name:       "database ahead within same major",
		dbVersions: []migrate.Version{v(2, 3, 0)},
		migrations: chain,
		version:    v(2, 3, 0),
	}, {
		name:       "database of newer major",
		dbVersions: []migrate.Version{v(3, 0, 0)},
		migrations: chain,
		version:    v(3, 0, 0),
		err:        ErrDatabaseIncompatible,
	}, {
		name:       "database below minimum version",
		dbVersions: []migrate.Version{v(1, 2, 3)},
		migrations: chain,
		version:    v(1, 2, 3),
		err:        ErrDatabaseIncompatible,
	}, {
		name:       "missing major migration",
		dbVersions: []migrate.Version{v(1, 2, 17)},
		migrations: func(calls *[]string) []Migration {
			return []Migration{testMigrationMinor{testMigration{v(2, 1, 0), calls, nil}}}
		},
		version: v(1, 2, 17),
		err:     ErrDatabaseIncompatible,
	}, {
		name: "minor migration not implementing MigrationMinor",
		migrations: func(calls *[]string) []Migration {
			return []Migration{testMigration{v(2, 1, 0), calls, nil}}
		},
		err: ErrInvalidMigration,
	}, {
		name: "duplicate versions",
		migrations: func(calls *[]string) []Migration {
			return []Migration{
				testMigration{v(2, 1, 1), calls, nil},
				testMigration{v(2, 1, 1), calls, nil},
			}
		},
		err: ErrInvalidMigration,
	}, {
		name:       "failed migration is not recorded",
		dbVersions: []migrate.Version{v(2, 0, 0)},
		migrations: func(calls *[]string) []Migration {
			return []Migration{
				testMigrationMinor{testMigration{v(2, 1, 0), calls, nil}},
				testMigration{v(2, 1, 1), calls, errors.New("boom")},
			}
		},
		calls:   []string{"2.0.0->2.1.0", "2.1.0->2.1.1"},
		version: v(2, 1, 0),
		err:     errors.New("boom"),
	}, {
		name:       "target version",
		dbVersions: []migrate.Version{v(1, 2, 17)},
		migrations: chain,
		opts:       []MigrationOption{WithTargetVersion(v(2, 1, 0))},
		calls:      []string{"1.2.17->2.0.0", "2.0.0->2.1.0"},
		version:    v(2, 1, 0),
	}, {
		name:       "target version of previous major",
		dbVersions: []migrate.Version{v(1, 2, 17)},
		migrations: chain,
		opts:       []MigrationOption{WithTargetVersion(v(1, 5, 0))},
		version:    v(1, 2, 17),
	}, {
		name:       "target version below database version",
		dbVersions: []migrate.Version{v(2, 1, 1)},
		migrations: chain,
		opts:       []MigrationOption{WithTargetVersion(v(2, 0, 0))},
		version:    v(2, 1, 1),
	}, {
		name:       "force from major",
		dbVersions: []migrate.Version{v(1, 2, 16), v(1, 2, 17), v(2, 0, 0), v(2, 1, 1)},
		migrations: chain,
		opts:       []MigrationOption{WithForceFromMajor(true)},
		calls:      []string{"1.2.17->2.0.0", "2.0.0->2.1.0", "2.1.0->2.1.1"},
		version:    v(2, 1, 1),
	}, {
		name:       "force from major on initialized database",
		dbVersions: []migrate.Version{v(2, 0, 0), v(2, 1, 0)},
		migrations: chain,
		opts:       []MigrationOption{WithForceFromMajor(true)},
		calls:      []string{"0.0.0->2.0.0", "2.0.0->2.1.0", "2.1.0->2.1.1"},
		version:    v(2, 1, 1),
	}, {
		name:       "force from major with target version",
		dbVersions: []migrate.Version{v(1, 2, 17), v(2, 0, 0)},
		migrations: chain,
		opts: []MigrationOption{
			WithForceFromMajor(true),
			WithTargetVersion(v(2, 1, 0)),
		},
		calls:   []string{"1.2.17->2.0.0", "2.0.0->2.1.0"},
		version: v(2, 1, 0),
	}, {
		name:       "force from major with database ahead of target",
		dbVersions: []migrate.Version{v(2, 3, 0)},
		migrations: chain,
		opts:       []MigrationOption{WithForceFromMajor(true)},
		version:    v(2, 3, 0),
		err:        ErrDatabaseIncompatible,
	}, {
		name:       "force from major on previous major database",
		dbVersions: []migrate.Version{v(1, 2, 17)},
		migrations: chain,
		opts:       []MigrationOption{WithForceFromMajor(true)},
		version:    v(1, 2, 17),
		err:        ErrDatabaseIncompatible,
	}, {
		name:       "force from major on empty database",
		migrations: chain,
		opts:       []MigrationOption{WithForceFromMajor(true)},
		err:        ErrDatabaseIncompatible,
	}, {
		name:       "force from major without major migration of current major",
		dbVersions: []migrate.Version{v(2, 1, 0)},
		migrations: func(calls *[]string) []Migration {
			return []Migration{
				testMigrationMajor{testMigration{v(1, 0, 0), calls, nil}, v(0, 0, 0)},
				testMigration{v(2, 1, 1), calls, nil},
			}
		},
		opts:    []MigrationOption{WithForceFromMajor(true)},
		version: v(2, 1, 0),
		err:     ErrDatabaseIncompatible,
	}, {
		name:       "force disabled",
		dbVersions: []migrate.Version{v(2, 1, 1)},
		migrations: chain,
		opts:       []MigrationOption{WithForceFromMajor(false)},
		version:    v(2, 1, 1),
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db := c.Database(dbName(t))
			defer db.Drop(ctx)
			for _, dbVersion := range tc.dbVersions {
				err := migrate.UpdateMigrationInfo(ctx, dbVersion, c, db.Name())
				if err != nil {
					t.Fatal(err)
				}
			}

			var calls []string
			err := ApplyMigrations(ctx, db, tc.migrations(&calls), tc.opts...)
			switch {
			case tc.err == nil && err != nil:
				t.Errorf("unexpected error: %s", err)
			case tc.err != nil && err == nil:
				t.Errorf("expected error %q", tc.err)
			case tc.err != nil && !errors.Is(err, tc.err) &&
				!strings.Contains(err.Error(), tc.err.Error()):
				t.Errorf("expected error %q, got %q", tc.err, err)
			}
			if strings.Join(calls, ",") != strings.Join(tc.calls, ",") {
				t.Errorf("expected calls %v, got %v", tc.calls, calls)
			}
			history, err := migrate.GetDatabaseMigrationInfo(ctx, db)
			if err != nil {
				t.Fatal(err)
			} else if version := latestVersion(history); version != tc.version {
				t.Errorf("expected database version %s, got %s", tc.version, version)
			}
		})
	}

	t.Run("concurrent migrators migrate once", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		db := c.Database(dbName(t))
		defer db.Drop(ctx)

		var ups atomic.Int32
		var wg sync.WaitGroup
		errs := make(chan error, 5)
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs <- ApplyMigrations(ctx, db, []Migration{
					countingMigration{v(1, 0, 0), &ups},
				})
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		}
		if n := ups.Load(); n != 1 {
			t.Errorf("expected migration to run once, ran %d times", n)
		}
	})
}

type countingMigration struct {
	version migrate.Version
	ups     *atomic.Int32
}

func (m countingMigration) Version() migrate.Version { return m.version }

func (countingMigration) MinimumVersion() migrate.Version { return migrate.Version{} }

func (m countingMigration) Up(ctx context.Context, db *mongo.Database, from migrate.Version) error {
	m.ups.Add(1)
	time.Sleep(100 * time.Millisecond)
	return nil
}
