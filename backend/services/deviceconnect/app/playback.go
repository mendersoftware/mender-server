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
	"time"

	"github.com/nats-io/nats.go"
)

const (
	DefaultPlaybackSleepIntervalMs = uint(100)
)

type Playback struct {
	sessionID         string
	deviceChan        chan *nats.Msg
	sleepMilliseconds uint
}

func NewPlayback(sessionID string, deviceChan chan *nats.Msg, sleepMilliseconds uint) *Playback {
	return &Playback{
		deviceChan:        deviceChan,
		sessionID:         sessionID,
		sleepMilliseconds: sleepMilliseconds,
	}
}

func (r *Playback) Write(d []byte) (n int, err error) {
	//now playback gets the msgpacked ProtoMsgs
	m := nats.Msg{
		Subject: "playback",
		Reply:   "no-reply",
		Data:    d,
		Sub:     nil,
	}
	time.Sleep(time.Duration(r.sleepMilliseconds) * time.Millisecond)
	r.deviceChan <- &m
	return len(d), nil
}
