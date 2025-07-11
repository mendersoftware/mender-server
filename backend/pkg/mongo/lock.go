package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/mendersoftware/mender-server/pkg/sync"
)

type distributedLock struct {
	coll   *mongo.Collection
	lockID string
	lock   lockDoc
}

func NewLockGenerator(db *mongo.Database, collName string) sync.DistributedLockGenerator {
	return func(lockID string) (sync.DistributedLock, error) {
		return NewLock(db, collName, lockID), nil
	}
}

// NewLock creates a DistributedLock using a single document in a collection.
// The implementation relies on the uniqueness of the primary index (_id), and
// uses a combination of timestamps with document versioning to prevent
// "stealing" locks aquired by other processes.
func NewLock(db *mongo.Database, collName, key string) sync.DistributedLock {
	coll := db.Collection(collName, options.
		Collection().
		SetWriteConcern(writeconcern.Majority()))
	return &distributedLock{
		coll:   coll,
		lockID: key,
	}
}

// lockDoc is the document describing the lock.
type lockDoc struct {
	// ID is the unique (resource) ID of the lock
	ID string `bson:"_id"`
	// Exp is the lock expiry time
	Exp time.Time `bson:"exp"`
	// Created is the time the lock was created (aquired)
	Created time.Time `bson:"created"`
	// Version is an attribute that is always incremented when aquiring the lock
	// this ensures (together with exp) that once lock is aquired it will not
	// release another process' if it expires.
	Version uint32 `bson:"v"`
}

func (l *distributedLock) TryLock(ctx context.Context) (bool, error) {
	const defaultTTL = time.Minute * 2
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(defaultTTL)
	}
	now := time.Now()
	var lock lockDoc
	err := l.coll.FindOneAndUpdate(ctx,
		bson.M{"_id": l.lockID, "exp": bson.M{"$lt": now}},
		bson.M{
			"$set": bson.M{"_id": l.lockID, "exp": deadline, "created": now},
			"$inc": bson.M{"v": 1},
		},
		options.FindOneAndUpdate().
			SetReturnDocument(options.After).
			SetUpsert(true)).
		Decode(&lock)
	if mongo.IsDuplicateKeyError(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to aquire lock: %w", err)
	}
	l.lock = lock
	return true, nil
}

func (l *distributedLock) Unlock(ctx context.Context) error {
	_, err := l.coll.DeleteOne(ctx, l.lock)
	return err
}
