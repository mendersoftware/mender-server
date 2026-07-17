//nolint:all // This is all test code
package opensource

import (
	"net/http"
	"time"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DecommissionSuite covers open-source device decommissioning. Enterprise-only
// tenant behavior is out of scope.
type DecommissionSuite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
}

func (i *BackendIntegrationSuite) TestDecommission() {
	suite.Run(i.T(), &DecommissionSuite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (s *DecommissionSuite) SetupSuite() {
	require := require.New(s.T())

	ctx := common.BasicAuthContext(s.T().Context(), s.User)
	token, _, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotEmpty(token)
	s.JWT = token
}

func (s *DecommissionSuite) TestOk() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	device, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	require.NotEmpty(device.ID)
	require.NotEmpty(device.Token)

	// check if the device can access the API by patching its own inventory
	devCtx := common.JWTAuthContext(ctx, device.Token)
	payload := []client.DeviceAttributeRequest{
		{Name: "mac", Value: client.StringAsAttributeValueRequest(client.PtrString("foo"))},
	}
	r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
		AssignAttributes(devCtx).
		DeviceAttributeRequest(payload).
		Execute()
	require.NoError(err)
	assert.Equal(http.StatusOK, r.StatusCode)

	// decommission
	_, err = s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementDecommissionDevice(ctx, device.ID).Execute()
	require.NoError(err)

	// check device is rejected
	_, r, err = device.CheckUpdate(ctx, s.APIClient, "bar", "foo")
	require.Error(err)
	assert.Equal(http.StatusUnauthorized, r.StatusCode)

	// check device gone from inventory -- this may take some time
	// because it's done as an async job (workflow)
	err = common.RetryUntil(ctx, 3*time.Minute, time.Second, func() (bool, error) {
		_, r, _ := s.APIClient.DeviceInventoryManagementAPIAPI.
			GetDeviceInventory(ctx, device.ID).Execute()
		return r != nil && r.StatusCode == http.StatusNotFound, nil
	})
	require.NoError(err, "device not removed from the inventory")

	// check device gone from deviceauth
	err = common.RetryUntil(ctx, time.Minute, time.Second, func() (bool, error) {
		_, r, _ := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementGetDevice(ctx, device.ID).Execute()
		return r != nil && r.StatusCode == http.StatusNotFound, nil
	})
	require.NoError(err, "device not removed from the deviceauth")
}
