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

package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"
	"github.com/pkg/errors"

	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

const (
	healthURL                          = "/api/v1/health"
	generateArtifactURL                = "/api/v1/workflow/generate_artifact"
	reindexReportingURL                = "/api/v1/workflow/reindex_reporting"
	reindexReportingDeploymentURL      = "/api/v1/workflow/reindex_reporting_deployment"
	reindexReportingDeploymentBatchURL = "/api/v1/workflow/reindex_reporting_deployment/batch"
	defaultTimeout                     = 5 * time.Second
)

type DeviceDeploymentShortInfo struct {
	ID           string
	DeviceID     string
	DeploymentID string
}

// Client is the workflows client
//
//go:generate ../../utils/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	StartGenerateArtifact(
		ctx context.Context,
		multipartGenerateImageMsg *model.MultipartGenerateImageMsg,
	) error
	StartReindexReporting(c context.Context, device string) error
	StartReindexReportingDeployment(c context.Context, device, deployment, id string) error
	StartReindexReportingDeploymentBatch(c context.Context, info []DeviceDeploymentShortInfo) error
}

// NewClient returns a new workflows client
func NewClient() Client {
	workflowsBaseURL := config.Config.GetString(dconfig.SettingWorkflows)
	return &client{
		baseURL:    workflowsBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
}

type client struct {
	baseURL    string
	httpClient *http.Client
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

func (c *client) StartGenerateArtifact(
	ctx context.Context,
	multipartGenerateImageMsg *model.MultipartGenerateImageMsg,
) error {
	l := log.FromContext(ctx)
	l.Debugf("Submit generate artifact: tenantID=%s, artifactID=%s",
		multipartGenerateImageMsg.TenantID, multipartGenerateImageMsg.ArtifactID)

	workflowsURL := c.baseURL + generateArtifactURL

	payload, _ := json.Marshal(multipartGenerateImageMsg)
	req, err := http.NewRequest("POST", workflowsURL, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to start workflow: generate_artifact")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			body = []byte("<failed to read>")
		}
		l.Errorf("generate artifact failed with status %v, response text: %s",
			res.StatusCode, body)
		return errors.New("failed to start workflow: generate_artifact")
	}
	return nil
}

func (c *client) StartReindexReporting(ctx context.Context, device string) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	tenantID := ""
	if ident := identity.FromContext(ctx); ident != nil {
		tenantID = ident.Tenant
	}
	wflow := ReindexWorkflow{
		RequestID: requestid.FromContext(ctx),
		TenantID:  tenantID,
		DeviceID:  device,
		Service:   ServiceDeployments,
	}
	payload, _ := json.Marshal(wflow)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		c.baseURL+reindexReportingURL,
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "workflows: failed to trigger reporting reindex")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 300 {
		return nil
	}

	if rsp.StatusCode == http.StatusNotFound {
		workflowURIparts := strings.Split(reindexReportingURL, "/")
		workflowName := workflowURIparts[len(workflowURIparts)-1]
		return errors.New(`workflows: workflow "` + workflowName + `" not defined`)
	}

	return errors.Errorf(
		"workflows: unexpected HTTP status from workflows service: %s",
		rsp.Status,
	)
}

func (c *client) StartReindexReportingDeployment(ctx context.Context,
	device, deployment, id string) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	tenantID := ""
	if ident := identity.FromContext(ctx); ident != nil {
		tenantID = ident.Tenant
	}
	wflow := ReindexDeploymentWorkflow{
		RequestID:    requestid.FromContext(ctx),
		TenantID:     tenantID,
		DeviceID:     device,
		DeploymentID: deployment,
		ID:           id,
		Service:      ServiceDeployments,
	}
	payload, _ := json.Marshal(wflow)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		c.baseURL+reindexReportingDeploymentURL,
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "workflows: failed to trigger reporting reindex deployment")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 300 {
		return nil
	}

	if rsp.StatusCode == http.StatusNotFound {
		workflowURIparts := strings.Split(reindexReportingDeploymentURL, "/")
		workflowName := workflowURIparts[len(workflowURIparts)-1]
		return errors.New(`workflows: workflow "` + workflowName + `" not defined`)
	}

	return errors.Errorf(
		"workflows: unexpected HTTP status from workflows service: %s",
		rsp.Status,
	)
}

func (c *client) StartReindexReportingDeploymentBatch(ctx context.Context,
	info []DeviceDeploymentShortInfo) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	tenantID := ""
	if ident := identity.FromContext(ctx); ident != nil {
		tenantID = ident.Tenant
	}
	reqID := requestid.FromContext(ctx)
	wflows := make([]ReindexDeploymentWorkflow, len(info))
	for i, d := range info {
		wflows[i] = ReindexDeploymentWorkflow{
			RequestID:    reqID,
			TenantID:     tenantID,
			DeviceID:     d.DeviceID,
			DeploymentID: d.DeploymentID,
			ID:           d.ID,
			Service:      ServiceDeployments,
		}
	}
	payload, _ := json.Marshal(wflows)
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		c.baseURL+reindexReportingDeploymentBatchURL,
		bytes.NewReader(payload),
	)
	if err != nil {
		return errors.Wrap(err, "workflows: error preparing HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "workflows: failed to trigger reporting reindex deployment")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 300 {
		return nil
	}

	if rsp.StatusCode == http.StatusNotFound {
		workflowURIparts := strings.Split(reindexReportingDeploymentURL, "/")
		workflowName := workflowURIparts[len(workflowURIparts)-1]
		return errors.New(`workflows: workflow "` + workflowName + `" not defined`)
	}

	return errors.Errorf(
		"workflows: unexpected HTTP status from workflows service: %s",
		rsp.Status,
	)
}
