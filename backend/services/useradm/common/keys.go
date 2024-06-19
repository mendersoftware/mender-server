// Copyright 2023 Northern.tech AS
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

package common

import (
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

const KeyIdZero = 0

var (
	ErrKeyIdNotFound  = errors.New("cant locate key by key id")
	ErrKeyIdCollision = errors.New("key id already loaded")
)

func KeyIdFromPath(privateKeyPath string, privateKeyFilenamePattern string) (keyId int) {
	fileName := filepath.Base(privateKeyPath)
	r, _ := regexp.Compile(privateKeyFilenamePattern)
	b := []byte(fileName)
	indices := r.FindAllSubmatchIndex(b, -1)
	keyId = KeyIdZero
	if len(indices) > 0 && len(indices[0]) > 3 {
		k, err := strconv.Atoi(string(b[indices[0][2]:indices[0][3]]))
		if err == nil {
			keyId = k
		}
	}
	return keyId
}
