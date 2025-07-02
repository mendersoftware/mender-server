// Copyright 2021 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

// newTestServer creates a new mock server that responds with the responses
// pushed onto the rspChan and pushes any requests received onto reqChan if
// the requests are consumed in the other end.
func newTestServer(
	rspChan <-chan *http.Response,
	reqChan chan<- *http.Request,
) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var rsp *http.Response
		select {
		case rsp = <-rspChan:
		default:
			panic("[PROG ERR] I don't know what to respond!")
		}
		if reqChan != nil {
			bodyClone := bytes.NewBuffer(nil)
			_, _ = io.Copy(bodyClone, r.Body)
			req := r.Clone(context.TODO())
			req.Body = ioutil.NopCloser(bodyClone)
			select {
			case reqChan <- req:
				// Only push request if test function is
				// popping from the channel.
			default:
			}
		}
		hdrs := w.Header()
		for k, v := range rsp.Header {
			for _, vv := range v {
				hdrs.Add(k, vv)
			}
		}
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestCheckHealth(t *testing.T) {
	t.Parallel()

	expiredCtx, cancel := context.WithDeadline(
		context.TODO(), time.Now().Add(-1*time.Second))
	defer cancel()
	defaultCtx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	testCases := []struct {
		Name string

		Ctx context.Context

		// Workflows response
		ResponseCode int
		ResponseBody interface{}

		Error error
	}{{
		Name: "ok",

		Ctx:          defaultCtx,
		ResponseCode: http.StatusOK,
	}, {
		Name: "error, expired deadline",

		Ctx:   expiredCtx,
		Error: errors.New(context.DeadlineExceeded.Error()),
	}, {
		Name: "error, workflows unhealthy",

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: rest.Error{
			Err:       "internal error",
			RequestID: "test",
		},

		Error: errors.New("internal error"),
	}, {
		Name: "error, bad response",

		Ctx: context.TODO(),

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: "foobar",

		Error: errors.New("health check HTTP error: 503 Service Unavailable"),
	}}

	responses := make(chan http.Response, 1)
	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		rsp := <-responses
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(serveHTTP))
	client := NewClient(srv.URL)
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			if tc.ResponseCode > 0 {
				rsp := http.Response{
					StatusCode: tc.ResponseCode,
				}
				if tc.ResponseBody != nil {
					b, _ := json.Marshal(tc.ResponseBody)
					rsp.Body = ioutil.NopCloser(bytes.NewReader(b))
				}
				responses <- rsp
			}

			err := client.CheckHealth(tc.Ctx)

			if tc.Error != nil {
				assert.Contains(t, err.Error(), tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubmitAuditLog(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX      context.Context
		AuditLog AuditLog

		URLNoise string // Sole purpose is to provide a bad URL

		Response *http.Response
		Error    error
	}{{
		Name: "ok",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			),
			"testing"),
		AuditLog: AuditLog{
			Action: ActionSetConfiguration,
			Actor: Actor{
				ID:   "4cd02655-d45e-464f-9790-e730286ff888",
				Type: ActorUser,
			},
			Object: Object{
				ID:   "4cd02655-d45e-464f-9790-e730286ff889",
				Type: ObjectDevice,
			},
			EventTS: time.Unix(1234567890, 0),
		},

		Response: &http.Response{
			StatusCode: 201,
		},
	}, {
		Name: "error, bad log",

		CTX: context.Background(),

		AuditLog: AuditLog{},

		Error: errors.New(`^workflows: invalid AuditLog entry: ` +
			`action: cannot be blank; ` +
			`actor: \(.+\); object: \(.+\)\.$`,
		),
	}, {
		Name: "error, invalid context",

		CTX: context.Background(),

		AuditLog: AuditLog{
			Action: ActionSetConfiguration,
			Actor: Actor{
				ID:   "4cd02655-d45e-464f-9790-e730286ff888",
				Type: ActorUser,
			},
			Object: Object{
				ID:   "4cd02655-d45e-464f-9790-e730286ff889",
				Type: ObjectDevice,
			},
		},

		Error: errors.New(`^workflows: Context lacking tenant identity$`),
	}, {
		Name: "error, bad request URL",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			),
			"testing"),
		AuditLog: AuditLog{
			Action: ActionSetConfiguration,
			Actor: Actor{
				ID:   "4cd02655-d45e-464f-9790-e730286ff888",
				Type: ActorUser,
			},
			Object: Object{
				ID:   "4cd02655-d45e-464f-9790-e730286ff889",
				Type: ObjectDevice,
			},
			EventTS: time.Unix(1234567890, 0),
		},
		URLNoise: "?####$%%%%",
		Error: errors.New(`^workflows: error preparing HTTP request: ` +
			`parse ".+": invalid URL escape "%%%"$`),
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx := requestid.WithContext(
				identity.WithContext(
					context.Background(),
					&identity.Identity{
						Tenant: "testing-mender-io",
					},
				), "testing",
			)
			ctx, cancel := context.WithCancel(ctx)
			cancel()
			return ctx
		}(),
		AuditLog: AuditLog{
			Action: ActionSetConfiguration,
			Actor: Actor{
				ID:   "4cd02655-d45e-464f-9790-e730286ff888",
				Type: ActorUser,
			},
			Object: Object{
				ID:   "4cd02655-d45e-464f-9790-e730286ff889",
				Type: ObjectDevice,
			},
			EventTS: time.Unix(1234567890, 0),
		},
		Error: errors.Errorf(`workflows: failed to submit auditlog: `+
			`Post ".+?%s": %s`, AuditlogsURI,
			context.Canceled.Error(),
		),
		Response: &http.Response{
			StatusCode: 201,
		},
	}, {
		Name: "error, auditlog does not exist",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			), "testing",
		),
		AuditLog: AuditLog{
			Action: ActionSetConfiguration,
			Actor: Actor{
				ID:   "4cd02655-d45e-464f-9790-e730286ff888",
				Type: ActorUser,
			},
			Object: Object{
				ID:   "4cd02655-d45e-464f-9790-e730286ff889",
				Type: ObjectDevice,
			},
			EventTS: time.Unix(1234567890, 0),
		},
		Error: errors.New(`^workflows: workflow "auditlogs" not defined$`),
		Response: &http.Response{
			StatusCode: 404,
		},
	}, {
		Name: "error, unexpected response",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			), "testing",
		),
		AuditLog: AuditLog{
			Action: ActionSetConfiguration,
			Actor: Actor{
				ID:   "4cd02655-d45e-464f-9790-e730286ff888",
				Type: ActorUser,
			},
			Object: Object{
				ID:   "4cd02655-d45e-464f-9790-e730286ff889",
				Type: ObjectDevice,
			},
			EventTS: time.Unix(1234567890, 0),
		},
		Error: errors.Errorf(`^workflows: unexpected HTTP status from `+
			`workflows service: %d`, http.StatusInternalServerError),
		Response: &http.Response{
			StatusCode: 500,
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			rspChan := make(chan *http.Response, 1)
			reqChan := make(chan *http.Request, 1)
			srv := newTestServer(rspChan, reqChan)
			c := NewClient(srv.URL+tc.URLNoise, ClientOptions{
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
			})
			if tc.Response != nil {
				select {
				case rspChan <- tc.Response:
				default:
					panic("[PROG ERR] Test case error (race)!")
				}
			}

			err := c.SubmitAuditLog(tc.CTX, tc.AuditLog)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
				var (
					req   *http.Request
					wflow AuditWorkflow
				)
				select {
				case req = <-reqChan:

				default:
					panic("[PROG ERR] bad test case!")
				}
				if !assert.NotNil(t, req.Body) {
					return
				}
				defer req.Body.Close()
				decoder := json.NewDecoder(req.Body)
				err := decoder.Decode(&wflow)
				if !assert.NoError(t, err) {
					return
				}
				if tc.AuditLog.EventTS.IsZero() {
					assert.WithinDuration(t,
						time.Now(),
						wflow.AuditLog.EventTS,
						time.Minute,
					)
				} else {
					assert.WithinDuration(t,
						tc.AuditLog.EventTS,
						wflow.AuditLog.EventTS,
						time.Second,
					)
				}
				wflow.AuditLog.EventTS = time.Time{}
				tc.AuditLog.EventTS = time.Time{}
				assert.Equal(t, tc.AuditLog, wflow.AuditLog)
				id := identity.FromContext(tc.CTX)
				if assert.NotNil(t, id) {
					assert.Equal(t, id.Tenant, wflow.TenantID)
				}
			}
		})
	}
}

