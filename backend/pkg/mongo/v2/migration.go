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
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	"github.com/mendersoftware/mender-server/pkg/sync"
)

// MigrationKind classifies a migration by the semantic version it produces.
// The kind is never declared by the migration itself; it is derived from
// Version() (see KindOf) so that the two can never disagree.
type MigrationKind uint8

const (
	// MigrationKindPatch (X.Y.Z, Z > 0) is a data-only fix-up that does not
	// change the shape of the schema, e.g. repairing inconsistent documents
	// or rebuilding an index with the same definition. Patches MUST be
	// idempotent and are safe to run while other instances are serving.
	MigrationKindPatch MigrationKind = iota
	// MigrationKindMinor (X.Y.0, Y > 0) is a backward compatible, additive
	// schema change: new collections, new indexes, new optional fields.
	// Service instances of the same major version (including older minors)
	// MUST keep working on the migrated database, which makes minor
	// migrations safe during rolling upgrades.
	MigrationKindMinor
	// MigrationKindMajor (X.0.0) is a breaking schema change. Older service
	// versions can no longer operate on the database, so all instances of the
	// previous major version must be stopped before the migration runs.
	MigrationKindMajor
)

func (k MigrationKind) String() string {
	switch k {
	case MigrationKindMajor:
		return "major"
	case MigrationKindMinor:
		return "minor"
	default:
		return "patch"
	}
}

// KindOf returns the kind of migration that produces version v.
func KindOf(v migrate.Version) MigrationKind {
	switch {
	case v.Patch > 0:
		return MigrationKindPatch
	case v.Minor > 0:
		return MigrationKindMinor
	default:
		return MigrationKindMajor
	}
}

// Migration is the common interface of all database migrations. A migration
// is a stateless step that brings db from version `from` to Version().
//
// The migrator records Version() in the migration info collection only after
// Up returns successfully; a failed migration leaves the recorded version
// unchanged. Up therefore MUST tolerate being re-run after a partial failure.
//
// Depending on KindOf(Version()) a migration MUST additionally implement
// MigrationMajor, MigrationMinor or MigrationPatch.
type Migration interface {
	// Version returns the database version after the migration is applied.
	Version() migrate.Version
	// Up applies the migration to db, which is currently at version from.
	Up(ctx context.Context, db *mongo.Database, from migrate.Version) error
}

// MigrationMajor is a breaking, irreversible migration (version X.0.0).
//
// There is no Down: a breaking change typically discards or reshapes data in
// ways that cannot be restored, so rolling back a major version means
// restoring a database backup.
type MigrationMajor interface {
	Migration
	// MinimumVersion returns the oldest version (from the previous major)
	// that Up can migrate from. Older databases must first be upgraded with
	// an older release of the service. The check does not apply to an empty
	// database (version 0.0.0), which Up MUST be able to initialize from
	// scratch.
	MinimumVersion() migrate.Version
}

// MigrationMinor is a backward compatible, reversible migration (X.Y.0).
type MigrationMinor interface {
	Migration
	// Down reverts the changes introduced by Up, leaving the database at the
	// previous minor version. Patches applied on top of this minor version
	// are considered reverted along with it. Down MUST be idempotent.
	Down(ctx context.Context, db *mongo.Database) error
}

// MigrationPatch is an idempotent data fix-up (X.Y.Z, Z > 0).
//
// It adds no methods to Migration; the stricter contract is that Up MUST be
// idempotent and MUST NOT change the schema, so the migrator may run it
// without the exclusive lock required by major migrations. A patch has no
// Down: it only corrects data, so reverting it would reintroduce the
// inconsistency it fixed.
type MigrationPatch = Migration

const migrationLockID = "migrate"

// MigrationLockTTL is the maximum duration ApplyMigrations holds the
// migration lock. Migrations still running when it expires are canceled so
// that they never run unprotected by the lock.
var MigrationLockTTL = time.Hour

var (
	ErrInvalidMigration     = errors.New("invalid migration")
	ErrDatabaseIncompatible = errors.New("database version is incompatible")
)

type migrationOptions struct {
	target         *migrate.Version
	forceFromMajor bool
}

// MigrationOption configures ApplyMigrations.
type MigrationOption func(*migrationOptions)

// WithTargetVersion sets the highest version to migrate to. Migrations of a
// higher version are skipped. By default all migrations are applied.
func WithTargetVersion(target migrate.Version) MigrationOption {
	return func(o *migrationOptions) {
		o.target = &target
	}
}

