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

package model

import (
	"errors"
	"time"
)

// Type info structure
type ArtifactUpdateTypeInfo struct {
	Type *string `json:"type" valid:"required"`
}

// Update file structure
type UpdateFile struct {
	// Image name
	Name string `json:"name" valid:"required"`

	// Image file checksum
	Checksum string `json:"checksum" valid:"optional"`

	// Image size
	Size int64 `json:"size" valid:"optional"`

	// Date build
	Date *time.Time `json:"date" valid:"optional"`
}

// Update structure
type Update struct {
	TypeInfo ArtifactUpdateTypeInfo `json:"type_info" valid:"required"`
	Files    []UpdateFile           `json:"files"`
	MetaData interface{}            `json:"meta_data,omitempty" valid:"optional"`
}

func (u Update) Match(update Update) bool {
	if len(u.Files) != len(update.Files) {
		return false
	}

	lFiles := make(map[string]UpdateFile, len(u.Files))
	for i, f := range u.Files {
		lFiles[f.Name] = u.Files[i]
	}
	for _, f := range update.Files {
		if _, ok := lFiles[f.Name]; !ok {
			return false
		}
	}
	return true
}

const maxUpdateFiles = 1024

func (u Update) Validate() error {
	if len(u.Files) > maxUpdateFiles {
		return errors.New("too large update files array")
	}

	return nil
}
