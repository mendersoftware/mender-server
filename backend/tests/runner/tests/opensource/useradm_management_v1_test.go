package opensource

import (
	oapiclient "github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UseradmManagementV1Suite struct {
	suite.Suite

	APIClient *oapiclient.APIClient
	User      common.User

	JWT string
}

func (i *BackendIntegrationSuite) TestUseradmManagementV1() {
	suite.Run(i.T(), &UseradmManagementV1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
	})
}

func (u *UseradmManagementV1Suite) SetupSuite() {
	require := require.New(u.T())
	ctx := common.BasicAuthContext(u.T().Context(), u.User)

	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(200, r.StatusCode)
	require.NotZero(len(token))

	u.JWT = token
}

func (u *UseradmManagementV1Suite) TestLogin() {
	require := require.New(u.T())
	ctx := common.BasicAuthContext(u.T().Context(), u.User)

	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(200, r.StatusCode)
	require.NotZero(len(token))
}

func (u *UseradmManagementV1Suite) TestMe() {
	require := require.New(u.T())
	ctx := common.JWTAuthContext(u.T().Context(), u.JWT)

	body, r, err := u.APIClient.UserAdministrationManagementAPIAPI.ShowMyUserSettings(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.NotNil(body)
	require.Equal(200, r.StatusCode)
}

func (u *UseradmManagementV1Suite) TestSelf() {
	require := require.New(u.T())
	ctx := common.JWTAuthContext(u.T().Context(), u.JWT)

	body, r, err := u.APIClient.UserAdministrationManagementAPIAPI.ShowOwnUserData(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(u.User.Username, body.Email)
}

func (u *UseradmManagementV1Suite) TestRemoveUser() {
	require := require.New(u.T())
	ctx := common.JWTAuthContext(u.T().Context(), u.JWT)

	r, err := u.APIClient.UserAdministrationManagementAPIAPI.RemoveUser(ctx, "id").Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(204, r.StatusCode)
}
