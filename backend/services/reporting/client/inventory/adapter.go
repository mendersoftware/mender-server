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

	"github.com/pkg/errors"

	inventory "github.com/mendersoftware/mender-server/pkg/api/internalapi/inventory"
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

// GetDevices uses the search endpoint to get devices by IDs.
func (a *clientAdapter) GetDevices(
	ctx context.Context,
	tid string,
	deviceIDs []string,
) ([]Device, error) {
	devices, err := a.client.GetDevices(ctx, tid, deviceIDs)
	if err != nil {
		return nil, err
	}

	// Convert inventory.InvDevice to local Device type
	result := make([]Device, len(devices))
	for i, d := range devices {
		attrs := make(DeviceAttributes, len(d.Attributes))
		for j, attr := range d.Attributes {
			attrs[j] = DeviceAttribute{
				Name:  attr.Name,
				Scope: string(attr.Scope),
				Value: attr.Value, // Keep as string - the Value is interface{}
			}
			if attr.Description != nil {
				attrs[j].Description = attr.Description
			}
		}
		result[i] = Device{
			ID:         DeviceID(d.ID),
			Attributes: attrs,
		}
	}

	return result, nil
}
