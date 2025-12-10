// Copyright 2025 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package inventory

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	inventory "github.com/mendersoftware/mender-server/pkg/api/internalapi/inventory"
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

// clientAdapter wraps the shared inventory client and implements the
// service-specific Client interface.
type clientAdapter struct {
	client *inventory.InventoryClient
}

// NewClientAdapter creates a new inventory client using the shared generated client.
func NewClientAdapter(baseURL string, timeout time.Duration) (Client, error) {
	opts := []inventory.InventoryClientOption{}
	if timeout > 0 {
		opts = append(opts, inventory.WithClientTimeout(timeout))
	}
	client, err := inventory.NewInventoryClient(baseURL, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create inventory client")
	}
	return &clientAdapter{client: client}, nil
}

// CheckHealth checks the inventory service health.
func (a *clientAdapter) CheckHealth(ctx context.Context) error {
	return a.client.CheckHealth(ctx)
}

// Search searches for devices in inventory.
func (a *clientAdapter) Search(
	ctx context.Context,
	tenantID string,
	searchParams model.SearchParams,
) ([]model.InvDevice, int, error) {
	// Convert model.FilterPredicate to inventory.FilterPredicate
	filters := make([]inventory.FilterPredicate, len(searchParams.Filters))
	for i, f := range searchParams.Filters {
		// Convert interface{} value to string (JSON-encoded if not a string)
		valueStr := interfaceToString(f.Value)
		filters[i] = inventory.FilterPredicate{
			Scope:     inventory.Scope(f.Scope),
			Attribute: f.Attribute,
			Type:      inventory.FilterPredicateType(f.Type),
			Value:     valueStr,
		}
	}

	params := inventory.SearchParams{
		Page:      searchParams.Page,
		PerPage:   searchParams.PerPage,
		Filters:   filters,
		DeviceIDs: searchParams.DeviceIDs,
	}

	devices, totalCount, err := a.client.Search(ctx, tenantID, params)
	if err != nil {
		return nil, -1, err
	}

	// Convert inventory.InvDevice to model.InvDevice
	result := make([]model.InvDevice, len(devices))
	for i, d := range devices {
		attrs := make([]model.DeviceAttribute, len(d.Attributes))
		for j, attr := range d.Attributes {
			attrs[j] = model.DeviceAttribute{
				Name:  attr.Name,
				Scope: string(attr.Scope),
				Value: attr.Value, // Keep as string - the model Value is interface{}
			}
			if attr.Description != nil {
				attrs[j].Description = attr.Description
			}
		}
		result[i] = model.InvDevice{
			ID:         d.ID,
			Attributes: attrs,
		}
	}

	return result, totalCount, nil
}

// GetDeviceGroups returns the groups a device belongs to.
func (a *clientAdapter) GetDeviceGroups(
	ctx context.Context,
	tenantID string,
	deviceID string,
) ([]string, error) {
	return a.client.GetDeviceGroups(ctx, tenantID, deviceID)
}

// interfaceToString converts an interface{} value to a string representation.
// If the value is already a string, it's returned as-is.
// Otherwise, it's JSON-encoded.
func interfaceToString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
