// Copyright 2022 Northern.tech AS
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
package tenant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/apiclient"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"

	ct "github.com/mendersoftware/mender-server/services/useradm/client/testing"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	c := NewClient(Config{TenantAdmAddr: "http://foo"})
	assert.NotNil(t, c)
}

func TestCheckHealth(t *testing.T) {
	t.Parallel()

	contextWithTimeout, cancel := context.WithTimeout(
		context.TODO(), time.Second*5,
	)
	defer cancel()
	expiredContext, cancel := context.WithDeadline(
		context.TODO(),
		time.Now().Add(-time.Second),
	)
	defer cancel()
	testCases := []struct {
		Name string

		Ctx   context.Context
		Error error

		ResponseCode int
		ResponseBody interface{}
	}{{
		Name: "ok",

		Ctx:          contextWithTimeout,
		ResponseCode: http.StatusNoContent,
	}, {
		Name: "ok, nil context",

		ResponseCode: http.StatusNoContent,
	}, {
		Name: "error, deadline expired",

		Ctx:   expiredContext,
		Error: context.DeadlineExceeded,
	}, {
		Name: "error, Tenantadm unhealty",

		Ctx:          context.Background(),
		Error:        errors.New("*COUGH! COUGH!*"),
		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: rest_utils.ApiError{
			Err: "*COUGH! COUGH!*",
		},
	}, {
		Name: "error, unexpected response",

		Ctx:          context.Background(),
		Error:        errors.New("service unhealthy: HTTP 503 Service Unavailable"),
		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: "bleh!",
	}}

	responses := make(chan http.Response, 1)
	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		var rsp http.Response
		select {
		case rsp = <-responses:
		default:
			t.Log("[PROG ERR] I don't know what to respond")
			t.FailNow()
		}
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			io.Copy(w, rsp.Body)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(serveHTTP))
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tadmClient := NewClient(Config{
				TenantAdmAddr: srv.URL,
			})

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

			err := tadmClient.CheckHealth(tc.Ctx)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Contains(t,
						err.Error(),
						tc.Error.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTenant(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		status    int
		inTenants []Tenant
		tenant    *Tenant
		err       error
	}{

		"ok": {
			status: http.StatusOK,
			inTenants: []Tenant{
				{
					ID:   "1234",
					Name: "tenant",
				},
			},

			tenant: &Tenant{
				ID:   "1234",
				Name: "tenant",
			},
			err: nil,
		},
		"ok - tenant not found": {
			// note no 404 here
			status:    http.StatusOK,
			inTenants: []Tenant{},
			tenant:    nil,
			err:       nil,
		},
		"error: generic": {
			status: http.StatusInternalServerError,
			err:    errors.New("GET /tenants request failed with unexpected status 500"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, rd := ct.NewMockServer(tc.status, tc.inTenants)

			c := NewClient(Config{
				TenantAdmAddr: s.URL,
			})

			tenant, err := c.GetTenant(context.Background(), "username", &apiclient.HttpApi{})
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, GetTenantsUri, rd.Url.Path)
				assert.Equal(t, "GET", rd.Method)
				assert.Equal(t, tc.tenant, tenant)
			}
			s.Close()
		})
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		status int
		err    error
	}{
		"ok": {
			status: http.StatusCreated,
			err:    nil,
		},
		"error: duplicate user": {
			status: http.StatusUnprocessableEntity,
			err:    ErrDuplicateUser,
		},
		"error: generic": {
			status: http.StatusInternalServerError,
			err:    errors.New("POST /users request failed with unexpected status 500"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, rd := ct.NewMockServer(tc.status, nil)

			c := NewClient(Config{
				TenantAdmAddr: s.URL,
			})

			user := &User{
				ID:       "foo",
				Name:     "foo@bar.com",
				TenantID: "foo-tenant",
			}

			err := c.CreateUser(context.Background(), user, &apiclient.HttpApi{})
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, UsersUri, rd.Url.Path)
				assert.Equal(t, "POST", rd.Method)
			}
			s.Close()
		})
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		status int
		err    error
	}{
		"ok": {
			status: http.StatusNoContent,
			err:    nil,
		},
		"error: generic": {
			status: http.StatusInternalServerError,
			err:    errors.New("DELETE /api/internal/v1/tenantadm/tenants/foo/users/bar request failed with unexpected status 500"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, rd := ct.NewMockServer(tc.status, nil)

			c := NewClient(Config{
				TenantAdmAddr: s.URL,
			})

			err := c.DeleteUser(context.Background(), "foo", "bar", &apiclient.HttpApi{})
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				uri := strings.Replace(TenantsUsersUri, "#tid", "foo", 1)
				uri = strings.Replace(uri, "#uid", "bar", 1)
				assert.Equal(t, uri, rd.Url.Path)
				assert.Equal(t, "DELETE", rd.Method)
			}
			s.Close()
		})
	}
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tid    string
		uid    string
		status int
		err    error
	}{
		"ok": {
			tid:    "foo",
			uid:    "bar",
			status: http.StatusNoContent,
			err:    nil,
		},
		"error: duplicate user": {
			tid:    "1",
			uid:    "2",
			status: http.StatusUnprocessableEntity,
			err:    ErrDuplicateUser,
		},
		"error: not found": {
			tid:    "bar",
			uid:    "baz",
			status: http.StatusNotFound,
			err:    ErrUserNotFound,
		},
		"error: generic": {
			tid:    "3",
			uid:    "4",
			status: http.StatusInternalServerError,
			err:    errors.New("PUT /tenants/:id/users/:id request failed with unexpected status 500"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, rd := ct.NewMockServer(tc.status, nil)

			c := NewClient(Config{
				TenantAdmAddr: s.URL,
			})

			up := &UserUpdate{
				Name: "foo@bar.com",
			}

			err := c.UpdateUser(context.Background(), tc.tid, tc.uid, up, &apiclient.HttpApi{})

			assert.Equal(t, "PUT", rd.Method)
			uri := strings.Replace(TenantsUsersUri, "#tid", tc.tid, 1)
			uri = strings.Replace(uri, "#uid", tc.uid, 1)
			assert.Equal(t, uri, rd.Url.Path)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
			s.Close()
		})
	}
}
