//nolint:all // This is all test code
package common

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

const (
	waitInventoryTimeout = 15 * time.Second
	waitInventoryPeriod  = 500 * time.Millisecond

	acceptStatusPollTimeout = 24 * 500 * time.Millisecond
	acceptStatusPollPeriod  = 500 * time.Millisecond

	attributeNameMAC    = "mac"
	attributeNameStatus = "status"
)

const (
	DefaultDeviceType string = "qemux86-64"
)

type Device struct {
	api *client.APIClient

	keys               *KeyPair
	tenantToken        string
	identityAttributes map[string]any

	// ID is the device id fetched from the server, set after first authset is submitted.
	ID string
	// Token is the device JWT, set every time an authset is accepted.
	Token string
	// MAC is a randomly generated MAC-address, set at creation and automatically added to IdentityAttributes.
	MAC string
}

func (d *Device) Keys() *KeyPair {
	return d.keys
}

type DeviceOption func(d *Device)

// WithKeys configures the keys (private and public) that the device should
// use instead of generating its own.
func WithKeys(k *KeyPair) DeviceOption {
	return func(d *Device) {
		d.keys = k
	}
}

// WithMAC configures the mac address the device should use instead of
// generating its own. It will also overwrite the "mac" identity attribute
// sent with auth requests.
func WithMAC(m string) DeviceOption {
	return func(d *Device) {
		d.MAC = m
		d.identityAttributes[attributeNameMAC] = d.MAC
	}
}

// WithIdentityAttributes configures identity attributes
// should be sent with auth requests in addition to the default
// "mac" identity attribute.
// If the passed attributes contains a "mac" attribute with a string
// value, this will overwrite the default "mac" attribute and also the
// `MAC` property of the device.
func WithIdentityAttributes(i map[string]any) DeviceOption {
	return func(d *Device) {
		for name, value := range i {
			d.identityAttributes[name] = value
			// Handle MAC being reconfigured
			if mac, ok := value.(string); ok && name == attributeNameMAC {
				d.MAC = mac
			}
		}
	}
}

// NewDevice creates a device with generated RSA keys, a random MAC and the default device type
func NewDevice(api *client.APIClient, tenantToken string, opts ...DeviceOption) (*Device, error) {
	device := &Device{
		api:                api,
		tenantToken:        tenantToken,
		identityAttributes: make(map[string]any),
	}

	for _, o := range opts {
		o(device)
	}

	if device.MAC == "" {
		// WithMAC option was not used, generate MAC
		mac, err := RandomMAC()
		if err != nil {
			return nil, fmt.Errorf("failed to generate random mac: %w", err)
		}
		device.identityAttributes[attributeNameMAC] = mac.String()
		device.MAC = mac.String()
	}

	if device.keys == nil {
		// WithKeys option was not passed, generate keys
		keys, err := NewKeyPair(KeyKindRSA)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key-pair: %w", err)
		}
		device.keys = keys
	}

	return device, nil
}

// NewAcceptedDevice creates a device with fresh RSA keys, a random MAC and the default device type
// before authenticating it with the server by sending and accepting an auth set.
// Supported option types are `DeviceOption` and `AuthRequestOption` - other types are ignored.
func NewAcceptedDevice(
	ctx context.Context,
	api *client.APIClient,
	tenantToken string,
	opts ...any,
) (*Device, error) {
	var (
		deviceOptions  []DeviceOption
		authReqOptions []AuthRequestOption
	)

	for _, o := range opts {
		switch oo := o.(type) {
		case DeviceOption:
			deviceOptions = append(deviceOptions, oo)
		case AuthRequestOption:
			authReqOptions = append(authReqOptions, oo)
		}
	}

	device, err := NewDevice(api, tenantToken, deviceOptions...)
	if err != nil {
		return nil, err
	}

	authorized, err := device.SubmitAuthRequest(ctx, authReqOptions...)
	if err != nil {
		return nil, err
	}

	if authorized {
		return nil, fmt.Errorf("invalid precondition encountered, new device is already authorized (accepted)")
	}

	if err := device.AcceptFirst(ctx); err != nil {
		return nil, err
	}

	// The first SubmitAuthRequest above ran before the device was
	// accepted, so it got a 401 and no token. Submit again now that the
	// device is accepted, so d.Token is actually populated.
	if _, err := device.SubmitAuthRequest(ctx, authReqOptions...); err != nil {
		return nil, err
	}

	return device, nil
}

