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
	"fmt"
	"strings"
	"time"
)

type Link struct {
	Uri      string            `json:"uri" bson:"-"`
	Expire   time.Time         `json:"expire,omitempty" bson:"expire"`
	Method   string            `json:"method,omitempty" bson:"-"`
	Header   map[string]string `json:"header,omitempty" bson:"-"`
	TenantID string            `json:"-" bson:"tenant_id"`
}

type UploadLink struct {
	ArtifactID string `json:"id" bson:"_id"`
	Link       `bson:"inline"`

	// Internal metadata
	IssuedAt  time.Time  `json:"-" bson:"issued_ts"`
	UpdatedTS time.Time  `json:"-" bson:"updated_ts"`
	Status    LinkStatus `json:"-" bson:"status"`
}

type LinkStatus uint32

const (
	LinkStatusPending LinkStatus = (iota << 4)
	LinkStatusProcessing
	LinkStatusCompleted
	LinkStatusAborted

	LinkStatusProcessedBit  = LinkStatus(1 << 7)
	LinkStatusProcessedMask = ^LinkStatus(LinkStatusProcessedBit)

	linkStatusPending    = "pending"
	linkStatusProcessing = "processing"
	linkStatusCompleted  = "completed"
	linkStatusAborted    = "aborted"
)

func (status LinkStatus) MarshalText() (b []byte, err error) {
	switch status & LinkStatusProcessedMask {
	case LinkStatusPending:
		b = []byte(linkStatusPending)
	case LinkStatusProcessing:
		b = []byte(linkStatusProcessing)
	case LinkStatusCompleted:
		b = []byte(linkStatusCompleted)
	case LinkStatusAborted:
		b = []byte(linkStatusAborted)
	default:
		err = fmt.Errorf("invalid link status value '%d'", status)
	}
	return b, err
}

func (status *LinkStatus) UnmarshalText(b []byte) error {
	var err error
	s := string(b)
	value := strings.ToLower(s)
	switch value {
	case linkStatusPending:
		*status = LinkStatusPending
	case linkStatusProcessing:
		*status = LinkStatusProcessing
	case linkStatusCompleted:
		*status = LinkStatusCompleted
	case linkStatusAborted:
		*status = LinkStatusAborted
	default:
		err = fmt.Errorf("invalid link status %q", s)
	}
	return err
}

func NewLink(uri string, expire time.Time) *Link {
	return &Link{
		Uri:    uri,
		Expire: expire,
	}
}
