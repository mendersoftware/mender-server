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
	limitFromContext RatelimitParamsFunc,
) rate.Limiter {
	return &fixedWindowRatelimiter{
		client:     client,
		paramsFunc: limitFromContext,
		nowFunc:    time.Now,
		keyPrefix:  keyPrefix,
		interval:   interval,
	}
}

type RatelimitParamsFunc func(context.Context) (burst uint64, eventID string, err error)

func FixedRatelimitParams(burst uint64) RatelimitParamsFunc {
	return func(ctx context.Context) (uint64, string, error) {
		return burst, "", nil
	}
}

type fixedWindowRatelimiter struct {
	client     Client
	paramsFunc RatelimitParamsFunc
	nowFunc    func() time.Time
	keyPrefix  string

	interval time.Duration
}

type simpleReservation struct {
	ok     bool
	tokens uint64
	delay  time.Duration
}

func (r *simpleReservation) OK() bool {
	return r.ok
}

func (r *simpleReservation) Delay() time.Duration {
	return r.delay
}

func (r *simpleReservation) Tokens() uint64 {
	return r.tokens
}

func epoch(t time.Time, interval time.Duration) int64 {
	return t.UnixMilli() / interval.Milliseconds()
}

func fixedWindowKey(prefix, eventID string, epoch int64) string {
	if prefix == "" {
		prefix = "ratelimit"
	}
	return fmt.Sprintf("%s:id:%s:e:%d:c", prefix, eventID, epoch)
}

func (rl *fixedWindowRatelimiter) Reserve(ctx context.Context) (rate.Reservation, error) {
	now := rl.nowFunc()
	burst, eventID, err := rl.paramsFunc(ctx)
	if err != nil {
		return nil, err
	}
	epoch := epoch(now, rl.interval)
	key := fixedWindowKey(rl.keyPrefix, eventID, epoch)
	count := uint64(1)

	err = rl.client.SetArgs(ctx, key, count, redis.SetArgs{
		TTL:  rl.interval,
		Mode: `NX`,
	}).Err()
	if errors.Is(err, redis.Nil) {
		count, err = rl.client.Incr(ctx, key).Uint64()
	}
	if err != nil {
		return nil, fmt.Errorf("redis: error computing rate limit: %w", err)
	}
	if count <= burst {
		return &simpleReservation{
			delay:  0,
			ok:     true,
			tokens: burst - count,
		}, nil
	}
	return &simpleReservation{
		delay: now.Sub(time.UnixMilli((epoch + 1) *
			rl.interval.Milliseconds())),
		ok:     false,
		tokens: 0,
	}, nil
}

func (rl *fixedWindowRatelimiter) Tokens(ctx context.Context) (uint64, error) {
	burst, eventID, err := rl.paramsFunc(ctx)
	if err != nil {
		return 0, err
	}
	count, err := rl.client.Get(ctx,
		fixedWindowKey(rl.keyPrefix,
			eventID,
			epoch(rl.nowFunc(), rl.interval),
		),
	).Uint64()
	if errors.Is(err, redis.Nil) {
		return burst, nil
	} else if err != nil {
		return 0, fmt.Errorf("redis: error getting free tokens: %w", err)
	} else if count > burst {
		return 0, nil
	}
	return burst - count, nil
}