type AuthRequestOption func(a *client.AuthRequest)

// SubmitAuthRequest signs and submits an auth request. It returns true (authorized)
// when deviceauth accepts the request (which means the device was
// accepted before), false (unauthorized) on 401 (the device is pending and must be accepted),
// and an error on anything else.
func (d *Device) SubmitAuthRequest(
	ctx context.Context, opts ...AuthRequestOption,
) (bool, error) {
	idData, err := json.Marshal(d.identityAttributes)
	if err != nil {
		return false, fmt.Errorf("failed to marshal id data: %w", err)
	}

	authRequest := client.AuthRequest{
		IdData: string(idData),
		Pubkey: d.keys.ExportPublicKeyPEM(),
	}

	if d.tenantToken != "" {
		authRequest.TenantToken = client.PtrString(d.tenantToken)
	}

	for _, o := range opts {
		o(&authRequest)
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
	signature, err := d.keys.Sign(authRequestData)
	if err != nil {
		return false, fmt.Errorf("failed to sign request data: %w", err)
	}

	token, r, err := d.api.DeviceAuthenticationDeviceAPIAPI.
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
	case http.StatusUnauthorized:
		// Pending
	default:
		return false, fmt.Errorf(
			"unexpected auth request status %d: %w", r.StatusCode, err,
		)
	}

	if d.ID == "" {
		// If this is the first authset, poll inventory for device ID
		err = RetryUntil(ctx, waitInventoryTimeout, waitInventoryPeriod, func() (bool, error) {
			filter := []client.FilterPredicate{
				{
					Scope:     client.IDENTITY,
					Attribute: attributeNameMAC,
					Type:      "$eq",
					Value:     client.AttributeValueRequest{String: client.PtrString(d.MAC)},
				},
			}

			devices, _, err := d.api.
				DeviceInventoryFiltersAndSearchManagementAPIAPI.
				InventoryV2SearchDeviceInventories(ctx).
				SearchParams(client.SearchParams{Filters: filter}).
				Execute()
			if err != nil {
				return false, fmt.Errorf("failed to get device inventory: %w", err)
			}

			if len(devices) > 0 {
				d.ID = devices[0].GetId()
				return true, nil
			}
			return false, nil
		})

		if err != nil {
			return false, fmt.Errorf("device with mac %s did not register with inventory in time: %w", d.MAC, err)
		}
	}

	return r.StatusCode == http.StatusOK, nil
}

// AcceptFirst accepts the device's first authset via the management API and
// waits until inventory reflects the accepted status.
// The provided context must contain the JWT of a user with access to fetch the device in
// the management API.
func (d *Device) AcceptFirst(ctx context.Context) error {
	device, err := d.GetServerDevice(ctx)
	if err != nil {
		return err
	}

	if len(device.AuthSets) < 1 {
		return errors.New("no authsets found for device")
	}

	return d.Accept(ctx, device.AuthSets[0])
}

// AcceptNewest accepts the newest of the device's authsets via the management API and
// waits until inventory reflects the accepted status.
// The provided context must contain the JWT of a user with access to the device in
// the management API.
func (d *Device) AcceptNewest(ctx context.Context) error {
	device, err := d.GetServerDevice(ctx)
	if err != nil {
		return err
	}

	var current client.AuthSet
	for _, authSet := range device.AuthSets {
		if authSet.GetStatus() == model.DevStatusPending && authSet.GetTs().After(current.GetTs()) {
			current = authSet
		}
	}

	if current.GetId() == "" {
		return errors.New("no pending authsets found for device")
	}

	return d.Accept(ctx, current)
}

