package rate

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// Limiter implements a rate limiting interface based on golang.org/x/time/rate
// but extending the interface with the ability to expose internal errors.
type Limiter interface {
	Reserve(ctx context.Context) (Reservation, error)
	Tokens(ctx context.Context) (uint64, error)
}

// Reservation is a point-in-time reservation of a ratelimit token. If the token
// is approved OK return true, otherways Delay will return the duration for next
// token(s) to become available. While Tokens return the number of available
// tokens after the reservation.
type Reservation interface {
	OK() bool
	Delay() time.Duration
	Tokens() uint64
}

type limiter rate.Limiter

func NewLimiter(limit int, interval time.Duration) Limiter {
	return (*limiter)(rate.NewLimiter(rate.Every(interval/time.Duration(limit)), limit))
}

func (lim *limiter) Reserve(context.Context) (Reservation, error) {
	now := time.Now()
	goLimiter := (*rate.Limiter)(lim)
	res := &reservation{
		reservation: goLimiter.ReserveN(now, 1),
		time:        now,
	}
	if res.OK() {
		res.tokens = uint64(goLimiter.TokensAt(now))
	}
	return res, nil
}

func (lim *limiter) Tokens(context.Context) (uint64, error) {
	goLimiter := (*rate.Limiter)(lim)
	if tokens := goLimiter.Tokens(); tokens > 0 {
		return uint64(tokens), nil
	}
	return 0, nil
}

type reservation struct {
	time        time.Time
	tokens      uint64
	reservation *rate.Reservation
}

func (r *reservation) OK() bool {
	return r.Delay() == 0
}

func (r *reservation) Delay() time.Duration {
	return r.reservation.DelayFrom(r.time)
}

func (r *reservation) Tokens() uint64 {
	return r.tokens
}
