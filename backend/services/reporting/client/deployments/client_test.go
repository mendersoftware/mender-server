// Copyright 2023 Northern.tech AS
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

package deployments

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

func newTestServer(
	t *testing.T,
	rspChan <-chan *http.Response,
	reqChan chan<- *http.Request,
) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var rsp *http.Response
		select {
		case rsp = <-rspChan:
			t.Log(rsp)
		default:
			panic("[PROG ERR] I don't know what to respond!")
		}
		if reqChan != nil {
			bodyClone := bytes.NewBuffer(nil)
			_, _ = io.Copy(bodyClone, r.Body)
			req := r.Clone(context.TODO())
			req.Body = io.NopCloser(bodyClone)
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

func TestGetDeployments(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX       context.Context
		TenantID  string
		DeviceIDs []string

		URLNoise     string
		ResponseCode int
		ResponseBody interface{}

		Res   []*DeviceDeployment
		Error error
	}{{
		Name: "ok, no devices",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: []DeviceDeployment{},
	}, {
		Name: "ok, single device",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: []DeviceDeployment{{
			ID: "c5e37ef5-160e-401a-aec3-9dbef94855c0",
			Device: &Device{
				Status: "success",
			},
		}},

		Res: []*DeviceDeployment{{
			ID: "c5e37ef5-160e-401a-aec3-9dbef94855c0",
			Device: &Device{
				Status: "success",
			},
		}},
	}, {
		Name: "ok, many devices",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceIDs: []string{
			"9acfe595-78ff-456a-843a-0fa08bfd7c7a",
			"e48cd1d6-e83a-46af-97c6-1e4a1cefced4",
			"8cddc8b1-f9d6-40dc-972c-d7789a289451",
			"3c4b7d97-4e72-41a8-9148-d89df494ecb6",
			"89b6a72f-0f9c-4d21-8c27-47609dee63df",
			"f8da794d-1e4c-42a2-9a5e-cdc1b072578a",
			"21d14a6e-ef13-49dd-9f26-0b431be64f2a",
			"2f2b32b6-d50b-4e63-9f76-748181230058",
			"d7996bce-af0b-4406-ba9f-4112d8971932",
			"dc005631-2bd0-4acb-bcc2-0057b233dff0",
			"04ea9ad7-4c63-462a-b44a-18a1287500ee",
			"c6ec945f-5f5e-4297-a1aa-dcd1d8c8d431",
			"f21a4010-9e47-4424-aa5b-89773619b5f1",
			"c61bddef-1a90-4cb1-a10b-c7ade4f03d82",
			"5e1158c6-df44-4ede-8419-58a16df7be6b",
			"1e0bd30a-bf51-445e-9387-231fd9b93bc1",
			"2d94b97a-61d7-488c-bbd3-51007fbe2be7",
			"9a7f7b78-c49e-438d-993d-1b472c028bc3",
			"f16e00cf-2cb7-481c-ab66-89c4b50c0c19",
			"44cd5dc0-179c-4659-b244-6fe0ee004efa",
			"09f75d35-f060-4384-accb-2a5464eed0fc",
		},

		ResponseCode: http.StatusOK,
		ResponseBody: [][]*DeviceDeployment{
			{{
				ID:     "48280477-0797-4e82-bf4b-8393def31c80",
				Device: &Device{Status: "success"},
			}, {
				ID:     "3ae8a28a-5b93-4e91-aa2a-06b67e5fede8",
				Device: &Device{Status: "success"},
			}, {
				ID:     "fdf6ac57-dd6c-4115-b96b-cf1117ae2be1",
				Device: &Device{Status: "success"},
			}, {
				ID:     "0d23d997-acc7-4754-b02a-7ccc975b1ed5",
				Device: &Device{Status: "success"},
			}, {
				ID:     "89b8fbd6-1a80-4cae-91d4-6dc81a4739d7",
				Device: &Device{Status: "success"},
			}, {
				ID:     "ad771081-fa21-4bd9-9f00-e85837614e5c",
				Device: &Device{Status: "success"},
			}, {
				ID:     "9c1e28f4-c89c-4f92-92f8-bebb8f20ec0a",
				Device: &Device{Status: "success"},
			}, {
				ID:     "f59c2dfb-e0e1-4183-8dd7-d025eac1f18d",
				Device: &Device{Status: "success"},
			}, {
				ID:     "6c528663-664d-47b1-aaf6-af67273ffb21",
				Device: &Device{Status: "success"},
			}, {
				ID:     "f67ee4d3-0371-4fb1-8ac7-599dce78bef3",
				Device: &Device{Status: "success"},
			}, {
				ID:     "dac6cdf3-bd5c-4e63-b4ca-da79c2ea25e5",
				Device: &Device{Status: "success"},
			}, {
				ID:     "90ea87be-1857-4586-873a-5a0d5aa6a312",
				Device: &Device{Status: "success"},
			}, {
				ID:     "60eab960-7265-4bdc-9f8d-406a09ce05c6",
				Device: &Device{Status: "success"},
			}, {
				ID:     "81153459-a1e8-4d4f-aa64-15e2c2e06d89",
				Device: &Device{Status: "success"},
			}, {
				ID:     "623181f3-e39d-4f41-896f-a3fcbb895cdc",
				Device: &Device{Status: "success"},
			}, {
				ID:     "68203804-f297-4ba8-8607-6981f7cbfbf8",
				Device: &Device{Status: "success"},
			}, {
				ID:     "f4c87f5b-8a48-4458-9c70-6f30dde5d28e",
				Device: &Device{Status: "success"},
			}, {
				ID:     "ef6c6c9d-9ea0-4a8d-a3eb-20e2307a1d96",
				Device: &Device{Status: "success"},
			}, {
				ID:     "9cbbf86a-81d1-4ed4-ad6e-c90032f570c5",
				Device: &Device{Status: "success"},
			}, {
				ID:     "9ff9fc00-2df0-4bfe-97f0-951312246fa0",
				Device: &Device{Status: "success"},
			}}, {{
				ID:     "e92bb2c7-9f25-4113-8478-322360077895",
				Device: &Device{Status: "success"},
			}},
		},

		Res: []*DeviceDeployment{{
			ID:     "48280477-0797-4e82-bf4b-8393def31c80",
			Device: &Device{Status: "success"},
		}, {
			ID:     "3ae8a28a-5b93-4e91-aa2a-06b67e5fede8",
			Device: &Device{Status: "success"},
		}, {
			ID:     "fdf6ac57-dd6c-4115-b96b-cf1117ae2be1",
			Device: &Device{Status: "success"},
		}, {
			ID:     "0d23d997-acc7-4754-b02a-7ccc975b1ed5",
			Device: &Device{Status: "success"},
		}, {
			ID:     "89b8fbd6-1a80-4cae-91d4-6dc81a4739d7",
			Device: &Device{Status: "success"},
		}, {
			ID:     "ad771081-fa21-4bd9-9f00-e85837614e5c",
			Device: &Device{Status: "success"},
		}, {
			ID:     "9c1e28f4-c89c-4f92-92f8-bebb8f20ec0a",
			Device: &Device{Status: "success"},
		}, {
			ID:     "f59c2dfb-e0e1-4183-8dd7-d025eac1f18d",
			Device: &Device{Status: "success"},
		}, {
			ID:     "6c528663-664d-47b1-aaf6-af67273ffb21",
			Device: &Device{Status: "success"},
		}, {
			ID:     "f67ee4d3-0371-4fb1-8ac7-599dce78bef3",
			Device: &Device{Status: "success"},
		}, {
			ID:     "dac6cdf3-bd5c-4e63-b4ca-da79c2ea25e5",
			Device: &Device{Status: "success"},
		}, {
			ID:     "90ea87be-1857-4586-873a-5a0d5aa6a312",
			Device: &Device{Status: "success"},
		}, {
			ID:     "60eab960-7265-4bdc-9f8d-406a09ce05c6",
			Device: &Device{Status: "success"},
		}, {
			ID:     "81153459-a1e8-4d4f-aa64-15e2c2e06d89",
			Device: &Device{Status: "success"},
		}, {
			ID:     "623181f3-e39d-4f41-896f-a3fcbb895cdc",
			Device: &Device{Status: "success"},
		}, {
			ID:     "68203804-f297-4ba8-8607-6981f7cbfbf8",
			Device: &Device{Status: "success"},
		}, {
			ID:     "f4c87f5b-8a48-4458-9c70-6f30dde5d28e",
			Device: &Device{Status: "success"},
		}, {
			ID:     "ef6c6c9d-9ea0-4a8d-a3eb-20e2307a1d96",
			Device: &Device{Status: "success"},
		}, {
			ID:     "9cbbf86a-81d1-4ed4-ad6e-c90032f570c5",
			Device: &Device{Status: "success"},
		}, {
			ID:     "9ff9fc00-2df0-4bfe-97f0-951312246fa0",
			Device: &Device{Status: "success"},
		}, {
			ID:     "e92bb2c7-9f25-4113-8478-322360077895",
			Device: &Device{Status: "success"},
		}},
	}, {
		Name: "ok, not found",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusNotFound,
	}, {
		Name: "error, context canceled",

		DeviceIDs: []string{
			"fddb6e88-0265-4948-befd-909b1f08ac2e",
		},
		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		Error: context.Canceled,
	}, {
		Name:     "error, nil context",
		CTX:      context.Background(),
		URLNoise: "#%%%",

		Error: errors.New("failed to create request"),
	}, {
		Name: "error, invalid response schema",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: []byte("bad response"),
		Error:        errors.New("failed to parse request body"),
	}, {
		Name: "error, unexpected status code",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusInternalServerError,
		ResponseBody: rest.Error{Err: "something went wrong..."},
		Error:        errors.New(`^GET .+ request failed with status 500`),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			rspChan := make(chan *http.Response, 1)

			rsp := &http.Response{
				StatusCode: tc.ResponseCode,
			}
			switch typ := tc.ResponseBody.(type) {
			case [][]*DeviceDeployment:
				rspChan = make(chan *http.Response, len(typ))
				for _, body := range typ {
					b, _ := json.Marshal(body)
					rspChan <- &http.Response{
						StatusCode: tc.ResponseCode,
						Body:       io.NopCloser(bytes.NewReader(b)),
					}
				}
				rsp = nil

			case []DeviceDeployment:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case rest.Error:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case []byte:
				rsp.Body = io.NopCloser(bytes.NewReader(typ))

			case nil:
				// pass

			default:
				panic("[PROG ERR] invalid ResponseBody type")
			}
			if rsp != nil {
				rspChan <- rsp
			}

			srv := newTestServer(t, rspChan, nil)
			defer srv.Close()

			client := NewClient(srv.URL + tc.URLNoise)

			dev, err := client.GetDeployments(tc.CTX, tc.TenantID, tc.DeviceIDs)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error message does not match expected pattern",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Res, dev)
			}

		})
	}
}

