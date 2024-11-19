package devauth

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/plan"
	"github.com/mendersoftware/mender-server/pkg/rate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	mstore "github.com/mendersoftware/mender-server/services/deviceauth/store/mocks"
)

type errLimiter struct {
	rate.Limiter
	err error
}

func (l errLimiter) Reserve(ctx context.Context) (rate.Reservation, error) {
	return nil, l.err
}

func TestCheckRateLimits(t *testing.T) {
	t.Parallel()

	t.Run("ok/token bucket", func(t *testing.T) {
		rateLimiter := rate.NewLimiter(1, time.Hour)
		d := new(DevAuth).WithRatelimits(rateLimiter, make(map[string]float64), 1.0)
		ctx := context.Background()

		err := d.checkRateLimits(ctx)
		if err != nil {
			t.Errorf("unexpected error on first rate limiter event: %s", err.Error())
		}
		err = d.checkRateLimits(ctx)
		if !errors.Is(err, cache.ErrTooManyRequests) {
			if err == nil {
				t.Errorf("expected error %q, received none", cache.ErrTooManyRequests.Error())
			} else {
				t.Errorf("unexpected error on second rate limiter event: %s", err.Error())
			}
		}
	})
	t.Run("error/unknown error propagated", func(t *testing.T) {
		var errExpected = errors.New("test error")
		d := new(DevAuth).WithRatelimits(errLimiter{err: errExpected}, map[string]float64{}, 1.0)
		errActual := d.checkRateLimits(context.Background())
		if !errors.Is(errActual, errExpected) {
			t.Errorf("unexpected error: %s", errActual.Error())
		}
	})
}

func TestRateLimitParamsFromContext(t *testing.T) {
	t.Parallel()

	type testCase struct {
		CTX     context.Context
		Store   func(t *testing.T) store.DataStore
		Weights map[string]float64

		ResultLimit   int64
		ResultEventID string
		ResultError   error
	}
	contextArgMatcher := mock.MatchedBy(func(context.Context) bool { return true })

	for name, _tc := range map[string]testCase{
		"ok/no tenant context": testCase{
			CTX: httpheader.WithContext(context.Background(),
				http.Header{
					"X-Forwarded-Uri": []string{"/api/devices/v1/foo/bar"},
				}, "X-Forwarded-Uri"),
			Store: func(t *testing.T) store.DataStore {
				ds := mstore.NewDataStore(t)
				ds.On("GetLimit", contextArgMatcher, model.LimitMaxDeviceCount).
					Return(&model.Limit{Name: model.LimitMaxDeviceCount, Value: 69}, nil)
				return ds
			},

			ResultEventID: fmtEventID("default", "foo/bar"),
			ResultLimit:   69,
		},
		"ok/with tenant context": testCase{
			CTX: identity.WithContext(
				httpheader.WithContext(
					context.Background(),
					http.Header{
						"X-Forwarded-Uri": []string{"/api/devices/v1/foo/bar"},
					}, "X-Forwarded-Uri"),
				&identity.Identity{
					Tenant: "1234",
				}),
			Store: func(t *testing.T) store.DataStore {
				ds := mstore.NewDataStore(t)
				ds.On("GetLimit", contextArgMatcher, model.LimitMaxDeviceCount).
					Return(&model.Limit{Name: model.LimitMaxDeviceCount, Value: 123}, nil)
				return ds
			},

			ResultEventID: fmtEventID("1234", "foo/bar"),
			ResultLimit:   123,
		},
		"ok/float and int overflow": testCase{
			CTX: identity.WithContext(
				httpheader.WithContext(
					context.Background(),
					http.Header{
						"X-Forwarded-Uri": []string{"/api/devices/v1/foo/bar"},
					}, "X-Forwarded-Uri"),
				&identity.Identity{
					Tenant: "1234",
					Plan:   plan.PlanEnterprise,
				}),
			Store: func(t *testing.T) store.DataStore {
				ds := mstore.NewDataStore(t)
				ds.On("GetLimit", contextArgMatcher, model.LimitMaxDeviceCount).
					Return(&model.Limit{Name: model.LimitMaxDeviceCount, Value: (1 << 61)}, nil)
				return ds
			},
			Weights: map[string]float64{
				plan.PlanEnterprise:   10.0,
				plan.PlanProfessional: 5.0,
				plan.PlanOpenSource:   2.0,
			},

			ResultEventID: fmtEventID("1234", "foo/bar"),
			ResultLimit:   rateLimitMax,
		},
	} {
		tc := _tc
		t.Run(name, func(t *testing.T) {
			ds := tc.Store(t)
			devauth := NewDevAuth(ds, nil, nil, Config{}).
				WithRatelimits(rate.NewLimiter(1, time.Minute), tc.Weights, 1.0)
			limit, eventID, err := devauth.RateLimitsFromContext(tc.CTX)
			assert.Equal(t, tc.ResultLimit, limit)
			assert.Equal(t, tc.ResultEventID, eventID)
			assert.ErrorIs(t, err, tc.ResultError)
		})
	}

}
