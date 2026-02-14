// Copyright 2026 Northern.tech AS
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

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	openapi "github.com/mendersoftware/mender-server/pkg/api/client"
)

const (
	workflowsHost = "workflows:8080"
)

// WorkflowsSuite tests workflow versioning functionality.
// Mirrors test_workflows.py::TestWorkflowMinVersion.
type WorkflowsSuite struct {
	suite.Suite
	Settings *TestSettings
	client   *openapi.APIClient
}

func (s *WorkflowsSuite) SetupSuite() {
	config := openapi.NewConfiguration()
	config.Host = workflowsHost
	config.Scheme = "http"
	config.HTTPClient = &http.Client{}
	s.client = openapi.NewAPIClient(config)
}

func (s *WorkflowsSuite) registerWorkflow(name string, version int) *http.Response {
	s.T().Helper()

	retries := int32(1)
	retryDelay := int32(4)
	schemaVersion := int32(1)
	desc := "some"

	wf := *openapi.NewWorkflow(name, int32(version), []openapi.Task{
		{
			Name:              "t1",
			Type:              "http",
			Retries:           &retries,
			RetryDelaySeconds: &retryDelay,
		},
	})
	wf.Schemaversion = &schemaVersion
	wf.Description = &desc
	wf.InputParameters = []string{"newid"}

	resp, err := s.client.WorkflowsOtherAPI.
		RegisterWorkflow(context.Background()).
		Workflow(wf).
		Execute()
	s.Require().NoError(err)
	return resp
}

// startWorkflow starts a workflow, optionally setting the X-Workflows-Min-Version
// header. Uses raw HTTP because:
// 1. The server accepts a flat map body ({"key": "value"}) rather than the
//    InputParameter array format ([{"name":"key","value":"value"}]) that the
//    generated client sends.
// 2. The generated client does not support the X-Workflows-Min-Version header.
func (s *WorkflowsSuite) startWorkflow(name, minVersion string) *http.Response {
	s.T().Helper()
	payload := map[string]string{"newid": "1"}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("http://%s/api/v1/workflow/%s", workflowsHost, name)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	if minVersion != "" {
		req.Header.Set("X-Workflows-Min-Version", minVersion)
	}

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *WorkflowsSuite) TestWorkflowMinVersion() {
	wfName := "wf1-" + uuid.New().String()
	wfVersion := 4

	s.T().Logf("creating workflow: %s/v%d", wfName, wfVersion)
	r := s.registerWorkflow(wfName, wfVersion)
	drainBody(r)
	s.Require().Equal(201, r.StatusCode, "creating workflow")

	s.T().Logf("starting: %s/v%d (exact version)", wfName, wfVersion)
	r = s.startWorkflow(wfName, fmt.Sprintf("%d", wfVersion))
	drainBody(r)
	s.Assert().Equal(201, r.StatusCode, "start with exact version")

	s.T().Logf("starting: %s/v%d (lower min version)", wfName, wfVersion-1)
	r = s.startWorkflow(wfName, fmt.Sprintf("%d", wfVersion-1))
	drainBody(r)
	s.Assert().Equal(201, r.StatusCode, "start with lower min version")

	s.T().Logf("attempting to start: %s/v%d (higher min version)", wfName, wfVersion+1)
	r = s.startWorkflow(wfName, fmt.Sprintf("%d", wfVersion+1))
	drainBody(r)
	s.Assert().Equal(404, r.StatusCode, "start with higher min version should fail")
}

func drainBody(r *http.Response) {
	if r != nil && r.Body != nil {
		io.ReadAll(r.Body)
		r.Body.Close()
	}
}
