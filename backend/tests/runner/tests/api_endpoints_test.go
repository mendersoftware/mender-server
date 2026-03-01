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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

// OS services whose API specs are tested.
var osRepos = []string{
	"deployments",
	"deviceauth",
	"deviceconfig",
	"deviceconnect",
	"inventory",
	"iot-manager",
	"useradm",
	"workflows",
}

// internalServiceHosts maps service names to their internal hostname:port.
var internalServiceHosts = map[string]string{
	"deployments":  "deployments:8080",
	"deviceauth":   "deviceauth:8080",
	"deviceconfig": "deviceconfig:8080",
	"deviceconnect": "deviceconnect:8080",
	"inventory":    "inventory:8080",
	"iot-manager":  "iot-manager:8080",
	"useradm":      "useradm:8080",
	"workflows":    "workflows:8080",
}

// apiEndpoint describes a single API endpoint parsed from a spec.
type apiEndpoint struct {
	Service    string
	Kind       string // "management", "devices", or "internal"
	Returns401 bool
	Method     string
	Path       string // absolute path, e.g. "/api/management/v1/deployments/..."
}

// APIEndpointsSuite tests all API endpoints from OpenAPI specs.
// Mirrors test_api_endpoints.py::TestAPIEndpoints.
type APIEndpointsSuite struct {
	suite.Suite
	Settings *TestSettings
}

func (s *APIEndpointsSuite) TestAPIEndpoints() {
	endpoints := s.loadAllEndpoints(osRepos)
	s.Require().NotEmpty(endpoints, "no API endpoints found — is /backend mounted?")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	for _, ep := range endpoints {
		name := fmt.Sprintf("%s/%s_%s_%s", ep.Service, ep.Kind, ep.Method, ep.Path)
		s.Run(name, func() {
			var baseURL string
			if ep.Kind == "management" || ep.Kind == "devices" {
				// Management and device endpoints go through the gateway
				baseURL = s.Settings.ServerURL
			} else {
				// Internal endpoints go directly to the service
				host := internalServiceHosts[ep.Service]
				baseURL = "http://" + host
			}
			url := baseURL + ep.Path

			req, err := http.NewRequest(strings.ToUpper(ep.Method), url, nil)
			s.Require().NoError(err)

			resp, err := client.Do(req)
			s.Require().NoError(err, "request to %s failed", url)
			defer resp.Body.Close()

			if ep.Returns401 {
				s.Assert().Equal(401, resp.StatusCode,
					"expected 401 for authenticated endpoint %s %s", ep.Method, ep.Path)
			} else {
				s.Assert().NotEqual(401, resp.StatusCode,
					"unexpected 401 for unauthenticated endpoint %s %s", ep.Method, ep.Path)
				s.Assert().True(resp.StatusCode >= 200 && resp.StatusCode < 500 && resp.StatusCode != 405,
					"unexpected status %d for %s %s", resp.StatusCode, ep.Method, ep.Path)
			}
		})
	}
}

func (s *APIEndpointsSuite) loadAllEndpoints(repos []string) []apiEndpoint {
	var all []apiEndpoint
	for _, repo := range repos {
		endpoints, err := getAPIEndpoints(repo)
		if err != nil {
			s.T().Logf("warning: could not load endpoints for %s: %v", repo, err)
			continue
		}
		all = append(all, endpoints...)
	}
	return all
}

func getAPIEndpoints(service string) ([]apiEndpoint, error) {
	// API specs are in the backend volume at /backend/docs/api/{service}/.
	// Files follow the pattern: devices_v1.yaml, management_v1.yaml, internal_v1.yaml.
	docsDir := filepath.Join("/backend/docs/api", service)
	files, err := filepath.Glob(filepath.Join(docsDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	var endpoints []apiEndpoint
	for _, file := range files {
		basename := filepath.Base(file)

		// Determine API kind from filename prefix; skip non-endpoint files
		var kind string
		switch {
		case strings.HasPrefix(basename, "management_"):
			kind = "management"
		case strings.HasPrefix(basename, "devices_"):
			kind = "devices"
		case strings.HasPrefix(basename, "internal_"):
			kind = "internal"
		default:
			// Skip schemas.yaml and other non-endpoint files
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", file, err)
		}

		var spec specFile
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}

		eps := parseSpec(spec, service, kind)
		endpoints = append(endpoints, eps...)
	}
	return endpoints, nil
}

// specFile represents a minimal OpenAPI spec (paths and security only).
type specFile struct {
	Paths    map[string]map[string]method `yaml:"paths"`
	Security []interface{}                `yaml:"security"`
}

type method struct {
	Security []interface{} `yaml:"security"`
}

func parseSpec(spec specFile, service, kind string) []apiEndpoint {
	globalSecurity := len(spec.Security) > 0

	var endpoints []apiEndpoint
	for path, methods := range spec.Paths {
		for m, def := range methods {
			trimmedPath := strings.TrimRight(path, "/")

			// Skip shutdown endpoint
			if strings.HasSuffix(trimmedPath, "/shutdown") {
				continue
			}

			returns401 := len(def.Security) > 0 ||
				globalSecurity ||
				strings.HasSuffix(trimmedPath, "/verify") ||
				strings.HasSuffix(trimmedPath, "/2faqr") ||
				strings.HasSuffix(trimmedPath, "/2faverify") ||
				strings.HasSuffix(trimmedPath, "/auth/magic/{id}")

			// auth_requests returns 400 before 401
			if strings.HasSuffix(trimmedPath, "/auth_requests") {
				returns401 = false
			}

			endpoints = append(endpoints, apiEndpoint{
				Service:    service,
				Kind:       kind,
				Returns401: returns401,
				Method:     m,
				Path:       path,
			})
		}
	}
	return endpoints
}
