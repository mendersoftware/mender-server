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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	apiclient "github.com/mendersoftware/mender-server/pkg/api/client"
	rest "github.com/mendersoftware/mender-server/pkg/rest.utils"
)

const (
	DefaultTimeout = 10 * time.Second
)

// ErrPreconditionsFailed is returned when the If-Unmodified-Since precondition fails.
var ErrPreconditionsFailed = errors.New("preconditions failed")

// InventoryClient wraps the generated client with higher-level methods
// that match the existing service client interfaces.
type InventoryClient struct {
	client  *ClientWithResponses
	timeout time.Duration
}

// InventoryClientOption is a function that configures the InventoryClient.
type InventoryClientOption func(*InventoryClient)

// WithClientTimeout sets the default timeout for requests.
func WithClientTimeout(timeout time.Duration) InventoryClientOption {
	return func(c *InventoryClient) {
		c.timeout = timeout
	}
}

// NewInventoryClient creates a new inventory client wrapper.
func NewInventoryClient(serverURL string, opts ...InventoryClientOption) (*InventoryClient, error) {
	httpClient := apiclient.NewHTTPClient()

	genClient, err := NewClientWithResponses(serverURL, WithHTTPClient(httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create inventory client")
	}

	ic := &InventoryClient{
		client:  genClient,
		timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(ic)
	}

	return ic, nil
}

// withTimeout returns a context with the client's default timeout if no deadline is set.
func (c *InventoryClient) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		return context.WithTimeout(ctx, c.timeout)
	}
	return ctx, func() {}
}

// CheckHealth checks the inventory service health.
func (c *InventoryClient) CheckHealth(ctx context.Context) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.client.InventoryInternalCheckHealthWithResponse(ctx)
	if err != nil {
		return err
	}

	if resp.StatusCode() >= http.StatusOK && resp.StatusCode() < 300 {
		return nil
	}

	// Parse error response
	if resp.JSON500 != nil {
		return &rest.Error{Err: resp.JSON500.Error, RequestID: deref(resp.JSON500.RequestID)}
	}

	return errors.Errorf("health check HTTP error: %s", resp.Status())
}

// DeviceInventoryUpdate represents a device update for status changes.
type DeviceInventoryUpdate struct {
	ID       string `json:"id"`
	Revision uint   `json:"revision"`
}

