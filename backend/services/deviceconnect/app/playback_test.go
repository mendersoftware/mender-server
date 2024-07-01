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

package app

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestNewPlayback(t *testing.T) {
	sessionID := "sessionID"
	deviceChan := make(chan *nats.Msg, 1)
	sleepMs := uint(100)
	r := NewPlayback(sessionID, deviceChan, sleepMs)
	assert.NotNil(t, r)
	assert.Equal(t, r.sessionID, sessionID)
	assert.Equal(t, r.deviceChan, deviceChan)
	assert.Equal(t, r.sleepMilliseconds, sleepMs)
}

func TestPlaybackWrite(t *testing.T) {
	deviceChan := make(chan *nats.Msg, 1)
	sessionID := "sessionID"

	testCases := []struct {
		Name      string
		Data      []byte
		SleepTime uint
	}{
		{
			Name: "ok",
			Data: []byte("some data"),
		},
		{
			Name:      "ok with sleep time",
			Data:      []byte("some data"),
			SleepTime: 2000,
		},
	}

	thresholdMs := uint(15)
	for _, tc := range testCases {
		r := NewPlayback(sessionID, deviceChan, tc.SleepTime)
		assert.NotNil(t, r)

		t0 := float64(time.Now().UTC().UnixNano()) * 0.000001
		n, err := r.Write(tc.Data)
		assert.NoError(t, err)
		assert.Equal(t, len(tc.Data), n)
		t1 := float64(time.Now().UTC().UnixNano()) * 0.000001
		<-deviceChan

		if tc.SleepTime > 0 {
			t.Logf("dt:%d", uint(t1-t0))
			assert.True(t, (t1-t0) > float64(tc.SleepTime-thresholdMs) && (t1-t0) < float64(thresholdMs+tc.SleepTime))
		}
	}
}
