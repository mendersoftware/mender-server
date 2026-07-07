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
	FreeUserSession(ctx context.Context, sessionID string, sessionTypes []string) error
	GetSessionRecording(ctx context.Context, id string, w io.Writer) (err error)
	SaveSessionRecording(ctx context.Context, id string, sessionBytes []byte) error
	GetRecorder(sessionID string) Recorder
	GetControlRecorder(sessionID string) Recorder
	DeleteTenant(ctx context.Context, tenantID string) error
	Shutdown(timeout time.Duration)
	ShutdownDone()
	RegisterConnectionCancelHandle(id string, cancel context.CancelFunc, exclusive bool) uint32
	UnregisterConnectionCancelHandle(handleID uint32)
}

type cancelHandle struct {
	id     string
	cancel context.CancelFunc
}

// app is an app object
type app struct {
	store            store.DataStore
	shutdownCancels  map[uint32]*cancelHandle
	shutdownCancelsM *sync.Mutex
	shutdownDone     chan struct{}
	Config
}

type Config struct{}

// NewApp initialize a new deviceconnect App
func New(ds store.DataStore, config ...Config) App {
	conf := Config{}
	return &app{
		store:            ds,
		Config:           conf,
		shutdownCancels:  make(map[uint32]*cancelHandle),
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

// FreeUserSession releases the session
func (a *app) FreeUserSession(
	ctx context.Context,
	sessionID string,
	sessionTypes []string,
) error {
	_, err := a.store.DeleteSession(ctx, sessionID)
	if err != nil {
		return err
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
	return nil
}

func (a *app) UploadFile(ctx context.Context, sess *model.Session, path string) error {
	return nil
}

func (a *app) Shutdown(timeout time.Duration) {
	a.shutdownCancelsM.Lock()
	defer a.shutdownCancelsM.Unlock()
	ticker := time.NewTicker(timeout / time.Duration(len(a.shutdownCancels)+1))
	for _, handle := range a.shutdownCancels {
		handle.cancel()
		<-ticker.C
	}
	<-ticker.C
	close(a.shutdownDone)
}

func (a *app) ShutdownDone() {
	<-a.shutdownDone
}

var shutdownID uint32

func (a *app) RegisterConnectionCancelHandle(
	id string,
	cancel context.CancelFunc,
	exclusive bool,
) uint32 {
	a.shutdownCancelsM.Lock()
	defer a.shutdownCancelsM.Unlock()
	if exclusive {
		for handleID, handle := range a.shutdownCancels {
			if handle.id == id {
				handle.cancel()
				delete(a.shutdownCancels, handleID)
			}
		}
	}
	handleID := atomic.AddUint32(&shutdownID, 1)
	a.shutdownCancels[handleID] = &cancelHandle{id: id, cancel: cancel}
	return handleID
}

func (a *app) UnregisterConnectionCancelHandle(handleID uint32) {
	a.shutdownCancelsM.Lock()
	defer a.shutdownCancelsM.Unlock()
	delete(a.shutdownCancels, handleID)
}

func (d *app) DeleteTenant(ctx context.Context, tenantID string) error {
	tenantCtx := identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantID,
	})
	return d.store.DeleteTenant(tenantCtx, tenantID)
}
