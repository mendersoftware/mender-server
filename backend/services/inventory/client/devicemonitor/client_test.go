// Copyright 2023 Northern.tech AS
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
package devicemonitor

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

func TestCheckAlerts(t *testing.T) {
	t.Parallel()

	defaultCtx := context.Background()
	expiredCtx, cancel := context.WithDeadline(defaultCtx, time.Now().Add(-1*time.Second))
	defer cancel()

	uuid, err := uuid.Parse("123e4567-e89b-12d3-a456-426652340000")
	assert.NoError(t, err)

	testCases := []struct {
		Name string

		Ctx      context.Context
		DeviceId string

		// Devicemonitor response
		ResponseCode   int
		ResponseBody   interface{}
		NumberOfAlerts int

		Error error
	}{
		{
			Name: "ok, no alerts",

			Ctx:          defaultCtx,
			DeviceId:     "foo",
			ResponseCode: http.StatusOK,
			ResponseBody: Alerts{},
		},
		{
			Name: "ok, alerts",

			Ctx:          defaultCtx,
			DeviceId:     "foo",
			ResponseCode: http.StatusOK,
			ResponseBody: Alerts{
				Alert{ID: uuid},
				Alert{ID: uuid},
			},
			NumberOfAlerts: 2,
		},
		{
			Name: "ok, alerts with tenant ID",

			Ctx:          identity.WithContext(defaultCtx, &identity.Identity{Tenant: "tenant"}),
			DeviceId:     "foo",
			ResponseCode: http.StatusOK,
			ResponseBody: Alerts{
				Alert{ID: uuid},
				Alert{ID: uuid},
			},
			NumberOfAlerts: 2,
		},
		{
			Name: "error, bad HTTP status",

			Ctx:          defaultCtx,
			DeviceId:     "foo",
			ResponseCode: http.StatusNotFound,
			Error:        errors.New("unexpected HTTP status from devicemonitor service: 404 Not Found"),
		},
		{
			Name: "error, bad response",

			Ctx:          defaultCtx,
			DeviceId:     "foo",
			ResponseCode: http.StatusOK,
			ResponseBody: "dummy",
			Error:        errors.New("json: cannot unmarshal string into Go value of type devicemonitor.Alerts"),
		},
		{
			Name: "error, expired deadline",

			Ctx:   expiredCtx,
			Error: errors.New(context.DeadlineExceeded.Error()),
		},
	}

	responses := make(chan http.Response, 1)
	serverHTTP := func(w http.ResponseWriter, r *http.Request) {
		rsp := <-responses
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(serverHTTP))
	client := NewClient(srv.URL, ClientOptions{Client: &http.Client{}})
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

			count, err := client.CheckAlerts(tc.Ctx, tc.DeviceId)

			if tc.Error != nil {
				assert.Contains(t, err.Error(), tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.NumberOfAlerts, count)
			}
		})
	}
}
