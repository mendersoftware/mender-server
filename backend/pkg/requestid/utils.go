// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package requestid

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/google/uuid"
)

type requestIdKeyType int

const (
	requestIdKey requestIdKeyType = 0
)

var idCounter uint64
var procRand [4]byte

var initOnce = &sync.Once{}

func lazyInit() {
	var processSeed [12]byte
	_, err := io.ReadFull(rand.Reader, processSeed[:])
	if err != nil {
		panic(fmt.Errorf("failed to initialize random seed: %w", err))
	}
	idCounter = binary.BigEndian.Uint64(processSeed[4:])
	copy(procRand[:], processSeed[:4])
}

var timeNow func() time.Time = time.Now

// New generates a new Request ID which is a modified version of UUID v7.
//
// Instead of always generating a new random value for the upper 32bits,
// it recycles a value generated once per process to save entropy.
//
//	 0               1               2               3
//	 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|                           unix_ts_ms                          |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|          unix_ts_ms           |  ver  |  count[30:42]         |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|var|                        count[0:30]                        |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|                            rand_const                         |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func New() uuid.UUID {
	initOnce.Do(lazyInit)
	var id uuid.UUID
	ts := timeNow().UnixMilli()
	binary.BigEndian.PutUint32(id[:], uint32(ts>>12))
	id[0] = byte(ts >> 40)
	id[1] = byte(ts >> 32)
	id[2] = byte(ts >> 24)
	id[3] = byte(ts >> 16)
	id[4] = byte(ts >> 8)
	id[5] = byte(ts)
	count := atomic.AddUint64(&idCounter, 1)
	id[6] = byte(count >> 38)
	id[7] = byte(count >> 30)
	id[8] = byte(count >> 24)
	id[9] = byte(count >> 16)
	id[10] = byte(count >> 8)
	id[11] = byte(count)
	id[6] = (id[6] & 0x0f) | 0x70 // Version 7
	id[8] = (id[8] & 0x3f) | 0x80 // Variant is 10
	copy(id[12:], procRand[:])
	return id
}

// GetReqId helper for retrieving current request Id
func GetReqId(r *rest.Request) string {
	return FromContext(r.Context())
}

// SetReqId is a helper for setting request ID in request context
func SetReqId(r *rest.Request, reqid string) *rest.Request {
	ctx := WithContext(r.Context(), reqid)
	r.Request = r.Request.WithContext(ctx)
	return r
}

// FromContext extracts current request Id from context.Context
func FromContext(ctx context.Context) string {
	val := ctx.Value(requestIdKey)
	if v, ok := val.(string); ok {
		return v
	}
	return ""
}

// WithContext adds request to context `ctx` and returns the resulting context.
func WithContext(ctx context.Context, reqid string) context.Context {
	return context.WithValue(ctx, requestIdKey, reqid)
}
