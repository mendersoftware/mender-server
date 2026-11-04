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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentConstructorValidate(t *testing.T) {

	t.Parallel()

	testCases := []struct {
		InputName         string
		InputArtifactName string
		InputDevices      []string
		InputAllDevices   bool
		InputGroup        string
		IsValid           bool
	}{
		{
			InputName:         "",
			InputArtifactName: "",
			InputDevices:      nil,
			IsValid:           false,
		},
		{
			InputName:         "something",
			InputArtifactName: "",
			InputDevices:      nil,
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "",
			InputDevices:      nil,
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      nil,
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{},
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{""},
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{"lala"},
			IsValid:           true,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{"f826484e-1157-4109-af21-304e6d711560"},
			IsValid:           true,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{},
			InputGroup:        "foo",
			IsValid:           true,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{},
			InputAllDevices:   true,
			IsValid:           true,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{"lala"},
			InputAllDevices:   true,
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{"lala"},
			InputGroup:        "foo",
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{"lala"},
			InputAllDevices:   true,
			IsValid:           false,
		},
	}

	for _, test := range testCases {

		dep := &DeploymentConstructor{}
		dep.Name = test.InputName
		dep.ArtifactName = test.InputArtifactName
		dep.Devices = test.InputDevices
		dep.Group = test.InputGroup
		dep.AllDevices = test.InputAllDevices

		err := dep.ValidateNew()

		if !test.IsValid {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}

}

func TestNewDeploymentFromConstructor(t *testing.T) {

	t.Parallel()

	dep, err := NewDeploymentFromConstructor(nil)
	assert.NoError(t, err)
	assert.NotNil(t, dep)

	con := &DeploymentConstructor{}

	dep, err = NewDeploymentFromConstructor(con)
	assert.NoError(t, err)

	assert.NotNil(t, dep)
	assert.Equal(t, con, dep.DeploymentConstructor)
}

func TestDeploymentValidate(t *testing.T) {

	t.Parallel()

	testCases := []struct {
		InputName         string
		InputArtifactName string
		InputDevices      []string
		IsValid           bool
	}{
		{
			InputName:         "",
			InputArtifactName: "",
			InputDevices:      nil,
			IsValid:           false,
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{"f826484e-1157-4109-af21-304e6d711560"},
			IsValid:           true,
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			pub := &DeploymentConstructor{}
			pub.Name = test.InputName
			pub.ArtifactName = test.InputArtifactName
			pub.Devices = test.InputDevices

			dep, err := NewDeploymentFromConstructor(pub)
			assert.NoError(t, err)

			err = dep.Validate()

			if !test.IsValid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestDeploymentConstructorChecksum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		InputName         string
		InputArtifactName string
		InputDevices      []string
		IsValid           bool
		Checksum          string
	}{
		{
			InputName:         "",
			InputArtifactName: "",
			InputDevices:      nil,
			IsValid:           false,
			Checksum:          "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
		},
		{
			InputName:         "f826484e-1157-4109-af21-304e6d711560",
			InputArtifactName: "f826484e-1157-4109-af21-304e6d711560",
			InputDevices:      []string{"f826484e-1157-4109-af21-304e6d711560"},
			IsValid:           true,
			Checksum:          "abd7a7e46d0c50fa4bffa4ccea26b63e54d9ee17b1d242f35d40a046e9f92e00",
		},
	}

	for i, test := range testCases {
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			pub := &DeploymentConstructor{}
			pub.Name = test.InputName
			pub.ArtifactName = test.InputArtifactName
			pub.Devices = test.InputDevices

			checksum := pub.Checksum()
			assert.Equal(t, test.Checksum, checksum)
		})
	}

}

