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
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

const (
	uriSearch = "/api/internal/v2/inventory/tenants/:id/filters/search"
)

//go:generate ../../utils/mockgen.sh
type Client interface {
	Search(
		ctx context.Context,
		tenantId string,
		filter model.SearchParams) ([]model.InvDevice, int, error)
}

type client struct {
	client  *http.Client
	uriBase string
}

func NewClient(uriBase string, timeout int) *client {
	return &client{
		uriBase: uriBase,
		client:  &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

func (c *client) Search(
	ctx context.Context,
	tenantId string,
	searchParams model.SearchParams) ([]model.InvDevice, int, error) {

	l := log.FromContext(ctx)
	l.Debugf("Search")

	repl := strings.NewReplacer(":id", tenantId)
	url := c.uriBase + repl.Replace(uriSearch)

	payload, _ := json.Marshal(searchParams)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, -1, err
	}
	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, -1, errors.Wrap(err, "search devices request failed")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, -1, errors.Errorf(
			"search devices request failed with unexpected status %v", rsp.StatusCode)
	}

	devs := []model.InvDevice{}
	if err := json.NewDecoder(rsp.Body).Decode(&devs); err != nil {
		return nil, -1, errors.Wrap(err, "error parsing search devices response")
	}

	totalCountStr := rsp.Header.Get("X-Total-Count")
	totalCount, err := strconv.Atoi(totalCountStr)
	if err != nil {
		return nil, -1, errors.Wrap(err, "error parsing X-Total-Count header")
	}

	return devs, totalCount, nil
}
