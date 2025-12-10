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
	"strconv"
	"time"

	"github.com/pkg/errors"

	inventory "github.com/mendersoftware/mender-server/pkg/api/internalapi/inventory"
	"github.com/mendersoftware/mender-server/pkg/config"

	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

const (
	defaultTimeout = 5 * time.Second
)

// Errors
var (
	ErrFilterNotFound = errors.New("Filter with given ID not found in the inventory.")
	ErrDevNotFound    = errors.New("Device with given ID not found in the inventory.")
)

// Client is the inventory client
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	Search(
		ctx context.Context,
		tenantId string,
		searchParams model.SearchParams,
	) ([]model.InvDevice, int, error)
	GetDeviceGroups(ctx context.Context, tenantId, deviceId string) ([]string, error)
}

// clientWrapper wraps the shared InventoryClient to implement the service-specific interface.
type clientWrapper struct {
	*inventory.InventoryClient
}

// NewClient returns a new inventory client using the shared generated client.
func NewClient() Client {
	var timeout time.Duration
	baseURL := config.Config.GetString(dconfig.SettingInventoryAddr)
	timeoutStr := config.Config.GetString(dconfig.SettingInventoryTimeout)

	t, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = defaultTimeout
	} else {
		timeout = time.Duration(t) * time.Second
	}

	client, err := inventory.NewInventoryClient(baseURL, inventory.WithClientTimeout(timeout))
	if err != nil {
		panic(errors.Wrap(err, "failed to create inventory client"))
	}
	return &clientWrapper{InventoryClient: client}
}

// Search searches for devices in inventory using the shared client.
func (c *clientWrapper) Search(
	ctx context.Context,
	tenantID string,
	searchParams model.SearchParams,
) ([]model.InvDevice, int, error) {
	// Convert model.FilterPredicate to inventory.FilterPredicate
	filters := make([]inventory.FilterPredicate, len(searchParams.Filters))
	for i, f := range searchParams.Filters {
		filters[i] = inventory.FilterPredicate{
			Scope:     inventory.Scope(f.Scope),
			Attribute: f.Attribute,
			Type:      inventory.FilterPredicateType(f.Type),
			Value:     interfaceToString(f.Value),
		}
	}

	params := inventory.SearchParams{
		Page:      searchParams.Page,
		PerPage:   searchParams.PerPage,
		Filters:   filters,
		DeviceIDs: searchParams.DeviceIDs,
	}

	devices, totalCount, err := c.InventoryClient.Search(ctx, tenantID, params)
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
				Value: attr.Value,
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

// interfaceToString converts an interface{} value to a string representation.
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
