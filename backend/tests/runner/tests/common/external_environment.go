package common

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	oapi "github.com/mendersoftware/mender-server/pkg/api"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	oapiclient "github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	MinimumSPStandardDevices = 1000
	MinimumSPMicroDevices    = 1000
	MinimumSPSystemDevices   = 1000
)

type ExternalEnvironmentConfig struct {
	URL  string
	User User
}

type ExternalEnvironment struct {
	config    ExternalEnvironmentConfig
	apiClient *oapiclient.APIClient

	cleanupTenants []Tenant
}

func (e *ExternalEnvironment) Setup(t *testing.T) {
	t.Logf("Using external testing environment: %s", e.config.URL)

	var (
		ctx = t.Context()
	)

	config, err := oapi.NewDefaultClientConfigurationFromURL(e.config.URL)
	require.NoError(t, err)
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	apiClient := client.NewAPIClient(config)

	t.Logf("Sanity checking external test environment...")

	token, err := login(ctx, e.config.User, apiClient)
	require.NoError(t, err)

	user, _, err := apiClient.UserAdministrationManagementAPIAPI.
		ShowOwnUserData(JWTAuthContext(ctx, token)).
		Execute()
	require.NoError(t, err)
	require.NotNil(t, user)

	e.apiClient = apiClient
	t.Logf("Looks ok, proceeding...")
}

func (e *ExternalEnvironment) Teardown(t *testing.T) {
}

func (e *ExternalEnvironment) User(ctx context.Context) (User, error) {
	return e.config.User, nil
}

func (e *ExternalEnvironment) TenantOfUser(ctx context.Context, user User) (Tenant, error) {
	return Tenant{}, nil
}

func (e *ExternalEnvironment) APIClient() *oapiclient.APIClient {
	return e.apiClient
}

func NewExternalEnvironment(config ExternalEnvironmentConfig) *ExternalEnvironment {
	return &ExternalEnvironment{
		config: config,
	}
}

func login(ctx context.Context, user User, client *client.APIClient) (string, error) {
	token, r, err := client.UserAdministrationManagementAPIAPI.Login(BasicAuthContext(ctx, user)).Execute()
	if err != nil {
		return "", errors.Wrap(err, "failed to login")
	}

	if r == nil || r.StatusCode != http.StatusOK {
		return "", errors.New("failed to login")
	}
	return token, nil
}
