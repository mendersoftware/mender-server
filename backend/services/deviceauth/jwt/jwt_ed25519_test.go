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
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/oid"
)

func TestNewJWTHandlerEd25519(t *testing.T) {
	privKey := loadEd25519PrivKey("./testdata/ed25519.pem", t)
	jwtHandler := NewJWTHandlerEd25519(privKey)

	assert.NotNil(t, jwtHandler)
}

func TestJWTHandlerEd25519GenerateToken(t *testing.T) {
	testCases := map[string]struct {
		privKey      *ed25519.PrivateKey
		claims       Claims
		expiresInSec int64
	}{
		"ok": {
			privKey: loadEd25519PrivKey("./testdata/ed25519.pem", t),
			claims: Claims{
				Issuer:  "Mender",
				Subject: oid.NewUUIDv5("foo"),
				ExpiresAt: Time{
					Time: time.Now().Add(time.Hour),
				},
			},
			expiresInSec: 3600,
		},
		"ok, with tenant": {
			privKey: loadEd25519PrivKey("./testdata/ed25519.pem", t),
			claims: Claims{
				Issuer:  "Mender",
				Subject: oid.NewUUIDv5("foo"),
				ExpiresAt: Time{
					Time: time.Now().Add(time.Hour),
				},
				Tenant: "foobar",
			},
			expiresInSec: 3600,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		jwtHandler := NewJWTHandlerEd25519(tc.privKey)

		raw, err := jwtHandler.ToJWT(&Token{
			Claims: tc.claims,
		})
		assert.NoError(t, err)

		parsed := parseGeneratedTokenEd25519(t, string(raw), tc.privKey)
		if assert.NotNil(t, parsed) {
			mc := parsed.Claims.(jwtgo.MapClaims)
			assert.Equal(t, tc.claims.Issuer, mc["iss"])
			assert.Equal(t, tc.claims.Subject.String(), mc["sub"])
			if tc.claims.Tenant != "" {
				assert.Equal(t, tc.claims.Tenant, mc["mender.tenant"])
			} else {
				assert.Nil(t, mc["mender.tenant"])
			}
		}
	}
}

func TestJWTHandlerEd25519FromJWT(t *testing.T) {

	key := loadEd25519PrivKey("./testdata/ed25519.pem", t)

	testCases := map[string]struct {
		privKey *ed25519.PrivateKey

		inToken string

		outToken Token
		outErr   error
	}{
		"ok (all claims)": {
			privKey: key,

			inToken: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdG" +
				"kiOiJiOTQ3NTMzNi1kZGU2LTU0OTctODA0NC01MWFhOW" +
				"RkYzAyZjgiLCJzdWIiOiJiY2E5NWFkYi1iNWYxLTU2NG" +
				"YtOTZhNy02MzU1YzUyZDFmYTciLCJhdWQiOiJNZW5kZX" +
				"IiLCJzY3AiOiJtZW5kZXIuKiIsImlzcyI6Ik1lbmRlci" +
				"IsImV4cCI6NDE0NzQ4MzY0NywiaWF0IjoxMjM0NTY3LC" +
				"JuYmYiOjEyMzQ1Njc4LCJtZW5kZXIudHJpYWwiOmZhbH" +
				"NlfQ.eOnpurEYseItJXycyjOyfTO-RI_MCSF1e79HG63" +
				"HzVoR2xLzrA044hQ_pUneqG1V30h67EhWZY1wspqBay-" +
				"Cw",

			outToken: Token{
				Claims: Claims{
					ID:       oid.NewUUIDv5("someid"),
					Subject:  oid.NewUUIDv5("foo"),
					Audience: "Mender",
					ExpiresAt: Time{
						Time: time.Unix(4147483647, 0),
					},
					IssuedAt: Time{
						Time: time.Unix(1234567, 0),
					},
					Issuer: "Mender",
					NotBefore: Time{
						Time: time.Unix(12345678, 0),
					},
					Scope: "mender.*",
				},
			},
		},
		"ok (some claims)": {
			privKey: key,

			inToken: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdG" +
				"kiOiJiOTQ3NTMzNi1kZGU2LTU0OTctODA0NC01MWFhOW" +
				"RkYzAyZjgiLCJzdWIiOiJiY2E5NWFkYi1iNWYxLTU2NG" +
				"YtOTZhNy02MzU1YzUyZDFmYTciLCJzY3AiOiJtZW5kZX" +
				"IudXNlcnMuaW5pdGlhbC5jcmVhdGUiLCJpc3MiOiJNZW" +
				"5kZXIiLCJleHAiOjQxNDc0ODM2NDcsImlhdCI6MTIzND" +
				"U2NywibmJmIjoxMjM0NTY3OCwibWVuZGVyLnRyaWFsIj" +
				"pmYWxzZX0.M2TiIXKt5vVYVznlzACUkD_PQCnfhedg3r" +
				"LpLAge3wI9Xq22t2KL0nc2c8GhQWXVV40M73zwf5p8rn" +
				"42PdGvCg",

			outToken: Token{
				Claims: Claims{
					ID:      oid.NewUUIDv5("someid"),
					Subject: oid.NewUUIDv5("foo"),
					ExpiresAt: Time{
						Time: time.Unix(4147483647, 0),
					},
					IssuedAt: Time{
						Time: time.Unix(1234567, 0),
					},
					NotBefore: Time{
						Time: time.Unix(12345678, 0),
					},
					Issuer: "Mender",
					Scope:  "mender.users.initial.create",
				},
			},
		},
		"error - token invalid": {
			privKey: key,

			inToken: "1234123412341234",

			outToken: Token{},
			outErr:   ErrTokenInvalid,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		jwtHandler := NewJWTHandlerEd25519(tc.privKey)

		token, err := jwtHandler.FromJWT(tc.inToken)
		if tc.outErr == nil {
			assert.NoError(t, err)
			assert.Equal(t, tc.outToken, *token)
		} else {
			assert.EqualError(t, tc.outErr, err.Error())
		}
	}
}

func TestJWTHandlerEd25519Validate(t *testing.T) {
	key := loadEd25519PrivKey("./testdata/ed25519.pem", t)

	testCases := map[string]struct {
		privKey *ed25519.PrivateKey

		inToken string

		outErr error
	}{
		"ok (all claims)": {
			privKey: key,

			inToken: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdG" +
				"kiOiJiOTQ3NTMzNi1kZGU2LTU0OTctODA0NC01MWFhOW" +
				"RkYzAyZjgiLCJzdWIiOiJiY2E5NWFkYi1iNWYxLTU2NG" +
				"YtOTZhNy02MzU1YzUyZDFmYTciLCJhdWQiOiJNZW5kZX" +
				"IiLCJzY3AiOiJtZW5kZXIuKiIsImlzcyI6Ik1lbmRlci" +
				"IsImV4cCI6NDE0NzQ4MzY0NywiaWF0IjoxMjM0NTY3LC" +
				"JuYmYiOjEyMzQ1Njc4LCJtZW5kZXIudHJpYWwiOmZhbH" +
				"NlfQ.eOnpurEYseItJXycyjOyfTO-RI_MCSF1e79HG63" +
				"HzVoR2xLzrA044hQ_pUneqG1V30h67EhWZY1wspqBay-" +
				"FCw",
		},
		"ok (some claims)": {
			privKey: key,

			inToken: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdG" +
				"kiOiJiOTQ3NTMzNi1kZGU2LTU0OTctODA0NC01MWFhOW" +
				"RkYzAyZjgiLCJzdWIiOiJiY2E5NWFkYi1iNWYxLTU2NG" +
				"YtOTZhNy02MzU1YzUyZDFmYTciLCJzY3AiOiJtZW5kZX" +
				"IudXNlcnMuaW5pdGlhbC5jcmVhdGUiLCJpc3MiOiJNZW" +
				"5kZXIiLCJleHAiOjQxNDc0ODM2NDcsImlhdCI6MTIzND" +
				"U2NywibmJmIjotNjIxMzU1OTY4MDAsIm1lbmRlci50cm" +
				"lhbCI6ZmFsc2V9.C2-Ws5A_Pm9H7v_TCVDXCgREWBeYZ" +
				"HCRuZmTFxBuv7TJhfGvGusu-SB6bZ3N9OPJYwfConmq1" +
				"YprlfFpAAgmDQ",
		},
		"error - bad claims": {
			privKey: key,

			inToken: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdG" +
				"kiOiJiOTQ3NTMzNi1kZGU2LTU0OTctODA0NC01MWFhOW" +
				"RkYzAyZjgiLCJzdWIiOiJiY2E5NWFkYi1iNWYxLTU2NG" +
				"YtOTZhNy02MzU1YzUyZDFmYTciLCJhdWQiOiJNZW5kZX" +
				"IiLCJzY3AiOiJtZW5kZXIuKiIsImV4cCI6NDE0NzQ4Mz" +
				"Y0NywiaWF0IjoxMjM0NTY3LCJuYmYiOjEyMzQ1Njc4LC" +
				"JtZW5kZXIudHJpYWwiOmZhbHNlfQ.T4PVYJvRSusq7MZ" +
				"5XaOo6mLW9GDKdqdWO8NUZOZZ-KJ69d1UDbKWFSs9PPx" +
				"cNwS5a0j8iiTA6m6-YW0nEvLWAg",

			outErr: errors.New("jwt: token invalid"),
		},
		"error - bad signature": {
			privKey: key,

			inToken: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdG" +
				"kiOiJiOTQ3NTMzNi1kZGU2LTU0OTctODA0NC01MWFhOW" +
				"RkYzAyZjgiLCJzdWIiOiJiY2E5NWFkYi1iNWYxLTU2NG" +
				"YtOTZhNy02MzU1YzUyZDFmYTciLCJhdWQiOiJNZW5kZX" +
				"IiLCJzY3AiOiJtZW5kZXIuKiIsImlzcyI6Ik1lbmRlci" +
				"IsImV4cCI6NDE0NzQ4MzY0NywiaWF0IjoxMjM0NTY3LC" +
				"JuYmYiOjEyMzQ1Njc4LCJtZW5kZXIudHJpYWwiOmZhbH" +
				"NlfQ.eOnpurEYseItJXycyjOyfTO-RI_MCSF1e79HG63" +
				"HzVoR2xLzrA044hQ_pUneqG1V30h67EhWZY1wspqBay-" +
				"XXX",

			outErr: ErrTokenInvalid,
		},
		"error - token invalid": {
			privKey: key,

			inToken: "1234123412341234",

			outErr: errors.New("token contains an invalid number of segments"),
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		jwtHandler := NewJWTHandlerEd25519(tc.privKey)

		err := jwtHandler.Validate(tc.inToken)
		if tc.outErr == nil {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
			assert.EqualError(t, tc.outErr, err.Error())
		}
	}
}

func loadEd25519PrivKey(path string, t *testing.T) *ed25519.PrivateKey {
	pemData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load key: %v", err)
	}

	block, _ := pem.Decode(pemData)
	assert.Equal(t, block.Type, pemHeaderPKCS8)

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	assert.NoError(t, err)

	retKey := key.(ed25519.PrivateKey)
	return &retKey
}

func parseGeneratedTokenEd25519(t *testing.T, token string, key *ed25519.PrivateKey) *jwtgo.Token {
	tokenParsed, err := jwtgo.Parse(token, func(token *jwtgo.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtgo.SigningMethodEd25519); !ok {
			return nil, errors.New("Unexpected signing method: " + token.Method.Alg())
		}
		return key.Public(), nil
	})

	if err != nil {
		t.Fatalf("can't parse token: %s", err.Error())
	}

	return tokenParsed
}
