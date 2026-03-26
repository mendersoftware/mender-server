package tests

import (
	"context"

	openapi "github.com/mendersoftware/mender-server/tests/runner/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UseradmManagementV1Suite struct {
	suite.Suite
	TestSettings

	User User
	JWT  string
}

func (i *BackendIntegrationSuite) TestUseradmManagementV1() {
	// We should probably have a better way of picking users (and tenants)...
	require := require.New(i.T())
	require.GreaterOrEqual(len(i.settings.Tenants), 1, "tests needs at least one user")
	require.GreaterOrEqual(len(i.settings.Tenants[0].Users), 1, "tests needs at least one user")

	suite.Run(i.T(), &UseradmManagementV1Suite{
		TestSettings: i.settings,
		User:         i.settings.Tenants[0].Users[0]},
	)
}

func (u *UseradmManagementV1Suite) SetupSuite() {
	require := require.New(u.T())
	ctx := basicAuthContext(u.T().Context(), u.User)

	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(200, r.StatusCode)
	require.NotZero(len(token))

	u.JWT = token
}

func (u *UseradmManagementV1Suite) TestLogin() {
	require := require.New(u.T())
	ctx := basicAuthContext(u.T().Context(), u.User)

	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(200, r.StatusCode)
	require.NotZero(len(token))
}

func (u *UseradmManagementV1Suite) TestMe() {
	require := require.New(u.T())
	ctx := jwtAuthContext(u.T().Context(), u.JWT)

	body, r, err := u.APIClient.UserAdministrationManagementAPIAPI.ShowMyUserSettings(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.NotNil(body)
	require.Equal(200, r.StatusCode)
}

func (u *UseradmManagementV1Suite) TestSelf() {
	require := require.New(u.T())
	ctx := jwtAuthContext(u.T().Context(), u.JWT)

	body, r, err := u.APIClient.UserAdministrationManagementAPIAPI.ShowOwnUserData(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(u.User.Username, body.Email)
}

func (u *UseradmManagementV1Suite) TestRemoveUser() {
	require := require.New(u.T())
	ctx := jwtAuthContext(u.T().Context(), u.JWT)

	r, err := u.APIClient.UserAdministrationManagementAPIAPI.RemoveUser(ctx, "id").Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(204, r.StatusCode)
}

func basicAuthContext(ctx context.Context, user User) context.Context {
	return context.WithValue(
		ctx,
		openapi.ContextBasicAuth,
		openapi.BasicAuth{UserName: user.Username, Password: user.Password},
	)
}

func jwtAuthContext(ctx context.Context, jwt string) context.Context {
	return context.WithValue(
		ctx,
		openapi.ContextAccessToken,
		jwt,
	)
}
