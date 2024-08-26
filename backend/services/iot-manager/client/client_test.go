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

package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

func init() {
	model.SetTrustedHostnames([]string{
		"localhost",
		"totally.legit",
		"*.azure-devices.net",
		"*.iot.*.amazonaws.com",
	})
}

func TestAddrIsGlobalUnicast(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Address            string
		ErrorAssertionFunc assert.ErrorAssertionFunc
	}{{
		Name:               "ok",
		Address:            "8.8.8.8",
		ErrorAssertionFunc: assert.NoError,
	}, {
		Name:               "ok/6",
		Address:            "2001:4860:4860:0:0:0:0:8888",
		ErrorAssertionFunc: assert.NoError,
	}, {
		Name:               "ok/port ipv4",
		Address:            "8.8.8.8:51",
		ErrorAssertionFunc: assert.NoError,
	}, {
		Name:               "ok/port ipv6",
		Address:            "[2001:4860:4860:0:0:0:0:8888]:51",
		ErrorAssertionFunc: assert.NoError,
	}, {
		Name:    "error/loopback address",
		Address: "127.0.0.1",
		ErrorAssertionFunc: func(t assert.TestingT, err error, _ ...interface{}) bool {
			var e net.InvalidAddrError
			return assert.ErrorAs(t, err, &e)
		},
	}, {
		Name:    "error/private address",
		Address: "192.168.0.0",
		ErrorAssertionFunc: func(t assert.TestingT, err error, _ ...interface{}) bool {
			var e net.InvalidAddrError
			return assert.ErrorAs(t, err, &e)
		},
	}, {
		Name:    "error/global multi-cast",
		Address: "224.1.2.3",
		ErrorAssertionFunc: func(t assert.TestingT, err error, _ ...interface{}) bool {
			var e net.InvalidAddrError
			return assert.ErrorAs(t, err, &e)
		},
	}, {
		Name:    "error/parse non-ip",
		Address: "this is not an IP address",
		ErrorAssertionFunc: func(t assert.TestingT, err error, _ ...interface{}) bool {
			var e *net.ParseError
			return assert.ErrorAs(t, err, &e)
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			err := addrIsGlobalUnicast("ip", tc.Address, nil)
			tc.ErrorAssertionFunc(t, err)
		})
	}
}

func TestNewWebhookRequest(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		CTX   context.Context
		Creds *model.Credentials
		Event model.WebhookEvent

		RequestValidationFunc func(
			t *testing.T,
			req *http.Request,
			self *testCase,
		) bool
		Error error
	}
	testCases := []testCase{
		{
			Name: "ok",

			CTX: context.Background(),
			Creds: &model.Credentials{
				Type: model.CredentialTypeHTTP,
				HTTP: &model.HTTPCredentials{
					URL: "http://localhost",
				},
			},
			Event: model.WebhookEvent{},
			RequestValidationFunc: func(
				t *testing.T,
				req *http.Request,
				self *testCase,
			) bool {
				ret := assert.NotContains(t, req.Header, ParamAlgorithmType)
				ret = ret && assert.NotContains(t, req.Header, ParamSignature)
				ret = assert.Contains(t, req.Header, HdrKeyContentType)
				b, _ := json.Marshal(self.Event)
				body, _ := io.ReadAll(req.Body)
				ret = ret && assert.JSONEq(t, string(b), string(body))
				return ret
			},
		},
		{
			Name: "ok/with secret",

			CTX: context.Background(),
			Creds: &model.Credentials{
				Type: model.CredentialTypeHTTP,
				HTTP: &model.HTTPCredentials{
					URL: "http://localhost",
					Secret: func() *model.HexSecret {
						s := model.HexSecret([]byte{0, 1, 2, 3})
						return &s
					}(),
				},
			},
			Event: model.WebhookEvent{},
			RequestValidationFunc: func(
				t *testing.T,
				req *http.Request,
				self *testCase,
			) bool {
				ret := true
				if ret = ret && assert.Contains(t,
					req.Header,
					ParamAlgorithmType,
				); ret {
					ret = ret && assert.Equal(t,
						AlgorithmTypeHMAC256,
						req.Header.Get(ParamAlgorithmType),
					)
				}
				ret = assert.Contains(t, req.Header, HdrKeyContentType)
				b, _ := json.Marshal(self.Event)
				body, _ := io.ReadAll(req.Body)
				ret = ret && assert.JSONEq(t, string(b), string(body))

				if r := assert.Contains(t, req.Header, ParamSignature); ret && r {
					signer := hmac.New(
						sha256.New,
						[]byte(*self.Creds.HTTP.Secret),
					)
					signer.Write(body)
					ret = ret && assert.Equal(t,
						req.Header.Get(ParamSignature),
						hex.EncodeToString(signer.Sum(nil)),
					)
				}
				ret = ret && assert.Contains(t, req.Header, ParamSignature)
				return ret
			},
		},
		{
			Name: "error/nil context",

			Creds: &model.Credentials{
				Type: model.CredentialTypeHTTP,
				HTTP: &model.HTTPCredentials{
					URL: "http://localhost",
					Secret: func() *model.HexSecret {
						s := model.HexSecret([]byte{0, 1, 2, 3})
						return &s
					}(),
				},
			},
			Event: model.WebhookEvent{},
			Error: errors.New("nil Context"),
		},
		{
			Name: "error/invalid event",

			Creds: &model.Credentials{
				Type: model.CredentialTypeHTTP,
				HTTP: &model.HTTPCredentials{
					URL: "http://localhost",
				},
			},
			Event: model.WebhookEvent{Data: func() {}},
			Error: &json.UnsupportedTypeError{Type: reflect.TypeOf(func() {})},
		},
		{
			Name: "error/invalid credential type",

			Creds: &model.Credentials{
				Type: model.CredentialTypeSAS,
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Name:     "foobar",
					Key:      crypto.String("1234"),
				},
			},
			Error: errors.New("invalid credentials"),
		},
		{
			Name: "error/invalid credentials",

			Creds: &model.Credentials{},
			Error: validation.Errors{"type": validation.ErrRequired},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			req, err := NewWebhookRequest(tc.CTX, tc.Creds, tc.Event)
			if tc.Error == nil {
				if assert.NoError(t, err) {
					tc.RequestValidationFunc(t, req, &tc)
				}
			} else {
				assert.ErrorContains(t, err, tc.Error.Error())
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	client := New()
	assert.NotNil(t, client.CheckRedirect, "CheckRedirect not overridden")
	transport := client.Transport.(*http.Transport)
	assert.NotNil(t, transport.DialContext, "Transport.DialContext not overridden")
	assert.NotNil(t, transport.DialTLSContext, "Transport.DialTLSContext not overridden")

	client.Transport = http.DefaultTransport
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Location", "http://localhost/redirect")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))
	rsp, err := client.Get(srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTemporaryRedirect, rsp.StatusCode)
}