// Accept accepts the provided authset. The authset must belong to the device and be pending.
// A list of the devices authsets can be retrieved by calling `GetServerDevice`.
// The provided context must contain the JWT of a user with access to the device in
// the management API.
func (d *Device) Accept(ctx context.Context, authSet client.AuthSet) error {
	if authSet.GetStatus() != model.DevStatusPending {
		return fmt.Errorf("failed to accept device: expected authset to be '%s' but was '%s'", model.DevStatusPending, authSet.GetStatus())
	}

	_, err := d.api.
		DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementSetAuthenticationStatus(ctx, d.ID, authSet.GetId()). // The server will enforce that d.ID owns authset.GetId()
		Status(client.Status{Status: model.DevStatusAccepted}).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to accept a device: %w", err)
	}

	err = RetryUntil(ctx, acceptStatusPollTimeout, acceptStatusPollPeriod, func() (bool, error) {
		inv, _, err := d.api.
			DeviceInventoryManagementAPIAPI.
			GetDeviceInventory(ctx, d.ID).
			Execute()
		if err != nil {
			return false, nil
		}

		accepted := slices.ContainsFunc(
			inv.Attributes,
			func(a client.AttributeResponse) bool {
				if a.GetScope() != client.IDENTITY || a.GetName() != attributeNameStatus {
					return false
				}
				return a.GetValue().String != nil &&
					*a.GetValue().String == model.DevStatusAccepted
			},
		)
		return accepted, nil
	})

	if err != nil {
		return fmt.Errorf("device with mac %s was not accepted in time: %w", d.MAC, err)
	}

	return nil
}

// GetServerDevice fetches the server side representation of the device from the deviceauth service.
// The provided context must contain the JWT of a user with access to the device in
// the management API.
func (d *Device) GetServerDevice(
	ctx context.Context,
) (*client.Device, error) {
	if d.ID == "" {
		return nil, fmt.Errorf("device has no id set")
	}

	device, _, err := d.api.
		DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementGetDevice(ctx, d.ID).
		Execute()
	return device, err
}

func RandomMAC() (net.HardwareAddr, error) {
	mac := make([]byte, 6)
	_, err := rand.Read(mac)
	if err != nil {
		return nil, err
	}

	mac[0] &= 0xfe // Set to Unicast
	mac[0] |= 0x02 // Set to Locally Administered

	return mac, nil
}

const (
	KeyKindRSA     = "rsa"
	KeyKindECP224  = "ec-p224"
	KeyKindECP256  = "ec-p256"
	KeyKindECP384  = "ec-p384"
	KeyKindECP521  = "ec-p521"
	KeyKindEd25519 = "ed25519"
)

// KeyPair is a signing key-pair for a device, abstracting over the key
// types deviceauth accepts (RSA, ECDSA P-224/256/384/521, Ed25519).
type KeyPair struct {
	Kind       string
	privateKey crypto.Signer
	publicKey  any
}

func (k *KeyPair) Sign(data []byte) (string, error) {
	switch privateKey := k.privateKey.(type) {
	case *rsa.PrivateKey:
		hash := sha256.Sum256(data)
		signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(signature), nil

	case *ecdsa.PrivateKey:
		hash := sha256.Sum256(data)
		signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(signature), nil

	case ed25519.PrivateKey:
		signature := ed25519.Sign(privateKey, data)
		return base64.StdEncoding.EncodeToString(signature), nil

	default:
		return "", fmt.Errorf("unsupported private key type %T", privateKey)
	}
}

// ExportPublicKeyPEM renders a public key (RSA, ECDSA or Ed25519) as the
// PKIX PEM string used in device auth requests. Panics if the key cannot be
// marshaled.
func (k *KeyPair) ExportPublicKeyPEM() string {
	pubASN1, err := x509.MarshalPKIXPublicKey(k.publicKey)
	if err != nil {
		panic(fmt.Errorf("failed to marshal public key: %w", err))
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
	return string(pubBytes)
}

// NewKeyPair generates a key-pair of the given kind
func NewKeyPair(kind string) (*KeyPair, error) {
	switch kind {
	case KeyKindRSA:
		privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			return nil, err
		}
		return &KeyPair{Kind: kind, privateKey: privateKey, publicKey: &privateKey.PublicKey}, nil

	case KeyKindECP224, KeyKindECP256, KeyKindECP384, KeyKindECP521:
		curves := map[string]elliptic.Curve{
			KeyKindECP224: elliptic.P224(),
			KeyKindECP256: elliptic.P256(),
			KeyKindECP384: elliptic.P384(),
			KeyKindECP521: elliptic.P521(),
		}

		privateKey, err := ecdsa.GenerateKey(curves[kind], rand.Reader)
		if err != nil {
			return nil, err
		}
		return &KeyPair{Kind: kind, privateKey: privateKey, publicKey: privateKey.PublicKey}, nil

	case KeyKindEd25519:
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		return &KeyPair{Kind: kind, privateKey: privateKey, publicKey: publicKey}, nil

	default:
		return nil, fmt.Errorf("unsupported key kind %q", kind)
	}
}
