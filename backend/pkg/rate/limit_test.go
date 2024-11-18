package rate

import (
	"context"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	limiter := NewLimiter(2, time.Hour)
	resApproved, err := limiter.Reserve(ctx)
	if err != nil {
		t.Errorf("unexpected error reserving time slot #1: %s", err.Error())
		t.FailNow()
	}
	_, err = limiter.Reserve(ctx)
	if err != nil {
		t.Errorf("unexpected error reserving time slot #2: %s", err.Error())
		t.FailNow()
	}
	resDenied, err := limiter.Reserve(ctx)
	if err != nil {
		t.Errorf("unexpected error reserving time slot #3: %s", err.Error())
		t.FailNow()
	}
	if !resApproved.OK() {
		t.Error("expected first reservation to be available")
	} else if resApproved.Delay() > 0 {
		t.Error("an approved reservation should not have a delay")
	} else if resApproved.Tokens() == 0 {
		t.Error("expected more tokens to be available after first reservation")
	}
	if resDenied.OK() {
		t.Error("reservation should not be available before 1h has passed")
	} else {
		if resDenied.Delay() == 0 {
			t.Error("a denied reservation should have a non-zero delay")
		}
		if resDenied.Tokens() > 0 {
			t.Errorf("a deneied reservation should not report free tokens, got: %d", resDenied.Tokens())
		}
	}
}
