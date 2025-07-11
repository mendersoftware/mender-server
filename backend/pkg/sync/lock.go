package sync

import (
	"context"
)

// DistributedLock provides an interface similar to sync.Locker with the
// addition of a context and non-blocking* TryLock instead of Lock.
// A DistributedLock should provide mutually exclusive access to the given
// lock for up to a given amount of time (determined by the Context.Deadline).
//
// *The implementation may actually block on network interfaces, but MUST not
// block on another process to release the lock.
type DistributedLock interface {
	// TryLock aquires a lease of the lock.
	//
	// The context SHOULD always have a deadline and the DistributedLock implementation
	// MUST always expire the lock after a certain period of time to avoid
	// permanent deadlocks.
	TryLock(ctx context.Context) (bool, error)

	// Unlock releases the lock (lease).
	//
	// The DistributedLock implementation MUST ensure that calling Unlock on an
	// expired lease does not release a lock aquired after the lock expires.
	Unlock(ctx context.Context) error
}

// DistributedLockGenerator initializes a new unique lock for the given
// resourceName.
type DistributedLockGenerator func(resourceName string) (DistributedLock, error)
