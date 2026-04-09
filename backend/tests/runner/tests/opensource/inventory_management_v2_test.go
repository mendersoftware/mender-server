package opensource

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	oapiclient "github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	modelinventory "github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InventoryManagementV2Suite struct {
	suite.Suite

	APIClient *oapiclient.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
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
	require.NotZero(len(token))
	require.Equal(200, r.StatusCode)
	u.JWT = token
}

func (u *InventoryManagementV2Suite) TestGetInventoryStatistics() {
	var (
		require = require.New(u.T())
		assert  = assert.New(u.T())
		ctx     = common.JWTAuthContext(u.T().Context(), u.JWT)
	)

	var macs []string
	for range 4 {
		mac, err := u.randomMAC()
		require.NoError(err, "failed to generate random mac")
		macs = append(macs, mac.String())
	}

	// Create auth requests for all macs
	for _, m := range macs {
		_, _, err := u.authRequest(ctx, u.Tenant.TenantToken, m)
		require.NoError(err)
	}

	// Accept half of them
	for _, m := range macs[:len(macs)/2] {
		_, err := u.acceptWait(ctx, m)
		require.NoError(err)
	}

	res, _, err := u.APIClient.
		DeviceInventoryFiltersAndSearchManagementAPIAPI.
		GetStatistics(ctx).
		Execute()
	require.NoError(err)

	statistics := res.GetDevicesByStatus()
	assert.Equal(int32(len(macs)/2), statistics.Accepted.Standard)
	assert.Equal(int32(0), statistics.Accepted.Micro)
	assert.Equal(int32(0), statistics.Accepted.System)

	assert.Equal(int32(len(macs)/2), statistics.Pending.Standard)
	assert.Equal(int32(0), statistics.Pending.Micro)
	assert.Equal(int32(0), statistics.Pending.System)
}

func (d *InventoryManagementV2Suite) authRequest(ctx context.Context, tenantToken *string, mac string) (string, bool, error) {
	privateKey, publicKey, err := d.generateKeys()
	if err != nil {
		return "", false, errors.Wrap(err, "failed to generate key-pair")
	}

	idData, err := json.Marshal(map[string]string{"mac": mac})
	if err != nil {
		return "", false, errors.Wrap(err, "failed to marshal id data")
	}

	authRequest := client.AuthRequest{
		IdData:      string(idData),
		TenantToken: tenantToken,
		Pubkey:      d.exportPublicKeyPEM(publicKey),
	}

	authRequestData, err := json.Marshal(authRequest)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to marshal auth request")
	}
	authRequestData = append(authRequestData, '\n')

	signature, err := d.signData(privateKey, authRequestData)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to sign request data")
	}

	token, r, err := d.APIClient.DeviceAuthenticationDeviceAPIAPI.DeviceAuthAuthenticateDevice(ctx).
		XMENSignature(signature).
		AuthRequest(authRequest).
		Execute()
	if err != nil && err.Error() != "401 Unauthorized" {
		return "", false, errors.Wrap(err, "failed to send auth request")
	}

	require.NotNil(d.T(), r)
	require.Contains(d.T(), []int{http.StatusOK, http.StatusUnauthorized}, r.StatusCode)

	return token, token != "", nil
}

func (d *InventoryManagementV2Suite) acceptWait(ctx context.Context, mac string) (*client.Device, error) {
	var (
		inventorym = d.APIClient.DeviceInventoryFiltersAndSearchManagementAPIAPI
		devauthm   = d.APIClient.DeviceAuthenticationManagementAPIAPI
	)

	getDeviceInventory := func() (client.DeviceInventory, error) {
		for range 10 {
			filter := []client.FilterPredicate{
				{
					Scope:     client.IDENTITY,
					Attribute: "mac",
					Type:      "$eq",
					Value:     client.StringAsFilterPredicateValue(&mac),
				},
			}

			devices, _, err := inventorym.
				InventoryV2SearchDeviceInventories(ctx).
				InventoryV2SearchDeviceInventoriesRequest(
					client.InventoryV2SearchDeviceInventoriesRequest{Filters: filter}).
				Execute()

			if err != nil {
				return client.DeviceInventory{}, errors.Wrap(err, "failed to get device inventory")
			}

			if len(devices) > 0 {
				return devices[0], nil
			}
			time.Sleep(500 * time.Millisecond)
		}

		return client.DeviceInventory{}, errors.New("failed to get device inventory (no results)")
	}

	deviceInventory, err := getDeviceInventory()
	if err != nil {
		return nil, err
	}

	device, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, deviceInventory.GetId()).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get device from device auth")
	}

	if len(device.AuthSets) < 1 {
		return nil, errors.New("no authsets found for device")
	}

	_, err = devauthm.DeviceAuthManagementSetAuthenticationStatus(
		ctx,
		device.GetId(),
		device.AuthSets[0].GetId()).
		Status(client.Status{Status: model.DevStatusAccepted}).
		Execute()

	if err != nil {
		return nil, errors.Wrap(err, "failed to accept a device")
	}

	maxIterations := 8
	for maxIterations > 0 {
		deviceInventory, err := getDeviceInventory()
		if err != nil {
			return nil, err
		}

		accepted := slices.ContainsFunc(
			deviceInventory.Attributes,
			func(a client.AttributeV2) bool {
				if a.GetScope() != client.IDENTITY || a.GetName() != "status" {
					return false
				}
				return a.GetValue().String != nil &&
					*a.GetValue().String == modelinventory.DeviceStatusAccepted
			},
		)

		if accepted {
			return device, nil
		}

		time.Sleep(500 * time.Millisecond)
		maxIterations--
	}

	return nil, fmt.Errorf("device with mac %s was not accepted in time", mac)
}

func (u *InventoryManagementV2Suite) generateKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func (u *InventoryManagementV2Suite) signData(privateKey *rsa.PrivateKey, data []byte) (string, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (u *InventoryManagementV2Suite) exportPublicKeyPEM(pubkey *rsa.PublicKey) string {
	pubASN1, _ := x509.MarshalPKIXPublicKey(pubkey)
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
	return string(pubBytes)
}

func (u *InventoryManagementV2Suite) randomMAC() (net.HardwareAddr, error) {
	mac := make([]byte, 6)
	_, err := rand.Read(mac)
	if err != nil {
		return nil, err
	}

	mac[0] &= 0xfe // Set to Unicast
	mac[0] |= 0x02 // Set to Locally Administered

	return mac, nil
}
