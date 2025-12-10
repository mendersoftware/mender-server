// Copyright 2025 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

package inventory

import (
	"context"
	"time"

	"github.com/pkg/errors"

	inventory "github.com/mendersoftware/mender-server/pkg/api/internalapi/inventory"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

// Re-export shared errors for backwards compatibility
var ErrPreconditionsFailed = inventory.ErrPreconditionsFailed

// Client is the inventory client interface.
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	SetDeviceStatus(
		ctx context.Context,
		tenantId string,
		deviceUpdates []model.DeviceInventoryUpdate,
		status string,
	) error
	SetDeviceIdentity(
		ctx context.Context,
		tenantId,
		deviceId string,
		idData map[string]interface{},
	) error
	SetDeviceIdentityIfUnmodifiedSince(
		ctx context.Context,
		tenantId,
		deviceId string,
		idData map[string]interface{},
		unmodifiedSince time.Time,
	) error
}

// clientWrapper wraps the shared InventoryClient to implement the service-specific interface.
type clientWrapper struct {
	*inventory.InventoryClient
}

// NewClient creates a new inventory client using the shared generated client.
// The skipVerify parameter is deprecated and ignored - TLS verification
// is handled by the shared client infrastructure.
func NewClient(urlBase string, skipVerify bool) Client {
	client, err := inventory.NewInventoryClient(urlBase)
	if err != nil {
		// This should not happen in normal circumstances
		panic(errors.Wrap(err, "failed to create inventory client"))
	}
	return &clientWrapper{InventoryClient: client}
}

// SetDeviceStatus converts the model type to the shared type and calls the underlying client.
func (c *clientWrapper) SetDeviceStatus(
	ctx context.Context,
	tenantID string,
	deviceUpdates []model.DeviceInventoryUpdate,
	status string,
) error {
	// Convert model.DeviceInventoryUpdate to inventory.DeviceInventoryUpdate
	updates := make([]inventory.DeviceInventoryUpdate, len(deviceUpdates))
	for i, du := range deviceUpdates {
		updates[i] = inventory.DeviceInventoryUpdate{
			ID:       du.Id,
			Revision: du.Revision,
		}
	}
	return c.InventoryClient.SetDeviceStatus(ctx, tenantID, updates, status)
}
