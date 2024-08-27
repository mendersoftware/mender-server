// Copyright 2022 Northern.tech AS
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

package devauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	common "github.com/mendersoftware/mender-server/services/iot-manager/client"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

const (
	URIInternal        = "/api/internal/v1/devauth"
	URIInternalDevices = URIInternal + "/tenants/:tenant/devices"
	URIInternalAlive   = URIInternal + "/alive"
)

var (
	ErrInvalidURL = errors.New("invalid URL format")
)

const DefaultTimeout = time.Second * 10

// Client interface exposing a portion of the internal deviceauth API.
//
//go:generate ../../utils/mockgen.sh
type Client interface {
	Ping(ctx context.Context) error
	GetDevices(context.Context, []string) ([]Device, error)
}

// Config provides initialization options for creating a new client.
type Config struct {
	// Client provides an option to override the http.Client used for
	// performing the requests.
	Client *http.Client
	// DevauthAddress is the base URI to the deviceauth service (requires
	// a minimal of an addressable hostname, to a full blown URL).
	DevauthAddress string
}

type urlValidator struct{}

func (urlValidator) Validate(v interface{}) error {
	if uri, ok := v.(string); ok {
		_, err := url.Parse(uri)
		if err != nil {
			return ErrInvalidURL
		}
		return nil
	}
	return ErrInvalidURL
}

func (conf Config) Validate() error {
	return validation.ValidateStruct(&conf,
		validation.Field(
			&conf.DevauthAddress, validation.Required,
		),
		validation.Field(
			&conf.DevauthAddress, validation.Required, urlValidator{},
		),
	)
}

type client struct {
	*http.Client
	uri string
}

// NewClient initializes a new client from the given configuration options.
func NewClient(config Config) (Client, error) {
	client := &client{
		Client: config.Client,
	}
	if client.Client == nil {
		client.Client = new(http.Client)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	client.uri = strings.TrimRight(config.DevauthAddress, "/")
	if !strings.Contains(client.uri, "://") {
		client.uri = "http://" + client.uri
	}

	return client, nil
}

// GET /api/internal/v1/devauth/alive
func (c *client) Ping(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	//nolint:errcheck
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.uri+URIInternalAlive, nil)
	rsp, err := c.Do(req)
	if err != nil {
		return errors.Wrap(err, "error checking deviceauth liveliness")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode < 300 {
		return nil
	}
	return errors.Errorf(
		"received bad status code from deviceauth liveliness probe: %s",
		rsp.Status,
	)
}

// GetDevice provides a functional handle to the API endpoint:
// GET /api/internal/v1/devauth/tenants/{tenantID}/devices
func (c *client) GetDevices(
	ctx context.Context,
	deviceIDs []string,
) ([]Device, error) {
	var tenantID string
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}

	// Prepare request
	repl := strings.NewReplacer(":tenant", tenantID)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.uri+repl.Replace(URIInternalDevices),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "devauth: error preparing request")
	}
	q := url.Values{
		"id":       deviceIDs,
		"per_page": []string{strconv.Itoa(len(deviceIDs))},
	}
	req.URL.RawQuery = q.Encode()

	if reqID := requestid.FromContext(ctx); reqID != "" {
		req.Header.Set(requestid.RequestIdHeader, reqID)
	}
	// Execute request
	rsp, err := c.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "devauth: error performing request")
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusOK:
		var devices []Device
		decoder := json.NewDecoder(rsp.Body)
		err := decoder.Decode(&devices)
		if err != nil {
			return nil, errors.Wrap(err,
				"devauth: error decoding HTTP response body",
			)
		}
		return devices, nil

	default:
		var err error
		apiErr := new(rest.Error)
		decoder := json.NewDecoder(rsp.Body)
		if decoder.Decode(apiErr) == nil {
			err = common.WrapHTTPError(apiErr, rsp.StatusCode)
		} else {
			err = common.NewHTTPError(rsp.StatusCode)
		}
		return nil, err
	}
}
