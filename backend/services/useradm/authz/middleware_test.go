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

package authz_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
	mt "github.com/mendersoftware/mender-server/pkg/testing"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	. "github.com/mendersoftware/mender-server/services/useradm/authz"
	mauthz "github.com/mendersoftware/mender-server/services/useradm/authz/mocks"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	mtest "github.com/mendersoftware/mender-server/services/useradm/utils/testing"
)

func TestAuthzMiddleware(t *testing.T) {

	testCases := []struct {
		Name string

		Header http.Header
		Cookie *http.Cookie

		action    *Action
		actionErr error

		authErr error

		checker mt.ResponseChecker
	}{{
		Name: "ok",
		Header: http.Header{"Authorization": []string{
			"Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAi" +
				"OjQxMDExMDQwNjksImlzcyI6Im1lbmRlciIsInN1YiI6" +
				"Ijc0YzYwMmI1LWJlMWYtNTgwMC04MTRkLWJlMzE0MzEy" +
				"MjMwZSIsImp0aSI6Ijc0YzYwMmI1LWJlMWYtNTgwMC04" +
				"MTRkLWJlMzE0MzEyMjMwZSIsInNjcCI6Im1lbmRlci5m" +
				"b28ifQ.KsL6UqOwqE1sdVqHWUj1DNlagv_Gk9AI76zBt" +
				"vlSH8cP-8EV24bD6xdHowoN1TPHdPnvwYsCP6u0KOdRw" +
				"MDHVkJojNEUyMNAZAAKNvyjdEBlKNIbLIrPvFZY4zIyP" +
				"wOljmFhdACahVtPsn3x8g4fkCcd29Bpy21jWs5Y8N4nA" +
				"yJdfCNOaXyQwhY6mAGodAyKy0YKUqJicbpU4rmYUzGGc" +
				"vQqtivUIDXswb80vfsuhOGYJKT6XJMPEeTCK4lPrLIyU" +
				"U5gXjxl0Ym_61MVvPIohOreimeDFMN0TbV_ljDzWxLlM" +
				"d2At6zLVWIOkFc0YLeZAsNlD3JxGqwVfKiWRfj_bg",
		}},
		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},

		actionErr: nil,

		authErr: nil,

		checker: mt.NewJSONResponse(
			http.StatusOK,
			nil,
			map[string]string{"foo": "bar"},
		),
	}, {
		Name: "ok, with cookie",
		Cookie: &http.Cookie{
			Name: "JWT",
			Value: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAi" +
				"OjQxMDExMDQwNjksImlzcyI6Im1lbmRlciIsInN1YiI6" +
				"Ijc0YzYwMmI1LWJlMWYtNTgwMC04MTRkLWJlMzE0MzEy" +
				"MjMwZSIsImp0aSI6Ijc0YzYwMmI1LWJlMWYtNTgwMC04" +
				"MTRkLWJlMzE0MzEyMjMwZSIsInNjcCI6Im1lbmRlci5m" +
				"b28ifQ.KsL6UqOwqE1sdVqHWUj1DNlagv_Gk9AI76zBt" +
				"vlSH8cP-8EV24bD6xdHowoN1TPHdPnvwYsCP6u0KOdRw" +
				"MDHVkJojNEUyMNAZAAKNvyjdEBlKNIbLIrPvFZY4zIyP" +
				"wOljmFhdACahVtPsn3x8g4fkCcd29Bpy21jWs5Y8N4nA" +
				"yJdfCNOaXyQwhY6mAGodAyKy0YKUqJicbpU4rmYUzGGc" +
				"vQqtivUIDXswb80vfsuhOGYJKT6XJMPEeTCK4lPrLIyU" +
				"U5gXjxl0Ym_61MVvPIohOreimeDFMN0TbV_ljDzWxLlM" +
				"d2At6zLVWIOkFc0YLeZAsNlD3JxGqwVfKiWRfj_bg",
			Path: "/",
		},
		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},

		actionErr: nil,

		authErr: nil,

		checker: mt.NewJSONResponse(
			http.StatusOK,
			nil,
			map[string]string{"foo": "bar"},
		),
	}, {
		Name: "error: basic auth",
		Header: http.Header{"Authorization": []string{
			"Basic eyJhbGciOiJSUzI1NiIsInRn:5cCI6IkpXVCJ9eyJleHA",
		}},
		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},

		actionErr: nil,

		authErr: nil,

		checker: mt.NewJSONResponse(
			http.StatusUnauthorized,
			nil,
			restError(ErrInvalidAuthHeader.Error()),
		),
	}, {
		Name: "error: missing token header",

		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},

		actionErr: nil,

		authErr: nil,

		checker: mt.NewJSONResponse(
			http.StatusUnauthorized,
			nil,
			restError(ErrAuthzNoAuth.Error()),
		),
	}, {
		Name: "error: resource id error",

		Header: http.Header{"Authorization": []string{
			"Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAi" +
				"OjQxMDExMDQwNjksImlzcyI6Im1lbmRlciIsInN1YiI6" +
				"Ijc0YzYwMmI1LWJlMWYtNTgwMC04MTRkLWJlMzE0MzEy" +
				"MjMwZSIsImp0aSI6Ijc0YzYwMmI1LWJlMWYtNTgwMC04" +
				"MTRkLWJlMzE0MzEyMjMwZSIsInNjcCI6Im1lbmRlci5m" +
				"b28ifQ.KsL6UqOwqE1sdVqHWUj1DNlagv_Gk9AI76zBt" +
				"vlSH8cP-8EV24bD6xdHowoN1TPHdPnvwYsCP6u0KOdRw" +
				"MDHVkJojNEUyMNAZAAKNvyjdEBlKNIbLIrPvFZY4zIyP" +
				"wOljmFhdACahVtPsn3x8g4fkCcd29Bpy21jWs5Y8N4nA" +
				"yJdfCNOaXyQwhY6mAGodAyKy0YKUqJicbpU4rmYUzGGc" +
				"vQqtivUIDXswb80vfsuhOGYJKT6XJMPEeTCK4lPrLIyU" +
				"U5gXjxl0Ym_61MVvPIohOreimeDFMN0TbV_ljDzWxLlM" +
				"d2At6zLVWIOkFc0YLeZAsNlD3JxGqwVfKiWRfj_bg",
		}},
		action: &Action{
			Resource: "",
			Method:   "GET",
		},
		actionErr: errors.New("can't identify resource"),

		authErr: nil,

		checker: mt.NewJSONResponse(
			http.StatusInternalServerError,
			nil,
			restError("internal error"),
		),
	}, {
		Name: "error: invalid token",
		Header: http.Header{"Authorization": []string{
			"Bearer dummy",
		}},
		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},
		actionErr: errors.New("can't identify resource"),

		authErr: ErrAuthzTokenInvalid,

		checker: mt.NewJSONResponse(
			http.StatusUnauthorized,
			nil,
			restError("invalid jwt"),
		),
	}, {
		Name: "error: forbidden resource",
		Header: http.Header{"Authorization": []string{
			"Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAi" +
				"OjQxMDExMDQwNjksImlzcyI6Im1lbmRlciIsInN1YiI6" +
				"Ijc0YzYwMmI1LWJlMWYtNTgwMC04MTRkLWJlMzE0MzEy" +
				"MjMwZSIsImp0aSI6Ijc0YzYwMmI1LWJlMWYtNTgwMC04" +
				"MTRkLWJlMzE0MzEyMjMwZSIsInNjcCI6Im1lbmRlci5m" +
				"b28ifQ.KsL6UqOwqE1sdVqHWUj1DNlagv_Gk9AI76zBt" +
				"vlSH8cP-8EV24bD6xdHowoN1TPHdPnvwYsCP6u0KOdRw" +
				"MDHVkJojNEUyMNAZAAKNvyjdEBlKNIbLIrPvFZY4zIyP" +
				"wOljmFhdACahVtPsn3x8g4fkCcd29Bpy21jWs5Y8N4nA" +
				"yJdfCNOaXyQwhY6mAGodAyKy0YKUqJicbpU4rmYUzGGc" +
				"vQqtivUIDXswb80vfsuhOGYJKT6XJMPEeTCK4lPrLIyU" +
				"U5gXjxl0Ym_61MVvPIohOreimeDFMN0TbV_ljDzWxLlM" +
				"d2At6zLVWIOkFc0YLeZAsNlD3JxGqwVfKiWRfj_bg",
		}},
		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},
		actionErr: nil,

		authErr: ErrAuthzTokenInvalid,

		checker: mt.NewJSONResponse(
			http.StatusUnauthorized,
			nil,
			restError(ErrAuthzTokenInvalid.Error()),
		),
	}, {
		Name: "error: unauthorized token",
		Header: http.Header{"Authorization": []string{
			"Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAi" +
				"OjQxMDExMDQwNjksImlzcyI6Im1lbmRlciIsInN1YiI6" +
				"Ijc0YzYwMmI1LWJlMWYtNTgwMC04MTRkLWJlMzE0MzEy" +
				"MjMwZSIsImp0aSI6Ijc0YzYwMmI1LWJlMWYtNTgwMC04" +
				"MTRkLWJlMzE0MzEyMjMwZSIsInNjcCI6Im1lbmRlci5m" +
				"b28ifQ.KsL6UqOwqE1sdVqHWUj1DNlagv_Gk9AI76zBt" +
				"vlSH8cP-8EV24bD6xdHowoN1TPHdPnvwYsCP6u0KOdRw" +
				"MDHVkJojNEUyMNAZAAKNvyjdEBlKNIbLIrPvFZY4zIyP" +
				"wOljmFhdACahVtPsn3x8g4fkCcd29Bpy21jWs5Y8N4nA" +
				"yJdfCNOaXyQwhY6mAGodAyKy0YKUqJicbpU4rmYUzGGc" +
				"vQqtivUIDXswb80vfsuhOGYJKT6XJMPEeTCK4lPrLIyU" +
				"U5gXjxl0Ym_61MVvPIohOreimeDFMN0TbV_ljDzWxLlM" +
				"d2At6zLVWIOkFc0YLeZAsNlD3JxGqwVfKiWRfj_bg",
		}},
		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},
		actionErr: nil,

		authErr: ErrAuthzUnauthorized,

		checker: mt.NewJSONResponse(
			http.StatusForbidden,
			nil,
			restError(ErrAuthzUnauthorized.Error()),
		),
	}, {
		Name: "error: authorizer internal error",
		Header: http.Header{"Authorization": []string{
			"Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAi" +
				"OjQxMDExMDQwNjksImlzcyI6Im1lbmRlciIsInN1YiI6" +
				"Ijc0YzYwMmI1LWJlMWYtNTgwMC04MTRkLWJlMzE0MzEy" +
				"MjMwZSIsImp0aSI6Ijc0YzYwMmI1LWJlMWYtNTgwMC04" +
				"MTRkLWJlMzE0MzEyMjMwZSIsInNjcCI6Im1lbmRlci5m" +
				"b28ifQ.KsL6UqOwqE1sdVqHWUj1DNlagv_Gk9AI76zBt" +
				"vlSH8cP-8EV24bD6xdHowoN1TPHdPnvwYsCP6u0KOdRw" +
				"MDHVkJojNEUyMNAZAAKNvyjdEBlKNIbLIrPvFZY4zIyP" +
				"wOljmFhdACahVtPsn3x8g4fkCcd29Bpy21jWs5Y8N4nA" +
				"yJdfCNOaXyQwhY6mAGodAyKy0YKUqJicbpU4rmYUzGGc" +
				"vQqtivUIDXswb80vfsuhOGYJKT6XJMPEeTCK4lPrLIyU" +
				"U5gXjxl0Ym_61MVvPIohOreimeDFMN0TbV_ljDzWxLlM" +
				"d2At6zLVWIOkFc0YLeZAsNlD3JxGqwVfKiWRfj_bg",
		}},
		action: &Action{
			Resource: "foo:bar",
			Method:   "GET",
		},
		actionErr: nil,

		authErr: errors.New("some internal error"),

		checker: mt.NewJSONResponse(
			http.StatusInternalServerError,
			nil,
			restError("internal error"),
		),
	},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			//setup api
			api := rest.NewApi()
			api.Use(
				&requestlog.RequestLogMiddleware{
					BaseLogger: &logrus.Logger{Out: io.Discard},
				},
				&requestid.RequestIdMiddleware{},
			)
			rest.ErrorFieldName = "error"

			ctx := mtest.ContextMatcher()

			//setup mocks
			a := &mauthz.Authorizer{}
			a.On("Authorize",
				ctx,
				mock.AnythingOfType("*jwt.Token"),
				tc.action.Resource,
				tc.action.Method).Return(tc.authErr)

			a.On("WithLog",
				mock.AnythingOfType("*log.Logger")).
				Return(a)

			resfunc := func(r *rest.Request) (*Action, error) {
				return tc.action, tc.actionErr
			}

			//finish setting up the middleware
			privkey := loadPrivKey("../crypto/private.pem", t)
			jwth := jwt.NewJWTHandlerRS256(privkey, 0)
			mw := AuthzMiddleware{
				Authz:       a,
				ResFunc:     resfunc,
				JWTHandlers: map[int]jwt.Handler{0: jwth},
			}
			api.Use(&mw)

			//setup dummy handler
			api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(map[string]string{"foo": "bar"})
			}))

			//test
			req, _ := http.NewRequest(tc.action.Method, "localhost", nil)
			for k, v := range tc.Header {
				for _, vv := range v {
					req.Header.Add(k, vv)
				}
			}
			req.Header.Set("X-MEN-RequestID", "test")
			if tc.Cookie != nil {
				req.AddCookie(tc.Cookie)
			}
			recorded := test.RunRequest(t, api.MakeHandler(), req)
			mt.CheckResponse(t, tc.checker, recorded)
		})
	}
}

func restError(status string) map[string]interface{} {
	return map[string]interface{}{"error": status, "request_id": "test"}
}

func loadPrivKey(path string, t *testing.T) *rsa.PrivateKey {
	pem_data, err := os.ReadFile(path)
	if err != nil {
		t.FailNow()
	}

	block, _ := pem.Decode(pem_data)

	if block == nil ||
		block.Type != "RSA PRIVATE KEY" {
		t.FailNow()
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.FailNow()
	}

	return key
}

func TestGetRequestToken(t *testing.T) {
	token := &jwt.Token{
		Claims: jwt.Claims{
			Subject:   oid.NewUUIDv5("foo"),
			Issuer:    "bar",
			ExpiresAt: &jwt.Time{Time: time.Unix(12345, 0)},
		},
	}

	env := map[string]interface{}{
		"authz_token": token,
	}

	outToken := GetRequestToken(env)
	assert.Equal(t, token, outToken)

}
