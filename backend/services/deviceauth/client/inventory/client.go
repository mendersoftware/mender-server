// Copyright 2023 Northern.tech AS
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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

const (
	urlHealth             = "/api/internal/v1/inventory/health"
	urlUpdateDeviceStatus = "/api/internal/v1/inventory/tenants/#tid/devices/status/"
	urlSetDeviceAttribute = "/api/internal/v1/inventory/tenants/#tid/device/" +
		"#did/attribute/scope/#scope"
	defaultTimeout = 10 * time.Second
)

var ErrPreconditionsFailed = errors.New("preconditions failed")

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

type client struct {
	client  *http.Client
	urlBase string
}

// NewClient creates a new inventory client.
// This now uses the shared generated client under the hood.
// The skipVerify parameter is deprecated and ignored - TLS verification
// is handled by the shared client infrastructure.
func NewClient(urlBase string, skipVerify bool) Client {
	adapter, err := NewClientAdapter(urlBase)
	if err != nil {
		// Fall back to legacy implementation if adapter creation fails
		// This should not happen in normal circumstances
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
		}
		return &client{
			client: &http.Client{
				Transport: tr,
			},
			urlBase: urlBase,
		}
	}
	return adapter
}

func (c *client) CheckHealth(ctx context.Context) error {
	var apiErr rest.Error

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(
		ctx, "GET",
		utils.JoinURL(c.urlBase, urlHealth), nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create healthcheck request: %w", err)
	}

	rsp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= http.StatusOK && rsp.StatusCode < 300 {
		return nil
	}
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(&apiErr)
	if err != nil {
		return errors.Errorf("health check HTTP error: %s", rsp.Status)
	}
	return &apiErr
}

func (c *client) SetDeviceStatus(
	ctx context.Context,
	tenantId string,
	deviceUpdates []model.DeviceInventoryUpdate,
	status string,
) error {
	l := log.FromContext(ctx)

	if len(deviceUpdates) < 1 {
		return errors.New("no devices to update")
	}
	body, err := json.Marshal(deviceUpdates)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize devices")
	}

	rd := bytes.NewReader(body)

	url := utils.JoinURL(c.urlBase, urlUpdateDeviceStatus+status)
	url = strings.Replace(url, "#tid", tenantId, 1)

	req, err := http.NewRequest(http.MethodPost, url, rd)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("X-MEN-Source", "deviceauth")
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	rsp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrapf(err, "failed to submit %s %s", req.Method, req.URL)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			body = []byte("<failed to read>")
		}
		l.Errorf("request %s %s failed with status %v, response: %s",
			req.Method, req.URL, rsp.Status, body)

		return errors.Errorf(
			"%s %s request failed with status %v", req.Method, req.URL, rsp.Status)
	}

	return nil
}

func (c *client) setDeviceIdentityIfUnmodifiedSince(
	ctx context.Context,
	tenantID,
	deviceID string,
	idData map[string]interface{},
	unmodifiedSince *time.Time,
) error {
	l := log.FromContext(ctx)

	if deviceID == "" {
		return errors.New("device id is needed")
	}

	attributes := make([]model.DeviceAttribute, len(idData))
	i := 0
	for name, value := range idData {
		attribute := model.DeviceAttribute{
			Name:        name,
			Description: nil,
			Value:       value,
			Scope:       "identity",
		}
		attributes[i] = attribute
		i++
	}

	if i < 1 {
		return errors.New("no attributes to update")
	}

	if i != len(idData) {
		attributes = attributes[:i]
	}

	body, err := json.Marshal(attributes)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize device attribute")
	}

	rd := bytes.NewReader(body)

	url := utils.JoinURL(c.urlBase, urlSetDeviceAttribute)
	url = strings.Replace(url, "#tid", tenantID, 1)
	url = strings.Replace(url, "#did", deviceID, 1)
	url = strings.Replace(url, "#scope", "identity", 1)

	req, err := http.NewRequest(http.MethodPatch, url, rd)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("X-MEN-Source", "deviceauth")
	req.Header.Set("Content-Type", "application/json")
	if unmodifiedSince != nil {
		// UTC is not part of RFC1123 / need to rename to GMT
		req.Header.Set("If-Unmodified-Since", unmodifiedSince.
			In(time.FixedZone("GMT", 0)).
			Format(time.RFC1123))
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	rsp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrapf(err, "failed to submit %s %s", req.Method, req.URL)
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusPreconditionFailed:
		return ErrPreconditionsFailed
	default:
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			body = []byte("<failed to read>")
		}
		l.Errorf("request %s %s failed with status %v, response: %s",
			req.Method, req.URL, rsp.Status, body)

		return errors.Errorf(
			"%s %s request failed with status %v", req.Method, req.URL, rsp.Status)
	}
}

// SetDeviceIdentity updates the inventory attributes under the identity scope.
// WARN: Make sure the attribute with name 'status' is handled properly as it
// should always reflect the current device status.
func (c *client) SetDeviceIdentity(
	ctx context.Context,
	tenantID,
	deviceID string,
	idData map[string]interface{},
) error {
	return c.setDeviceIdentityIfUnmodifiedSince(ctx, tenantID, deviceID, idData, nil)
}

// SetDeviceIdentityIfUnmodifiedSince updates the inventory attributes under
// the identity scope if the inventory did not change since unmodifiedSince.
// WARN: Make sure the attribute with name 'status' is handled properly as it
// should always reflect the current device status.
func (c *client) SetDeviceIdentityIfUnmodifiedSince(
	ctx context.Context,
	tenantID,
	deviceID string,
	idData map[string]interface{},
	unmodifiedSince time.Time,
) error {
	return c.setDeviceIdentityIfUnmodifiedSince(ctx, tenantID, deviceID, idData, &unmodifiedSince)
}