func TestDeploymentMarshalJSON(t *testing.T) {

	t.Parallel()

	dep, err := NewDeployment()
	assert.NoError(t, err)
	dep.Name = "Region: NYC"
	dep.ArtifactName = "App 123"
	dep.Devices = []string{"Device 123"}
	dep.Id = "14ddec54-30be-49bf-aa6b-97ce271d71f5"
	deviceCount := int64(1337)
	dep.DeviceCount = &deviceCount
	dep.Status = DeploymentStatusInProgress
	dep.Statistics = DeploymentStatistics{
		TotalSize: 10,
	}
	b, _ := json.Marshal(Stats{})
	dep.Statistics.Status = Stats{}

	j, err := dep.MarshalJSON()
	assert.NoError(t, err)

	// date format may be slightly different on different platforms
	expectedJSON := `
	{
		"name":"Region: NYC",
		"artifact_name":"App 123",
        "created":"` + dep.Created.Format(time.RFC3339Nano) + `",
		"id":"14ddec54-30be-49bf-aa6b-97ce271d71f5",
		"statistics":{"status":` + string(b) + `,"total_size": 10},
		"status":"inprogress",
		"device_count":1337,
		"type":"software"
	}`

	assert.JSONEq(t, expectedJSON, string(j))
}

func TestDeploymentIs(t *testing.T) {
	d, err := NewDeployment()
	assert.NoError(t, err)
	d.MaxDevices = 1

	assert.False(t, d.IsNotPending())
	assert.False(t, d.IsFinished())

	// check all active statuses
	active := []DeviceDeploymentStatus{
		DeviceDeploymentStatusRebooting,
		DeviceDeploymentStatusInstalling,
		DeviceDeploymentStatusDownloading,
		DeviceDeploymentStatusPauseBeforeInstall,
		DeviceDeploymentStatusPauseBeforeReboot,
		DeviceDeploymentStatusPauseBeforeCommit,
	}
	for _, as := range active {
		t.Logf("checking in-progress deployment stat %s", as)
		d.Statistics.Status = Stats{}
		d.Statistics.Status.Set(as, 1)
		assert.True(t, d.IsNotPending())
		assert.False(t, d.IsFinished())
	}

	finished := []DeviceDeploymentStatus{
		DeviceDeploymentStatusSuccess,
		DeviceDeploymentStatusFailure,
		DeviceDeploymentStatusNoArtifact,
		DeviceDeploymentStatusAlreadyInst,
		DeviceDeploymentStatusAborted,
	}
	for _, as := range finished {
		t.Logf("checking finished deployment stat %s", as)
		d.Statistics.Status = Stats{}
		d.Statistics.Status.Set(as, 1)
		assert.True(t, d.IsFinished())
		assert.True(t, d.IsNotPending())
	}

	pending := []DeviceDeploymentStatus{
		DeviceDeploymentStatusPending,
	}
	for _, as := range pending {
		t.Logf("checking pending deployment stat %s", as)
		d.Statistics.Status = Stats{}
		d.Statistics.Status.Set(as, 1)
		assert.False(t, d.IsFinished())
		assert.False(t, d.IsNotPending())
	}
}

