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
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	mongohelper "github.com/mendersoftware/mender-server/tests/runner/helpers/mongo"
	"github.com/mendersoftware/mender-server/tests/runner/helpers/useradm"

	openapi "github.com/mendersoftware/mender-server/tests/runner/client"
)

var ServerURL string

func TestMain(m *testing.M) {
	flag.StringVar(&ServerURL, "server-url", "traefik", "server URL (hostname or hostname:port)")
	flag.Parse()

	// Wait for all services to be healthy
	services := []ServiceHealth{
		{Name: "deployments", Path: "/api/internal/v1/deployments/health"},
		{Name: "deviceauth", Path: "/api/internal/v1/devauth/health"},
		{Name: "deviceconfig", Path: "/api/internal/v1/deviceconfig/health"},
		{Name: "deviceconnect", Path: "/api/internal/v1/deviceconnect/health"},
		{Name: "inventory", Path: "/api/internal/v1/inventory/health"},
		{Name: "iot-manager", Path: "/api/internal/v1/iot-manager/health"},
		{Name: "useradm", Path: "/api/internal/v1/useradm/health"},
		{Name: "workflows", Path: "/api/v1/health"},
	}

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	log.Println("Waiting for services to become healthy...")
	for _, svc := range services {
		if err := waitForHealth(httpClient, svc, 120*time.Second); err != nil {
			log.Fatalf("Service %s did not become healthy: %v", svc.Name, err)
		}
		log.Printf("Service %s is healthy", svc.Name)
	}

	os.Exit(m.Run())
}

// ServiceHealth defines a service to health-check.
type ServiceHealth struct {
	Name string
	Path string
}

func waitForHealth(client *http.Client, svc ServiceHealth, timeout time.Duration) error {
	url := fmt.Sprintf("http://%s:8080%s", svc.Name, svc.Path)
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("status %d", resp.StatusCode)
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for %s: %v", svc.Name, lastErr)
}

// TestSettings holds configuration shared across all test suites.
type TestSettings struct {
	ServerURL     string
	HTTPClient    *http.Client
	Configuration *openapi.Configuration
	Client        *openapi.APIClient
}

// BaseIntegrationSuite provides common setup for all integration test suites:
// - Per-test MongoDB cleanup
// - User creation and login helpers
// - Shared HTTP client and OpenAPI client
type BaseIntegrationSuite struct {
	suite.Suite
	Settings    *TestSettings
	mongoClient *mongohelper.Client
}

func (s *BaseIntegrationSuite) SetupSuite() {
	mc, err := mongohelper.NewClient("mender-mongo:27017")
	if err != nil {
		s.T().Fatalf("Failed to connect to mongo: %v", err)
	}
	s.mongoClient = mc
}

func (s *BaseIntegrationSuite) TearDownSuite() {
	if s.mongoClient != nil {
		s.mongoClient.Close()
	}
}

// SetupTest cleans the database before each test for isolation.
func (s *BaseIntegrationSuite) SetupTest() {
	if err := s.mongoClient.Cleanup(); err != nil {
		s.T().Fatalf("Failed to clean mongo: %v", err)
	}
}

// CreateUser creates a test user via docker exec into the useradm container.
func (s *BaseIntegrationSuite) CreateUser(name, password string) *useradm.User {
	s.T().Helper()
	user, err := useradm.CreateUser(name, password, "")
	s.Require().NoError(err, "creating user %s", name)
	return user
}

// Login logs in via the management API and returns a JWT token.
func (s *BaseIntegrationSuite) Login(email, password string) string {
	s.T().Helper()
	token, err := useradm.Login(s.Settings.Client, email, password)
	s.Require().NoError(err, "logging in as %s", email)
	return token
}

// BackendIntegrationSuite is the top-level test suite that dispatches
// to per-service nested suites.
type BackendIntegrationSuite struct {
	suite.Suite
	settings *TestSettings
}

func (i *BackendIntegrationSuite) SetupSuite() {
	config := openapi.NewConfiguration()
	config.Host = ServerURL
	config.Scheme = "https"
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	i.settings = &TestSettings{
		ServerURL:     "https://" + ServerURL,
		Configuration: config,
		Client:        openapi.NewAPIClient(config),
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func TestBackendIntegrationSuite(t *testing.T) {
	suite.Run(t, new(BackendIntegrationSuite))
}

// --- Nested suite registration ---

func (i *BackendIntegrationSuite) TestHealth() {
	suite.Run(i.T(), &HealthSuite{Settings: i.settings})
}

func (i *BackendIntegrationSuite) TestAPIEndpoints() {
	suite.Run(i.T(), &APIEndpointsSuite{Settings: i.settings})
}
