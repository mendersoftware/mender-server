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
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrDeviceDeploymentStatusMismatch = errors.New(
		"model active state does not match status",
	)
)

// DeviceDeploymentStatus is an enumerated type showing the status of a device within a deployment
type DeviceDeploymentStatus int32

// Deployment statuses
const (
	// The following statuses are distributed evenly by incrementing the
	// enum counter's second byte.
	// NOTE: when adding new statuses into the list, use the mean value between the
	//       neighbouring values and append the value AFTER the following list and extend
	//       the DeviceDeploymentStatus<type>Str constant and allStatuses variable as well
	//       as the MarshalText and UnmarshalText interface functions.
	//       See example below.
	// WARN: DO NOT CHANGE ANY OF THE FOLLOWING VALUES.
	DeviceDeploymentStatusNull DeviceDeploymentStatus = iota << 8 // i=0... {i * 2^8}
	DeviceDeploymentStatusFailure
	DeviceDeploymentStatusAborted
	DeviceDeploymentStatusPauseBeforeInstall
	DeviceDeploymentStatusPauseBeforeCommit
	DeviceDeploymentStatusPauseBeforeReboot
	DeviceDeploymentStatusDownloading
	DeviceDeploymentStatusInstalling
	DeviceDeploymentStatusRebooting
	DeviceDeploymentStatusPending
	DeviceDeploymentStatusSuccess
	DeviceDeploymentStatusNoArtifact
	DeviceDeploymentStatusAlreadyInst
	DeviceDeploymentStatusDecommissioned
	// DeviceDeploymentStatusNew = (DeviceDeploymentStatusSuccess +
	// DeviceDeploymentStatusNoArtifact) / 2

	DeviceDeploymentStatusActiveLow  = DeviceDeploymentStatusPauseBeforeInstall
	DeviceDeploymentStatusActiveHigh = DeviceDeploymentStatusPending

	DeviceDeploymentStatusFailureStr            = "failure"
	DeviceDeploymentStatusAbortedStr            = "aborted"
	DeviceDeploymentStatusPauseBeforeInstallStr = "pause_before_installing"
	DeviceDeploymentStatusPauseBeforeCommitStr  = "pause_before_committing"
	DeviceDeploymentStatusPauseBeforeRebootStr  = "pause_before_rebooting"
	DeviceDeploymentStatusDownloadingStr        = "downloading"
	DeviceDeploymentStatusInstallingStr         = "installing"
	DeviceDeploymentStatusRebootingStr          = "rebooting"
	DeviceDeploymentStatusPendingStr            = "pending"
	DeviceDeploymentStatusSuccessStr            = "success"
	DeviceDeploymentStatusNoArtifactStr         = "noartifact"
	DeviceDeploymentStatusAlreadyInstStr        = "already-installed"
	DeviceDeploymentStatusDecommissionedStr     = "decommissioned"
	// DeviceDeploymentStatusNew = "lorem-ipsum"
)

const (
	DeviceDeploymentStatusPauseStr    = "pause"
	DeviceDeploymentStatusActiveStr   = "active"
	DeviceDeploymentStatusFinishedStr = "finished"
)

func NewStatus(status string) DeviceDeploymentStatus {
	var stat DeviceDeploymentStatus
	_ = stat.UnmarshalText([]byte(status))
	return stat
}

var allStatuses = []DeviceDeploymentStatus{
	DeviceDeploymentStatusFailure,
	DeviceDeploymentStatusPauseBeforeInstall,
	DeviceDeploymentStatusPauseBeforeCommit,
	DeviceDeploymentStatusPauseBeforeReboot,
	DeviceDeploymentStatusDownloading,
	DeviceDeploymentStatusInstalling,
	DeviceDeploymentStatusRebooting,
	DeviceDeploymentStatusPending,
	DeviceDeploymentStatusSuccess,
	DeviceDeploymentStatusAborted,
	DeviceDeploymentStatusNoArtifact,
	DeviceDeploymentStatusAlreadyInst,
	DeviceDeploymentStatusDecommissioned,
	// DeviceDeploymentStatusNew
}

