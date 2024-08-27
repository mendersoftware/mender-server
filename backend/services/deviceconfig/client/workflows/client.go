// Copyright 2021 Northern.tech AS
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

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

const (
	HealthCheckURI              = "/api/v1/health"
	AuditlogsURI                = "/api/v1/workflow/emit_auditlog"
	DeployDeviceConfigurationRI = "/api/v1/workflow/deploy_device_configuration"
)

const (
	defaultTimeout = time.Duration(5) * time.Second
)

// Client is the workflows client
//
//go:generate ../../x/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	SubmitAuditLog(ctx context.Context, log AuditLog) error
	DeployConfiguration(ctx context.Context, tenantID string, deviceID string,
		deploymentID uuid.UUID, configuration []byte,
		retries uint, updateControlMap map[string]interface{}) error
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
		ctx, "GET", c.url+HealthCheckURI, nil,
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

func (c *client) SubmitAuditLog(ctx context.Context, log AuditLog) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	if log.EventTS.IsZero() {
		log.EventTS = time.Now()
	}
	if err := log.Validate(); err != nil {
		return errors.Wrap(err, "workflows: invalid AuditLog entry")
	}
	id := identity.FromContext(ctx)
	if id == nil || id.Tenant == "" {
		return errors.New("workflows: Context lacking tenant identity")
	}
	wflow := AuditWorkflow{
		RequestID: requestid.FromContext(ctx),
		TenantID:  id.Tenant,
		AuditLog:  log,
	}
	payload, _ := json.Marshal(wflow)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		c.url+AuditlogsURI,
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Add("Content-Type", "application/json")
	rsp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "workflows: failed to submit auditlog")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 300 {
		return nil
	}

	if rsp.StatusCode == http.StatusNotFound {
		return errors.New(`workflows: workflow "auditlogs" not defined`)
	}

	return errors.Errorf(
		"workflows: unexpected HTTP status from workflows service: %s",
		rsp.Status,
	)
}

func (c *client) DeployConfiguration(ctx context.Context, tenantID string, deviceID string,
	deploymentID uuid.UUID, configuration []byte, retries uint,
	updateControlMap map[string]interface{}) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	wflow := DeployConfigurationWorkflow{
		RequestID:        requestid.FromContext(ctx),
		TenantID:         tenantID,
		DeviceID:         deviceID,
		DeploymentID:     deploymentID,
		Configuration:    string(configuration),
		Retries:          retries,
		UpdateControlMap: updateControlMap,
	}

	payload, _ := json.Marshal(wflow)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		c.url+DeployDeviceConfigurationRI,
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Add("Content-Type", "application/json")
	rsp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "workflows: failed to deploy configuration")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 300 {
		return nil
	}

	if rsp.StatusCode == http.StatusNotFound {
		return errors.New(`workflows: workflow "deploy_device_configuration" not defined`)
	}

	return errors.Errorf(
		"workflows: unexpected HTTP status from workflows service: %s",
		rsp.Status,
	)
}
