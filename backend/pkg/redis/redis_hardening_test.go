package redis

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func newTestHardeningHook() *clusterHardeningHook {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"127.0.0.1:1"},
	})
	return newClusterHardeningHook(rdb, 0)
}

func TestClusterHardeningHookRecord(t *testing.T) {
	t.Run("success resets the counter", func(t *testing.T) {
		h := newTestHardeningHook()
		h.consecutiveErrs = 3
		h.record(nil)
		if h.consecutiveErrs != 0 {
			t.Fatalf("expected counter reset to 0, got %d", h.consecutiveErrs)
		}
	})

	qualifying := []struct {
		name string
		err  error
	}{
		{"cluster down", errors.New("CLUSTERDOWN The cluster is down")},
		{"loading", errors.New("LOADING Redis is loading the dataset in memory")},
		{"master down", errors.New("MASTERDOWN Link with MASTER is down")},
		{"try again", errors.New("TRYAGAIN try again")},
		{"network timeout", &net.DNSError{IsTimeout: true}},
	}
	for _, tc := range qualifying {
		t.Run(tc.name+" increments", func(t *testing.T) {
			h := newTestHardeningHook()
			h.record(tc.err)
			if h.consecutiveErrs != 1 {
				t.Fatalf("expected counter to increment to 1, got %d", h.consecutiveErrs)
			}
		})
	}

	nonQualifying := []struct {
		name string
		err  error
	}{
		{"redis.Nil", redis.Nil},
		{"WRONGTYPE", errors.New("WRONGTYPE Operation against a key")},
	}
	for _, tc := range nonQualifying {
		t.Run(tc.name+" is a no-op, not a reset", func(t *testing.T) {
			h := newTestHardeningHook()
			h.consecutiveErrs = 3
			h.record(tc.err)
			if h.consecutiveErrs != 3 {
				t.Fatalf("expected counter unchanged at 3, got %d", h.consecutiveErrs)
			}
		})
	}

	t.Run("forces a reload after the threshold, rate-limited", func(t *testing.T) {
		h := newTestHardeningHook()
		qualifyingErr := errors.New("CLUSTERDOWN The cluster is down")
		for i := 0; i < clusterForceReloadAfterN; i++ {
			h.record(qualifyingErr)
		}
		if h.lastForcedReload.IsZero() {
			t.Fatal("expected a forced reload to have been triggered")
		}
		first := h.lastForcedReload
		h.record(qualifyingErr)
		if !h.lastForcedReload.Equal(first) {
			t.Fatal("expected the cooldown to suppress an immediate second forced reload")
		}
	})
}

func TestClusterHardeningHookDialHook(t *testing.T) {
	h := newTestHardeningHook()
	called := false
	next := func(ctx context.Context, network, addr string) (net.Conn, error) {
		called = true
		return nil, nil
	}
	_, _ = h.DialHook(next)(context.Background(), "tcp", "addr")
	if !called {
		t.Fatal("expected DialHook to pass through to next unchanged")
	}
}

func TestClusterHardeningHookProcessHook(t *testing.T) {
	t.Run("retries on timeout up to the configured limit, then records the failure", func(t *testing.T) {
		h := newTestHardeningHook()
		h.retries = 2
		attempts := 0
		next := func(ctx context.Context, cmd redis.Cmder) error {
			attempts++
			return &net.DNSError{IsTimeout: true}
		}
		cmd := redis.NewStatusCmd(context.Background())
		err := h.ProcessHook(next)(context.Background(), cmd)
		if err == nil {
			t.Fatal("expected an error")
		}
		if want := h.retries + 1; attempts != want {
			t.Fatalf("expected %d attempts, got %d", want, attempts)
		}
		if h.consecutiveErrs != 1 {
			t.Fatalf("expected the hook to have recorded the failure, consecutiveErrs=%d", h.consecutiveErrs)
		}
	})

	t.Run("does not retry a non-timeout, non-TRYAGAIN error", func(t *testing.T) {
		h := newTestHardeningHook()
		h.retries = 2
		attempts := 0
		next := func(ctx context.Context, cmd redis.Cmder) error {
			attempts++
			return errors.New("WRONGTYPE Operation against a key")
		}
		cmd := redis.NewStatusCmd(context.Background())
		_ = h.ProcessHook(next)(context.Background(), cmd)
		if attempts != 1 {
			t.Fatalf("expected exactly 1 attempt (no retry), got %d", attempts)
		}
	})

	t.Run("succeeding resets the counter", func(t *testing.T) {
		h := newTestHardeningHook()
		h.consecutiveErrs = 3
		next := func(ctx context.Context, cmd redis.Cmder) error {
			return nil
		}
		cmd := redis.NewStatusCmd(context.Background())
		if err := h.ProcessHook(next)(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if h.consecutiveErrs != 0 {
			t.Fatalf("expected counter reset to 0, got %d", h.consecutiveErrs)
		}
	})
}

func TestClusterHardeningHookProcessPipelineHook(t *testing.T) {
	h := newTestHardeningHook()
	next := func(ctx context.Context, cmds []redis.Cmder) error {
		return errors.New("CLUSTERDOWN The cluster is down")
	}
	if err := h.ProcessPipelineHook(next)(context.Background(), nil); err == nil {
		t.Fatal("expected an error")
	}
	if h.consecutiveErrs != 1 {
		t.Fatalf("expected the hook to have recorded the failure, consecutiveErrs=%d", h.consecutiveErrs)
	}
}

func TestClientFromConnectionStringClusterMode(t *testing.T) {
	// A closed local port: exercises the isCluster branch of
	// ClientFromConnectionString (parsing, hook attachment, background
	// reload startup) without needing a live cluster - Ping is expected
	// to fail, but everything before it already ran.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ClientFromConnectionString(ctx, "redis://127.0.0.1:1?addr=127.0.0.1:1")
	if err == nil {
		t.Fatal("expected an error connecting through a closed port")
	}
}
