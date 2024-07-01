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

package iothub

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	common "github.com/mendersoftware/mender-server/services/iot-manager/client"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func init() {
	model.SetTrustedHostnames([]string{"*.azure-devices.net", "localhost"})
}

var (
	externalCS       *model.ConnectionString
	externalDeviceID string
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (rt RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

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
	const testKey = "_TESTING"
	client := NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	mod, err := client.GetDeviceTwin(ctx, externalCS, externalDeviceID)
	if assert.NoError(t, err) {
		assert.Equal(t, externalDeviceID, mod.DeviceID)
	}
	if t.Failed() {
		t.FailNow()
	}

	var nextValue uint32
	if cur, ok := mod.Properties.Desired[testKey].(float64); ok {
		nextValue = uint32(cur) + 1
	}

	err = client.UpdateDeviceTwin(ctx, externalCS, externalDeviceID, &DeviceTwinUpdate{
		Properties: UpdateProperties{
			Desired: map[string]interface{}{
				testKey: nextValue,
			},
		},
	})

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	modUpdated, err := client.GetDeviceTwin(ctx, externalCS, externalDeviceID)
	if assert.NoError(t, err) {
		value, ok := modUpdated.Properties.Desired[testKey].(float64)
		if assert.True(t, ok, "Updated twin does not contain update value") {
			assert.Equal(t, nextValue, uint32(value), "property does not match update")
		}
	}

}

func TestUpsertDevice(t *testing.T) {
	t.Parallel()
	cs := &model.ConnectionString{
		HostName: "localhost",
		Key:      crypto.String("secret"),
		Name:     "gimmeAccessPls",
	}
	deviceID := "6c985f61-5093-45eb-8ece-7dfe97a6de7b"
	testCases := []struct {
		Name string

		Updates []*Device
		ConnStr *model.ConnectionString

		RSPCode int
		RSPBody interface{}

		RTError error

		Error error
	}{{
		Name: "ok",

		Updates: []*Device{{
			Auth: &Auth{
				Type: AuthTypeSymmetric,
				SymmetricKey: &SymmetricKey{
					Primary:   Key("foo"),
					Secondary: Key("bar"),
				},
			},
			ETag: "qwerty",
		}, nil},
		ConnStr: cs,
		RSPCode: http.StatusOK,
	}, {
		Name: "error/invalid connection string",

		ConnStr: &model.ConnectionString{
			Name: "bad",
		},
		Error: errors.New("failed to prepare request: invalid connection string"),
	}, {
		Name: "error/internal roundtrip error",

		ConnStr: cs,
		RTError: errors.New("idk"),
		Error:   errors.New("failed to execute request:.*idk"),
	}, {
		Name: "error/bad status code",

		ConnStr: cs,

		RSPCode: http.StatusInternalServerError,
		Error:   common.NewHTTPError(http.StatusInternalServerError),
	}, {
		Name: "error/malformed response",

		ConnStr: cs,

		RSPBody: []byte("imagine a device in this reponse pls"),

		RSPCode: http.StatusOK,
		Error:   errors.New("iothub: failed to decode updated device"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			w := httptest.NewRecorder()
			httpClient := &http.Client{
				Transport: RoundTripperFunc(func(
					r *http.Request,
				) (*http.Response, error) {
					if tc.RTError != nil {
						return nil, tc.RTError
					}
					w.WriteHeader(tc.RSPCode)
					switch typ := tc.RSPBody.(type) {
					case []byte:
						w.Write(typ)
					case nil:
						dev := mergeDevices(tc.Updates...)
						b, _ := json.Marshal(dev)
						w.Write(b)
					default:
						b, _ := json.Marshal(typ)
						w.Write(b)
					}

					return w.Result(), nil
				}),
			}
			client := NewClient(NewOptions(nil).
				SetClient(httpClient))

			dev, err := client.UpsertDevice(ctx, tc.ConnStr, deviceID, tc.Updates...)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				expected := mergeDevices(tc.Updates...)
				expected.DeviceID = deviceID
				assert.Equal(t, expected, dev)
			}

		})
	}
}