func TestDeploymentGetStatus(t *testing.T) {

	tests := map[string]struct {
		Stats        Stats
		OutputStatus DeploymentStatus
	}{
		"Single NoArtifact": {
			Stats: Stats{
				NoArtifact: 1,
			},
			OutputStatus: "finished",
		},
		"Single Success": {
			Stats: Stats{
				Success: 1,
			},
			OutputStatus: "finished",
		},
		"Success + NoArtifact": {
			Stats: Stats{
				Success:    1,
				NoArtifact: 1,
			},
			OutputStatus: "finished",
		},
		"Failed + NoArtifact": {
			Stats: Stats{
				Failure:    1,
				NoArtifact: 1,
			},
			OutputStatus: "finished",
		},
		"Failed + AlreadyInst": {
			Stats: Stats{
				Failure:          1,
				AlreadyInstalled: 1,
			},
			OutputStatus: "finished",
		},
		"Failed + Aborted": {
			Stats: Stats{
				Failure: 1,
				Aborted: 1,
			},
			OutputStatus: "finished",
		},
		"Rebooting + NoArtifact": {
			Stats: Stats{
				Rebooting:  1,
				NoArtifact: 1,
			},
			OutputStatus: "inprogress",
		},
		"Rebooting + Installing": {
			Stats: Stats{
				Rebooting:  1,
				Installing: 1,
			},
			OutputStatus: "inprogress",
		},
		"Rebooting + Pending": {
			Stats: Stats{
				Rebooting: 1,
				Pending:   1,
			},
			OutputStatus: "inprogress",
		},
		"All paused states": {
			Stats: Stats{
				PauseBeforeInstall: 1,
				PauseBeforeCommit:  1,
				PauseBeforeReboot:  1,
			},
			OutputStatus: "inprogress",
		},
		"Some paused states": {
			Stats: Stats{
				Installing:         1,
				PauseBeforeInstall: 1,
				PauseBeforeCommit:  0,
				PauseBeforeReboot:  1,
			},
			OutputStatus: "inprogress",
		},
		"Pending": {
			Stats: Stats{
				Pending: 1,
			},
			OutputStatus: "pending",
		},
		"Empty": {
			OutputStatus: "pending",
		},
		//verify we count 'already-installed' towards 'inprogress'
		"pending + already-installed": {
			Stats: Stats{
				Pending:          1,
				AlreadyInstalled: 1,
			},
			OutputStatus: "inprogress",
		},
		//verify we count 'already-installed' towards 'finished'
		"already-installed + finished": {
			Stats: Stats{
				Pending:          0,
				AlreadyInstalled: 1,
			},
			OutputStatus: "finished",
		},
	}

	for name, test := range tests {

		t.Run(name, func(t *testing.T) {

			dep, err := NewDeployment()
			assert.NoError(t, err)
			dep.Statistics.Status = test.Stats

			dep.MaxDevices = dep.Statistics.Status.Failure +
				dep.Statistics.Status.Aborted +
				dep.Statistics.Status.PauseBeforeInstall +
				dep.Statistics.Status.PauseBeforeCommit +
				dep.Statistics.Status.PauseBeforeReboot +
				dep.Statistics.Status.Downloading +
				dep.Statistics.Status.Installing +
				dep.Statistics.Status.Rebooting +
				dep.Statistics.Status.Pending +
				dep.Statistics.Status.Success +
				dep.Statistics.Status.NoArtifact +
				dep.Statistics.Status.AlreadyInstalled +
				dep.Statistics.Status.Decommissioned

			assert.Equal(t, test.OutputStatus, dep.GetStatus())
		})
	}

}

func TestFuzzyGetStatus(t *testing.T) {

	rand := func(min int, max int) int64 {
		rand.Seed(time.Now().UTC().UnixNano())
		return int64(min + rand.Intn(max-min))
	}

	max := 3

	for i := 0; i < 1000; i++ {
		dep, err := NewDeployment()
		assert.NoError(t, err)

		dep.Statistics.Status.Downloading = rand(0, max)
		dep.Statistics.Status.Installing = rand(0, max)
		dep.Statistics.Status.Rebooting = rand(0, max)
		dep.Statistics.Status.Pending = rand(0, max)
		dep.Statistics.Status.Success = rand(0, max)
		dep.Statistics.Status.Failure = rand(0, max)
		dep.Statistics.Status.NoArtifact = rand(0, max)
		dep.Statistics.Status.AlreadyInstalled = rand(0, max)
		dep.Statistics.Status.Aborted = rand(0, max)
		dep.Statistics.Status.Decommissioned = rand(0, max)

		pending := 0
		inprogress := 0
		finished := 0

		if dep.GetStatus() == "pending" {
			pending++
		}

		if dep.GetStatus() == "finished" {
			finished++
		}

		if dep.GetStatus() == "inprogress" {
			inprogress++
		}

		exp_stats := pending + inprogress + finished
		assert.Equal(t, 1, exp_stats, dep.Statistics.Status)
	}
}
