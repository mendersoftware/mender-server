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
package keys

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRsaPrivateKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		privKeyPath string
		err         string
	}{
		{
			privKeyPath: "testdata/private.pem",
			err:         "",
		},
		{
			privKeyPath: "wrong_path",
			err:         ErrMsgPrivKeyReadFailed + ": open wrong_path: no such file or directory",
		},
		{
			privKeyPath: "testdata/private_broken.pem",
			err:         ErrMsgPrivKeyNotPEMEncoded,
		},
		{
			privKeyPath: "testdata/public.pem",
			err:         "unknown server private key type; got: PUBLIC KEY, want: RSA PRIVATE KEY",
		},
	}

	for i, tc := range testCases {
		tcVar := tc
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()
			key, err := LoadRSAPrivate(tcVar.privKeyPath)
			if tcVar.err != "" {
				assert.EqualError(t, err, tcVar.err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
			}
		})
	}
}