// WithForceFromMajor, if force is true, re-applies the latest major migration
// (not exceeding the target version) and all following migrations, even if
// the database has already been migrated past them. The major migration is
// passed the last version recorded before it, or 0.0.0 if the database was
// initialized by it, and MUST therefore cope with already migrated data.
//
// Forcing is only allowed within the major version the database is currently
// at; upgrading from a previous major version, or initializing an empty
// database, must be done without forcing. Forcing is refused with
// ErrDatabaseIncompatible if the database is of another major version than the
// latest major migration, or if the database is ahead of the target version,
// as re-applying the major migration could undo changes made by the newer
// migrations.
func WithForceFromMajor(force bool) MigrationOption {
	return func(o *migrationOptions) {
		o.forceFromMajor = force
	}
}

// ApplyMigrations brings db up to the target version, which defaults to the
// highest version in migrations (see WithTargetVersion).
//
// Migrations are applied in version order while holding a distributed lock,
// so concurrently starting instances only migrate once; the others wait for
// the lock and find the database up to date. Each applied version is recorded
// in the migration info collection as soon as its migration succeeds.
//
// An empty database starts from the latest major migration, which MUST be
// able to initialize the database from scratch. A database that is ahead of
// the migrations within the same major version is left untouched, since
// minor and patch migrations are backward compatible. A database of a newer
// major version, or one that is too old for a major migration (see
// MigrationMajor.MinimumVersion), results in ErrDatabaseIncompatible.
func ApplyMigrations(
	ctx context.Context,
	db *mongo.Database,
	migrations []Migration,
	opts ...MigrationOption,
) error {
	var o migrationOptions
	for _, opt := range opts {
		opt(&o)
	}
	l := log.FromContext(ctx).F(log.Ctx{"db": db.Name()})
	migrations = slices.Clone(migrations)
	slices.SortFunc(migrations, func(a, b Migration) int {
		return compareVersions(a.Version(), b.Version())
	})
	if err := validateMigrations(migrations); err != nil {
		return err
	}
	if o.target != nil {
		n := len(migrations)
		migrations = slices.DeleteFunc(migrations, func(m Migration) bool {
			return migrate.VersionIsLess(*o.target, m.Version())
		})
		if skipped := n - len(migrations); skipped > 0 {
			l.Infof("skipping %d migration(s) above target version %s", skipped, *o.target)
		}
	}
	if len(migrations) == 0 {
		return nil
	}
	target := migrations[len(migrations)-1].Version()

	ctx, cancel := context.WithTimeout(ctx, MigrationLockTTL)
	defer cancel()
	lock := NewLock(db, migrate.DbMigrationsColl, migrationLockID)
	if err := waitForLock(ctx, lock); err != nil {
		return err
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if err := lock.Unlock(unlockCtx); err != nil {
			l.Warnf("failed to release migration lock: %s", err)
		}
	}()

	// Read the history only after acquiring the lock: another instance may
	// have migrated the database while we were waiting.
	history, err := migrate.GetDatabaseMigrationInfo(ctx, db)
	if err != nil {
		return err
	}
	current := latestVersion(history)
	if current.Major > target.Major {
		return fmt.Errorf("%w: database version %s is newer than %s",
			ErrDatabaseIncompatible, current, target)
	}

	var pending []Migration
	if o.forceFromMajor {
		pending, current, err = forcedMigrations(migrations, history, target)
		if err != nil {
			return err
		}
		l.Warnf("forcing migrations from version %s to %s", current, target)
	} else if !migrate.VersionIsLess(current, target) {
		l.Infof("database version %s is up to date (target %s)", current, target)
		return nil
	} else {
		pending = pendingMigrations(migrations, current)
	}

	for _, m := range pending {
		v := m.Version()
		l.Infof("applying %s migration from version %s to %s", KindOf(v), current, v)
		if err := applyMigration(ctx, db, m, current); err != nil {
			return fmt.Errorf("failed to migrate from version %s to %s: %w", current, v, err)
		}
		current = v
	}
	l.Infof("database migrated to version %s", current)
	return nil
}

func compareVersions(a, b migrate.Version) int {
	if migrate.VersionIsLess(a, b) {
		return -1
	} else if migrate.VersionIsLess(b, a) {
		return 1
	}
	return 0
}

