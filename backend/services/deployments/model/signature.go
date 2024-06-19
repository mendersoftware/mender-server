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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	ParamExpire    = "x-men-expire"
	ParamSignature = "x-men-signature"
	ParamTenantID  = "tenant_id"
)

var ErrLinkExpired = errors.New("URL expired")

type RequestSignature struct {
	*http.Request
	Secret []byte
}

func NewRequestSignature(req *http.Request, secret []byte) *RequestSignature {
	return &RequestSignature{
		Request: req,
		Secret:  secret,
	}
}

func (sig *RequestSignature) SetExpire(expire time.Time) {
	q := sig.URL.Query()
	q.Set(ParamExpire, expire.UTC().Format(time.RFC3339))
	sig.URL.RawQuery = q.Encode()
}

// Validate validates the request parameters - assumes that the signature is
// already signed.
func (sig *RequestSignature) Validate() error {
	q := sig.URL.Query()
	if err := validation.Validate(q, validation.Map(
		validation.Key(ParamExpire, validation.Required),
		validation.Key(ParamSignature, validation.Required),
	).AllowExtraKeys()); err != nil {
		return err
	}
	ts, err := time.Parse(time.RFC3339, q.Get(ParamExpire))
	if err != nil {
		return errors.Errorf("parameter '%s' is not a valid timestamp", ParamExpire)
	}
	if time.Now().After(ts) {
		return ErrLinkExpired
	}
	return nil
}

// PresignURL generates and assign the request signature parameter and returning
// the resulting URL.
func (sig *RequestSignature) PresignURL() string {
	signature := sig.HMAC256()
	signature64 := base64.RawURLEncoding.EncodeToString(signature)

	q := sig.URL.Query()
	q.Set(ParamSignature, signature64)
	sig.URL.RawQuery = q.Encode()
	return sig.URL.String()
}

func (sig *RequestSignature) Bytes() []byte {
	// Bytes returns the byte digest for the HMAC256
	// The format is similar to s3 signed request with
	// <Method>\n<Canonical URI>\n<Canonical parameters>\n[<Canonical headers>]\n
	q := sig.URL.Query()
	return []byte(fmt.Sprintf(
		"%s\n%s\n%s=%s\n%s=%s\n",
		sig.Method, sig.URL.Path,
		ParamExpire, q.Get(ParamExpire),
		ParamTenantID, q.Get(ParamTenantID),
	))
}

// VerifyHMAC256 verifies the request signature with the parameter.
func (sig *RequestSignature) VerifyHMAC256() bool {
	//nolint:errcheck
	q := sig.URL.Query()
	sign, _ := base64.RawURLEncoding.
		DecodeString(q.Get(ParamSignature))
	return hmac.Equal(sig.HMAC256(), sign)
}

//nolint:errcheck
func (sig *RequestSignature) HMAC256() []byte {
	hash := hmac.New(sha256.New, sig.Secret)
	hash.Write(sig.Bytes())
	return hash.Sum(nil)
}
