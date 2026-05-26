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

package app

import (
	"context"
	"fmt"
	"io"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"

	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
	"github.com/mendersoftware/mender-server/services/deviceconnect/store"
)

// App errors
var (
	ErrDeviceNotFound     = errors.New("device not found")
	ErrDeviceNotConnected = errors.New("device not connected")
)

// App interface describes app objects
//
//nolint:lll
//go:generate ../../../utils/mockgen.sh
type App interface {
	HealthCheck(ctx context.Context) error
	ProvisionDevice(ctx context.Context, tenantID string, device *model.Device) error
	GetDevice(ctx context.Context, tenantID, deviceID string) (*model.Device, error)
	DeleteDevice(ctx context.Context, tenantID, deviceID string) error
	SetDeviceConnected(ctx context.Context, tenantID, deviceID string) (int64, error)
	SetDeviceDisconnected(ctx context.Context, tenantID, deviceID string, version int64) error
	PrepareUserSession(ctx context.Context, sess *model.Session) error
	LogUserSession(ctx context.Context, sess *model.Session, sessionType string) error
	FreeUserSession(ctx context.Context, sessionID string, sessionTypes []string) error
	GetSessionRecording(ctx context.Context, id string, w io.Writer) (err error)
	SaveSessionRecording(ctx context.Context, id string, sessionBytes []byte) error
	GetRecorder(sessionID string) Recorder
	GetControlRecorder(sessionID string) Recorder
	DownloadFile(ctx context.Context, sess *model.Session, path string) error
	UploadFile(ctx context.Context, sess *model.Session, path string) error
	DeleteTenant(ctx context.Context, tenantID string) error
	Shutdown(timeout time.Duration)
	ShutdownDone()
	RegisterShutdownCancel(context.CancelFunc) uint32
	UnregisterShutdownCancel(uint32)
}

// app is an app object
type app struct {
	store            store.DataStore
	workflows        client.WorkflowsOtherAPI
	shutdownCancels  map[uint32]context.CancelFunc
	shutdownCancelsM *sync.Mutex
	shutdownDone     chan struct{}
	Config
}

type Config struct {
	HaveAuditLogs bool
}

// NewApp initialize a new deviceconnect App
func New(ds store.DataStore, wf client.WorkflowsOtherAPI, config ...Config) App {
	conf := Config{}
	for _, cfgIn := range config {
		if cfgIn.HaveAuditLogs {
			conf.HaveAuditLogs = true
		}
	}
	return &app{
		store:            ds,
		workflows:        wf,
		Config:           conf,
		shutdownCancels:  make(map[uint32]context.CancelFunc),
		shutdownCancelsM: &sync.Mutex{},
		shutdownDone:     make(chan struct{}),
	}
}

// HealthCheck performs a health check and returns an error if it fails
func (a *app) HealthCheck(ctx context.Context) error {
	return a.store.Ping(ctx)
}

// ProvisionDevice provisions a new tenant
func (a *app) ProvisionDevice(
	ctx context.Context,
	tenantID string,
	device *model.Device,
) error {
	return a.store.ProvisionDevice(ctx, tenantID, device.ID)
}

// GetDevice returns a device
func (a *app) GetDevice(
	ctx context.Context,
	tenantID string,
	deviceID string,
) (*model.Device, error) {
	device, err := a.store.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return nil, err
	} else if device == nil {
		return nil, ErrDeviceNotFound
	}
	return device, nil
}

// DeleteDevice provisions a new tenant
func (a *app) DeleteDevice(ctx context.Context, tenantID, deviceID string) error {
	return a.store.DeleteDevice(ctx, tenantID, deviceID)
}

func (a *app) SetDeviceConnected(
	ctx context.Context,
	tenantID string,
	deviceID string,
) (int64, error) {
	return a.store.SetDeviceConnected(ctx, tenantID, deviceID)
}
func (a *app) SetDeviceDisconnected(
	ctx context.Context,
	tenantID string,
	deviceID string,
	version int64,
) error {
	return a.store.SetDeviceDisconnected(ctx, tenantID, deviceID, version)
}

