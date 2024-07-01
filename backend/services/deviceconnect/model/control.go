// Copyright 2021 Northern.tech AS
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
	"time"

	"github.com/google/uuid"
)

type ControlData struct {
	ID        uuid.UUID `json:"-" bson:"_id"`
	SessionID string    `json:"session_id" bson:"session_id"`
	Control   []byte    `json:"control" bson:"control"`
	CreatedTs time.Time `json:"created_ts" bson:"created_ts"`
	ExpireTs  time.Time `json:"expire_ts" bson:"expire_ts"`
}
