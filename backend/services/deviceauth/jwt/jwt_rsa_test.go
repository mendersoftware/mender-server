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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v4"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewJWTHandlerRS256(t *testing.T) {
	privKey := loadRSAPrivKey("./testdata/rsa.pem", t)
	jwtHandler := NewJWTHandlerRS256(privKey)

	assert.NotNil(t, jwtHandler)
}

func TestJWTHandlerRS256GenerateToken(t *testing.T) {
	testCases := map[string]struct {
		privKey      *rsa.PrivateKey
		claims       Claims
		expiresInSec int64
	}{
		"ok": {
			privKey: loadRSAPrivKey("./testdata/rsa.pem", t),
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
			privKey: loadRSAPrivKey("./testdata/rsa.pem", t),
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
		jwtHandler := NewJWTHandlerRS256(tc.privKey)

		raw, err := jwtHandler.ToJWT(&Token{
			Claims: tc.claims,
		})
		assert.NoError(t, err)

		parsed := parseGeneratedTokenRS256(t, string(raw), tc.privKey)
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

func TestJWTHandlerRS256FromJWT(t *testing.T) {

	key := loadRSAPrivKey("./testdata/rsa.pem", t)

	testCases := map[string]struct {
		privKey         *rsa.PrivateKey
		fallbackPrivKey *rsa.PrivateKey

		inToken string

		outToken Token
		outErr   error
	}{
		"ok (all claims)": {
			privKey: key,

			inToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdW" +
				"QiOiJNZW5kZXIiLCJleHAiOjQxNDc0ODM2NDcsImp0aS" +
				"I6ImI5NDc1MzM2LWRkZTYtNTQ5Ny04MDQ0LTUxYWE5ZG" +
				"RjMDJmOCIsImlhdCI6MTIzNDU2NywiaXNzIjoiTWVuZG" +
				"VyIiwibmJmIjoxMjM0NTY3OCwic3ViIjoiYmNhOTVhZG" +
				"ItYjVmMS01NjRmLTk2YTctNjM1NWM1MmQxZmE3Iiwic2" +
				"NwIjoibWVuZGVyLioifQ.bEvw5q8Ohf_3DOw77EDeOTq" +
				"99_JKUDz1YhCpJ5NaKPtMGmTksZIDoc6vk_lFyrPWzXm" +
				"lmbiCB8bEYI2-QGe2XwTnCkWm8YPxTFJw3UriZLt-5Pw" +
				"cEBDPG8FqTMtFaRjcbH-E7W7m_KT_Tm6fm93Vvqv_z6a" +
				"JiCOL7e16sLC0DQCJ2nZ4OleztNDkP4rCOgtBuSbhOaR" +
				"E_zhSsLf2Dj4Dlt5DVqDd8kqUBmA9-Sn9m5BeCUs023_" +
				"W4FWOH4NJpqyxjO0jXGoncvZu0AYPqHSbJ9J6Oucvc4y" +
				"lpbrCHN4diQ39s2egWzRbrSORsr-IL3hb1PZTINzLlQE" +
				"6Wol2S-I8ag",

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

			inToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleH" +
				"AiOjQxNDc0ODM2NDcsImp0aSI6ImI5NDc1MzM2LWRkZT" +
				"YtNTQ5Ny04MDQ0LTUxYWE5ZGRjMDJmOCIsImlhdCI6MT" +
				"IzNDU2NywiaXNzIjoiTWVuZGVyIiwic3ViIjoiYmNhOT" +
				"VhZGItYjVmMS01NjRmLTk2YTctNjM1NWM1MmQxZmE3Ii" +
				"wic2NwIjoibWVuZGVyLnVzZXJzLmluaXRpYWwuY3JlYX" +
				"RlIn0.qzW1QfnvfB384DfOyX6LC4jsTSVEWwsyb-vSeA" +
				"ebfHdJquX2BfQ6_1ZGtqyCC7mOhMrXeJv1gmprpkOxKw" +
				"hPBexS-U1gOc_aO7Oi7uPl1HQRhMw9SM2QamOOVGmLi5" +
				"1uVg9ZEQhvnN7s-w4girnmGyhnPWV58CorJtW4t1Dgyr" +
				"6fG_v8wtrGt-rMb7uMLmEQMjIqcUBa6mlU1sVBEPTeGb" +
				"KvR6kSJ727UW91y7krTcQUdNN4rv2CfG7ETlPsrUgMvr" +
				"GUPqoq_ygbLX3kDZveVzTE2CQdI7PpAO14UZQxRBfff5" +
				"ewyW4P0ulYRj0mPF5NmsHwbADoAjILoA5uSWW9Dg",

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
		jwtHandler := NewJWTHandlerRS256(tc.privKey)

		token, err := jwtHandler.FromJWT(tc.inToken)
		if tc.outErr == nil {
			assert.NoError(t, err)
			assert.Equal(t, tc.outToken, *token)
		} else {
			assert.EqualError(t, tc.outErr, err.Error())
		}
	}
}

func TestJWTHandlerRS256Validate(t *testing.T) {
	key := loadRSAPrivKey("./testdata/rsa.pem", t)

	testCases := map[string]struct {
		privKey *rsa.PrivateKey

		inToken string

		outErr error
	}{
		"ok (all claims)": {
			privKey: key,

			inToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdW" +
				"QiOiJNZW5kZXIiLCJleHAiOjQxNDc0ODM2NDcsImp0aS" +
				"I6ImI5NDc1MzM2LWRkZTYtNTQ5Ny04MDQ0LTUxYWE5ZG" +
				"RjMDJmOCIsImlhdCI6MTIzNDU2NywiaXNzIjoiTWVuZG" +
				"VyIiwibmJmIjoxMjM0NTY3OCwic3ViIjoiYmNhOTVhZG" +
				"ItYjVmMS01NjRmLTk2YTctNjM1NWM1MmQxZmE3Iiwic2" +
				"NwIjoibWVuZGVyLioifQ.bEvw5q8Ohf_3DOw77EDeOTq" +
				"99_JKUDz1YhCpJ5NaKPtMGmTksZIDoc6vk_lFyrPWzXm" +
				"lmbiCB8bEYI2-QGe2XwTnCkWm8YPxTFJw3UriZLt-5Pw" +
				"cEBDPG8FqTMtFaRjcbH-E7W7m_KT_Tm6fm93Vvqv_z6a" +
				"JiCOL7e16sLC0DQCJ2nZ4OleztNDkP4rCOgtBuSbhOaR" +
				"E_zhSsLf2Dj4Dlt5DVqDd8kqUBmA9-Sn9m5BeCUs023_" +
				"W4FWOH4NJpqyxjO0jXGoncvZu0AYPqHSbJ9J6Oucvc4y" +
				"lpbrCHN4diQ39s2egWzRbrSORsr-IL3hb1PZTINzLlQE" +
				"6Wol2S-I8ag",
		},
		"ok (some claims)": {
			privKey: key,

			inToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleH" +
				"AiOjQxNDc0ODM2NDcsImp0aSI6ImI5NDc1MzM2LWRkZT" +
				"YtNTQ5Ny04MDQ0LTUxYWE5ZGRjMDJmOCIsImlhdCI6MT" +
				"IzNDU2NywiaXNzIjoiTWVuZGVyIiwic3ViIjoiYmNhOT" +
				"VhZGItYjVmMS01NjRmLTk2YTctNjM1NWM1MmQxZmE3Ii" +
				"wic2NwIjoibWVuZGVyLnVzZXJzLmluaXRpYWwuY3JlYX" +
				"RlIn0.qzW1QfnvfB384DfOyX6LC4jsTSVEWwsyb-vSeA" +
				"ebfHdJquX2BfQ6_1ZGtqyCC7mOhMrXeJv1gmprpkOxKw" +
				"hPBexS-U1gOc_aO7Oi7uPl1HQRhMw9SM2QamOOVGmLi5" +
				"1uVg9ZEQhvnN7s-w4girnmGyhnPWV58CorJtW4t1Dgyr" +
				"6fG_v8wtrGt-rMb7uMLmEQMjIqcUBa6mlU1sVBEPTeGb" +
				"KvR6kSJ727UW91y7krTcQUdNN4rv2CfG7ETlPsrUgMvr" +
				"GUPqoq_ygbLX3kDZveVzTE2CQdI7PpAO14UZQxRBfff5" +
				"ewyW4P0ulYRj0mPF5NmsHwbADoAjILoA5uSWW9Dg",
		},

		"error - bad claims": {
			privKey: key,

			inToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGk" +
				"iOm51bGwsInN1YiI6ImJjYTk1YWRiLWI1ZjEtNTY0Zi0" +
				"5NmE3LTYzNTVjNTJkMWZhNyIsImV4cCI6MTY5NjQxNjU" +
				"2NCwiaWF0IjotNjIxMzU1OTY4MDAsIm5iZiI6LTYyMTM" +
				"1NTk2ODAwLCJtZW5kZXIudHJpYWwiOmZhbHNlfQ.LNRO" +
				"1CzYVkOqU_-ikNva-MvFyvZTVLbR8irmecbPKPij-6cv" +
				"h_DymOdbtupRCpABq2XFfLGAz68AOqGhc0Utp_AL-EY7" +
				"kSH-QbPVdlFvnZO_T-gPHxOY2wNoZqnyusr-cpiRR413" +
				"lySS5t5ZPsghFtlCSFHITdZ11sin79C1JJxd3cUnhjXj" +
				"P-wL7YJmsfFR9KfSL4AEPtpDsQ98gPhcnqPRCBuLSFcU" +
				"d3_w-pbc7PkbM0A_nO2jrwCJCaHvjMMvL9FHIZ2-xfUW" +
				"qDB13KkPo0BrVwHLvhykLlCuhshNaNugzH0Tb4djrM__" +
				"NCKofdozu3DowLLjesXp7oIYWRAKUQ",

			outErr: errors.New("jwt: token invalid"),
		},
		"error - bad signature": {
			privKey: key,

			inToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdW" +
				"QiOiJNZW5kZXIiLCJleHAiOjQxNDc0ODM2NDcsImp0aS" +
				"I6ImI5NDc1MzM2LWRkZTYtNTQ5Ny04MDQ0LTUxYWE5ZG" +
				"RjMDJmOCIsImlhdCI6MTIzNDU2NywiaXNzIjoiTWVuZG" +
				"VyIiwibmJmIjoxMjM0NTY3OCwic3ViIjoiYmNhOTVhZG" +
				"ItYjVmMS01NjRmLTk2YTctNjM1NWM1MmQxZmE3Iiwic2" +
				"NwIjoibWVuZGVyLioifQ.bEvw5q8Ohf_3DOw77EDeOTq" +
				"99_JKUDz1YhCpJ5NaKPtMGmTksZIDoc6vk_lFyrPWzXm" +
				"lmbiCB8bEYI2-QGe2XwTnCkWm8YPxTFJw3UriZLt-5Pw" +
				"cEBDPG8FqTMtFaRjcbH-E7W7m_KT_Tm6fm93Vvqv_z6a" +
				"JiCOL7e16sLC0DQCJ2nZ4OleztNDkP4rCOgtBuSbhOaR" +
				"E_zhSsLf2Dj4Dlt5DVqDd8kqUBmA9-Sn9m5BeCUs023_" +
				"W4FWOH4NJpqyxjO0jXGoncvZu0AYPqHSbJ9J6Oucvc4y" +
				"lpbrCHN4diQ39s2egWzRbrSORsr-IL3hb1PZTINzLlQE" +
				"6Wol2S-I8XX",
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
		jwtHandler := NewJWTHandlerRS256(tc.privKey)

		err := jwtHandler.Validate(tc.inToken)
		if tc.outErr == nil {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
			assert.EqualError(t, tc.outErr, err.Error())
		}
	}
}

func loadRSAPrivKey(path string, t *testing.T) *rsa.PrivateKey {
	pemData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load key: %v", err)
	}

	block, _ := pem.Decode(pemData)
	assert.Equal(t, block.Type, pemHeaderPKCS1)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	assert.NoError(t, err)

	return key
}

func parseGeneratedTokenRS256(t *testing.T, token string, key *rsa.PrivateKey) *jwtgo.Token {
	tokenParsed, err := jwtgo.Parse(token, func(token *jwtgo.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtgo.SigningMethodRSA); !ok {
			return nil, errors.New("Unexpected signing method: " + token.Method.Alg())
		}
		return &key.PublicKey, nil
	})

	if err != nil {
		t.Fatalf("can't parse token: %s", err.Error())
	}

	return tokenParsed
}
