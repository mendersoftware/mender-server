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
	paramsFromContext RatelimitParamsFunc,
) rate.Limiter {
	return &fixedWindowRatelimiter{
		client:     client,
		paramsFunc: paramsFromContext,
		nowFunc:    time.Now,
	}
}

type RatelimitParams struct {
	Burst     uint64
	Interval  time.Duration
	KeyPrefix string
}

type RatelimitParamsFunc func(context.Context) (*RatelimitParams, error)

func FixedRatelimitParams(params RatelimitParams) RatelimitParamsFunc {
	return func(ctx context.Context) (*RatelimitParams, error) {
		return &params, nil
	}
}

type fixedWindowRatelimiter struct {
	client     Client
	paramsFunc RatelimitParamsFunc
	nowFunc    func() time.Time
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

func fixedWindowKey(prefix string, epoch int64) string {
	if prefix == "" {
		prefix = "ratelimit"
	}
	return fmt.Sprintf("%s:e:%d:c", prefix, epoch)
}

func (rl *fixedWindowRatelimiter) Reserve(ctx context.Context) (rate.Reservation, error) {
	now := rl.nowFunc()
	params, err := rl.paramsFunc(ctx)
	if err != nil {
		return nil, err
	} else if params == nil {
		return &simpleReservation{
			ok: true,
		}, nil
	}
	epoch := epoch(now, params.Interval)
	key := fixedWindowKey(params.KeyPrefix, epoch)
	count := uint64(1)

	err = rl.client.SetArgs(ctx, key, count, redis.SetArgs{
		TTL:  params.Interval,
		Mode: `NX`,
	}).Err()
	if errors.Is(err, redis.Nil) {
		count, err = rl.client.Incr(ctx, key).Uint64()
	}
	if err != nil {
		return nil, fmt.Errorf("redis: error computing rate limit: %w", err)
	}
	if count <= params.Burst {
		return &simpleReservation{
			delay:  0,
			ok:     true,
			tokens: params.Burst - count,
		}, nil
	}
	return &simpleReservation{
		delay: now.Sub(time.UnixMilli((epoch + 1) *
			params.Interval.Milliseconds())),
		ok:     false,
		tokens: 0,
	}, nil
}

func (rl *fixedWindowRatelimiter) Tokens(ctx context.Context) (uint64, error) {
	params, err := rl.paramsFunc(ctx)
	if err != nil {
		return 0, err
	}
	count, err := rl.client.Get(ctx,
		fixedWindowKey(params.KeyPrefix, epoch(rl.nowFunc(), params.Interval)),
	).Uint64()
	if errors.Is(err, redis.Nil) {
		return params.Burst, nil
	} else if err != nil {
		return 0, fmt.Errorf("redis: error getting free tokens: %w", err)
	} else if count > params.Burst {
		return 0, nil
	}
	return params.Burst - count, nil
}
