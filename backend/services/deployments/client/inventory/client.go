// Copyright 2023 Northern.tech AS
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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

const (
	healthURL          = "/api/internal/v1/inventory/health"
	searchURL          = "/api/internal/v2/inventory/tenants/:tenantId/filters/search"
	getDeviceGroupsURL = "/api/internal/v1/inventory/tenants/:tenantId/devices/:deviceId/groups"

	defaultTimeout = 5 * time.Second

	hdrTotalCount = "X-Total-Count"
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

// NewClient returns a new inventory client.
// This now uses the shared generated client under the hood.
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

	adapter, err := NewClientAdapter(baseURL, timeout)
	if err != nil {
		// Fall back to legacy implementation if adapter creation fails
		return &client{
			baseURL:    baseURL,
			httpClient: &http.Client{Timeout: timeout},
		}
	}
	return adapter
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

func (c *client) CheckHealth(ctx context.Context) error {
	var (
		apiErr rest.Error
		client http.Client
	)

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	req, _ := http.NewRequestWithContext(
		ctx, "GET", c.baseURL+healthURL, nil,
	)

	rsp, err := client.Do(req)
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

func (c *client) Search(
	ctx context.Context,
	tenantId string,
	searchParams model.SearchParams,
) ([]model.InvDevice, int, error) {
	l := log.FromContext(ctx)
	l.Debugf("Search")

	repl := strings.NewReplacer(":tenantId", tenantId)
	url := c.baseURL + repl.Replace(searchURL)

	payload, _ := json.Marshal(searchParams)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, -1, err
	}
	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, -1, errors.Wrap(err, "search devices request failed")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, -1, errors.Errorf(
			"search devices request failed with unexpected status %v",
			rsp.StatusCode,
		)
	}

	devs := []model.InvDevice{}
	if err := json.NewDecoder(rsp.Body).Decode(&devs); err != nil {
		return nil, -1, errors.Wrap(err, "error parsing search devices response")
	}

	totalCountStr := rsp.Header.Get(hdrTotalCount)
	totalCount, err := strconv.Atoi(totalCountStr)
	if err != nil {
		return nil, -1, errors.Wrap(err, "error parsing "+hdrTotalCount+" header")
	}

	return devs, totalCount, nil
}

func (c *client) GetDeviceGroups(ctx context.Context, tenantId, deviceId string) ([]string, error) {
	repl := strings.NewReplacer(":tenantId", tenantId, ":deviceId", deviceId)
	url := c.baseURL + repl.Replace(getDeviceGroupsURL)

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(
		ctx, "GET", url, nil,
	)
	if err != nil {
		return nil, err
	}

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "get device groups request failed")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		if rsp.StatusCode == http.StatusNotFound {
			return []string{}, nil
		}
		return nil, errors.Errorf(
			"get device groups request failed with unexpected status: %v",
			rsp.StatusCode,
		)
	}

	res := model.DeviceGroups{}
	if err := json.NewDecoder(rsp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "error parsing device groups response")
	}

	return res.Groups, nil
}
