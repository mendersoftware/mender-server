//nolint:all // This is all test code
package common

import (
	"context"
	"fmt"
	"time"
)

// RetryUntil calls fn repeatedly, waiting period between attempts, until fn
// reports success (true, nil), fn returns an error (which aborts the retry
// loop immediately), ctx is done, or the timeout budget runs out.
//
// fn decides for itself whether an error is fatal (return it, aborting the
// loop) or should just be treated as "not ready yet" (swallow it and return
// false, nil).
//
// Returns nil on success, fn's error if it aborted the loop, ctx.Err() if
// ctx was done, or a timeout error if the budget ran out without success.
func RetryUntil(ctx context.Context, timeout, period time.Duration, fn func() (bool, error)) error {
	deadline := time.Now().Add(timeout)
	for {
		ok, err := fn()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for condition", timeout)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(period):
		}
	}
}
