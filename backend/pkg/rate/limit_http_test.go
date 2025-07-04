package rate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/identity"
)

type eventLimiter struct {
	limiters map[string]Limiter
	quota    int
	interval time.Duration
	mu       sync.Mutex
}

func NewEventLimiter(q int, i time.Duration) EventLimiter {
	return &eventLimiter{
		quota:    q,
		interval: i,
		limiters: make(map[string]Limiter),
	}
}

func (lim *eventLimiter) ReserveEvent(ctx context.Context, eventID string) (Reservation, error) {
	fmt.Fprintf(os.Stderr, "ReserveEvent(ctx, %q)\n", eventID)
	lim.mu.Lock()
	if lim.limiters == nil {
		if lim.limiters == nil {
			lim.limiters = make(map[string]Limiter)
		}
	}
	l, ok := lim.limiters[eventID]
	if !ok {
		l = NewLimiter(lim.quota, lim.interval)
		lim.limiters[eventID] = l
	}
	lim.mu.Unlock()
	return l.Reserve(ctx)
}

func newHTTPLimiterForTesting(t *testing.T) *HTTPLimiter {
	t.Helper()
	httpLimiter, err := NewHTTPLimiter(NewEventLimiter(10, time.Hour*24), `{{ .Header.Get "X-Userid" }}`)
	if err != nil {
		t.Errorf("unexpected error initializing HTTPLimiter: %s", err)
		return nil
	}
	httpLimiter.AddRateLimitGroup(NewEventLimiter(1, time.Hour), "slow", `slow`)
	httpLimiter.AddRateLimitGroup(NewEventLimiter(1, time.Hour*24), "superslow", `superslow`)
	httpLimiter.AddRateLimitGroup(NewEventLimiter(1, time.Microsecond), "fast", `fast`)
	httpLimiter.AddRateLimitGroup(NewEventLimiter(5, time.Hour), "sub", `{{with .Identity }}{{.Subject}}{{end}}`)

	httpLimiter.MatchHTTPPattern("/slow/", "slow")
	httpLimiter.MatchHTTPPattern("POST /slow/superslow", "superslow")
	httpLimiter.MatchHTTPPattern("/fast/", "fast")
	httpLimiter.MatchHTTPPattern("/group/", `{{ .Header.Get "X-Test-Group"}}`)
	httpLimiter.MatchHTTPPattern("/subject/", `sub`)
	httpLimiter.MatchHTTPPattern("/bad_template/", `{{.Foobar}}`)
	return httpLimiter
}

func TestHTTPLimiter(t *testing.T) {
	t.Parallel()

	limiter := newHTTPLimiterForTesting(t)
	limiter.WithTemplateDataFunc(func(r *http.Request) any {
		return r
	})

	t.Run("ServeHTTP", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest("GET", "http://localhost/", nil)
		for range 10 {
			w := httptest.NewRecorder()
			limiter.ServeHTTP(w, req)
			if w.Code != 200 {
				t.Errorf("did not expect ServeHTTP to write response: received status %d", w.Code)
				t.Errorf("response body: %s", w.Body)
			}
		}
		// The 10th request should trip the rate limiter
		w := httptest.NewRecorder()
		limiter.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("expected status code %d, received: %d", http.StatusTooManyRequests, w.Code)
		}

		w = httptest.NewRecorder()
		req.Header.Set("X-Userid", "another user")
		limiter.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, received: %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Reserve", func(t *testing.T) {
		t.Parallel()
		reqFast, _ := http.NewRequest("GET", "http://localhost/fast/apis", nil)
		res, _ := limiter.Reserve(reqFast)
		if !res.OK() {
			t.Error("unexpected rate limit on first request to fast rate limiter")
		} else {
			time.Sleep(time.Microsecond)
			res, _ = limiter.Reserve(reqFast)
			if !res.OK() {
				t.Error("unexpected rate limit after waiting on fast request")
			}
		}

		reqSuperSlow, _ := http.NewRequest("POST", "http://localhost/slow/superslow", nil)
		res, _ = limiter.Reserve(reqSuperSlow)
		if !res.OK() {
			t.Errorf("unexpected rate limit for superslow request")
		} else {
			res, _ = limiter.Reserve(reqSuperSlow)
			if res.OK() {
				t.Errorf("request to superslow should get rate limited")
			} else if res.Delay() < time.Hour*23 && res.Delay() > time.Hour*25 {
				t.Errorf("superslow request delay is not close to 24h: actual: %s", res.Delay())
			}
		}
	})
	const jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9" +
		"lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0." +
		"KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30"
	t.Run("MiddlewareGin", func(t *testing.T) {
		limiterGin := newHTTPLimiterForTesting(t)
		handler := gin.New()
		handler.Use(accesslog.Middleware(), identity.Middleware(), limiterGin.MiddlewareGin)
		handler.Handle("GET", "/subject/foo", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
		req, _ := http.NewRequest("GET", "http://localhost/subject/foo", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)
		for range 5 {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("unexpected status code: %d, expected: %d", w.Code, http.StatusOK)
			}
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("unexpected status code: %d, expected: %d", w.Code, http.StatusTooManyRequests)
		}
		req.URL.Path = "/bad_template/"
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("unexpected status code: %d, expected: %d", w.Code, http.StatusInternalServerError)
			t.Error("expected template execution to fail")
		}
	})
}