func TestDeployConfiguration(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX              context.Context
		tenantID         string
		deviceID         string
		deploymentID     uuid.UUID
		configuration    []byte
		retries          uint
		updateControlMap map[string]interface{}

		URLNoise string // Sole purpose is to provide a bad URL

		Response *http.Response
		Error    error
	}{{
		Name: "ok",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			),
			"testing"),

		tenantID:         "tenantID",
		deviceID:         uuid.New().String(),
		deploymentID:     uuid.New(),
		configuration:    []byte("{\"key\":\"value\"}"),
		retries:          1,
		updateControlMap: map[string]interface{}{"foo": "bar"},

		Response: &http.Response{
			StatusCode: 201,
		},
	}, {
		Name: "error, bad request URL",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			),
			"testing"),

		tenantID:      "tenantID",
		deviceID:      uuid.New().String(),
		deploymentID:  uuid.New(),
		configuration: []byte("{\"key\":\"value\"}"),
		retries:       1,

		URLNoise: "?####$%%%%",
		Error: errors.New(`^workflows: error preparing HTTP request: ` +
			`parse ".+": invalid URL escape "%%%"$`),
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx := requestid.WithContext(
				identity.WithContext(
					context.Background(),
					&identity.Identity{
						Tenant: "testing-mender-io",
					},
				), "testing",
			)
			ctx, cancel := context.WithCancel(ctx)
			cancel()
			return ctx
		}(),

		tenantID:      "tenantID",
		deviceID:      uuid.New().String(),
		deploymentID:  uuid.New(),
		configuration: []byte("{\"key\":\"value\"}"),
		retries:       1,

		Error: errors.Errorf(`workflows: failed to deploy configuration: `+
			`Post ".+?%s": %s`, DeployDeviceConfigurationRI,
			context.Canceled.Error(),
		),
		Response: &http.Response{
			StatusCode: 201,
		},
	}, {
		Name: "error, deploy_device_configuration does not exist",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			), "testing",
		),

		tenantID:      "tenantID",
		deviceID:      uuid.New().String(),
		deploymentID:  uuid.New(),
		configuration: []byte("{\"key\":\"value\"}"),
		retries:       1,

		Error: errors.New(`^workflows: workflow "deploy_device_configuration" not defined$`),
		Response: &http.Response{
			StatusCode: 404,
		},
	}, {
		Name: "error, unexpected response",

		CTX: requestid.WithContext(
			identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "testing-mender-io",
				},
			), "testing",
		),

		tenantID:      "tenantID",
		deviceID:      uuid.New().String(),
		deploymentID:  uuid.New(),
		configuration: []byte("{\"key\":\"value\"}"),
		retries:       1,

		Error: errors.Errorf(`^workflows: unexpected HTTP status from `+
			`workflows service: %d`, http.StatusInternalServerError),
		Response: &http.Response{
			StatusCode: 500,
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			rspChan := make(chan *http.Response, 1)
			reqChan := make(chan *http.Request, 1)
			srv := newTestServer(rspChan, reqChan)
			c := NewClient(srv.URL+tc.URLNoise, ClientOptions{
				Client: &http.Client{
					Timeout: defaultTimeout,
				},
			})
			if tc.Response != nil {
				select {
				case rspChan <- tc.Response:
				default:
					panic("[PROG ERR] Test case error (race)!")
				}
			}

			err := c.DeployConfiguration(tc.CTX, tc.tenantID, tc.deviceID, tc.deploymentID, tc.configuration, tc.retries, tc.updateControlMap)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
				var (
					req   *http.Request
					wflow DeployConfigurationWorkflow
				)
				select {
				case req = <-reqChan:

				default:
					panic("[PROG ERR] bad test case!")
				}
				if !assert.NotNil(t, req.Body) {
					return
				}
				defer req.Body.Close()
				decoder := json.NewDecoder(req.Body)
				err := decoder.Decode(&wflow)
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, tc.tenantID, wflow.TenantID)
				assert.Equal(t, tc.deviceID, wflow.DeviceID)
				assert.Equal(t, tc.deploymentID, wflow.DeploymentID)
				assert.Equal(t, string(tc.configuration), wflow.Configuration)
				assert.Equal(t, tc.retries, wflow.Retries)
			}
		})
	}
}
