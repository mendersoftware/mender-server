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
package authz

import (
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"

	"github.com/mendersoftware/mender-server/services/useradm/jwt"
)

const (
	// token's key in request.Env
	ReqToken = "authz_token"
)

// AuthzMiddleware checks the authorization on a given request.
// It retrieves the token + requested resource and action, and delegates the authz check to an
// Authorizer.
type AuthzMiddleware struct {
	Authz              Authorizer
	ResFunc            ResourceActionExtractor
	JWTHandlers        map[int]jwt.Handler
	JWTFallbackHandler jwt.Handler
}

// Action combines info about the requested resourd + http method.
type Action struct {
	Resource string
	Method   string
}

// ResourceActionExtractor extracts Actions from requests.
type ResourceActionExtractor func(r *rest.Request) (*Action, error)

// MiddlewareFunc makes AuthzMiddleware implement the Middleware interface.
func (mw *AuthzMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		l := log.FromContext(r.Context())

		//get token, no token header = http 401
		tokstr, err := ExtractToken(r.Request)
		if err != nil {
			rest_utils.RestErrWithLog(w, r, l, err, http.StatusUnauthorized)
			return
		}

		keyId := jwt.GetKeyId(tokstr)

		if _, ok := mw.JWTHandlers[keyId]; !ok {
			// we have not found the corresponding handler for the key by id
			// on purpose we do not return common.ErrKeyIdNotFound -- to not allow
			// the enumeration attack
			rest_utils.RestErrWithLog(w, r, l, ErrAuthzTokenInvalid, http.StatusUnauthorized)
			return
		}

		// parse token, insert into env
		token, err := mw.JWTHandlers[keyId].FromJWT(tokstr)
		if err != nil && mw.JWTFallbackHandler != nil {
			token, err = mw.JWTFallbackHandler.FromJWT(tokstr)
		}
		if err != nil {
			rest_utils.RestErrWithLog(w, r, l, ErrAuthzTokenInvalid, http.StatusUnauthorized)
			return
		}

		r.Env[ReqToken] = token

		// extract resource action
		action, err := mw.ResFunc(r)
		if err != nil {
			rest_utils.RestErrWithLogInternal(w, r, l, err)
			return
		}

		ctx := r.Context()

		//authorize, no authz = http 403
		err = mw.Authz.Authorize(ctx, token, action.Resource, action.Method)
		if err != nil {
			if err == ErrAuthzUnauthorized {
				rest_utils.RestErrWithLog(w, r, l,
					ErrAuthzUnauthorized, http.StatusForbidden)
			} else if err == ErrAuthzTokenInvalid {
				rest_utils.RestErrWithLog(w, r, l,
					ErrAuthzTokenInvalid, http.StatusUnauthorized)
			} else {
				rest_utils.RestErrWithLogInternal(w, r, l, err)
			}
			return
		}

		h(w, r)
	}
}

// extracts JWT from authorization header
func ExtractToken(req *http.Request) (string, error) {
	const authHeaderName = "Authorization"
	auth := req.Header.Get(authHeaderName)
	if auth != "" {
		auths := strings.Fields(auth)
		if !strings.EqualFold(auths[0], "Bearer") || len(auths) < 2 {
			return "", ErrInvalidAuthHeader
		}
		return auths[1], nil
	}
	cookie, err := req.Cookie("JWT")
	if err != nil {
		return "", ErrAuthzNoAuth
	}
	auth = cookie.Value
	if auth == "" {
		return "", ErrAuthzNoAuth
	}
	return auth, nil
}

func GetRequestToken(env map[string]interface{}) *jwt.Token {
	return env[ReqToken].(*jwt.Token)
}
