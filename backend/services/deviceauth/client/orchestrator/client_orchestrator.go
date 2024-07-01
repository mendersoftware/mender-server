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

package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

const (
	// orchestrator endpoint
	DeviceDecommissioningOrchestratorUri = "/api/v1/workflow/decommission_device"
	ProvisionDeviceOrchestratorUri       = "/api/v1/workflow/provision_device"
	UpdateDeviceStatusOrchestratorUri    = "/api/v1/workflow/update_device_status"
	UpdateDeviceInventoryOrchestratorUri = "/api/v1/workflow/update_device_inventory"
	HealthURI                            = "/api/v1/health"
	DeviceLimitWarningURI                = "/api/v1/workflow/device_limit_email"
	ReindexReportingURI                  = "/api/v1/workflow/reindex_reporting"
	ReindexReportingBatchURI             = "/api/v1/workflow/reindex_reporting/batch"
	// default request timeout, 10s?
	defaultReqTimeout = time.Duration(10) * time.Second
)

// Config conveys client configuration
type Config struct {
	// Orchestrator host
	OrchestratorAddr string
	// Request timeout
	Timeout time.Duration
}

// ClientRunner is an interface of orchestrator client
//
//go:generate ../../utils/mockgen.sh
type ClientRunner interface {
	CheckHealth(ctx context.Context) error
	SubmitDeviceDecommisioningJob(ctx context.Context, req DecommissioningReq) error
	SubmitProvisionDeviceJob(ctx context.Context, req ProvisionDeviceReq) error
	SubmitUpdateDeviceStatusJob(ctx context.Context, req UpdateDeviceStatusReq) error
	SubmitDeviceLimitWarning(ctx context.Context, devWarn DeviceLimitWarning) error
	SubmitUpdateDeviceInventoryJob(ctx context.Context, req UpdateDeviceInventoryReq) error
	SubmitReindexReporting(c context.Context, device string) error
	SubmitReindexReportingBatch(c context.Context, devices []string) error
}

// Client is an opaque implementation of orchestrator client. Implements
// ClientRunner interface
type Client struct {
	conf Config
	http http.Client
}

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
		utils.JoinURL(c.conf.OrchestratorAddr, HealthURI), nil,
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

func (co *Client) SubmitDeviceDecommisioningJob(
	ctx context.Context,
	decommissioningReq DecommissioningReq,
) error {

	l := log.FromContext(ctx)

	l.Debugf("Submit decommissioning job for device: %s", decommissioningReq.DeviceId)

	DecommissioningReqJson, err := json.Marshal(decommissioningReq)
	if err != nil {
		return errors.Wrapf(err, "failed to submit device decommissioning job")
	}

	contentReader := bytes.NewReader(DecommissioningReqJson)

	req, err := http.NewRequest(
		http.MethodPost,
		utils.JoinURL(co.conf.OrchestratorAddr, DeviceDecommissioningOrchestratorUri),
		contentReader)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	// set the device admission request timeout
	ctx, cancel := context.WithTimeout(ctx, co.conf.Timeout)
	defer cancel()

	rsp, err := co.http.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrapf(err, "failed to submit decommissioning job")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			body = []byte("<failed to read>")
		}
		l.Errorf("decommision request %s %s failed with status %v, response text: %s",
			req.Method, req.URL, rsp.Status, body)

		return errors.Errorf(
			"submit decommissioning request failed with status %v", rsp.Status)
	}
	return nil
}

func (co *Client) SubmitProvisionDeviceJob(
	ctx context.Context,
	provisionDeviceReq ProvisionDeviceReq,
) error {

	l := log.FromContext(ctx)

	l.Debugf("Submit provision device job for device: %s", provisionDeviceReq.DeviceID)

	ProvisionDeviceReqJson, err := json.Marshal(provisionDeviceReq)
	if err != nil {
		return errors.Wrapf(err, "failed to submit provision device job")
	}

	contentReader := bytes.NewReader(ProvisionDeviceReqJson)

	req, err := http.NewRequest(
		http.MethodPost,
		utils.JoinURL(co.conf.OrchestratorAddr, ProvisionDeviceOrchestratorUri),
		contentReader)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	// set the device admission request timeout
	ctx, cancel := context.WithTimeout(ctx, co.conf.Timeout)
	defer cancel()

	rsp, err := co.http.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrapf(err, "failed to submit provision device job")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			body = []byte("<failed to read>")
		}
		l.Errorf("provision device request %s %s failed with status %v, response text: %s",
			req.Method, req.URL, rsp.Status, body)

		return errors.Errorf(
			"submit provision device request failed with status %v", rsp.Status)
	}
	return nil
}