// SetDeviceStatus updates the status of multiple devices in inventory.
func (c *InventoryClient) SetDeviceStatus(
	ctx context.Context,
	tenantID string,
	deviceUpdates []DeviceInventoryUpdate,
	status string,
) error {
	if len(deviceUpdates) < 1 {
		return errors.New("no devices to update")
	}

	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	// Convert to the generated type
	updates := make([]DeviceUpdate, len(deviceUpdates))
	for i, du := range deviceUpdates {
		updates[i] = DeviceUpdate{
			ID:       du.ID,
			Revision: int(du.Revision),
		}
	}

	body, err := json.Marshal(updates)
	if err != nil {
		return errors.Wrap(err, "failed to serialize device updates")
	}

	resp, err := c.client.UpdateStatusOfDevicesWithBodyWithResponse(
		ctx,
		tenantID,
		status,
		"application/json",
		bytes.NewReader(body),
		addMenderSourceHeader("deviceauth"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to update device status")
	}

	if resp.StatusCode() != http.StatusOK {
		return parseErrorResponse(resp.HTTPResponse, resp.Body)
	}

	return nil
}

// DeviceAttribute represents an inventory attribute.
type DeviceAttribute struct {
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Value       interface{} `json:"value"`
	Scope       string      `json:"scope"`
}

// SetDeviceIdentity updates the inventory attributes under the identity scope.
func (c *InventoryClient) SetDeviceIdentity(
	ctx context.Context,
	tenantID string,
	deviceID string,
	idData map[string]interface{},
) error {
	return c.setDeviceIdentityIfUnmodifiedSince(ctx, tenantID, deviceID, idData, nil)
}

// SetDeviceIdentityIfUnmodifiedSince updates the inventory attributes under
// the identity scope if the inventory did not change since unmodifiedSince.
func (c *InventoryClient) SetDeviceIdentityIfUnmodifiedSince(
	ctx context.Context,
	tenantID string,
	deviceID string,
	idData map[string]interface{},
	unmodifiedSince time.Time,
) error {
	return c.setDeviceIdentityIfUnmodifiedSince(ctx, tenantID, deviceID, idData, &unmodifiedSince)
}

func (c *InventoryClient) setDeviceIdentityIfUnmodifiedSince(
	ctx context.Context,
	tenantID string,
	deviceID string,
	idData map[string]interface{},
	unmodifiedSince *time.Time,
) error {
	if deviceID == "" {
		return errors.New("device id is needed")
	}

	attributes := make([]DeviceAttribute, 0, len(idData))
	for name, value := range idData {
		attributes = append(attributes, DeviceAttribute{
			Name:  name,
			Value: value,
			Scope: "identity",
		})
	}

	if len(attributes) < 1 {
		return errors.New("no attributes to update")
	}

	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	body, err := json.Marshal(attributes)
	if err != nil {
		return errors.Wrap(err, "failed to serialize attributes")
	}

	var params *UpdateInventoryForADeviceParams
	if unmodifiedSince != nil {
		ifUnmodifiedSince := unmodifiedSince.In(time.FixedZone("GMT", 0)).Format(time.RFC1123)
		params = &UpdateInventoryForADeviceParams{
			IfUnmodifiedSince: &ifUnmodifiedSince,
		}
	}

	resp, err := c.client.UpdateInventoryForADeviceWithBodyWithResponse(
		ctx,
		tenantID,
		deviceID,
		"identity",
		params,
		"application/json",
		bytes.NewReader(body),
		addMenderSourceHeader("deviceauth"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to update device identity")
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusPreconditionFailed:
		return ErrPreconditionsFailed
	default:
		return parseErrorResponse(resp.HTTPResponse, resp.Body)
	}
}

// SearchParams contains parameters for the device search.
type SearchParams struct {
	Page      int               `json:"page,omitempty"`
	PerPage   int               `json:"per_page,omitempty"`
	Filters   []FilterPredicate `json:"filters,omitempty"`
	Sort      []SortCriteria    `json:"sort,omitempty"`
	DeviceIDs []string          `json:"id,omitempty"`
	Text      string            `json:"text,omitempty"`
}

// InvDevice represents a device returned from search.
type InvDevice struct {
	ID         string        `json:"id"`
	Attributes []AttributeV2 `json:"attributes,omitempty"`
	UpdatedTS  string        `json:"updated_ts,omitempty"`
}

// Search searches for devices in the inventory.
func (c *InventoryClient) Search(
	ctx context.Context,
	tenantID string,
	params SearchParams,
) ([]InvDevice, int, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	reqBody := InventoryInternalV2SearchDeviceInventoriesJSONRequestBody{
		Page:    &params.Page,
		PerPage: &params.PerPage,
		Filters: &params.Filters,
		Sort:    &params.Sort,
	}

	if len(params.DeviceIDs) > 0 {
		reqBody.DeviceIds = &params.DeviceIDs
	}
	if params.Text != "" {
		reqBody.Text = &params.Text
	}

	resp, err := c.client.InventoryInternalV2SearchDeviceInventoriesWithResponse(
		ctx,
		tenantID,
		reqBody,
	)
	if err != nil {
		return nil, -1, errors.Wrap(err, "search devices request failed")
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, -1, parseErrorResponse(resp.HTTPResponse, resp.Body)
	}

	// Parse response body
	var devices []InvDevice
	if err := json.Unmarshal(resp.Body, &devices); err != nil {
		return nil, -1, errors.Wrap(err, "failed to parse search response")
	}

	// Get total count from header
	totalCount := 0
	if resp.HTTPResponse != nil {
		if countStr := resp.HTTPResponse.Header.Get("X-Total-Count"); countStr != "" {
			if cnt, err := strconv.Atoi(countStr); err == nil {
				totalCount = cnt
			}
		}
	}

	return devices, totalCount, nil
}

// GetDeviceGroups returns the groups a device belongs to.
func (c *InventoryClient) GetDeviceGroups(
	ctx context.Context,
	tenantID string,
	deviceID string,
) ([]string, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.client.GetDeviceGroupsWithResponse(ctx, tenantID, deviceID)
	if err != nil {
		return nil, errors.Wrap(err, "get device groups request failed")
	}

	if resp.StatusCode() == http.StatusNotFound {
		return []string{}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, parseErrorResponse(resp.HTTPResponse, resp.Body)
	}

	// Parse response body manually since the generated code doesn't have JSON200
	var groups Groups
	if err := json.Unmarshal(resp.Body, &groups); err != nil {
		return nil, errors.Wrap(err, "failed to parse device groups response")
	}

	return groups.Groups, nil
}

// GetDevices retrieves devices by their IDs using the search endpoint.
func (c *InventoryClient) GetDevices(
	ctx context.Context,
	tenantID string,
	deviceIDs []string,
) ([]InvDevice, error) {
	devices, _, err := c.Search(ctx, tenantID, SearchParams{
		DeviceIDs: deviceIDs,
		Page:      1,
		PerPage:   len(deviceIDs),
	})
	return devices, err
}

// Helper functions

func addMenderSourceHeader(source string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("X-MEN-Source", source)
		return nil
	}
}

func parseErrorResponse(resp *http.Response, body []byte) error {
	var apiErr Error
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != "" {
		return &rest.Error{Err: apiErr.Error, RequestID: deref(apiErr.RequestID)}
	}
	return errors.Errorf("request failed with status %s", resp.Status)
}

func deref[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}
