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
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/useradm/common"
)

// JWTHandlerEd25519 is an Ed25519-specific JWTHandler
type JWTHandlerEd25519 struct {
	privKey map[int]*ed25519.PrivateKey
}

func NewJWTHandlerEd25519(privKey *ed25519.PrivateKey, keyId int) *JWTHandlerEd25519 {
	return &JWTHandlerEd25519{
		privKey: map[int]*ed25519.PrivateKey{keyId: privKey},
	}
}

func (j *JWTHandlerEd25519) ToJWT(token *Token) (string, error) {
	//generate
	jt := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &token.Claims)
	jt.Header["kid"] = token.KeyId
	if _, exists := j.privKey[token.KeyId]; !exists {
		return "", common.ErrKeyIdNotFound
	}
	//sign
	data, err := jt.SignedString(j.privKey[token.KeyId])
	return data, err
}

func (j *JWTHandlerEd25519) FromJWT(tokstr string) (*Token, error) {
	jwttoken, err := jwt.ParseWithClaims(tokstr, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			keyId := common.KeyIdZero
			if _, ok := token.Header["kid"]; ok {
				if _, isFloat := token.Header["kid"].(float64); isFloat {
					keyId = int(token.Header["kid"].(float64))
				}
				if _, isInt := token.Header["kid"].(int64); isInt {
					keyId = int(token.Header["kid"].(int64))
				}
				if _, isInt := token.Header["kid"].(int); isInt {
					keyId = token.Header["kid"].(int)
				}
			}
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, errors.New("unexpected signing method: " + token.Method.Alg())
			}
			if _, exists := j.privKey[keyId]; !exists {
				return nil, errors.New("cannot find the key with id " + strconv.Itoa(keyId))
			}
			return j.privKey[keyId].Public(), nil
		},
	)

	if err == nil {
		token := Token{}
		if claims, ok := jwttoken.Claims.(*Claims); ok && jwttoken.Valid {
			token.Claims = *claims
			return &token, nil
		}
	}

	return nil, ErrTokenInvalid
}
