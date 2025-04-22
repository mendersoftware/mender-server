package devauth

import (
	"context"
	"fmt"
	"strings"

	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

func (d *DevAuth) checkRateLimits(ctx context.Context) error {
	if d.rateLimiter != nil {
		rsp, err := d.rateLimiter.Reserve(ctx)
		if err != nil {
			return err
		} else if !rsp.OK() {
			return cache.ErrTooManyRequests
		}
	}
	return nil
}

const rateLimitMax = int64(1 << 50)

func fmtEventID(tenantID, event string) string {
	return fmt.Sprintf("tenant:%s:ratelimit:%s", tenantID, event)
}

// rateLimitFromContext returns the burst quota given the context
func (d *DevAuth) RateLimitsFromContext(ctx context.Context) (
	limit int64,
	eventID string,
	err error,
) {
	var tenantID string = "default"
	var weight float64 = d.rateLimiterWeightDefault
	id := identity.FromContext(ctx)
	if id != nil {
		tenantID = id.Tenant
		plan := id.Plan
		if w, ok := d.rateLimiterWeights[plan]; ok {
			weight = w
		}
	}
	origUri := ctxhttpheader.FromContext(ctx, "X-Forwarded-Uri")
	origUri = purgeUriArgs(origUri)
	if splitPath := strings.SplitN(origUri, "/", 5); len(splitPath) == 5 {
		// discard `/api/devices/v*/`
		origUri = splitPath[4]
	}
	lim, err := d.GetLimit(ctx, model.LimitMaxDeviceCount)
	if err != nil {
		return -1, "", err
	} else if lim.Value == 0 {
		return -1, "", nil
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
		limit = int64(limitf64)
	}
	return limit, fmtEventID(tenantID, origUri), nil
}
