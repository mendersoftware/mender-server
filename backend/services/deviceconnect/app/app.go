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
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deviceconnect/client/workflows"
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
//go:generate ../utils/mockgen.sh
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
	GetRecorder(ctx context.Context, sessionID string) io.Writer
	GetControlRecorder(ctx context.Context, sessionID string) io.Writer
	DownloadFile(ctx context.Context, userID string, deviceID string, path string) error
	UploadFile(ctx context.Context, userID string, deviceID string, path string) error
	DeleteTenant(ctx context.Context, tenantID string) error
	Shutdown(timeout time.Duration)
	ShutdownDone()
	RegisterShutdownCancel(context.CancelFunc) uint32
	UnregisterShutdownCancel(uint32)
}

// app is an app object
type app struct {
	store            store.DataStore
	workflows        workflows.Client
	shutdownCancels  map[uint32]context.CancelFunc
	shutdownCancelsM *sync.Mutex
	shutdownDone     chan struct{}
	Config
}

type Config struct {
	HaveAuditLogs bool
}

// NewApp initialize a new deviceconnect App
func New(ds store.DataStore, wf workflows.Client, config ...Config) App {
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

	device, err := a.store.GetDevice(ctx, sess.TenantID, sess.DeviceID)
	if err != nil {
		return err
	} else if device == nil {
		return ErrDeviceNotFound
	} else if device.Status != model.DeviceStatusConnected {
		return ErrDeviceNotConnected
	}

	err = a.store.AllocateSession(ctx, sess)
	if err != nil {
		return err
	}

	return nil
}

// LogUserSession logs a new user session
func (a *app) LogUserSession(
	ctx context.Context,
	sess *model.Session,
	sessionType string,
) error {
	if !a.HaveAuditLogs {
		return nil
	}
	var change string
	var action workflows.Action
	if sessionType == model.SessionTypePortForward {
		change = "User requested a new port forwarding session"
		action = workflows.ActionPortForwardOpen
	} else if sessionType == model.SessionTypeTerminal {
		change = "User requested a new terminal session"
		action = workflows.ActionTerminalOpen
	} else {
		return errors.New("unknown session type: " + sessionType)
	}
	err := a.workflows.SubmitAuditLog(ctx, workflows.AuditLog{
		Action: action,
		Actor: workflows.Actor{
			ID:   sess.UserID,
			Type: workflows.ActorUser,
		},
		Object: workflows.Object{
			ID:   sess.DeviceID,
			Type: workflows.ObjectDevice,
		},
		Change: change,
		MetaData: map[string][]string{
			"session_id": {sess.ID},
		},
		EventTS: time.Now(),
	})
	if err != nil {
		err = errors.Wrap(err, "failed to submit audit log")
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
	if a.HaveAuditLogs {
		for _, sessionType := range sessionTypes {
			var action workflows.Action
			if sessionType == model.SessionTypePortForward {
				action = workflows.ActionPortForwardClose
			} else if sessionType == model.SessionTypeTerminal {
				action = workflows.ActionTerminalClose
			} else {
				continue
			}
			err = a.workflows.SubmitAuditLog(ctx, workflows.AuditLog{
				Action: action,
				Actor: workflows.Actor{
					ID:   sess.UserID,
					Type: workflows.ActorUser,
				},
				Object: workflows.Object{
					ID:   sess.DeviceID,
					Type: workflows.ObjectDevice,
				},
				MetaData: map[string][]string{
					"session_id": {sess.ID},
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to submit audit log")
			}
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

func (a app) GetRecorder(ctx context.Context, sessionID string) io.Writer {
	return NewRecorder(ctx, sessionID, a.store)
}

func (a app) GetControlRecorder(ctx context.Context, sessionID string) io.Writer {
	return NewControlRecorder(ctx, sessionID, a.store)
}

func (a *app) DownloadFile(ctx context.Context, userID string, deviceID string, path string) error {
	return a.submitFileTransferAuditlog(ctx, userID, deviceID, path,
		workflows.ActionDownloadFile, "User downloaded a file from the device")
}

func (a *app) UploadFile(ctx context.Context, userID string, deviceID string, path string) error {
	return a.submitFileTransferAuditlog(ctx, userID, deviceID, path,
		workflows.ActionUploadFile, "User uploaded a file to the device")
}

func (a *app) submitFileTransferAuditlog(ctx context.Context, userID string, deviceID string,
	path string, action workflows.Action, change string) error {
	if a.HaveAuditLogs {
		err := a.workflows.SubmitAuditLog(ctx, workflows.AuditLog{
			Action: action,
			Actor: workflows.Actor{
				ID:   userID,
				Type: workflows.ActorUser,
			},
			Object: workflows.Object{
				ID:   deviceID,
				Type: workflows.ObjectDevice,
			},
			Change: change,
			MetaData: map[string][]string{
				"path": {path},
			},
			EventTS: time.Now(),
		})
		if err != nil {
			return errors.Wrap(err,
				"failed to submit audit log for file transfer",
			)
		}
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
