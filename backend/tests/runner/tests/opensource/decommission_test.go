//nolint:all // This is all test code
package opensource

import (
	"context"
	"net/http"
	"time"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/utils/types"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DecommissionSuite ports the OS device decommissioning test from
// mender-server/backend/tests/integration/test_decommission.py.
// Enterprise-only branches (useExistingTenant()) are intentionally not
// ported.
type DecommissionSuite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string

	// deviceID tracks the device this suite creates, so TearDownSuite can
	// decommission it even if a mid-test failure happens before the
	// happy path's own decommission call runs.
	deviceID string
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
	token, r, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.Equal(http.StatusOK, r.StatusCode)
	require.NotEmpty(token)
	s.JWT = token
}

// TearDownSuite decommissions the device this suite created, tolerating
// it already being gone (the happy path in TestOk already decommissions
// it, so this is only needed if a mid-test failure happens first).
func (s *DecommissionSuite) TearDownSuite() {
	if s.JWT == "" || s.deviceID == "" {
		return
	}
	ctx := common.JWTAuthContext(context.Background(), s.JWT)
	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementDecommissionDevice(ctx, s.deviceID).Execute()
	if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
		s.T().Logf("failed to decommission device %s in teardown: %v", s.deviceID, err)
	}
}

func (s *DecommissionSuite) TestOk() {
	// ported from test_decommission.py::TestDeviceDecomissioning::test_ok
	//
	// The Python original stages this by hand (preauthorize/authset,
	// accept, submit auth, poll for provisioning) using its own
	// fixtures. common.NewAcceptedDevice covers the same ground: it
	// creates the device, submits the auth request, accepts the first
	// authset, and polls inventory until the device shows up as
	// "accepted" -- then submits the auth request again so the device
	// ends up with a valid token, matching the state the Python test is
	// in right before it patches inventory attributes.
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	device, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	require.NotEmpty(device.ID)
	require.NotEmpty(device.Token)
	s.deviceID = device.ID

	// check if the device can access the API by patching its own inventory
	devCtx := common.JWTAuthContext(ctx, device.Token)
	payload := []client.DeviceAttributeRequest{
		{Name: "mac", Value: client.StringAsAttributeValueRequest(types.Pointer("foo"))},
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
