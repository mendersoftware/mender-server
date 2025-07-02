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

package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

const (
	ReindexURI = "/api/v1/workflow/reindex_reporting/batch"
	HealthURI  = "/api/v1/health"
)

const (
	defaultTimeout = time.Duration(5) * time.Second
)

// Client is the workflows client
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	StartReindex(c context.Context, deviceIDs []model.DeviceID) error
}

type ClientOptions struct {
	Client *http.Client
}

// NewClient returns a new workflows client
func NewClient(url string, opts ...ClientOptions) Client {
	// Initialize default options
	var clientOpts = ClientOptions{
		Client: &http.Client{},
	}
	// Merge options
	for _, opt := range opts {
		if opt.Client != nil {
			clientOpts.Client = opt.Client
		}
	}

	return &client{
		url:    strings.TrimSuffix(url, "/"),
		client: *clientOpts.Client,
	}
}

type client struct {
	url    string
	client http.Client
}

func (c *client) StartReindex(ctx context.Context, deviceIDs []model.DeviceID) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	tenantID := ""
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	wflow := make([]ReindexWorkflow, len(deviceIDs))
	for i, deviceID := range deviceIDs {
		wflow[i] = ReindexWorkflow{
			RequestID: requestid.FromContext(ctx),
			TenantID:  tenantID,
			DeviceID:  string(deviceID),
			Service:   ServiceInventory,
		}
	}
	payload, _ := json.Marshal(wflow)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		c.url+ReindexURI,
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "workflows: failed to submit reindex job")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 300 {
		return nil
	} else if rsp.StatusCode == http.StatusNotFound {
		return errors.New(`workflows: workflow "reindex_reporting" not defined`)
	}

	return errors.Errorf(
		"workflows: unexpected HTTP status from workflows service: %s",
		rsp.Status,
	)
}

func (c *client) CheckHealth(ctx context.Context) error {
	var (
		apiErr rest.Error
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
		ctx, "GET",
		utils.JoinURL(c.url, HealthURI), nil,
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