// PrepareUserSession prepares a new user session
func (a *app) PrepareUserSession(
	ctx context.Context,
	sess *model.Session,
) error {
	if sess == nil {
		return errors.New("nil Session")
	}
	if sess.ID == "" {
		sessID, err := uuid.NewRandom()
		if err != nil {
			return errors.Wrap(err, "failed to generate session ID")
		}
		sess.ID = sessID.String()
	}
	if err := sess.Validate(); err != nil {
		return errors.Wrap(err, "app: cannot create invalid Session")
	}

	err := a.store.AllocateSession(ctx, sess)
	if err != nil {
		return err
	}

	return nil
}

type Action string

const (
	ActionTerminalOpen     Action = "open_terminal"
	ActionTerminalClose    Action = "close_terminal"
	ActionPortForwardOpen  Action = "open_portforward"
	ActionPortForwardClose Action = "close_portforward"
	ActionDownloadFile     Action = "download_file"
	ActionUploadFile       Action = "upload_file"
)

type ActorType string

const (
	ActorUser ActorType = "user"
)

type Actor struct {
	ID             string    `json:"id"`
	Type           ActorType `json:"type"`
	Email          string    `json:"email,omitempty"`
	DeviceIdentity string    `json:"identity_data,omitempty"`
}

func (a Actor) Validate() error {
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

type ObjectType string

const ObjectDevice ObjectType = "device"

type Object struct {
	ID   string     `json:"id"`
	Type ObjectType `json:"type"`
}

func (o Object) Validate() error {
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
	Action   Action              `json:"action"`
	Actor    Actor               `json:"actor"`
	Object   Object              `json:"object"`
	Change   string              `json:"change,omitempty"`
	MetaData map[string][]string `json:"meta,omitempty"`
	EventTS  time.Time           `json:"time,omitempty"`
}

func (l AuditLog) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Actor, validation.Required),
		validation.Field(&l.Action, validation.In(
			ActionTerminalOpen, ActionTerminalClose,
			ActionPortForwardOpen, ActionPortForwardClose,
			ActionDownloadFile, ActionUploadFile,
		), validation.Required),
		validation.Field(&l.Object, validation.Required),
		validation.Field(&l.EventTS, validation.Required),
	)
}

const workflowSubmitAuditlog = "emit_auditlog"

func (a *app) submitAuditLog(
	ctx context.Context,
	action Action,
	change string,
	sess *model.Session,
	extraMeta map[string][]string,
) error {
	if !a.HaveAuditLogs {
		return nil
	}
	log := AuditLog{
		Action: action,
		Actor: Actor{
			ID:   sess.UserID,
			Type: ActorUser,
		},
		Object: Object{
			ID:   sess.DeviceID,
			Type: ObjectDevice,
		},
		Change: change,
		MetaData: map[string][]string{
			"session_id": {sess.ID},
		},
		EventTS: time.Now(),
	}
	maps.Copy(log.MetaData, extraMeta)
	err := log.Validate()
	if err != nil {
		return err
	}
	//nolint:bodyclose
	_, _, err = a.workflows.StartWorkflow(ctx, workflowSubmitAuditlog).
		RequestBody(map[string]any{
			"auditlog":   log,
			"tenant_id":  sess.TenantID,
			"request_id": requestid.FromContext(ctx),
		}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to submit audit log: %w", err)
	}

	return nil
}