func (co *Client) SubmitUpdateDeviceStatusJob(
	ctx context.Context,
	updateDeviceStatusReq UpdateDeviceStatusReq,
) error {
	l := log.FromContext(ctx)

	l.Debugf("Submit update device status job for devices: %v", updateDeviceStatusReq.Devices)

	UpdateDeviceStatusReqJson, err := json.Marshal(updateDeviceStatusReq)
	if err != nil {
		return errors.Wrapf(err, "failed to submit update device status job")
	}

	contentReader := bytes.NewReader(UpdateDeviceStatusReqJson)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		utils.JoinURL(co.conf.OrchestratorAddr, UpdateDeviceStatusOrchestratorUri),
		contentReader)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	// set the device admission request timeout
	ctx, cancel := context.WithTimeout(ctx, co.conf.Timeout)
	defer cancel()

	rsp, err := co.http.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrapf(err, "failed to submit update device status job")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			body = []byte("<failed to read>")
		}
		l.Errorf("update device status request %s %s failed with status %v, response text: %s",
			req.Method, req.URL, rsp.Status, body)

		return errors.Errorf(
			"submit update device status request failed with status %v", rsp.Status)
	}
	return nil
}

func (co *Client) SubmitDeviceLimitWarning(
	ctx context.Context,
	devWarn DeviceLimitWarning,
) error {
	if err := devWarn.Validate(); err != nil {
		return errors.Wrap(err,
			"workflows: [internal] invalid request argument",
		)
	}

	bodyJSON, _ := json.Marshal(devWarn)

	req, err := http.NewRequestWithContext(ctx, "POST",
		utils.JoinURL(co.conf.OrchestratorAddr, DeviceLimitWarningURI),
		bytes.NewReader(bodyJSON),
	)
	if err != nil {
		return errors.Wrap(err,
			"workflows: error preparing device limit warning request",
		)
	}
	rsp, err := co.http.Do(req)
	if err != nil {
		return errors.Wrap(err,
			"workflows: error sending device limit warning request",
		)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= 400 {
		var (
			apiErr    = new(rest_utils.ApiError)
			jsDecoder = json.NewDecoder(rsp.Body)
		)
		err := jsDecoder.Decode(apiErr)
		if err != nil {
			return errors.Errorf(
				"workflows: unexpected HTTP response: %s",
				rsp.Status,
			)
		}
		return apiErr
	}
	return nil
}

func (co *Client) SubmitUpdateDeviceInventoryJob(
	ctx context.Context,
	updateDeviceInventoryReq UpdateDeviceInventoryReq,
) error {
	l := log.FromContext(ctx)

	l.Debugf("Submit update device inventory job for device: %q", updateDeviceInventoryReq.DeviceId)

	UpdateDeviceInventoryReqJson, err := json.Marshal(updateDeviceInventoryReq)
	if err != nil {
		return errors.Wrapf(err, "failed to submit update device inventory job")
	}

	contentReader := bytes.NewReader(UpdateDeviceInventoryReqJson)

	req, err := http.NewRequest(
		http.MethodPost,
		utils.JoinURL(co.conf.OrchestratorAddr, UpdateDeviceInventoryOrchestratorUri),
		contentReader)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	// set the workflows client request timeout
	ctx, cancel := context.WithTimeout(ctx, co.conf.Timeout)
	defer cancel()

	rsp, err := co.http.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrapf(err, "failed to submit update device inventory job")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			body = []byte("<failed to read>")
		}
		l.Errorf("update device inventory request %s %s failed with status %v, response text: %s",
			req.Method, req.URL, rsp.Status, body)

		return errors.Errorf(
			"submit update device inventory request failed with status %v", rsp.Status)
	}
	return nil
}

func (co *Client) SubmitReindexReporting(ctx context.Context, device string) error {
	ctx, cancel := context.WithTimeout(ctx, co.conf.Timeout)
	defer cancel()

	tenantID := ""
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	wflow := ReindexReportingWorkflow{
		RequestID: requestid.FromContext(ctx),
		TenantID:  tenantID,
		DeviceID:  device,
		Service:   ServiceDeviceauth,
	}
	payload, _ := json.Marshal(wflow)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		utils.JoinURL(co.conf.OrchestratorAddr, ReindexReportingURI),
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := co.http.Do(req)
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

func (co *Client) SubmitReindexReportingBatch(ctx context.Context, devices []string) error {
	ctx, cancel := context.WithTimeout(ctx, co.conf.Timeout)
	defer cancel()

	tenantID := ""
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	reqID := requestid.FromContext(ctx)
	wflows := make([]ReindexReportingWorkflow, len(devices))
	for i, device := range devices {
		wflows[i] = ReindexReportingWorkflow{
			RequestID: reqID,
			TenantID:  tenantID,
			DeviceID:  device,
			Service:   ServiceDeviceauth,
		}
	}
	payload, _ := json.Marshal(wflows)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		utils.JoinURL(co.conf.OrchestratorAddr, ReindexReportingBatchURI),
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := co.http.Do(req)
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
