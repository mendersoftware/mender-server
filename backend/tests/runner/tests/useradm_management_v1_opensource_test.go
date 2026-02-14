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
	"context"

	openapi "github.com/mendersoftware/mender-server/pkg/api/client"
)

type UseradmManagementV1Suite struct {
	BaseIntegrationSuite
}

func (u *UseradmManagementV1Suite) SetupSuite() {
	u.BaseIntegrationSuite.SetupSuite()
}

func (u *UseradmManagementV1Suite) TearDownSuite() {
	u.BaseIntegrationSuite.TearDownSuite()
}

func (u *UseradmManagementV1Suite) loginTestUser() (context.Context, string) {
	u.T().Helper()
	user := u.CreateUser("test-useradm@mender.io", "test-password-123!")
	_ = user
	token := u.Login("test-useradm@mender.io", "test-password-123!")
	ctx := context.WithValue(context.Background(), openapi.ContextAccessToken, token)
	return ctx, token
}

func (u *UseradmManagementV1Suite) TestLogin() {
	u.CreateUser("test-login@mender.io", "test-password-123!")
	auth := openapi.BasicAuth{
		UserName: "test-login@mender.io",
		Password: "test-password-123!",
	}
	ctx := context.WithValue(context.Background(), openapi.ContextBasicAuth, auth)
	token, r, err := u.Settings.Client.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	u.Require().NoError(err)
	u.Require().NotNil(r)
	u.Assert().Equal(200, r.StatusCode)
	u.Assert().NotZero(len(token))
}

func (u *UseradmManagementV1Suite) TestMe() {
	ctx, _ := u.loginTestUser()
	body, r, err := u.Settings.Client.UserAdministrationManagementAPIAPI.ShowMyUserSettings(ctx).Execute()
	u.Require().NoError(err)
	u.Require().NotNil(r)
	u.Assert().Equal(200, r.StatusCode)
	u.Assert().NotNil(body)
}

func (u *UseradmManagementV1Suite) TestSelf() {
	ctx, _ := u.loginTestUser()
	body, r, err := u.Settings.Client.UserAdministrationManagementAPIAPI.ShowOwnUserData(ctx).Execute()
	u.Require().NoError(err)
	u.Require().NotNil(r)
	u.Assert().Equal("test-useradm@mender.io", body.Email)
}

func (u *UseradmManagementV1Suite) TestRemoveUser() {
	ctx, _ := u.loginTestUser()
	r, err := u.Settings.Client.UserAdministrationManagementAPIAPI.RemoveUser(ctx, "id").Execute()
	u.Require().NoError(err)
	u.Require().NotNil(r)
	u.Assert().Equal(204, r.StatusCode)
}
