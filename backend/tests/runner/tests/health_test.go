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
	"context"
	"net/http"

	"github.com/stretchr/testify/suite"

	openapi "github.com/mendersoftware/mender-server/pkg/api/client"
)

// HealthSuite tests the internal health check endpoints
// of all backend services. Mirrors test_health.py::TestHealthCheck.
type HealthSuite struct {
	suite.Suite
	Settings *TestSettings
}

func (s *HealthSuite) newInternalClient(host string) *openapi.APIClient {
	config := openapi.NewConfiguration()
	config.Host = host
	config.Scheme = "http"
	config.HTTPClient = &http.Client{}
	return openapi.NewAPIClient(config)
}

func (s *HealthSuite) TestHealthDeployments() {
	c := s.newInternalClient("deployments:8080")
	_, err := c.DeploymentsInternalAPIInternalAPIAPI.
		DeploymentsInternalCheckHealth(context.Background()).Execute()
	s.Require().NoError(err)
}

func (s *HealthSuite) TestHealthDeviceAuth() {
	c := s.newInternalClient("deviceauth:8080")
	_, err := c.DeviceAuthenticationInternalAPIAPI.
		DeviceAuthInternalCheckHealth(context.Background()).Execute()
	s.Require().NoError(err)
}

func (s *HealthSuite) TestHealthDeviceConfig() {
	c := s.newInternalClient("deviceconfig:8080")
	_, err := c.DeviceConfigureInternalAPIAPI.
		DeviceConfigInternalCheckHealth(context.Background()).Execute()
	s.Require().NoError(err)
}

func (s *HealthSuite) TestHealthDeviceConnect() {
	c := s.newInternalClient("deviceconnect:8080")
	_, err := c.DeviceConnectInternalAPIAPI.
		DeviceConnectInternalCheckHealth(context.Background()).Execute()
	s.Require().NoError(err)
}

func (s *HealthSuite) TestHealthInventory() {
	c := s.newInternalClient("inventory:8080")
	_, err := c.DeviceInventoryInternalAPIAPI.
		InventoryInternalCheckHealth(context.Background()).Execute()
	s.Require().NoError(err)
}

func (s *HealthSuite) TestHealthIoTManager() {
	// IoT Manager internal API is not in the generated client spec;
	// use raw HTTP.
	resp, err := http.Get("http://iot-manager:8080/api/internal/v1/iot-manager/health")
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Assert().Less(resp.StatusCode, 300, "iot-manager health check failed")
}

func (s *HealthSuite) TestHealthUseradm() {
	c := s.newInternalClient("useradm:8080")
	_, err := c.UserAdministrationAndAuthenticationInternalAPIAPI.
		UseradmCheckHealth(context.Background()).Execute()
	s.Require().NoError(err)
}

func (s *HealthSuite) TestHealthWorkflows() {
	// The generated client constructs /health instead of /api/v1/health
	// for this endpoint (generator bug), so use raw HTTP like IoT Manager.
	resp, err := http.Get("http://workflows:8080/api/v1/health")
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Assert().Equal(http.StatusNoContent, resp.StatusCode, "workflows health check failed")
}
