package opensource

import (
	"net/http"
	"os"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DeploymentsManagementV1Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
}

func (i *BackendIntegrationSuite) TestDeploymentsManagementV1() {
	suite.Run(i.T(), &DeploymentsManagementV1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (u *DeploymentsManagementV1Suite) SetupTest() {
	require := require.New(u.T())

	ctx := common.BasicAuthContext(u.T().Context(), u.User)
	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.NotEmpty(token)
	require.Equal(http.StatusOK, r.StatusCode)
	u.JWT = token
}

func (u *DeploymentsManagementV1Suite) TestUploadArtifact() {
	u.Run("Success/MissingTypeInfo", func() {
		// Ensuring server backwards compatibility with artifacts created with
		// older versions of mender-artifact. Ref:
		// 	 https://github.com/mendersoftware/mender-artifact/commit/c25764218c6e48677ab0b5b1736ccaf531c62e42
		var (
			ctx     = common.JWTAuthContext(u.T().Context(), u.JWT)
			require = require.New(u.T())
		)
		file, err := os.Open("../data/missing-type-info.mender")
		require.NoError(err)
		r, err := u.APIClient.DeploymentsManagementAPIAPI.
			UploadArtifact(ctx).
			Artifact(file).
			Execute()
		require.NoError(err)
		require.NotNil(r)
		require.Equal(http.StatusCreated, r.StatusCode)
	})
}
