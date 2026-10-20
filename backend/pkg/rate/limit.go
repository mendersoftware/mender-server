package rate

import (
	"context"
	"math"
	"time"

	"golang.org/x/time/rate"
)

// Limiter implements a rate limiting interface based on golang.org/x/time/rate
// but extending the interface with the ability to expose internal errors.
type Limiter interface {
	Reserve(ctx context.Context) (Reservation, error)
	ReserveN(ctx context.Context, weight float64) (Reservation, error)
}

var _ Limiter = LimiterFunc(nil)

type LimiterFunc func(context.Context, float64) (Reservation, error)

func (f LimiterFunc) Reserve(ctx context.Context) (Reservation, error) {
	return f(ctx, 1)
}

func (f LimiterFunc) ReserveN(ctx context.Context, weight float64) (Reservation, error) {
	return f(ctx, weight)
}

// EventLimiter allows grouping limiters by eventID.
type EventLimiter interface {
	ReserveEvent(ctx context.Context, eventID string, weight float64) (Reservation, error)
}

// Reservation is a point-in-time reservation of a ratelimit token. If the token
// is approved OK return true, otherways Delay will return the duration for next
// token(s) to become available. While Tokens return the number of available
// tokens after the reservation.
type Reservation interface {
	OK() bool
	Delay() time.Duration
	Tokens() float64
}

type limiter rate.Limiter

func NewLimiter(limit int, interval time.Duration) Limiter {
	return (*limiter)(rate.NewLimiter(rate.Every(interval/time.Duration(limit)), limit))
}

func (lim *limiter) Reserve(ctx context.Context) (Reservation, error) {
	return lim.ReserveN(ctx, 1)
}

func (lim *limiter) ReserveN(ctx context.Context, weight float64) (Reservation, error) {
	now := time.Now()
	goLimiter := (*rate.Limiter)(lim)

	tokensToConsume := int(math.Ceil(weight))
	if tokensToConsume < 0 {
		tokensToConsume = 0
	}

	res := &reservation{
		reservation: goLimiter.ReserveN(now, tokensToConsume),
		time:        now,
	}
	if res.OK() {
		res.tokens = goLimiter.TokensAt(now)
	}
	return res, nil
}

type reservation struct {
	time        time.Time
	tokens      float64
	reservation *rate.Reservation
}

func (r *reservation) OK() bool {
	return r.Delay() == 0
}

func (r *reservation) Delay() time.Duration {
	return r.reservation.DelayFrom(r.time)
}

func (r *reservation) Tokens() float64 {
	return r.tokens
}
