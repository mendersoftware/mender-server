package opensource

import (
	"context"
	"net/http"
	"time"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InventoryManagementV2Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string

	createdDevices []string
}

func (i *BackendIntegrationSuite) TestInventoryManagementV2() {
	suite.Run(i.T(), &InventoryManagementV2Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (u *InventoryManagementV2Suite) SetupSuite() {
	require := require.New(u.T())

	ctx := common.BasicAuthContext(u.T().Context(), u.User)
	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.NotEmpty(token)
	require.Equal(http.StatusOK, r.StatusCode)
	u.JWT = token
}

func (u *InventoryManagementV2Suite) TearDownSuite() {
	if u.JWT == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ctx = common.JWTAuthContext(ctx, u.JWT)
	for _, id := range u.createdDevices {
		_, _ = u.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementDecommissionDevice(ctx, id).Execute()
	}
}

func (u *InventoryManagementV2Suite) getStatistics(
	ctx context.Context,
) (client.DeviceStatusStatistics, error) {
	res, _, err := u.APIClient.
		DeviceInventoryFiltersAndSearchManagementAPIAPI.
		GetStatistics(ctx).
		Execute()
	if err != nil {
		return client.DeviceStatusStatistics{}, err
	}
	return res.GetDevicesByStatus(), nil
}

func (u *InventoryManagementV2Suite) TestGetInventoryStatistics() {
	var (
		require = require.New(u.T())
		assert  = assert.New(u.T())
		ctx     = common.JWTAuthContext(u.T().Context(), u.JWT)
	)

	// The environment is shared and never reset, and earlier suites'
	// device decommissions drain asynchronously, so absolute counts are
	// not stable here. Assert deltas against a baseline instead, and
	// allow a short settle window for in-flight removals.
	baseline, err := u.getStatistics(ctx)
	require.NoError(err)

	var devices []*common.Device
	for range 4 {
		device, err := common.NewDevice()
		require.NoError(err, "failed to create device identity")
		devices = append(devices, device)
	}

	// Create auth requests for all devices
	for _, d := range devices {
		_, err := d.SubmitAuthRequest(ctx, u.APIClient, u.Tenant.TenantToken)
		require.NoError(err)
	}

	// Accept half of them
	for _, d := range devices[:len(devices)/2] {
		err := d.Accept(ctx, u.APIClient)
		require.NoError(err)
	}

	// Track everything for teardown; pending devices also appear in
	// inventory, so their ids can be resolved the same way.
	for _, d := range devices {
		dd, err := d.WaitInventory(ctx, u.APIClient)
		require.NoError(err)
		u.createdDevices = append(u.createdDevices, dd.GetId())
	}

	wantAccepted := baseline.Accepted.Standard + int32(len(devices)/2)
	wantPending := baseline.Pending.Standard + int32(len(devices)/2)
	var statistics client.DeviceStatusStatistics
	err = common.RetryUntil(ctx, 15*time.Second, time.Second, func() (bool, error) {
		var err error
		statistics, err = u.getStatistics(ctx)
		if err != nil {
			return false, err
		}
		return statistics.Accepted.Standard == wantAccepted &&
			statistics.Pending.Standard == wantPending, nil
	})
	require.NoError(err)

	assert.Equal(wantAccepted, statistics.Accepted.Standard)
	assert.Equal(baseline.Accepted.Micro, statistics.Accepted.Micro)
	assert.Equal(baseline.Accepted.System, statistics.Accepted.System)

	assert.Equal(wantPending, statistics.Pending.Standard)
	assert.Equal(baseline.Pending.Micro, statistics.Pending.Micro)
	assert.Equal(baseline.Pending.System, statistics.Pending.System)
}
