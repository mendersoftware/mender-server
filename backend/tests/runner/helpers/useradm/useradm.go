// Copyright 2025 Northern.tech AS
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

package useradm

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	openapi "github.com/mendersoftware/mender-server/tests/runner/client"
)

const (
	Host = "useradm:8080"
)

// User holds test user data.
type User struct {
	ID       string
	Name     string
	Password string
	Token    string
}

// NewInternalClient creates an OpenAPI client configured for direct
// access to the useradm internal API (bypassing the gateway).
func NewInternalClient(host string, httpClient *http.Client) *openapi.APIClient {
	config := openapi.NewConfiguration()
	config.Host = host
	config.Scheme = "http"
	config.HTTPClient = httpClient
	return openapi.NewAPIClient(config)
}

// CreateUser creates a user via docker exec into the useradm container.
// This mirrors the Python CliUseradm.create_user approach.
func CreateUser(name, password string, namespace string) (*User, error) {
	if namespace == "" {
		namespace = "backend-tests"
	}
	container := namespace + "-useradm-1"

	args := []string{
		"exec", container,
		"useradm", "create-user",
		"--username", name,
		"--password", password,
	}
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("create-user failed: %w: %s", err, string(out))
	}

	uid := strings.TrimSpace(string(out))
	return &User{ID: uid, Name: name, Password: password}, nil
}

// Login performs a login via the management API using the generated OpenAPI client.
func Login(client *openapi.APIClient, email, password string) (string, error) {
	auth := openapi.BasicAuth{
		UserName: email,
		Password: password,
	}
	ctx := context.WithValue(context.Background(), openapi.ContextBasicAuth, auth)

	token, resp, err := client.UserAdministrationManagementAPIAPI.
		Login(ctx).
		Execute()
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status %d", resp.StatusCode)
	}
	return token, nil
}

// CreateUserInternal creates a user via the useradm internal API
// using the generated OpenAPI client.
// tenantID can be empty for OSS (single-tenant) mode.
func CreateUserInternal(client *openapi.APIClient, email, password, tenantID string) (*User, error) {
	if tenantID == "" {
		tenantID = "default"
	}

	userNew := *openapi.NewUserNewInternal(email, password)

	resp, err := client.UserAdministrationAndAuthenticationInternalAPIAPI.
		CreateUserInternal(context.Background(), tenantID).
		UserNewInternal(userNew).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("create user internal failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create user internal failed with status %d", resp.StatusCode)
	}

	return &User{Name: email, Password: password}, nil
}
