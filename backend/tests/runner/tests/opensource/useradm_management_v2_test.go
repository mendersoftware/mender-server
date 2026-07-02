package opensource

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	oapiclient "github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UseradmManagementV2Suite struct {
	suite.Suite

	APIClient *oapiclient.APIClient
	User      common.User

	JWT string
}

func (i *BackendIntegrationSuite) TestUseradmManagementV2() {
	suite.Run(i.T(), &UseradmManagementV2Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
	})
}

func (u *UseradmManagementV2Suite) SetupSuite() {
	require := require.New(u.T())
	ctx := common.BasicAuthContext(u.T().Context(), u.User)

	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(200, r.StatusCode)
	require.NotZero(len(token))

	u.JWT = token
}

func (u *UseradmManagementV2Suite) login(user common.User) (string, error) {
	ctx := common.BasicAuthContext(u.T().Context(), user)

	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	if err != nil {
		return "", err
	}

	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected status code: %d", r.StatusCode)
	}

	return token, nil
}

func (u *UseradmManagementV2Suite) createUser(user common.User) error {
	ctx := common.JWTAuthContext(u.T().Context(), u.JWT)

	r, err := u.APIClient.UserAdministrationManagementAPIAPI.
		CreateUserManagement(ctx).
		UserNew(*oapiclient.NewUserNew(user.Username, user.Password)).
		Execute()
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status code: %d", r.StatusCode)
	}

	return nil
}

func (u *UseradmManagementV2Suite) TestUpdateOwnUser() {
	require := require.New(u.T())
	email := fmt.Sprintf("user-%s@docker.mender.io", uuid.New().String())
	password := "correcthorsebatterystaple"

	user := common.User{
		Username: email,
		Password: password,
	}

	err := u.createUser(user)
	require.NoError(err)

	token, err := u.login(user)
	require.NoError(err)
	require.NotZero(len(token))

	// Update user with new password
	ctx := common.JWTAuthContext(u.T().Context(), token)
	newPassword := "password123"

	userUpdate := oapiclient.NewOwnUserUpdateV2()
	userUpdate.SetCurrentPassword(user.Password)
	userUpdate.SetPassword(newPassword)

	r, err := u.APIClient.UserAdministrationV2ManagementAPIAPI.UseradmV2UpdateOwnUserData(ctx).OwnUserUpdateV2(*userUpdate).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(204, r.StatusCode)

	// Try to login with the old password
	ctx = common.BasicAuthContext(u.T().Context(), user)

	token, r, err = u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.Error(err)
	require.NotNil(r)
	require.Equal(401, r.StatusCode)

	user.Password = newPassword

	// Login with the new password
	ctx = common.BasicAuthContext(u.T().Context(), user)

	token, r, err = u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(200, r.StatusCode)
	require.NotZero(len(token))
}
