// Copyright 2024 Northern.tech AS
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
package tenant

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"

	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

const (
	TenantHealthURI = "/api/internal/v1/tenantadm/health"
	// devices endpoint
	TenantVerifyUri = "/api/internal/v1/tenantadm/tenants/verify"
	TenantGetUri    = "/api/internal/v1/tenantadm/tenants/#tid"
	TenantUsersURI  = "/api/internal/v1/tenantadm/tenants/users"
	// default request timeout, 10s?
	defaultReqTimeout = time.Duration(10) * time.Second
)

const (
	MsgErrTokenVerificationFailed = "tenant token verification failed"
)

func IsErrTokenVerificationFailed(e error) bool {
	return strings.HasPrefix(e.Error(), MsgErrTokenVerificationFailed)
}

func MakeErrTokenVerificationFailed(apiErr error) error {
	return errors.Wrap(apiErr, MsgErrTokenVerificationFailed)
}

// ClientConfig conveys client configuration
type Config struct {
	// Tenant administrator service address
	TenantAdmAddr string
	// Request timeout
	Timeout time.Duration
}

// ClientRunner is an interface of inventory client
//
//go:generate ../../../../utils/mockgen.sh
type ClientRunner interface {
	CheckHealth(ctx context.Context) error
	VerifyToken(ctx context.Context, token string) (*Tenant, error)
	GetTenant(ctx context.Context, tid string) (*Tenant, error)
	GetTenantUsers(ctx context.Context, tenantID string) ([]User, error)
}

// Client is an opaque implementation of tenant administrator client. Implements
// ClientRunner interface
type Client struct {
	conf Config
	http http.Client
}

// NewClient creates a client with given config.
func NewClient(c Config) *Client {
	if c.Timeout == 0 {
		c.Timeout = defaultReqTimeout
	}

	return &Client{
		conf: c,
		http: http.Client{
			Timeout: c.Timeout,
		},
	}
}

func (c *Client) CheckHealth(ctx context.Context) error {
	var (
		apiErr rest_utils.ApiError
	)

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.conf.Timeout)
		defer cancel()
	}
	req, _ := http.NewRequestWithContext(
		ctx, "GET",
		utils.JoinURL(c.conf.TenantAdmAddr, TenantHealthURI), nil,
	)

	rsp, err := c.http.Do(req)
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

// VerifyToken will execute a request to tenenatadm's endpoint for token
// verification. Returns nil if verification was successful.
func (tc *Client) VerifyToken(ctx context.Context, token string) (*Tenant, error) {

	l := log.FromContext(ctx)

	// TODO sanitize token

	url := utils.JoinURL(tc.conf.TenantAdmAddr, TenantVerifyUri)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request to tenant administrator")
	}

	// tenant token is passed in Authorization header
	req.Header.Add("Authorization", "Bearer "+token)

	ctx, cancel := context.WithTimeout(ctx, tc.conf.Timeout)
	defer cancel()

	rsp, err := tc.http.Do(req.WithContext(ctx))
	if err != nil {
		l.Errorf("tenantadm request failed: %v", err)
		return nil, errors.Wrap(err, "request to verify token failed")
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {

	case http.StatusUnauthorized: // 401, verification result negative
		apiErr := rest_utils.ParseApiError(rsp.Body)
		if !rest_utils.IsApiError(apiErr) {
			return nil, errors.Errorf("failed to parse tenantadm api error response")
		}

		return nil, MakeErrTokenVerificationFailed(apiErr)

	case http.StatusOK: // 200, token verified
		tenant := Tenant{}
		if err := json.NewDecoder(rsp.Body).Decode(&tenant); err != nil {
			return nil, errors.Wrap(err, "error parsing tenant verification response")
		}
		return &tenant, nil
	default:
		return nil, errors.Errorf("token verification request returned unexpected status %v",
			rsp.StatusCode)
	}
}

// GetTenant will retrieve a single tenant
// verification. Returns nil if verification was successful.
func (tc *Client) GetTenant(ctx context.Context, tid string) (*Tenant, error) {

	l := log.FromContext(ctx)

	repl := strings.NewReplacer("#tid", tid)
	uri := repl.Replace(TenantGetUri)

	req, err := http.NewRequest(http.MethodGet,
		utils.JoinURL(tc.conf.TenantAdmAddr, uri),
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request to tenantadm")
	}

	ctx, cancel := context.WithTimeout(ctx, tc.conf.Timeout)
	defer cancel()

	rsp, err := tc.http.Do(req.WithContext(ctx))
	if err != nil {
		l.Errorf("tenantadm request failed: %v", err)
		return nil, errors.Wrap(err, "request to get tenant failed")
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNotFound:
		return nil, nil
	case http.StatusOK:
		tenant := Tenant{}
		if err := json.NewDecoder(rsp.Body).Decode(&tenant); err != nil {
			return nil, errors.Wrap(err, "error parsing tenant")
		}
		return &tenant, nil
	default:
		return nil, errors.Errorf("getting tenant resulted in unexpected code: %v",
			rsp.StatusCode)
	}
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"name"`
	TenantID string `json:"tenant_id"`
}

func (tc *Client) GetTenantUsers(ctx context.Context, tenantID string) ([]User, error) {
	var ret []User
	if tenantID == "" {
		return nil, errors.New("tenantadm: [internal] bad argument " +
			"tenantID: cannot be empty")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		utils.JoinURL(tc.conf.TenantAdmAddr, TenantUsersURI),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "tenantadm: failed to prepare request")
	}
	q := req.URL.Query()
	q.Add("tenant_id", tenantID)
	req.URL.RawQuery = q.Encode()

	rsp, err := tc.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "tenantadm: error sending user request")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= 400 {
		var APIErr = new(rest_utils.ApiError)
		jsDecoder := json.NewDecoder(rsp.Body)
		err = jsDecoder.Decode(APIErr)
		if err != nil {
			return nil, errors.Errorf(
				"tenantadm: unexpected HTTP status: %s",
				rsp.Status,
			)
		}
		return nil, errors.Wrapf(APIErr,
			"tenantadm: HTTP error (%s) on user request",
			rsp.Status,
		)
	}
	jsDecoder := json.NewDecoder(rsp.Body)
	err = jsDecoder.Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err,
			"tenantadm: error decoding response payload",
		)
	}
	return ret, nil
}
