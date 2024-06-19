// Copyright 2020 Northern.tech AS
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

package utils

import (
	"crypto"
	"crypto/dsa"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	test "github.com/mendersoftware/mender-server/services/deviceauth/utils/testing"
)

func TestVerifyAuthReqSign(t *testing.T) {
	t.Parallel()

	/* Initialize a DSA key-pair for testing */
	var DSATestKey = &dsa.PrivateKey{}
	err := dsa.GenerateParameters(
		&DSATestKey.Parameters, rand.Reader, dsa.L1024N160,
	)
	if err != nil {
		panic(err)
	}
	err = dsa.GenerateKey(DSATestKey, rand.Reader)
	if err != nil {
		panic(err)
	}

	testCases := []struct {
		Name      string
		Content   string
		PrivKey   crypto.PrivateKey
		PubKey    crypto.PublicKey
		Signature []byte
		Err       string
	}{
		{
			Name: "OK, RSA",
			//correctly signed, matching keypair
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: test.DecodePrivKey([]byte(TestRSAPrivate)),
		},
		{
			Name: "Error: key-pair mismatch",
			//mismatched keypair
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: test.DecodePrivKey([]byte(TestRSAPrivatInvalid)),
			PubKey:  test.DecodePubKey([]byte(TestRSAPublic)),
			Err:     "verification failed: crypto/rsa: verification error",
		},
		{
			Name: "Error: invalid RSA signature",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: test.DecodePrivKey([]byte(TestRSAPrivate)),
			Signature: []byte(`
bVvcxjp9xckT33yYzyxE2ozX5Be8fwa7CobC8/0QJfbvJTwvxQm8GCU9bDLaE2Mr50LJj8YkXOPz
3mHize+CGrxBcf5vTSfySuDx4fnoybhyYSJBIyoJsmo0fghk7Bb1PgVV0UY8NVcAS0ziKzTxR4m4
DOrbMJKQIAUYYnX1xy4LX0EUUYGWFHZvOmH0L2tzLKlo9lQu+28PpaDVVp75ygn/yGNZ4mJeVsq0
e6qbGJPhLYhn4hC8euK//NvLWKbTokVJ9hvVWjY/so4jWaI3zWukcfkjYWzxv6lNY+hhfph413G3
5UDTlT6pt8iIknNKRwkYnODoeJ36AStisE+Byg==
`),
			Err: ErrMsgVerify + ": crypto/rsa: verification error",
		},
		{
			Name: "OK, ecdsa",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: test.DecodePrivKey([]byte(TestECDSAPrivate)),
		},
		{
			Name: "Error: invalid ECDSA signature",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: test.DecodePrivKey([]byte(TestECDSAPrivate)),
			Signature: []byte(`
MIGIAkIBpTxA1RZXcprHprcBNamyAK8/pvY6ZssbRaHSkdQp5WTqRY5QmSr3/Y86u7xAWYNjJeRY
C2lW2/fQafXcV+nJyJsCQgCDL/4r8S6ekh75Tx1EAKlEjXKbRzsABDIMTORVTk7f0ShGpWBbpSjH
1M2w7bpwWnLjq4FAodttsdooMzNqeOZxng==
`),
			Err: ErrMsgVerify,
		},
		{
			Name: "Error: invalid ECDSA signature format",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey:   test.DecodePrivKey([]byte(TestECDSAPrivate)),
			Signature: []byte(`Zm9vYmFyYmF6`),
			Err: ErrMsgVerify + ": asn1: structure error: " +
				"tags don't match (16 vs {class:1 tag:6 " +
				"length:111 isCompound:true}) {optional:false " +
				"explicit:false application:false " +
				"private:false defaultValue:<nil> tag:<nil> " +
				"stringType:0 timeType:0 set:false " +
				"omitEmpty:false}  @2",
		},
		{
			Name: "OK, ed25519",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: test.DecodePrivKey([]byte(TestED25519Private)),
		},
		{
			Name: "Error: invalid ED25519 signature",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: test.DecodePrivKey([]byte(TestED25519Private)),
			Signature: []byte(`
MIGIAkIBpTxA1RZXcprHprcBNamyAK8/pvY6ZssbRaHSkdQp5WTqRY5QmSr3/Y86u7xAWYNjJeRY
C2lW2/fQafXcV+nJyJsCQgCDL/4r8S6ekh75Tx1EAKlEjXKbRzsABDIMTORVTk7f0ShGpWBbpSjH
1M2w7bpwWnLjq4FAodttsdooMzNqeOZxng==
`),
			Err: ErrMsgVerify,
		},
		{
			Name: "Error: signature not PEM",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey:   test.DecodePrivKey([]byte(TestRSAPrivate)),
			Signature: []byte(`føøbårbæz`),
			Err:       ErrMsgVerify + ": illegal base64 data at input byte 1",
		},
		{
			Name: "Error: public key algorithm not supported",
			Content: `{
				"id_data": {"mac": "deadbeef"},
				"tenant_token": "token"
				"seq_no": 1
			}`,
			PrivKey: DSATestKey,
			PubKey:  DSATestKey.PublicKey,
			Signature: []byte(`
MIGIAkIBpTxA1RZXcprHprcBNamyAK8/pvY6ZssbRaHSkdQp5WTqRY5QmSr3/Y86u7xAWYNjJeRY
C2lW2/fQafXcV+nJyJsCQgCDL/4r8S6ekh75Tx1EAKlEjXKbRzsABDIMTORVTk7f0ShGpWBbpSjH
1M2w7bpwWnLjq4FAodttsdooMzNqeOZxng==
`),
			Err: ErrMsgVerify + ": public key algorithm (dsa.PublicKey) not supported",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			if tc.Signature == nil {
				tc.Signature = test.AuthReqSign(
					[]byte(tc.Content), tc.PrivKey, t,
				)
			}
			if tc.PubKey == nil {
				tc.PubKey = tc.PrivKey.(interface{ Public() crypto.PublicKey }).Public()
			}

			err := VerifyAuthReqSign(
				string(tc.Signature),
				tc.PubKey,
				[]byte(tc.Content),
			)

			if tc.Err != "" {
				assert.EqualError(t, err, tc.Err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParsePubKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		pubkey     string
		errMatcher func(t *testing.T, theErr error)
	}{
		"ok": {
			pubkey: TestRSAPublic,
		},
		"error, bad pem block": {
			pubkey: TestRSAPublicBadPEM,
			errMatcher: func(t *testing.T, err error) {
				if assert.Error(t, err) {
					assert.EqualError(t, err, "cannot decode public key")
				}
			},
		},
		"error, pem ok, but bad key content": {
			pubkey: TestRSAPublicBadContent,
			errMatcher: func(t *testing.T, err error) {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(),
						"cannot decode public key: "+
							"asn1: structure "+
							"error: tags don't match",
					)
				}
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %s", i), func(t *testing.T) {
			t.Parallel()

			key, err := ParsePubKey(tc.pubkey)

			if tc.errMatcher != nil {
				tc.errMatcher(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
			}
		})
	}
}

func TestSerializePubKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		pubKey crypto.PublicKey
		out    string
		err    error
	}{
		"ok": {
			pubKey: test.DecodePubKey([]byte(TestRSAPublic)),
			out:    TestRSAPublic,
		},
		"error, unrecognized key": {
			pubKey: test.DecodePubKey([]byte(TestDSAPublic)),
			err:    errors.New("unrecognized public key type"),
		},
		"error, corrupt public key struct": {
			pubKey: &rsa.PublicKey{},
			err:    errors.New("asn1: structure error: empty integer"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %s", i), func(t *testing.T) {
			t.Parallel()

			out, err := SerializePubKey(tc.pubKey)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.Equal(t, "", out)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.out, out)
			}
		})
	}
}
