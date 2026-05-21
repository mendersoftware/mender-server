// Copyright 2026 Northern.tech AS
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

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type AuditLogAction string

const (
	ActionSetConfiguration    AuditLogAction = "set_configuration"
	ActionDeployConfiguration AuditLogAction = "deploy_configuration"
)

type AuditLogActorType string

const (
	ActorUser AuditLogActorType = "user"
)

type AuditLogActor struct {
	ID             string            `json:"id"`
	Type           AuditLogActorType `json:"type"`
	Email          string            `json:"email,omitempty"`
	DeviceIdentity string            `json:"identity_data,omitempty"`
}

func (a AuditLogActor) Validate() error {
	err := validation.ValidateStruct(&a,
		validation.Field(&a.ID, validation.Required),
		validation.Field(&a.Type,
			validation.In(ActorUser),
			validation.Required,
		),
	)
	if err != nil {
		return err
	}

	switch a.Type {
	case ActorUser:
		err = validation.ValidateStruct(&a,
			validation.Field(&a.Email, is.EmailFormat),
			validation.Field(&a.DeviceIdentity, validation.Empty),
		)
	}
	return err
}

type AuditLogObjectType string

const ObjectDevice AuditLogObjectType = "device"

type AuditLogObject struct {
	ID   string             `json:"id"`
	Type AuditLogObjectType `json:"type"`
}

func (o AuditLogObject) Validate() error {
	err := validation.ValidateStruct(&o,
		validation.Field(&o.ID, validation.Required),
		validation.Field(&o.Type,
			validation.Required,
			validation.In(ObjectDevice),
		),
	)
	return err
}

type AuditLog struct {
	Action   AuditLogAction      `json:"action"`
	Actor    AuditLogActor       `json:"actor"`
	Object   AuditLogObject      `json:"object"`
	Change   string              `json:"change,omitempty"`
	MetaData map[string][]string `json:"meta,omitempty"`
	EventTS  time.Time           `json:"time,omitempty"`
}

func (l AuditLog) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Actor, validation.Required),
		validation.Field(&l.Action, validation.In(
			ActionSetConfiguration,
			ActionDeployConfiguration,
		), validation.Required),
		validation.Field(&l.Object, validation.Required),
		validation.Field(&l.EventTS, validation.Required),
	)
}
