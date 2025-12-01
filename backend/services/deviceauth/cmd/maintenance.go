package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/inventory"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store/mongo"
)

// fixupIDData2InventoryData makes a best-effort compatibility conversion
// from deviceauth identity data to inventory data.
// Since the data comes form JSON, all numbers are float64.
// All types that are not string, float64 or pure arrays of these types are
// stringified.
func fixupIDData2InventoryData(idData map[string]any) {
	for key, attr := range idData {
		switch t := attr.(type) {
		case []any:
			var isString bool
		SliceLoop:
			for i, elem := range t {
				switch te := elem.(type) {
				case float64:
					if isString {
						t[i] = fmt.Sprint(te)
					}
				case string:
					if i == 0 {
						isString = true
					} else if !isString {
						f, err := strconv.ParseFloat(te, 64)
						if err != nil {
							// The slice is not pure, discard it
							// and continue.
							delete(idData, key)
							break SliceLoop
						}
						t[i] = f
					}
				default:
					t[i] = fmt.Sprint(elem)
				}
			}

		case string:
		case float64:

		default:
			idData[key] = fmt.Sprint(attr)
		}
	}
}

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
		if dev.IdDataStruct == nil {
			dev.IdDataStruct = map[string]any{
				"status": dev.Status,
			}
		} else {
			dev.IdDataStruct["status"] = dev.Status
		}
		fixupIDData2InventoryData(dev.IdDataStruct)
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
