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
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
)

func TestTokenMarshal(t *testing.T) {
	tok := &Token{
		Claims: Claims{
			ID:      oid.NewUUIDv5("foo"),
			Subject: oid.NewUUIDv5("valid-subject"),
			ExpiresAt: &Time{
				Time: time.Now().Add(time.Hour),
			},
		},
	}

	res, err := tok.MarshalJWT(func(toSign *Token) (string, error) {
		assert.Equal(t, tok, toSign)
		return "signed", nil
	})
	assert.Equal(t, []byte("signed"), res)
	assert.NoError(t, err)

	res, err = tok.MarshalJWT(func(toSign *Token) (string, error) {
		assert.Equal(t, tok, toSign)
		return "", errors.New("failed")
	})
	assert.Empty(t, res)
	assert.Error(t, err)
}

func TestTokenUnmarshal(t *testing.T) {
	tokin := []byte("some-fake-jwt")
	tok := &Token{
		Claims: Claims{
			ID:      oid.NewUUIDv5("foo"),
			Subject: oid.NewUUIDv5("valid-subject"),
			ExpiresAt: &Time{
				Time: time.Now().Add(time.Hour),
			},
		},
	}

	unTok := &Token{}

	err := unTok.UnmarshalJWT(tokin, func(toUnpack string) (*Token, error) {
		assert.Equal(t, string(tokin), toUnpack)
		return tok, nil
	})
	assert.Equal(t, unTok, tok)
	assert.NoError(t, err)

	unTok = &Token{}

	err = unTok.UnmarshalJWT(tokin, func(toUnpack string) (*Token, error) {
		assert.Equal(t, string(tokin), toUnpack)
		return nil, errors.New("failed")
	})
	assert.Error(t, err)

	unTok = &Token{}
	// make sure that the token is updated if UnpackFunc returns both
	// a token & an error
	err = unTok.UnmarshalJWT(tokin, func(toUnpack string) (*Token, error) {
		assert.Equal(t, string(tokin), toUnpack)
		return tok, errors.New("failed")
	})
	assert.Equal(t, unTok, tok)
	assert.Error(t, err)
}
