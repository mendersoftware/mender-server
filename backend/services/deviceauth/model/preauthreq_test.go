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
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	pubKeyRSA = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4vkQl77vcl2bqFh+M4iH
pDMtBB1kuSwC/CHWmVm2jZx03ztCWx4evnLpT3ikBzIZNKYwng3HyUYPqzUAZ+ev
NXMai6saCLTq1k2apmadTaQte2c9MB62Zk+mXYGbO3FC5pihPjmQThP93Y6KKVtr
lyAbHN3B0YfScVENxa+V7TnadmOiwyHb0ej4G+a+mIsGz4CjvDet40+xxgt/XLwG
r9QiLrk7u9igG3Ub+kOE2my03lIAiPQicJsUlkI6p1G8fcejBEoGpsqqipUf7i3u
0SIVYT0kKrKSeHZY1DuzBuajxlT5Y1mojX2eaZXmqsVAOAJ0oU8oxTo013qKxiKj
2QIDAQAB
-----END PUBLIC KEY-----
`
)

func TestPreAuthReq(t *testing.T) {
	req := PreAuthReq{
		DeviceId:  "b5c98eda-ce7e-4b6e-9e12-3578e9d243f7",
		AuthSetId: "c7acb245-3ef4-4922-b28d-2d1ab8bfb930",
		IdData:    `{"mac":"00:11:22:33:44:55"}`,
		PubKey:    pubKeyRSA,
	}
	b, err := json.Marshal(req)
	assert.NoError(t, err)
	res, err := ParsePreAuthReq(bytes.NewReader(b))
	if assert.NoError(t, err) {
		assert.Equal(t, req, *res)
	}

	// Parse malformed request
	_, err = ParsePreAuthReq(bytes.NewReader([]byte("foobar")))
	assert.Error(t, err)

	// Parse invalid request
	req.PubKey = "<imagine a public key here>"
	b, _ = json.Marshal(req)
	_, err = ParsePreAuthReq(bytes.NewReader(b))
	assert.EqualError(t, err, "cannot decode public key")

	// Request invalid
	req.DeviceId = ""
	err = req.Validate()
	assert.EqualError(t, err, "device_id: cannot be blank.")
}
