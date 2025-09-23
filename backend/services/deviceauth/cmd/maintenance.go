package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/inventory"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store/mongo"
)

func MaintenanceSyncDeviceInventory(
	ctx context.Context,
	db *mongo.DataStoreMongo,
	inv inventory.Client,
	rateLimit *rate.Limiter,
) error {
	var totalCount int64
	startTS := time.Now().Add(-time.Second)
	logger := log.FromContext(ctx)
	for dev, err := range db.ListAllDevices(ctx,
		"tenant_id",
		model.DevKeyId,
		model.DevKeyIdDataStruct,
		model.DevKeyStatus,
	) {
		if err != nil {
			return fmt.Errorf("error listing all devices: %w", err)
		}
		if rateLimit != nil {
			err = rateLimit.Wait(ctx)
			if err != nil {
				return err
			}
		}
		dev.IdDataStruct["status"] = dev.Status
		err := inv.SetDeviceIdentityIfUnmodifiedSince(
			ctx, dev.TenantID,
			dev.Id, dev.IdDataStruct,
			startTS,
		)
		if errors.Is(err, inventory.ErrPreconditionsFailed) {
			err = nil
		}
		if err != nil {
			return fmt.Errorf("error updating inventory data: %w", err)
		}
		totalCount++
		if totalCount%1000 == 0 {
			logger.Infof("processed %d devices", totalCount)
		}
	}
	logger.Infof("finished processing %d devices", totalCount)
	return nil
}