func (stat DeviceDeploymentStatus) MarshalText() ([]byte, error) {
	switch stat {
	case DeviceDeploymentStatusFailure:
		return []byte(DeviceDeploymentStatusFailureStr), nil
	case DeviceDeploymentStatusPauseBeforeInstall:
		return []byte(DeviceDeploymentStatusPauseBeforeInstallStr), nil
	case DeviceDeploymentStatusPauseBeforeCommit:
		return []byte(DeviceDeploymentStatusPauseBeforeCommitStr), nil
	case DeviceDeploymentStatusPauseBeforeReboot:
		return []byte(DeviceDeploymentStatusPauseBeforeRebootStr), nil
	case DeviceDeploymentStatusDownloading:
		return []byte(DeviceDeploymentStatusDownloadingStr), nil
	case DeviceDeploymentStatusInstalling:
		return []byte(DeviceDeploymentStatusInstallingStr), nil
	case DeviceDeploymentStatusRebooting:
		return []byte(DeviceDeploymentStatusRebootingStr), nil
	case DeviceDeploymentStatusPending:
		return []byte(DeviceDeploymentStatusPendingStr), nil
	case DeviceDeploymentStatusSuccess:
		return []byte(DeviceDeploymentStatusSuccessStr), nil
	case DeviceDeploymentStatusAborted:
		return []byte(DeviceDeploymentStatusAbortedStr), nil
	case DeviceDeploymentStatusNoArtifact:
		return []byte(DeviceDeploymentStatusNoArtifactStr), nil
	case DeviceDeploymentStatusAlreadyInst:
		return []byte(DeviceDeploymentStatusAlreadyInstStr), nil
	case DeviceDeploymentStatusDecommissioned:
		return []byte(DeviceDeploymentStatusDecommissionedStr), nil
	//case DeviceDeploymentStatusNew:
	//	return []byte(DeviceDeploymentStatusNewStr), nil
	case 0:
		return nil, errors.New("invalid status: variable not initialized")
	}
	return nil, errors.New("invalid status")
}

func (stat DeviceDeploymentStatus) String() string {
	ret, err := stat.MarshalText()
	if err != nil {
		return "invalid"
	}
	return string(ret)
}

func (stat *DeviceDeploymentStatus) UnmarshalText(b []byte) error {
	s := string(b)
	switch s {
	case DeviceDeploymentStatusFailureStr:
		*stat = DeviceDeploymentStatusFailure
	case DeviceDeploymentStatusPauseBeforeInstallStr:
		*stat = DeviceDeploymentStatusPauseBeforeInstall
	case DeviceDeploymentStatusPauseBeforeCommitStr:
		*stat = DeviceDeploymentStatusPauseBeforeCommit
	case DeviceDeploymentStatusPauseBeforeRebootStr:
		*stat = DeviceDeploymentStatusPauseBeforeReboot
	case DeviceDeploymentStatusDownloadingStr:
		*stat = DeviceDeploymentStatusDownloading
	case DeviceDeploymentStatusInstallingStr:
		*stat = DeviceDeploymentStatusInstalling
	case DeviceDeploymentStatusRebootingStr:
		*stat = DeviceDeploymentStatusRebooting
	case DeviceDeploymentStatusPendingStr:
		*stat = DeviceDeploymentStatusPending
	case DeviceDeploymentStatusSuccessStr:
		*stat = DeviceDeploymentStatusSuccess
	case DeviceDeploymentStatusAbortedStr:
		*stat = DeviceDeploymentStatusAborted
	case DeviceDeploymentStatusNoArtifactStr:
		*stat = DeviceDeploymentStatusNoArtifact
	case DeviceDeploymentStatusAlreadyInstStr:
		*stat = DeviceDeploymentStatusAlreadyInst
	case DeviceDeploymentStatusDecommissionedStr:
		*stat = DeviceDeploymentStatusDecommissioned
	//case DeviceDeploymentStatusNewStr:
	//	*stat = DeviceDeploymentStatusNew
	default:
		return errors.Errorf("invalid status for device '%s'", s)
	}
	return nil
}

func (stat DeviceDeploymentStatus) Active() bool {
	return stat >= DeviceDeploymentStatusActiveLow &&
		stat <= DeviceDeploymentStatusActiveHigh
}

// DeviceDeploymentStatus is a helper type for reporting status changes through
// the layers
type DeviceDeploymentState struct {
	// status reported by device
	Status DeviceDeploymentStatus
	// substate reported by device
	SubState string `json:",omitempty" bson:",omitempty"`
	// finish time
	FinishTime *time.Time `json:",omitempty" bson:",omitempty"`
}

func (state DeviceDeploymentState) Validate() error {
	return validation.ValidateStruct(&state,
		validation.Field(&state.Status, validation.Required),
	)
}

type DeviceDeployment struct {
	// Active says whether the device's deployment status is in an active
	// state - in progress or pending.
	Active bool `json:"-" bson:"active"`

	// Internal field of initial creation of deployment
	Created *time.Time `json:"created" bson:"created"`

	// Internal field of start of the deployment (after /next call)
	Started *time.Time `json:"started,omitempty" bson:"started"`

	// Update finish time
	Finished *time.Time `json:"finished,omitempty" bson:"finished,omitempty"`

	// Logical deletion time
	Deleted *time.Time `json:"deleted,omitempty" bson:"deleted,omitempty"`

	// Status
	Status DeviceDeploymentStatus `json:"status" bson:"status"`

	// Device id
	DeviceId string `json:"id" bson:"deviceid"`

	// Deployment id
	DeploymentId string `json:"-" bson:"deploymentid"`

	// ID
	Id string `json:"-" bson:"_id"`

	// Assigned software image
	Image *Image `json:"image,omitempty"`

	// deployments/next request from the device
	Request *DeploymentNextRequest `json:"-"`

	// Presence of deployment log
	IsLogAvailable bool `json:"log" bson:"log"`

	// Device reported substate
	SubState string `json:"substate,omitempty" bson:"substate,omitempty"`
}

