package opensource

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
)

type BackendIntegrationSuite struct {
	suite.Suite
	environment *common.ComposeEnvironment

	user   common.User
	tenant common.Tenant
}

func (i *BackendIntegrationSuite) SetupSuite() {
	i.environment.Setup(i.T())

	user, err := i.environment.User(i.T().Context())
	require.NoError(i.T(), err, "couldn't get user dependency from provider")
	i.user = user

	tenant, err := i.environment.TenantOfUser(i.T().Context(), i.user)
	require.NoError(i.T(), err, "couldn't get user dependency from provider")
	i.tenant = tenant
}

func (i *BackendIntegrationSuite) TearDownSuite() {
	i.environment.Teardown(i.T())
}

func TestOpenSource(t *testing.T) {
	config := common.ConfigFromEnv()
	suite.Run(t, &BackendIntegrationSuite{
		environment: common.NewComposeEnvironment(
			config,
		),
	})
}
