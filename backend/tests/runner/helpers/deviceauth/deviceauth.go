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

package deviceauth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	openapi "github.com/mendersoftware/mender-server/pkg/api/client"
	testcrypto "github.com/mendersoftware/mender-server/tests/runner/helpers/crypto"
)

// Device represents a test device.
type Device struct {
	ID           string
	IdentityData map[string]string
	PubKeyPEM    string
	PrivKeyPEM   string
	Status       string
	Token        string
	AuthsetID    string
}

// MakePendingDevice creates a device with pending status via the device auth API
// using the generated OpenAPI client.
func MakePendingDevice(client *openapi.APIClient, utoken string) (*Device, error) {
	privPEM, pubPEM, err := testcrypto.GenerateRSAKeypair()
	if err != nil {
		return nil, fmt.Errorf("generating keypair: %w", err)
	}

	mac := make([]byte, 2)
	if _, err := rand.Read(mac); err != nil {
		return nil, err
	}
	idData := map[string]string{
		"mac": fmt.Sprintf("de:ad:be:ef:%02x:%02x", mac[0], mac[1]),
	}

	idDataJSON, err := json.Marshal(idData)
	if err != nil {
		return nil, err
	}

	authReqBody, sig, err := buildSignedAuthRequest(idData, pubPEM, privPEM, "")
	if err != nil {
		return nil, err
	}

	// The generated client's AuthRequest model uses typed fields,
	// but we need to sign the exact JSON body that gets sent.
	// The signature must match the wire format, so we build the body
	// ourselves and pass it via the model fields.
	authReq := *openapi.NewAuthRequest(string(idDataJSON), pubPEM)

	// We need to sign the serialized auth request body. The generated
	// client serializes internally, but the signature must cover the
	// exact bytes. We pre-compute the signature from our own serialization
	// and trust that the field values match.
	_ = authReqBody // signature was computed from this

	_, resp, err := client.DeviceAuthenticationDeviceAPIAPI.
		DeviceAuthAuthenticateDevice(context.Background()).
		XMENSignature(sig).
		AuthRequest(authReq).
		Execute()
	if err != nil {
		// A 401 response is expected for a new pending device;
		// the generated client treats non-2xx as errors.
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			// Expected: device is pending
		} else {
			return nil, fmt.Errorf("auth request failed: %w", err)
		}
	}
	if resp != nil {
		resp.Body.Close()
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return nil, fmt.Errorf("expected 401 for pending device, got %d", resp.StatusCode)
	}

	devID, authsetID, err := findDeviceByPubKey(client, utoken, pubPEM)
	if err != nil {
		return nil, fmt.Errorf("finding device: %w", err)
	}

	return &Device{
		ID:           devID,
		IdentityData: idData,
		PubKeyPEM:    pubPEM,
		PrivKeyPEM:   privPEM,
		Status:       "pending",
		AuthsetID:    authsetID,
	}, nil
}

// MakeAcceptedDevice creates a device and accepts its auth set.
func MakeAcceptedDevice(client *openapi.APIClient, utoken string) (*Device, error) {
	dev, err := MakePendingDevice(client, utoken)
	if err != nil {
		return nil, err
	}

	err = changeAuthsetStatus(client, utoken, dev.ID, dev.AuthsetID, "accepted")
	if err != nil {
		return nil, fmt.Errorf("accepting authset: %w", err)
	}

	// Re-submit auth request to obtain device token
	idDataJSON, _ := json.Marshal(dev.IdentityData)
	_, sig, err := buildSignedAuthRequest(
		dev.IdentityData,
		dev.PubKeyPEM,
		dev.PrivKeyPEM,
		"",
	)
	if err != nil {
		return nil, err
	}

	authReq := *openapi.NewAuthRequest(string(idDataJSON), dev.PubKeyPEM)

	token, resp, err := client.DeviceAuthenticationDeviceAPIAPI.
		DeviceAuthAuthenticateDevice(context.Background()).
		XMENSignature(sig).
		AuthRequest(authReq).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"expected 200 for accepted device, got %d",
			resp.StatusCode,
		)
	}

	dev.Token = token
	dev.Status = "accepted"
	return dev, nil
}

func buildSignedAuthRequest(
	idData map[string]string,
	pubPEM, privPEM, tenantToken string,
) (body []byte, sig string, err error) {
	idDataJSON, err := json.Marshal(idData)
	if err != nil {
		return nil, "", err
	}

	authReq := map[string]string{
		"id_data": string(idDataJSON),
		"pubkey":  pubPEM,
	}
	if tenantToken != "" {
		authReq["tenant_token"] = tenantToken
	}

	body, err = json.Marshal(authReq)
	if err != nil {
		return nil, "", err
	}

	sig, err = testcrypto.SignAuthRequest(body, privPEM)
	if err != nil {
		return nil, "", err
	}
	return body, sig, nil
}

func findDeviceByPubKey(
	client *openapi.APIClient,
	utoken, pubKeyPEM string,
) (devID, authsetID string, err error) {
	ctx := context.WithValue(context.Background(), openapi.ContextAccessToken, utoken)

	devices, resp, err := client.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementListDevices(ctx).
		Execute()
	if err != nil {
		return "", "", fmt.Errorf("list devices request failed: %w", err)
	}
	defer resp.Body.Close()

	for _, d := range devices {
		for _, a := range d.AuthSets {
			if a.Pubkey != nil && *a.Pubkey == pubKeyPEM {
				if d.Id == nil || a.Id == nil {
					return "", "", fmt.Errorf("device or authset ID is nil")
				}
				return *d.Id, *a.Id, nil
			}
		}
	}
	return "", "", fmt.Errorf("device with given public key not found")
}

func changeAuthsetStatus(
	client *openapi.APIClient,
	utoken, devID, authsetID, status string,
) error {
	ctx := context.WithValue(context.Background(), openapi.ContextAccessToken, utoken)

	statusModel := *openapi.NewStatus(status)

	resp, err := client.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementSetAuthenticationStatus(ctx, devID, authsetID).
		Status(statusModel).
		Execute()
	if err != nil {
		return fmt.Errorf("change authset status failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(
			"change authset status failed with status %d",
			resp.StatusCode,
		)
	}
	return nil
}
