//nolint:all // This is all test code
package common

import (
	"context"
	"errors"
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
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		ok, err := fn()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("timed out after %s waiting for condition", timeout)
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
