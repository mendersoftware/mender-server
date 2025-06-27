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
	keyPrefix := fmt.Sprintf("%s_%x", strings.ToLower(t.Name()), tMicro)
	rateLimiter := NewFixedWindowRateLimiter(client, keyPrefix,
		time.Minute,
		1)

	// Freeze time to avoid time to progress to next window.
	nowFrozen := time.Now()
	rateLimiter.nowFunc = func() time.Time { return nowFrozen }

	count, _ := rateLimiter.client.Get(ctx, fixedWindowKey(
		keyPrefix, "", epoch(nowFrozen, time.Minute),
	)).Int64()
	if count != 0 {
		t.Errorf("expected zero count after initialization, actual: %d", count)
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

	count, err = rateLimiter.client.Get(ctx, fixedWindowKey(
		keyPrefix, "", epoch(nowFrozen, time.Minute),
	)).Int64()
	if err != nil {
		t.Errorf("unexpected error retrieving remaining tokens: %s", err.Error())
	} else if count != 2 {
		t.Errorf("expected count to be 2 after two calls to Reserve, actual: %d", count)
	}

	if reservations[0].Tokens() != 0 {
		t.Errorf("there should be no tokens left after first event")
	} else if reservations[1].Tokens() != 0 {
		t.Errorf("there should be no tokens left after second event")
	}
}
