package redis

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mendersoftware/mender-server/pkg/rate"
)

func TestFixedWindowRatelimit(t *testing.T) {
	requireRedis(t)
	t.Parallel()

	ctx := context.Background()

	client, err := ClientFromConnectionString(ctx, RedisURL)
	if err != nil {
		t.Errorf("could not connect to redis (%s): is redis running?",
			RedisURL)
		t.FailNow()
	}
	tMicro := time.Now().UnixMicro()
	params := FixedRatelimitParams(RatelimitParams{
		Burst:     1,
		Interval:  time.Hour,
		KeyPrefix: fmt.Sprintf("%s_%x", strings.ToLower(t.Name()), tMicro),
	})
	rateLimiter := NewFixedWindowRateLimiter(client, params)

	// Freeze time to avoid time to progress to next window.
	nowFrozen := time.Now()
	rateLimiter.(*fixedWindowRatelimiter).nowFunc = func() time.Time { return nowFrozen }

	if tokens, _ := rateLimiter.Tokens(ctx); tokens != 1 {
		t.Errorf("expected token available after initialization, actual: %d", tokens)
	}

	var reservations [2]rate.Reservation
	for i := 0; i < len(reservations); i++ {
		reservations[i], err = rateLimiter.Reserve(ctx)
		if err != nil {
			t.Errorf("unexpected error reserving rate limit: %s", err.Error())
			t.FailNow()
		}
	}
	if !reservations[0].OK() {
		t.Errorf("expected first event to pass, but didn't")
	}
	if reservations[1].OK() {
		t.Errorf("expected the second event to block, but didn't")
	}
	if remaining, err := rateLimiter.Tokens(ctx); err != nil {
		t.Errorf("unexpected error retrieving remaining tokens: %s", err.Error())
	} else if remaining != 0 {
		t.Errorf("expected 0 tokens remaining, actual: %d", remaining)
	}

	if reservations[0].Tokens() != 0 {
		t.Errorf("there should be no tokens left after first event")
	} else if reservations[1].Tokens() != 0 {
		t.Errorf("there should be no tokens left after second event")
	}
}
