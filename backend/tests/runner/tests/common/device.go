//nolint:all // This is all test code
package common

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/utils/types"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	modelinventory "github.com/mendersoftware/mender-server/services/inventory/model"
)

// Device is a test device identity: a key-pair plus id-data, with the
// deviceauth id and device JWT filled in as it goes through onboarding.
// PrivateKey/PublicKey may hold any of the key types deviceauth accepts
// (RSA, ECDSA, Ed25519); see KeyPair for a constructor covering all of them.
type Device struct {
	PrivateKey crypto.Signer
	PublicKey  any
	MAC        string
	IDData     string

	// ID is the deviceauth device id, set by Accept.
	ID string
	// Token is the device JWT, set by SubmitAuthRequest once accepted.
	Token string
}

// NewDeviceFromKeyPair builds a Device identity from an already-generated
// key-pair and identity data, without a MAC (callers that need WaitInventory
// to find the device by its MAC should set d.MAC themselves).
func NewDeviceFromKeyPair(kp *KeyPair, idData string) *Device {
	return &Device{
		PrivateKey: kp.Private,
		PublicKey:  kp.Public,
		IDData:     idData,
	}
}

// NewDevice creates a device identity with fresh keys and a random MAC.
func NewDevice() (*Device, error) {
	kp, err := NewKeyPair(KeyKindRSA)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key-pair: %w", err)
	}

	mac, err := RandomMAC()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random mac: %w", err)
	}

	idData, err := json.Marshal(map[string]string{"mac": mac.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal id data: %w", err)
	}

	device := NewDeviceFromKeyPair(kp, string(idData))
	device.MAC = mac.String()
	return device, nil
}

// SubmitAuthRequest signs and submits an auth request. It returns true
// with d.Token set when deviceauth accepts the request (the device was
// accepted before), false on 401 (pending), and an error on anything
// else.
func (d *Device) SubmitAuthRequest(
	ctx context.Context, api *client.APIClient, tenantToken *string,
) (bool, error) {
	authRequest := client.AuthRequest{
		IdData:      d.IDData,
		TenantToken: tenantToken,
		Pubkey:      ExportPublicKeyPEM(d.PublicKey),
	}

	authRequestData, err := json.Marshal(authRequest)
	if err != nil {
		return false, fmt.Errorf("failed to marshal auth request: %w", err)
	}
	// The signature is verified over authRequestData including this
	// trailing newline. This only works because the generated client's
	// own JSON encoder also appends a trailing newline when it serializes
	// the same AuthRequest body for the actual HTTP request -- so the
	// bytes we sign here end up matching the bytes deviceauth receives.
	// If the client's encoding ever changes, this signature would stop
	// verifying.
	authRequestData = append(authRequestData, '\n')

	signature, err := SignAuthRequest(d.PrivateKey, authRequestData)
	if err != nil {
		return false, fmt.Errorf("failed to sign request data: %w", err)
	}

	token, r, err := api.DeviceAuthenticationDeviceAPIAPI.
		DeviceAuthAuthenticateDevice(ctx).
		XMENSignature(signature).
		AuthRequest(authRequest).
		Execute()
	if r == nil {
		return false, fmt.Errorf("no response from auth request: %w", err)
	}
	switch r.StatusCode {
	case http.StatusOK:
		d.Token = token
		return true, nil
	case http.StatusUnauthorized:
		return false, nil
	default:
		return false, fmt.Errorf(
			"unexpected auth request status %d: %w", r.StatusCode, err,
		)
	}
}

// CheckUpdate asks deployments for the device's next update as the device
// itself, authenticated with its JWT, reporting artifactName and deviceType
// as the currently installed artifact and device type. It returns the raw
// response so callers can assert on the status code (200 update available,
// 204 nothing to deploy).
func (d *Device) CheckUpdate(
	ctx context.Context, api *client.APIClient, artifactName, deviceType string,
) (*client.DeploymentInstructions, *http.Response, error) {
	return api.DeploymentsDeviceAPIAPI.
		CheckUpdate(JWTAuthContext(ctx, d.Token)).
		ArtifactName(artifactName).
		DeviceType(deviceType).
		Execute()
}

