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

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

// JWTHandlerRS256 is an RS256-specific JWTHandler
type JWTHandlerRS256 struct {
	privKey *rsa.PrivateKey
}

func NewJWTHandlerRS256(privKey *rsa.PrivateKey) *JWTHandlerRS256 {
	return &JWTHandlerRS256{
		privKey: privKey,
	}
}

func (j *JWTHandlerRS256) ToJWT(token *Token) (string, error) {
	//generate
	jt := jwt.NewWithClaims(jwt.SigningMethodRS256, &token.Claims)

	//sign
	data, err := jt.SignedString(j.privKey)
	return data, err
}

func (j *JWTHandlerRS256) FromJWT(tokstr string) (*Token, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	jwttoken, _, err := parser.ParseUnverified(tokstr, &Claims{})
	if err == nil {
		token := Token{}
		if claims, ok := jwttoken.Claims.(*Claims); ok {
			token.Claims = *claims
			return &token, nil
		}
	}

	return nil, ErrTokenInvalid
}

func (j *JWTHandlerRS256) Validate(tokstr string) error {
	jwttoken, err := jwt.ParseWithClaims(tokstr, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method: " + token.Method.Alg())
			}
			return &j.privKey.PublicKey, nil
		},
	)

	// our Claims return Mender-specific validation errors
	// go-jwt will wrap them in a generic ValidationError - unwrap and return directly
	if jwttoken != nil && !jwttoken.Valid {
		return ErrTokenInvalid
	} else if err != nil {
		err, ok := err.(*jwt.ValidationError)
		if ok && err.Inner != nil {
			return err.Inner
		} else {
			return err
		}
	}

	return nil
}
