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
	"encoding/json"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pkg/errors"
)

var (
	ErrInvalidDeploymentLog = errors.New("invalid deployment log")
	ErrInvalidLogMessage    = errors.New("invalid log message")
)

type LogMessage struct {
	Timestamp *time.Time `json:"timestamp" valid:"required"`
	Level     string     `json:"level" valid:"required"`
	Message   string     `json:"message" valid:"required"`
}

func (l LogMessage) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Timestamp, validation.Required),
		validation.Field(&l.Level, validation.Required),
		validation.Field(&l.Message, validation.Required),
	)
}

func (l *LogMessage) UnmarshalJSON(raw []byte) error {
	type logMessage LogMessage
	if err := json.Unmarshal(raw, (*logMessage)(l)); err != nil {
		return err
	}

	if err := l.Validate(); err != nil {
		return err
	}
	return nil
}

func (l LogMessage) String() string {
	return fmt.Sprintf("%s %s: %s", l.Timestamp.UTC().String(), l.Level, l.Message)
}

type DeploymentLog struct {
	// skip these 2 field when (un)marshaling to/from JSON
	DeviceID     string `json:"-" valid:"required"`
	DeploymentID string `json:"-" valid:"uuidv4,required"`

	Messages []LogMessage `json:"messages" valid:"required"`
}

func (d *DeploymentLog) UnmarshalJSON(raw []byte) error {
	type AuxDeploymentLog DeploymentLog

	var adl AuxDeploymentLog

	if err := json.Unmarshal(raw, &adl); err != nil {
		return err
	}

	if adl.Messages == nil || len(adl.Messages) == 0 {
		return errors.Wrapf(ErrInvalidDeploymentLog, "no messages")
	}

	d.Messages = adl.Messages
	return nil
}

func (d DeploymentLog) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.DeviceID, validation.Required),
		validation.Field(&d.DeploymentID, validation.Required, is.UUID),
		validation.Field(&d.Messages, validation.Required),
	)
}
