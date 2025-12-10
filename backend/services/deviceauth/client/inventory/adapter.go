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

// clientAdapter wraps the shared inventory client and implements the
// service-specific Client interface.
type clientAdapter struct {
	client *inventory.InventoryClient
}

// NewClientAdapter creates a new inventory client using the shared generated client.
func NewClientAdapter(urlBase string) (Client, error) {
	client, err := inventory.NewInventoryClient(urlBase)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create inventory client")
	}
	return &clientAdapter{client: client}, nil
}

// CheckHealth checks the inventory service health.
func (a *clientAdapter) CheckHealth(ctx context.Context) error {
	return a.client.CheckHealth(ctx)
}

// SetDeviceStatus updates the status of devices in the inventory.
func (a *clientAdapter) SetDeviceStatus(
	ctx context.Context,
	tenantID string,
	deviceUpdates []model.DeviceInventoryUpdate,
	status string,
) error {
	// Convert service-specific type to shared type
	updates := make([]inventory.DeviceInventoryUpdate, len(deviceUpdates))
	for i, du := range deviceUpdates {
		updates[i] = inventory.DeviceInventoryUpdate{
			ID:       du.Id,
			Revision: du.Revision,
		}
	}
	return a.client.SetDeviceStatus(ctx, tenantID, updates, status)
}

// SetDeviceIdentity updates the inventory attributes under the identity scope.
func (a *clientAdapter) SetDeviceIdentity(
	ctx context.Context,
	tenantID string,
	deviceID string,
	idData map[string]interface{},
) error {
	return a.client.SetDeviceIdentity(ctx, tenantID, deviceID, idData)
}

// SetDeviceIdentityIfUnmodifiedSince updates the inventory attributes under
// the identity scope if the inventory did not change since unmodifiedSince.
func (a *clientAdapter) SetDeviceIdentityIfUnmodifiedSince(
	ctx context.Context,
	tenantID string,
	deviceID string,
	idData map[string]interface{},
	unmodifiedSince time.Time,
) error {
	return a.client.SetDeviceIdentityIfUnmodifiedSince(ctx, tenantID, deviceID, idData, unmodifiedSince)
}

// Map the shared ErrPreconditionsFailed to the local one
func init() {
	// Ensure error compatibility - both packages define the same error semantically
	// The shared client returns inventory.ErrPreconditionsFailed which equals our ErrPreconditionsFailed
}
