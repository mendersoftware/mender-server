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

package devicemonitor

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
)

const (
	AlertsURI = "/api/internal/v1/devicemonitor/tenants/#tenant_id/devices/#device_id/alerts/latest"
)

const (
	defaultTimeout = time.Duration(5) * time.Second
)

// Client is the devicemonitor client
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	CheckAlerts(c context.Context, device string) (int, error)
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

func (c *client) CheckAlerts(ctx context.Context, device string) (int, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	id := identity.FromContext(ctx)
	tenant := ""
	if id != nil {
		tenant = id.Tenant
	}
	repl := strings.NewReplacer("#device_id", device,
		"#tenant_id", tenant)
	req, err := http.NewRequestWithContext(ctx,
		"GET",
		c.url+repl.Replace(AlertsURI),
		nil,
	)
	if err != nil {
		return -1, errors.Wrap(err, "devicemonitor: error preparing HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.client.Do(req)
	if err != nil {
		return -1, errors.Wrap(err, "devicemonitor: failed to get alerts for the device")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return -1, errors.Errorf(
			"devicemonitor: unexpected HTTP status from devicemonitor service: %s",
			rsp.Status,
		)
	}
	alerts := Alerts{}
	if err := json.NewDecoder(rsp.Body).Decode(&alerts); err != nil {
		return -1, errors.Wrap(err, "error parsing alerts")
	}

	return len(alerts), nil
}
