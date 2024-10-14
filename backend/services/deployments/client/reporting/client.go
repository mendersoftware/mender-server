// Copyright 2022 Northern.tech AS
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

package reporting

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/rest_utils"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

const (
	uriInternal       = "/api/internal/v1/reporting"
	uriInternalSearch = uriInternal + "/tenants/:tenant_id/devices/search"
	uriIInternalAlive = uriInternal + "/alive"

	defaultTimeout = 5 * time.Second

	hdrTotalCount = "X-Total-Count"
)

// Client is the reporting client
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	Search(
		ctx context.Context,
		tenantId string,
		searchParams model.SearchParams,
	) ([]model.InvDevice, int, error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a new reporting client
func NewClient(baseURL string) Client {
	return &client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
}

func (c *client) CheckHealth(ctx context.Context) error {
	var (
		apiErr rest_utils.ApiError
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
		ctx, "GET", c.baseURL+uriIInternalAlive, nil,
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
	repl := strings.NewReplacer(":tenant_id", tenantId)
	url := c.baseURL + repl.Replace(uriInternalSearch)

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
			"search devices request failed with unexpected status: %v",
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
