// Copyright 2022 Northern.tech AS
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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/services/iot-manager/app"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iothub"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	mocks_store "github.com/mendersoftware/mender-server/services/iot-manager/store/mocks"
)

const (
	APIURLDeviceTwin = "/devices/:id/twin"

	HdrKeyAuthz = "Authorization"
)

var (
	externalCS       *model.ConnectionString
	externalDeviceID string
)

func parseConnString(connection string) error {
	var err error
	externalCS, err = model.ParseConnectionString(connection)
	return err
}

func init() {
	flag.Func("test.connection-string",
		"Connection string for external iothub "+
			"(overwrite with env var TEST_CONNECTION_STRING).",
		parseConnString)
	flag.StringVar(&externalDeviceID,
		"test.device-id",
		"",
		"The id of a device on the iothub pointed to by connection-string"+
			" (overwrite with env TEST_DEVICE_ID).")
	cStr, ok := os.LookupEnv("TEST_CONNECTION_STRING")
	if ok {
		externalCS, _ = model.ParseConnectionString(cStr)
	}
	idStr, ok := os.LookupEnv("TEST_DEVICE_ID")
	if ok {
		externalDeviceID = idStr
	}

	testing.Init()
}

// TestIOTHubExternal runs against a real IoT Hub using the provided command line
// arguments / environment variables. The test updates fields in the device's
// desired state, so it's important that the hub-device is not used by a real
// device.
func TestIOTHubExternal(t *testing.T) {
	if externalCS == nil {
		t.Skip("test.connection-string is not provided or valid")
		return
	} else if externalDeviceID == "" {
		t.Skip("test.device-id is not provided nor valid")
		return
	}

	// The following gets the device and updates (increments)
	// the "desired" property "_TESTING" and checks the expected
	// value.
	authz := "Bearer " + GenerateJWT(identity.Identity{
		Subject: "7e57dc61-cd13-4d8a-beee-3cfa885c9cae",
		IsUser:  true,
	})
	w := httptest.NewRecorder()
	const testKey = "_TESTING"
	integrationId := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))

	store := &mocks_store.DataStore{}
	store.On("GetDeviceByIntegrationID",
		contextMatcher,
		externalDeviceID,
		integrationId,
	).Return(&model.Device{}, nil)
	store.On("GetIntegrationById",
		contextMatcher,
		integrationId,
	).Return(&model.Integration{
		Provider: model.ProviderIoTHub,
		Credentials: model.Credentials{
			ConnectionString: externalCS,
		},
	}, nil)
	defer store.AssertExpectations(t)

	iotHubClient := iothub.NewClient()
	app := app.New(store, nil, nil).WithIoTHub(iotHubClient)

	handler := NewRouter(app)
	srv := httptest.NewServer(handler)
	client := srv.Client()
	client.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, srv.Listener.Addr().String())
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	url := strings.ReplaceAll(APIURLDeviceStateIntegration, ":id", externalDeviceID)
	url = strings.ReplaceAll(url, ":integrationId", integrationId.String())
	uri := "http://localhost" + APIURLManagement + url
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header.Set(HdrKeyAuthz, authz)
	rsp, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	defer rsp.Body.Close()

	var deviceState model.DeviceState
	dec := json.NewDecoder(rsp.Body)
	err = dec.Decode(&deviceState)
	require.NoError(t, err)
	var nextValue uint32
	if cur, ok := deviceState.Desired[testKey].(float64); ok {
		nextValue = uint32(cur) + 1
	}
	deviceState.Desired[testKey] = nextValue
	b, _ := json.Marshal(deviceState)
	req, _ = http.NewRequestWithContext(ctx, http.MethodPut, uri, bytes.NewReader(b))
	req.Header.Set(HdrKeyAuthz, authz)
	require.NoError(t, err)
	rspPatch, err := client.Do(req)
	require.NoError(t, err)
	defer rspPatch.Body.Close()
	assert.Equal(t, http.StatusOK, rspPatch.StatusCode)

	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header.Set(HdrKeyAuthz, authz)
	rspGet, err := client.Do(req)
	require.NoError(t, err)
	defer rspGet.Body.Close()
	assert.Equal(t, http.StatusOK, rspGet.StatusCode)
	dec = json.NewDecoder(rspGet.Body)
	_ = dec.Decode(&deviceState)
	if updatedFloat, ok := deviceState.Desired[testKey].(float64); assert.True(t, ok) {
		assert.Equal(t, nextValue, uint32(updatedFloat))
	}
}