func TestGetLatestFinishedDeployment(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX       context.Context
		TenantID  string
		DeviceIDs []string

		URLNoise     string
		ResponseCode int
		ResponseBody interface{}

		Res   []LastDeviceDeployment
		Error error
	}{{
		Name: "ok, no devices",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: GetLastDeviceDeploymentRsp{
			DeviceDeploymentLastStatuses: []LastDeviceDeployment{},
		},
	}, {
		Name: "ok",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: GetLastDeviceDeploymentRsp{
			DeviceDeploymentLastStatuses: []LastDeviceDeployment{
				{
					DeviceID:               "c5e37ef5-160e-401a-aec3-9dbef94855c0",
					DeviceDeploymentStatus: "success",
				},
			},
		},

		Res: []LastDeviceDeployment{
			{
				DeviceID:               "c5e37ef5-160e-401a-aec3-9dbef94855c0",
				DeviceDeploymentStatus: "success",
			},
		},
	}, {
		Name: "ok, not found",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusNotFound,
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		Error: context.Canceled,
	}, {
		Name:     "error, nil context",
		CTX:      context.Background(),
		URLNoise: "#%%%",

		Error: errors.New("failed to create request"),
	}, {
		Name: "error, invalid response schema",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusOK,
		ResponseBody: []byte("bad response"),
		Error:        errors.New("failed to parse request body"),
	}, {
		Name: "error, unexpected status code",

		CTX:       context.Background(),
		TenantID:  "123456789012345678901234",
		DeviceIDs: []string{"9acfe595-78ff-456a-843a-0fa08bfd7c7a"},

		ResponseCode: http.StatusInternalServerError,
		ResponseBody: rest.Error{Err: "something went wrong..."},
		Error:        errors.New(`^POST .+ request failed with status 500`),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			rspChan := make(chan *http.Response, 1)
			srv := newTestServer(t, rspChan, nil)
			defer srv.Close()

			client := NewClient(srv.URL + tc.URLNoise)

			rsp := &http.Response{
				StatusCode: tc.ResponseCode,
			}

			switch typ := tc.ResponseBody.(type) {
			case GetLastDeviceDeploymentRsp:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case rest.Error:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case []byte:
				rsp.Body = io.NopCloser(bytes.NewReader(typ))

			case nil:
				// pass

			default:
				panic("[PROG ERR] invalid ResponseBody type")
			}
			rspChan <- rsp
			devs, err := client.GetLatestFinishedDeployment(tc.CTX, tc.TenantID, tc.DeviceIDs)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error message does not match expected pattern",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Res, devs)
			}

		})
	}
}