// validateMigrations checks that the sorted migrations have unique versions
// and implement the interface required by their kind.
func validateMigrations(migrations []Migration) error {
	for i, m := range migrations {
		v := m.Version()
		if v == (migrate.Version{}) {
			return fmt.Errorf("%w: version 0.0.0 is reserved for empty databases",
				ErrInvalidMigration)
		} else if i > 0 && v == migrations[i-1].Version() {
			return fmt.Errorf("%w: duplicate migrations to version %s", ErrInvalidMigration, v)
		}
		switch KindOf(v) {
		case MigrationKindMajor:
			mm, ok := m.(MigrationMajor)
			if !ok {
				return fmt.Errorf("%w: migration to version %s must implement MigrationMajor",
					ErrInvalidMigration, v)
			}
			if minVersion := mm.MinimumVersion(); minVersion.Major >= v.Major {
				return fmt.Errorf(
					"%w: minimum version %s of migration to %s is not of a previous major",
					ErrInvalidMigration, minVersion, v)
			}
		case MigrationKindMinor:
			if _, ok := m.(MigrationMinor); !ok {
				return fmt.Errorf("%w: migration to version %s must implement MigrationMinor",
					ErrInvalidMigration, v)
			}
		}
	}
	return nil
}

// waitForLock polls lock until it is acquired or ctx is done. The lock expires
// at the deadline of ctx.
func waitForLock(ctx context.Context, lock sync.DistributedLock) error {
	const retryInterval = time.Second
	for {
		ok, err := lock.TryLock(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire migration lock: %w", err)
		} else if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for migration lock: %w", context.Cause(ctx))
		case <-time.After(retryInterval):
		}
	}
}

// latestVersion returns the latest version in history as returned by
// migrate.GetDatabaseMigrationInfo, or 0.0.0 if the database has never been
// migrated.
func latestVersion(history []migrate.MigrationEntry) migrate.Version {
	if len(history) == 0 {
		return migrate.Version{}
	}
	// GetDatabaseMigrationInfo sorts by version in descending order.
	return history[0].Version
}

// forcedMigrations returns the sorted migrations from the latest major
// migration onwards, together with the version to migrate from: the latest
// version recorded before that major version.
func forcedMigrations(
	migrations []Migration,
	history []migrate.MigrationEntry,
	target migrate.Version,
) ([]Migration, migrate.Version, error) {
	if current := latestVersion(history); migrate.VersionIsLess(target, current) {
		return nil, current, fmt.Errorf(
			"%w: cannot force migrations, database version %s is ahead of target %s",
			ErrDatabaseIncompatible, current, target)
	}
	start := -1
	for i, m := range migrations {
		if KindOf(m.Version()) == MigrationKindMajor {
			start = i
		}
	}
	if start < 0 {
		return nil, migrate.Version{}, fmt.Errorf(
			"%w: cannot force migrations without a major migration", ErrInvalidMigration)
	}
	major := migrations[start].Version().Major
	if current := latestVersion(history); current.Major != major {
		return nil, current, fmt.Errorf(
			"%w: cannot force migrations of major version %d, database version is %s",
			ErrDatabaseIncompatible, major, current)
	}
	for _, entry := range history {
		if entry.Version.Major < major {
			return migrations[start:], entry.Version, nil
		}
	}
	return migrations[start:], migrate.Version{}, nil
}

// pendingMigrations returns the sorted migrations that must be applied to a
// database at version current.
func pendingMigrations(migrations []Migration, current migrate.Version) []Migration {
	start := 0
	for i, m := range migrations {
		v := m.Version()
		if current == (migrate.Version{}) {
			// Major migrations initialize an empty database from scratch,
			// so there is no need to run the ones preceding the latest.
			if KindOf(v) == MigrationKindMajor {
				start = i
			}
		} else if !migrate.VersionIsLess(current, v) {
			start = i + 1
		}
	}
	return migrations[start:]
}

// applyMigration runs m on a database at version from and records the new
// version. Crossing a major version requires a major migration that supports
// upgrading from version from.
func applyMigration(
	ctx context.Context,
	db *mongo.Database,
	m Migration,
	from migrate.Version,
) error {
	v := m.Version()
	if v.Major != from.Major {
		mm, ok := m.(MigrationMajor)
		if !ok {
			return fmt.Errorf("%w: missing major migration to version %d.0.0",
				ErrDatabaseIncompatible, v.Major)
		}
		minVersion := mm.MinimumVersion()
		if from != (migrate.Version{}) && migrate.VersionIsLess(from, minVersion) {
			return fmt.Errorf("%w: version %s or later is required, "+
				"upgrade using an older release first", ErrDatabaseIncompatible, minVersion)
		}
	}
	if err := m.Up(ctx, db, from); err != nil {
		return err
	}
	return migrate.UpdateMigrationInfo(ctx, v, db.Client(), db.Name())
}