// LogUserSession logs a new user session
func (a *app) LogUserSession(
	ctx context.Context,
	sess *model.Session,
	sessionType string,
) error {
	var change string
	var action Action
	if sessionType == model.SessionTypePortForward {
		change = "User requested a new port forwarding session"
		action = ActionPortForwardOpen
	} else if sessionType == model.SessionTypeTerminal {
		change = "User requested a new terminal session"
		action = ActionTerminalOpen
	} else {
		return errors.New("unknown session type: " + sessionType)
	}
	err := a.submitAuditLog(ctx, action, change, sess, nil)
	if err != nil {
		_, e := a.store.DeleteSession(ctx, sess.ID)
		if e != nil {
			err = errors.Errorf(
				"%s: failed to clean up session state: %s",
				err.Error(), e.Error(),
			)
		}
		return err
	}
	return nil
}

// FreeUserSession releases the session
func (a *app) FreeUserSession(
	ctx context.Context,
	sessionID string,
	sessionTypes []string,
) error {
	sess, err := a.store.DeleteSession(ctx, sessionID)
	if err != nil {
		return err
	}
	for _, sessionType := range sessionTypes {
		var action Action
		if sessionType == model.SessionTypePortForward {
			action = ActionPortForwardClose
		} else if sessionType == model.SessionTypeTerminal {
			action = ActionTerminalClose
		} else {
			continue
		}
		err = a.submitAuditLog(ctx, action, "", sess, nil)
		if err != nil {
			return errors.Wrap(err, "failed to submit audit log")
		}
	}
	return nil
}

func (a *app) GetSessionRecording(ctx context.Context, id string, w io.Writer) (err error) {
	err = a.store.WriteSessionRecords(ctx, id, w)
	return err
}

func (a *app) SaveSessionRecording(ctx context.Context, id string, sessionBytes []byte) error {
	err := a.store.InsertSessionRecording(ctx, id, sessionBytes)
	return err
}

func (a app) GetRecorder(sessionID string) Recorder {
	return NewRecorder(sessionID, a.store)
}

func (a app) GetControlRecorder(sessionID string) Recorder {
	return NewControlRecorder(sessionID, a.store)
}

func (a *app) DownloadFile(ctx context.Context, sess *model.Session, path string) error {
	return a.submitFileTransferAuditlog(ctx,
		ActionDownloadFile, sess, path,
		"User downloaded a file from the device")
}

func (a *app) UploadFile(ctx context.Context, sess *model.Session, path string) error {
	return a.submitFileTransferAuditlog(ctx,
		ActionUploadFile, sess, path,
		"User uploaded a file to the device")
}

func (a *app) submitFileTransferAuditlog(
	ctx context.Context,
	action Action,
	sess *model.Session,
	path string,
	change string,
) error {
	err := a.submitAuditLog(ctx, action, change, sess, map[string][]string{"path": {path}})
	if err != nil {
		return errors.Wrap(err,
			"failed to submit audit log for file transfer",
		)
	}
	return nil
}

func (a *app) Shutdown(timeout time.Duration) {
	a.shutdownCancelsM.Lock()
	defer a.shutdownCancelsM.Unlock()
	ticker := time.NewTicker(timeout / time.Duration(len(a.shutdownCancels)+1))
	for _, cancel := range a.shutdownCancels {
		cancel()
		<-ticker.C
	}
	<-ticker.C
	close(a.shutdownDone)
}

func (a *app) ShutdownDone() {
	<-a.shutdownDone
}

var shutdownID uint32

func (a *app) RegisterShutdownCancel(cancel context.CancelFunc) uint32 {
	a.shutdownCancelsM.Lock()
	defer a.shutdownCancelsM.Unlock()
	id := atomic.AddUint32(&shutdownID, 1)
	a.shutdownCancels[id] = cancel
	return id
}

func (a *app) UnregisterShutdownCancel(id uint32) {
	a.shutdownCancelsM.Lock()
	defer a.shutdownCancelsM.Unlock()
	delete(a.shutdownCancels, id)
}

func (d *app) DeleteTenant(ctx context.Context, tenantID string) error {
	tenantCtx := identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantID,
	})
	return d.store.DeleteTenant(tenantCtx, tenantID)
}
