package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mendersoftware/mender-server/pkg/rate"
)

func NewFixedWindowRateLimiter(
	client Client,
	keyPrefix string,
	interval time.Duration,
	quota int64,
) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		client:    client,
		nowFunc:   time.Now,
		keyPrefix: keyPrefix,
		interval:  interval,
		quota:     quota,
	}
}

var (
	_ rate.Limiter      = &FixedWindowRateLimiter{}
	_ rate.EventLimiter = &FixedWindowRateLimiter{}
)

// FixedWindowRateLimiter implements a version of the algorithm described
// at https://redis.io/glossary/rate-limiting/
type FixedWindowRateLimiter struct {
	client    Client
	nowFunc   func() time.Time
	keyPrefix string

	interval time.Duration
	quota    int64
}

// simpleReservation is a straight forward implementation of the
// rate.Reservation interface.
type simpleReservation struct {
	// tokens count the number of tokens (events) remaining after
	// reservation has been made.
	tokens int64
	// delay is the time the client would need to wait for a token (event)
	// to become available. If 0 or negative, the reservation is accepted
	// that is, (*simpleReservation).OK() == true.
	delay time.Duration
}

func (r *simpleReservation) OK() bool {
	return r.delay <= 0
}

func (r *simpleReservation) Delay() time.Duration {
	return r.delay
}

func (r *simpleReservation) Tokens() int64 {
	return r.tokens
}

func epoch(t time.Time, interval time.Duration) int64 {
	return t.UnixMilli() / interval.Milliseconds()
}

func fixedWindowKey(prefix, eventID string, epoch int64) string {
	if prefix == "" {
		prefix = "ratelimit"
	}
	if eventID == "" {
		return fmt.Sprintf("%s:e:%d:c", prefix, epoch)
	} else {
		return fmt.Sprintf("%s:%s:e:%d:c", prefix, eventID, epoch)
	}
}

func (rl *FixedWindowRateLimiter) ReserveEvent(
	ctx context.Context,
	eventID string,
) (rate.Reservation, error) {
	now := rl.nowFunc()
	epoch := epoch(now, rl.interval)
	key := fixedWindowKey(rl.keyPrefix, eventID, epoch)
	count := int64(1)

	err := rl.client.SetArgs(ctx, key, count, redis.SetArgs{
		TTL:  rl.interval,
		Mode: `NX`,
	}).Err()
	if errors.Is(err, redis.Nil) {
		count, err = rl.client.Incr(ctx, key).Result()
	}
	if err != nil {
		return nil, fmt.Errorf("redis: error computing rate limit: %w", err)
	}
	if count <= rl.quota {
		return &simpleReservation{
			delay:  0,
			tokens: rl.quota - count,
		}, nil
	}
	return &simpleReservation{
		delay: now.Sub(time.UnixMilli((epoch + 1) *
			rl.interval.Milliseconds())),
		tokens: 0,
	}, nil
}

func (rl *FixedWindowRateLimiter) Reserve(ctx context.Context) (rate.Reservation, error) {
	return rl.ReserveEvent(ctx, "")
}
