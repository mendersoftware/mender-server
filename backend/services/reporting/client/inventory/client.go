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

// Client is the inventory client interface.
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	// GetDevices uses the search endpoint to get devices just by ids (not filters)
	GetDevices(ctx context.Context, tid string, deviceIDs []string) ([]Device, error)
}

// clientWrapper wraps the shared InventoryClient to implement the service-specific interface.
type clientWrapper struct {
	*inventory.InventoryClient
}

// NewClient creates a new inventory client using the shared generated client.
func NewClient(urlBase string) Client {
	client, err := inventory.NewInventoryClient(urlBase)
	if err != nil {
		panic(errors.Wrap(err, "failed to create inventory client"))
	}
	return &clientWrapper{InventoryClient: client}
}

// GetDevices retrieves devices by their IDs using the search endpoint.
func (c *clientWrapper) GetDevices(
	ctx context.Context,
	tid string,
	deviceIDs []string,
) ([]Device, error) {
	devices, err := c.InventoryClient.GetDevices(ctx, tid, deviceIDs)
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
				Value: attr.Value,
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
