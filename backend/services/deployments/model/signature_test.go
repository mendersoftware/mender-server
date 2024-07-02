// Copyright 2021 Northern.tech AS
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

package model

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRequestSignatureValidate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Request *RequestSignature

		Error error // Validation error
	}{{
		Name: "ok",

		Request: func() *RequestSignature {
			req, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			sig := NewRequestSignature(req, []byte("test"))
			sig.SetExpire(time.Now().Add(time.Hour))
			sig.PresignURL()
			return sig
		}(),
	}, {
		Name: "error, missing signature",

		Request: func() *RequestSignature {
			req, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			sig := NewRequestSignature(req, []byte("test"))
			return sig
		}(),
		Error: errors.Errorf(
			"%s: required key is missing; %s: required key is missing.",
			ParamExpire, ParamSignature,
		),
	}, {
		Name: "error, malformed timestamp",

		Request: func() *RequestSignature {
			req, _ := http.NewRequest(
				http.MethodGet, fmt.Sprintf(
					"https://localhost?%s=foobar&%s=barbaz",
					ParamExpire, ParamSignature,
				), nil,
			)
			sig := NewRequestSignature(req, []byte("test"))
			return sig
		}(),
		Error: errors.Errorf(
			"parameter '%s' is not a valid timestamp", ParamExpire,
		),
	}, {
		Name: "error, timestamp expired",

		Request: func() *RequestSignature {
			req, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			sig := NewRequestSignature(req, []byte("test"))
			sig.SetExpire(time.Now().Add(-time.Hour))
			sig.PresignURL()
			return sig
		}(),
		Error: ErrLinkExpired,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := tc.Request.Validate()
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, tc.Request.VerifyHMAC256())
			}
		})
	}
}
