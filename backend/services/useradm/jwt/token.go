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
package jwt

import (
	"time"

	"github.com/pkg/errors"
)

// SignFunc will sign and encode token.
type SignFunc func(token *Token) (string, error)

// UnpackFunc will decode token
type UnpackFunc func(s string) (*Token, error)

// Token wrapper
type Token struct {
	Claims `bson:"inline"`
	// LastUsed is the token last usage timestamp.
	LastUsed *time.Time `json:"last_used,omitempty" bson:"last_used,omitempty"`
	// TokenName holds the name of the token
	TokenName *string `json:"name,omitempty" bson:"name,omitempty"`
	// KeyId is the field that corresponds to "kid" in the JWT Header, we use it
	// to identify the key which was used to sign the token. It allows the private key rotation
	// as long as you keep the private keys we can use the new ones (the highest id)
	// to issue new tokens, while verify the old tokens with the keys used to issue them
	KeyId int `json:"key_id,omitempty" bson:"key_id,omitempty"`
}

// MarshalJWT marshals Token into JWT comaptible format. `sign` provides means
// for generating a signed JWT token.
func (t *Token) MarshalJWT(sign SignFunc) ([]byte, error) {
	if sign == nil {
		panic("no signature helper")
	}

	signed, err := sign(t)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sign token")
	}
	return []byte(signed), nil
}

// UnmarshalJWT unmarshals raw JWT data into Token. UnpackFunc does the
// actual heavy-lifting of parsing and deserializing base64'ed JWT. Returns an
// error if `unpack` failed, however if `unpack` returns a token `t` will be
// updated as well (may happen if token is valid wrt. to structure & signature,
// but expired).
func (t *Token) UnmarshalJWT(raw []byte, unpack UnpackFunc) error {
	tok, err := unpack(string(raw))
	if tok != nil {
		*t = *tok
	}
	return err
}