func TestDeleteDevice(t *testing.T) {
	t.Parallel()
	cs := &model.ConnectionString{
		HostName: "localhost",
		Key:      crypto.String("secret"),
		Name:     "gimmeAccessPls",
	}
	deviceID := "6c985f61-5093-45eb-8ece-7dfe97a6de7b"
	testCases := []struct {
		Name string

		ConnStr *model.ConnectionString

		RSPCode int
		RTError error

		Error error
	}{{
		Name: "ok",

		ConnStr: cs,
		RSPCode: http.StatusOK,
	}, {
		Name: "error/invalid connection string",

		ConnStr: &model.ConnectionString{
			Name: "bad",
		},
		Error: errors.New("failed to prepare request: invalid connection string"),
	}, {
		Name: "error/internal roundtrip error",

		ConnStr: cs,
		RTError: errors.New("idk"),
		Error:   errors.New("failed to execute request:.*idk"),
	}, {
		Name: "error/bad status code",

		ConnStr: cs,

		RSPCode: http.StatusInternalServerError,
		Error:   common.NewHTTPError(http.StatusInternalServerError),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			w := httptest.NewRecorder()
			httpClient := &http.Client{
				Transport: RoundTripperFunc(func(
					r *http.Request,
				) (*http.Response, error) {
					if tc.RTError != nil {
						return nil, tc.RTError
					}
					w.WriteHeader(tc.RSPCode)

					return w.Result(), nil
				}),
			}
			client := NewClient(NewOptions(nil).
				SetClient(httpClient))

			err := client.DeleteDevice(ctx, tc.ConnStr, deviceID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestGetDevice(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		DeviceID string
		ConnStr  *model.ConnectionString
		RSPCode  int
		RSPBody  interface{}

		RTError error
		Error   error
	}{{
		Name: "ok",

		DeviceID: "141c6d55-5d96-4b60-b00a-47cdb9a49aeb",
		ConnStr: &model.ConnectionString{
			HostName: "localhost",
			Name:     "swellHub",
			Key:      crypto.String("password123"),
		},
		RSPCode: http.StatusOK,
		RSPBody: &Device{
			Auth: &Auth{
				Type: AuthTypeSymmetric,
				SymmetricKey: &SymmetricKey{
					Primary:   Key("foobar"),
					Secondary: Key("barbaz"),
				},
			},
			DeviceID:     "141c6d55-5d96-4b60-b00a-47cdb9a49aeb",
			GenerationID: "such api",
			ETag:         "much fields",
			Status:       StatusEnabled,
		},
	}, {
		Name: "error, bad connection string",

		DeviceID: "141c6d55-5d96-4b60-b00a-47cdb9a49aeb",
		ConnStr: &model.ConnectionString{
			Name: "namelessHub",
			Key:  crypto.String("password123"),
		},
		RSPCode: http.StatusOK,
		RSPBody: &Device{
			Auth: &Auth{
				Type: AuthTypeSymmetric,
				SymmetricKey: &SymmetricKey{
					Primary:   Key("foobar"),
					Secondary: Key("barbaz"),
				},
			},
			DeviceID:     "141c6d55-5d96-4b60-b00a-47cdb9a49aeb",
			GenerationID: "such api",
			ETag:         "much fields",
			Status:       StatusEnabled,
		},
		Error: errors.New("iothub: failed to prepare request"),
	}, {
		Name: "error, roundtrip error",

		DeviceID: "141c6d55-5d96-4b60-b00a-47cdb9a49aeb",
		ConnStr: &model.ConnectionString{
			HostName: "localhost",
			Name:     "namelessHub",
			Key:      crypto.String("password123"),
		},
		RTError: errors.New("internal error"),
		Error:   errors.New("iothub: failed to execute request:.*internal error"),
	}, {
		Name: "error, bad status code",

		DeviceID: "141c6d55-5d96-4b60-b00a-47cdb9a49aeb",
		ConnStr: &model.ConnectionString{
			HostName: "localhost",
			Name:     "swellHub",
			Key:      crypto.String("password123"),
		},
		RSPCode: http.StatusInternalServerError,
		RSPBody: rest.Error{Err: "internal error"},
		Error:   common.NewHTTPError(http.StatusInternalServerError),
	}, {
		Name: "error, malformed response",

		DeviceID: "141c6d55-5d96-4b60-b00a-47cdb9a49aeb",
		ConnStr: &model.ConnectionString{
			HostName: "localhost",
			Name:     "swellHub",
			Key:      crypto.String("password123"),
		},
		RSPCode: http.StatusOK,
		RSPBody: []byte("here's your device..."),

		Error: errors.New("iothub: failed to decode device"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			w := httptest.NewRecorder()
			httpClient := &http.Client{
				Transport: RoundTripperFunc(func(
					r *http.Request,
				) (*http.Response, error) {
					if tc.RTError != nil {
						return nil, tc.RTError
					}
					assert.Equal(t, "/devices/"+tc.DeviceID, r.URL.Path)

					w.WriteHeader(tc.RSPCode)
					switch t := tc.RSPBody.(type) {
					case []byte:
						_, _ = w.Write(t)
					default:
						b, _ := json.Marshal(t)
						_, _ = w.Write(b)
					}

					return w.Result(), nil
				}),
			}
			client := NewClient(NewOptions().SetClient(httpClient))
			dev, err := client.GetDevice(ctx, tc.ConnStr, tc.DeviceID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"unexpected error message content",
					)
				}
			} else if assert.NoError(t, err) {
				res := new(Device)
				if assert.IsType(t, res, tc.RSPBody, "Bad test case") {
					assert.Equal(t, tc.RSPBody, dev)
				}
			}
		})
	}
}

func TestGetDeviceTwins(t *testing.T) {
	t.Parallel()
	cs := &model.ConnectionString{
		HostName: "localhost",
		Key:      crypto.String("secret"),
		Name:     "gibDevice",
	}
	testCases := []struct {
		Name string

		CTX       context.Context
		ConnStr   *model.ConnectionString
		DeviceIDs []string

		ClientError error
		RspCode     int
		RspBody     interface{}

		Error error
	}{{
		Name: "ok",

		CTX:     context.Background(),
		ConnStr: cs,
		DeviceIDs: []string{
			"d9b4b693-fb1c-44fa-9e84-30e6aa0ee0c5",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb4",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb3",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb2",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb1",
		},

		RspCode: http.StatusOK,
		RspBody: []DeviceTwin{{
			DeviceID: "d9b4b693-fb1c-44fa-9e84-30e6aa0ee0c5",
		}, {
			DeviceID: "53a57499-51a6-41ee-a1ed-7abb994e8bb4",
		}, {
			DeviceID: "53a57499-51a6-41ee-a1ed-7abb994e8bb3",
		}, {
			DeviceID: "53a57499-51a6-41ee-a1ed-7abb994e8bb2",
		}, {
			DeviceID: "53a57499-51a6-41ee-a1ed-7abb994e8bb1",
		}},
	}, {
		Name: "ok/no devices",

		CTX:       context.Background(),
		ConnStr:   cs,
		DeviceIDs: []string{},
		RspBody:   []DeviceTwin{},
	}, {
		Name: "error/nil context",
		DeviceIDs: []string{
			"d9b4b693-fb1c-44fa-9e84-30e6aa0ee0c5",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb4",
		},
		ConnStr: cs,

		Error: errors.New("iothub: failed to prepare request:.*nil Context"),
	}, {
		Name: "error/client error",

		CTX:     context.Background(),
		ConnStr: cs,
		DeviceIDs: []string{
			"d9b4b693-fb1c-44fa-9e84-30e6aa0ee0c5",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb4",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb3",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb2",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb1",
		},

		ClientError: errors.New("internal error"),
		Error:       errors.New("iothub: failed to fetch device twins:.*internal error"),
	}, {
		Name: "error/bad status",

		CTX:     context.Background(),
		ConnStr: cs,
		DeviceIDs: []string{
			"d9b4b693-fb1c-44fa-9e84-30e6aa0ee0c5",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb4",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb3",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb2",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb1",
		},

		RspCode: http.StatusInternalServerError,
		RspBody: rest.Error{
			Err: "internal error",
		},

		Error: common.NewHTTPError(http.StatusInternalServerError),
	}, {
		Name: "error/corrupted response body",

		CTX:     context.Background(),
		ConnStr: cs,
		DeviceIDs: []string{
			"d9b4b693-fb1c-44fa-9e84-30e6aa0ee0c5",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb4",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb3",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb2",
			"53a57499-51a6-41ee-a1ed-7abb994e8bb1",
		},

		RspCode: http.StatusOK,
		RspBody: []byte(`{"almost": "json"`),
		Error:   errors.New("iothub: failed to decode API response"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			httpClient := &http.Client{
				Transport: RoundTripperFunc(func(
					req *http.Request,
				) (*http.Response, error) {
					if tc.ClientError != nil {
						return nil, tc.ClientError
					}
					// Validate request body
					var body struct {
						Query string `json:"query"`
					}
					dec := json.NewDecoder(req.Body)
					err := dec.Decode(&body)
					if assert.NoError(t, err) {
						q := fmt.Sprintf(
							"SELECT * FROM devices WHERE "+
								"devices.deviceid IN ['%s']",
							strings.Join(tc.DeviceIDs, "','"),
						)
						assert.Equal(t, q, body.Query)
					}

					w.WriteHeader(tc.RspCode)
					switch t := tc.RspBody.(type) {
					case []byte:
						_, _ = w.Write(t)
					default:
						b, _ := json.Marshal(t)
						_, _ = w.Write(b)
					}

					return w.Result(), nil
				}),
			}
			client := NewClient(NewOptions().SetClient(httpClient))
			res, err := client.GetDeviceTwins(tc.CTX, tc.ConnStr, tc.DeviceIDs)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.RspBody, res)
			}
		})
	}
}