// waitInventoryTimeout/waitInventoryPeriod: the deviceauth->inventory
// provisioning workflow is async and this wait runs for every device the
// whole suite creates, so the budget errs on the generous side for loaded
// CI runners.
const (
	waitInventoryTimeout = 15 * time.Second
	waitInventoryPeriod  = 500 * time.Millisecond

	// V3: named for symmetry with waitInventoryTimeout/waitInventoryPeriod
	// above -- see the budget comment on Accept's status poll.
	acceptStatusPollTimeout = 24 * 500 * time.Millisecond
	acceptStatusPollPeriod  = 500 * time.Millisecond
)

// WaitInventory polls the inventory v2 search until the device shows up
// by its identity MAC.
func (d *Device) WaitInventory(
	ctx context.Context, api *client.APIClient,
) (client.DeviceInventoryResponse, error) {
	inventorym := api.DeviceInventoryFiltersAndSearchManagementAPIAPI

	var result client.DeviceInventoryResponse
	err := RetryUntil(ctx, waitInventoryTimeout, waitInventoryPeriod, func() (bool, error) {
		filter := []client.FilterPredicate{
			{
				Scope:     client.IDENTITY,
				Attribute: "mac",
				Type:      "$eq",
				Value: client.AttributeValueRequest{
					String: types.Pointer(d.MAC),
				},
			},
		}

		devices, _, err := inventorym.
			InventoryV2SearchDeviceInventories(ctx).
			SearchParams(client.SearchParams{Filters: filter}).
			Execute()
		if err != nil {
			return false, fmt.Errorf("failed to get device inventory: %w", err)
		}

		if len(devices) > 0 {
			result = devices[0]
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return client.DeviceInventoryResponse{}, err
	}
	return result, nil
}

// Accept accepts the device's first authset via the management API and
// waits until inventory reflects the accepted status. It sets d.ID.
func (d *Device) Accept(ctx context.Context, api *client.APIClient) error {
	devauthm := api.DeviceAuthenticationManagementAPIAPI

	deviceInventory, err := d.WaitInventory(ctx, api)
	if err != nil {
		return err
	}

	device, _, err := devauthm.
		DeviceAuthManagementGetDevice(ctx, deviceInventory.GetId()).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to get device from device auth: %w", err)
	}

	if len(device.AuthSets) < 1 {
		return errors.New("no authsets found for device")
	}
	d.ID = device.GetId()

	_, err = devauthm.DeviceAuthManagementSetAuthenticationStatus(
		ctx,
		device.GetId(),
		device.AuthSets[0].GetId()).
		Status(client.Status{Status: model.DevStatusAccepted}).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to accept a device: %w", err)
	}

	// WaitInventory above already located the device once (and set d.ID).
	// Rather than re-running that whole MAC search on every iteration --
	// which previously nested a 30 x 500ms search inside a 24-iteration
	// countdown, an up to ~6 minute worst case (24 x 15s) -- poll the
	// already-known device's "status" attribute directly. The budget
	// matches the previous countdown loop's (24 x 500ms = 12s).
	err = RetryUntil(ctx, acceptStatusPollTimeout, acceptStatusPollPeriod, func() (bool, error) {
		inv, _, err := api.DeviceInventoryManagementAPIAPI.
			GetDeviceInventory(ctx, d.ID).Execute()
		if err != nil {
			return false, nil
		}

		accepted := slices.ContainsFunc(
			inv.Attributes,
			func(a client.AttributeResponse) bool {
				if a.GetScope() != client.IDENTITY || a.GetName() != "status" {
					return false
				}
				return a.GetValue().String != nil &&
					*a.GetValue().String == modelinventory.DeviceStatusAccepted
			},
		)
		return accepted, nil
	})
	if err != nil {
		return fmt.Errorf("device with mac %s was not accepted in time: %w", d.MAC, err)
	}
	return nil
}

// NewAcceptedDevice creates a device, submits its auth request and
// accepts it — the full onboarding most tests need.
func NewAcceptedDevice(
	ctx context.Context, api *client.APIClient, tenantToken *string,
) (*Device, error) {
	device, err := NewDevice()
	if err != nil {
		return nil, err
	}

	if _, err := device.SubmitAuthRequest(ctx, api, tenantToken); err != nil {
		return nil, err
	}

	if err := device.Accept(ctx, api); err != nil {
		return nil, err
	}

	// The first SubmitAuthRequest above ran before the device was
	// accepted, so it got a 401 and no token. Submit again now that the
	// device is accepted, so d.Token is actually populated.
	if _, err := device.SubmitAuthRequest(ctx, api, tenantToken); err != nil {
		return nil, err
	}

	return device, nil
}
