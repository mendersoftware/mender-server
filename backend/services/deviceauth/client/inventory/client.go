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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"

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
}

type client struct {
	client  *http.Client
	urlBase string
}

func NewClient(urlBase string, skipVerify bool) *client {
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

func (c *client) CheckHealth(ctx context.Context) error {
	var apiErr rest_utils.ApiError

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	req, _ := http.NewRequestWithContext(
		ctx, "GET",
		utils.JoinURL(c.urlBase, urlHealth), nil,
	)

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

func (c *client) SetDeviceIdentity(
	ctx context.Context,
	tenantId,
	deviceId string,
	idData map[string]interface{},
) error {
	l := log.FromContext(ctx)

	if deviceId == "" {
		return errors.New("device id is needed")
	}

	attributes := make([]model.DeviceAttribute, len(idData))
	i := 0
	for name, value := range idData {
		if name == "status" {
			//we have to forbid the client to override attribute status in identity scope
			//since it stands for status of a device (as in: accepted, rejected, preauthorized)
			continue
		}
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
	url = strings.Replace(url, "#tid", tenantId, 1)
	url = strings.Replace(url, "#did", deviceId, 1)
	url = strings.Replace(url, "#scope", "identity", 1)

	req, err := http.NewRequest(http.MethodPatch, url, rd)
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
