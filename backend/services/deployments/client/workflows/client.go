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
	"context"
	"encoding/json"
	"io"
	"net/http"
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
	healthURL           = "/api/v1/health"
	generateArtifactURL = "/api/v1/workflow/generate_artifact"
	defaultTimeout      = 5 * time.Second
)

type DeviceDeploymentShortInfo struct {
	ID           string
	DeviceID     string
	DeploymentID string
}

// Client is the workflows client
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	StartGenerateArtifact(
		ctx context.Context,
		multipartGenerateImageMsg *model.MultipartGenerateImageMsg,
	) error
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
