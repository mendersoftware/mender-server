package tests

import (
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	openapi "github.com/mendersoftware/mender-server/tests/runner/client"
)

type UseradmManagementV1Suite struct {
	suite.Suite
	settings *TestSettings
}

func (u *UseradmManagementV1Suite) SetupSuite() {
	u.T().Logf("user adm management v1 open source suite init")
	auth := openapi.BasicAuth{
		UserName: u.settings.Username,
		Password: u.settings.Password,
	}
	ctx := context.WithValue(context.Background(), openapi.ContextBasicAuth, auth)
	token, r, err := u.settings.client.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	assert.NoError(u.T(), err)
	assert.NotNil(u.T(), r)
	assert.NotZero(u.T(), len(token))
	assert.Equal(u.T(), 200, r.StatusCode)
	u.settings.jwt = token
}

func (u *UseradmManagementV1Suite) TestLogin() {
	u.T().Logf("login test starting")
	auth := openapi.BasicAuth{
		UserName: u.settings.Username,
		Password: u.settings.Password,
	}
	ctx := context.WithValue(context.Background(), openapi.ContextBasicAuth, auth)
	token, r, err := u.settings.client.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	assert.NoError(u.T(), err)
	assert.NotNil(u.T(), r)
	assert.NotZero(u.T(), len(token))
	assert.Equal(u.T(), 200, r.StatusCode)
	u.settings.jwt = token
	u.T().Logf("test passed with %d jwt len=%d", r.StatusCode, len(token))
}

func (u *UseradmManagementV1Suite) TestMe() {
	u.T().Logf("me test starting")
	ctx := context.WithValue(context.Background(), openapi.ContextAccessToken, u.settings.jwt)
	body, r, err := u.settings.client.UserAdministrationManagementAPIAPI.ShowMyUserSettings(ctx).Execute()
	assert.NoError(u.T(), err)
	assert.NotNil(u.T(), r)
	assert.NotNil(u.T(), body)
	assert.Equal(u.T(), 200, r.StatusCode)
	u.T().Logf("test passed with data: %+v", body)
}

func (u *UseradmManagementV1Suite) TestSelf() {
	u.T().Logf("self test starting")
	ctx := context.WithValue(context.Background(), openapi.ContextAccessToken, u.settings.jwt)
	body, r, err := u.settings.client.UserAdministrationManagementAPIAPI.ShowOwnUserData(ctx).Execute()
	assert.NoError(u.T(), err)
	assert.NotNil(u.T(), r)
	assert.Equal(u.T(), u.settings.Username, body.Email)
}

func (u *UseradmManagementV1Suite) TestRemoveUser() {
	u.T().Logf("remove user test starting")
	ctx := context.WithValue(context.Background(), openapi.ContextAccessToken, u.settings.jwt)
	r, err := u.settings.client.UserAdministrationManagementAPIAPI.RemoveUser(ctx, "id").Execute()
	assert.NoError(u.T(), err)
	assert.NotNil(u.T(), r)
	assert.Equal(u.T(), 204, r.StatusCode)
}

func (u *UseradmManagementV1Suite) TearDownSuite() {
	u.T().Logf("user adm management v1 open source suite teardown")
}