func NewDeviceDeployment(deviceId, deploymentId string) *DeviceDeployment {

	now := time.Now()

	uid, err := uuid.NewRandom()
	if err != nil {
		panic(errors.Wrap(err, "failed to generate random uuid (v4)"))
	}
	id := uid.String()

	return &DeviceDeployment{
		Active:         true,
		Status:         DeviceDeploymentStatusPending,
		DeviceId:       deviceId,
		DeploymentId:   deploymentId,
		Id:             id,
		Created:        &now,
		IsLogAvailable: false,
	}
}

func (d DeviceDeployment) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.Created, validation.Required),
		validation.Field(&d.Status, validation.Required, deviceDeploymentStatusValidator{}),
		validation.Field(&d.DeviceId, validation.Required),
		validation.Field(&d.DeploymentId, validation.Required, is.UUID),
		validation.Field(&d.Id, validation.Required, is.UUID),
	)
	if err != nil {
		return err
	}
	if d.Status.Active() {
		if !d.Active {
			return ErrDeviceDeploymentStatusMismatch
		}
	} else if d.Active {
		return ErrDeviceDeploymentStatusMismatch
	}
	return nil
}

// Deployment statistics wrapper, each value carries a count of deployments
// aggregated by state.
type Stats map[string]int

func NewDeviceDeploymentStats() Stats {

	s := make(Stats, len(allStatuses))

	// populate statuses with 0s
	for _, k := range allStatuses {
		s[k.String()] = 0
	}

	return s
}

func (s Stats) Set(status DeviceDeploymentStatus, count int) {
	key := status.String()
	s[key] = count
}

func (s Stats) Inc(status DeviceDeploymentStatus) {
	var count int
	key := status.String()
	count = s[key]
	count++
	s[key] = count
}

func (s Stats) Get(status DeviceDeploymentStatus) int {
	key := status.String()
	return s[key]
}

func IsDeviceDeploymentStatusFinished(status DeviceDeploymentStatus) bool {
	if status == DeviceDeploymentStatusFailure || status == DeviceDeploymentStatusSuccess ||
		status == DeviceDeploymentStatusNoArtifact || status == DeviceDeploymentStatusAlreadyInst ||
		status == DeviceDeploymentStatusAborted || status == DeviceDeploymentStatusDecommissioned {
		return true
	}
	return false
}

// ActiveDeploymentStatuses lists statuses that represent deployment in active state (not finished).
func ActiveDeploymentStatuses() []DeviceDeploymentStatus {
	return []DeviceDeploymentStatus{
		DeviceDeploymentStatusPending,
		DeviceDeploymentStatusDownloading,
		DeviceDeploymentStatusInstalling,
		DeviceDeploymentStatusRebooting,
		DeviceDeploymentStatusPauseBeforeInstall,
		DeviceDeploymentStatusPauseBeforeCommit,
		DeviceDeploymentStatusPauseBeforeReboot,
	}
}

func InactiveDeploymentStatuses() []DeviceDeploymentStatus {
	return []DeviceDeploymentStatus{
		DeviceDeploymentStatusAlreadyInst,
		DeviceDeploymentStatusSuccess,
		DeviceDeploymentStatusFailure,
		DeviceDeploymentStatusNoArtifact,
		DeviceDeploymentStatusAlreadyInst,
		DeviceDeploymentStatusAborted,
		DeviceDeploymentStatusDecommissioned,
	}
}

// InstalledDeviceDeployment describes a deployment currently installed on the
// device, usually reported by a device
type InstalledDeviceDeployment struct {
	ArtifactName string            `json:"artifact_name"`
	DeviceType   string            `json:"device_type"`
	Provides     map[string]string `json:"artifact_provides,omitempty"`
}

// DeploymentNextRequest holds a deployments/next request
type DeploymentNextRequest struct {
	DeviceProvides   *InstalledDeviceDeployment `json:"device_provides"`
	UpdateControlMap bool                       `json:"update_control_map"`
}

func (i *DeploymentNextRequest) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.DeviceProvides,
			validation.Required,
		),
	)
}

func (i *DeploymentNextRequest) String() string {
	j, err := json.Marshal(i)
	if err != nil {
		return "invalid request format"
	}
	return string(j)
}

func (i *InstalledDeviceDeployment) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.ArtifactName,
			validation.Required, lengthIn1To4096,
		),
		validation.Field(&i.DeviceType,
			validation.Required, lengthIn1To4096,
		),
		validation.Field(&i.Provides, validation.Each(lengthIn1To4096)),
	)
}
