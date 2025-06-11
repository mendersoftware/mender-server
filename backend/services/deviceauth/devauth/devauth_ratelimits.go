package devauth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

// ErrNoRatelimits is returned by RateLimitsFromContext when there are no limits.
var ErrNoRatelimits = errors.New("no ratelimits")

func (d *DevAuth) checkRateLimits(ctx context.Context) error {
	if d.rateLimiter != nil {
		rsp, err := d.rateLimiter.Reserve(ctx)
		if err != nil {
			if errors.Is(err, ErrNoRatelimits) {
				return nil
			}
			return err
		} else if !rsp.OK() {
			return cache.ErrTooManyRequests
		}
	}
	return nil
}

const rateLimitMax = uint64(1 << 50)

func fmtEventID(tenantID, deviceID, event string) string {
	return fmt.Sprintf("tenant:%s:dev:%s:event:%s", tenantID, deviceID, event)
}

// rateLimitFromContext returns the burst quota given the context
func (d *DevAuth) RateLimitsFromContext(ctx context.Context) (
	limit uint64,
	eventID string,
	err error,
) {
	var (
		tenantID string = "default"
		deviceID string = "default"
	)
	var weight float64 = d.rateLimiterWeightDefault
	id := identity.FromContext(ctx)
	if id != nil {
		tenantID = id.Tenant
		plan := id.Plan
		if w, ok := d.rateLimiterWeights[plan]; ok {
			weight = w
		}
		deviceID = id.Subject
	}
	origUri := ctxhttpheader.FromContext(ctx, "X-Forwarded-Uri")
	origUri = purgeUriArgs(origUri)
	if splitPath := strings.SplitN(origUri, "/", 5); len(splitPath) == 5 {
		// discard `/api/devices/v*/`
		origUri = splitPath[4]
	}
	lim, err := d.GetLimit(ctx, model.LimitMaxDeviceCount)
	if err != nil {
		return 0, "", err
	} else if lim.Value == 0 {
		return 0, "", ErrNoRatelimits
	}
	var limitf64 float64
	if lim.Value >= uint64(rateLimitMax) {
		// overflow protection: 1 << 50 is practically unlimited
		limitf64 = float64(rateLimitMax)
	} else {
		limitf64 = float64(lim.Value)
	}
	limitf64 *= weight
	if limitf64 > float64(rateLimitMax) {
		limit = rateLimitMax
	} else {
		limit = uint64(limitf64)
	}
	return limit, fmtEventID(tenantID, deviceID, origUri), nil
}
