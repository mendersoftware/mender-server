// Copyright 2024 Northern.tech AS
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
package accesslog

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/sirupsen/logrus"

	"github.com/mendersoftware/mender-server/pkg/netutils"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
)

const (
	StatusClientClosedConnection = 499

	// nolint:lll
	DefaultLogFormat = "%t %S\033[0m \033[36;1m%Dμs\033[0m \"%r\" \033[1;30m%u \"%{User-Agent}i\"\033[0m"
	SimpleLogFormat  = "%s %Dμs %r %u %{User-Agent}i"

	envProxyDepth = "ACCESSLOG_PROXY_DEPTH"
)

// AccesLogMiddleware uses logger from requestlog and adds a fixed set
// of fields to every accesslog records.
type AccessLogMiddleware struct {
	// Format is not used but kept for historical use.
	// FIXME(QA-673): Remove unused attributes and properties from package.
	Format AccessLogFormat // nolint:unused

	ClientIPHook func(req *http.Request) net.IP
	DisableLog   func(statusCode int, r *rest.Request) bool

	recorder *rest.RecorderMiddleware
}

func getClientIPFromEnv() func(r *http.Request) net.IP {
	if proxyDepthEnv, ok := os.LookupEnv(envProxyDepth); ok {
		proxyDepth, err := strconv.ParseUint(proxyDepthEnv, 10, 8)
		if err == nil {
			return func(r *http.Request) net.IP {
				return netutils.GetIPFromXFFDepth(r, int(proxyDepth))
			}
		}
	}
	return nil
}

const MaxTraceback = 32

func collectTrace() string {
	var (
		trace     [MaxTraceback]uintptr
		traceback strings.Builder
	)
	// Skip 4
	// = accesslog.LogFunc
	// + accesslog.collectTrace
	// + runtime.Callers
	// + runtime.gopanic
	n := runtime.Callers(4, trace[:])
	frames := runtime.CallersFrames(trace[:n])
	for frame, more := frames.Next(); frame.PC != 0 &&
		n >= 0; frame, more = frames.Next() {
		funcName := frame.Function
		if funcName == "" {
			fmt.Fprint(&traceback, "???\n")
		} else {
			fmt.Fprintf(&traceback, "%s@%s:%d",
				frame.Function,
				path.Base(frame.File),
				frame.Line,
			)
		}
		if more {
			fmt.Fprintln(&traceback)
		}
		n--
	}
	return traceback.String()
}

func (mw *AccessLogMiddleware) LogFunc(
	ctx context.Context, startTime time.Time,
	w rest.ResponseWriter, r *rest.Request) {
	fields := logrus.Fields{
		"type": r.Proto,
		"ts": startTime.
			Truncate(time.Millisecond).
			Format(time.RFC3339Nano),
		"method":    r.Method,
		"path":      r.URL.Path,
		"useragent": r.UserAgent(),
		"qs":        r.URL.RawQuery,
	}
	if mw.ClientIPHook != nil {
		fields["clientip"] = mw.ClientIPHook(r.Request)
	}
	lc := fromContext(ctx)
	if lc != nil {
		lc.addFields(fields)
	}
	statusCode, _ := r.Env["STATUS_CODE"].(int)
	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			statusCode = StatusClientClosedConnection
		}
	default:
	}

	if panic := recover(); panic != nil {
		trace := collectTrace()
		fields["panic"] = panic
		fields["trace"] = trace
		// Wrap in recorder middleware to make sure the response is recorded
		mw.recorder.MiddlewareFunc(func(w rest.ResponseWriter, r *rest.Request) {
			rest.Error(w, "Internal Server Error", http.StatusInternalServerError)
		})(w, r)
		statusCode = http.StatusInternalServerError
	} else if mw.DisableLog != nil && mw.DisableLog(statusCode, r) {
		return
	}
	rspTime := time.Since(startTime)
	// We do not need more than 3 digit fraction
	if rspTime > time.Second {
		rspTime = rspTime.Round(time.Millisecond)
	} else if rspTime > time.Millisecond {
		rspTime = rspTime.Round(time.Microsecond)
	}
	fields["responsetime"] = rspTime.String()
	fields["byteswritten"], _ = r.Env["BYTES_WRITTEN"].(int64)
	fields["status"] = statusCode

	logger := requestlog.GetRequestLogger(r)
	var level logrus.Level = logrus.InfoLevel
	if statusCode >= 500 {
		level = logrus.ErrorLevel
	} else if statusCode >= 300 {
		level = logrus.WarnLevel
	}
	logger.WithFields(fields).
		Log(level)
}

// MiddlewareFunc makes AccessLogMiddleware implement the Middleware interface.
func (mw *AccessLogMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	if mw.ClientIPHook == nil {
		// If not set, try get it from env
		mw.ClientIPHook = getClientIPFromEnv()
	}

	// This middleware depends on RecorderMiddleware to work
	mw.recorder = new(rest.RecorderMiddleware)
	return func(w rest.ResponseWriter, r *rest.Request) {
		ctx := r.Request.Context()
		startTime := time.Now()
		ctx = withContext(ctx, &logContext{maxErrors: DefaultMaxErrors})
		r.Request = r.Request.WithContext(ctx)
		defer mw.LogFunc(ctx, startTime, w, r)
		// call the handler inside recorder context
		mw.recorder.MiddlewareFunc(h)(w, r)
	}
}
