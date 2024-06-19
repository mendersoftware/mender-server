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

// Package log provides a thin wrapper over logrus, with a definition
// of a global root logger, its setup functions and convenience wrappers.
//
// The wrappers are introduced to reduce verbosity:
// - logrus.Fields becomes log.Ctx
// - logrus.WithFields becomes log.F(), defined on a Logger type
//
// The usage scenario in a multilayer app is as follows:
// - a new Logger is created in the upper layer with an initial context (request id, api method...)
// - it is passed to lower layer which automatically includes the context, and can further enrich it
// - result - logs across layers are tied together with a common context
//
// Note on concurrency:
// - all Loggers in fact point to the single base log, which serializes logging with its mutexes
// - all context is copied - each layer operates on an independent copy

package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// log is a global logger instance
	Log = logrus.New()
)

const (
	envLogFormat        = "LOG_FORMAT"
	envLogLevel         = "LOG_LEVEL"
	envLogDisableCaller = "LOG_DISABLE_CALLER_CONTEXT"

	logFormatJSON    = "json"
	logFormatJSONAlt = "ndjson"

	logFieldCaller    = "caller"
	logFieldCallerFmt = "%s@%s:%d"

	pkgSirupsen = "github.com/sirupsen/logrus"
)

type loggerContextKeyType int

const (
	loggerContextKey loggerContextKeyType = 0
)

// ContextLogger interface for components which support
// logging with context, via setting a logger to an exisiting one,
// thereby inheriting its context.
type ContextLogger interface {
	UseLog(l *Logger)
}

// init initializes the global logger to sane defaults.
func init() {
	var opts Options
	switch strings.ToLower(os.Getenv(envLogFormat)) {
	case logFormatJSON, logFormatJSONAlt:
		opts.Format = FormatJSON
	default:
		opts.Format = FormatConsole
	}
	opts.Level = Level(logrus.InfoLevel)
	if lvl := os.Getenv(envLogLevel); lvl != "" {
		logLevel, err := logrus.ParseLevel(lvl)
		if err == nil {
			opts.Level = Level(logLevel)
		}
	}
	opts.TimestampFormat = time.RFC3339
	opts.DisableCaller, _ = strconv.ParseBool(os.Getenv(envLogDisableCaller))
	Configure(opts)

	Log.ExitFunc = func(int) {}
}

type Level logrus.Level

const (
	LevelPanic = Level(logrus.PanicLevel)
	LevelFatal = Level(logrus.FatalLevel)
	LevelError = Level(logrus.ErrorLevel)
	LevelWarn  = Level(logrus.WarnLevel)
	LevelInfo  = Level(logrus.InfoLevel)
	LevelDebug = Level(logrus.DebugLevel)
	LevelTrace = Level(logrus.TraceLevel)
)

type Format int

const (
	FormatConsole Format = iota
	FormatJSON
)

type Options struct {
	TimestampFormat string

	Level Level

	DisableCaller bool

	Format Format

	Output io.Writer
}

func Configure(opts Options) {
	Log = logrus.New()

	if opts.Output != nil {
		Log.SetOutput(opts.Output)
	}
	Log.SetLevel(logrus.Level(opts.Level))

	if !opts.DisableCaller {
		Log.AddHook(ContextHook{})
	}

	var formatter logrus.Formatter

	switch opts.Format {
	case FormatConsole:
		formatter = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: opts.TimestampFormat,
		}
	case FormatJSON:
		formatter = &logrus.JSONFormatter{
			TimestampFormat: opts.TimestampFormat,
		}
	}
	Log.Formatter = formatter
}

// Setup allows to override the global logger setup.
func Setup(debug bool) {
	if debug {
		Log.Level = logrus.DebugLevel
	}
}

// Ctx short for log context, alias for the more verbose logrus.Fields.
type Ctx map[string]interface{}

// Logger is a wrapper for logrus.Entry.
type Logger struct {
	*logrus.Entry
}

// New returns a new Logger with a given context, derived from the global Log.
func New(ctx Ctx) *Logger {
	return NewFromLogger(Log, ctx)
}

// NewEmpty returns a new logger with empty context
func NewEmpty() *Logger {
	return New(Ctx{})
}

// NewFromLogger returns a new Logger derived from a given logrus.Logger,
// instead of the global one.
func NewFromLogger(log *logrus.Logger, ctx Ctx) *Logger {
	return &Logger{log.WithFields(logrus.Fields(ctx))}
}

// NewFromLogger returns a new Logger derived from a given logrus.Logger,
// instead of the global one.
func NewFromEntry(log *logrus.Entry, ctx Ctx) *Logger {
	return &Logger{log.WithFields(logrus.Fields(ctx))}
}

// F returns a new Logger enriched with new context fields.
// It's a less verbose wrapper over logrus.WithFields.
func (l *Logger) F(ctx Ctx) *Logger {
	return &Logger{l.Entry.WithFields(logrus.Fields(ctx))}
}

func (l *Logger) Level() logrus.Level {
	return l.Entry.Logger.Level
}

type ContextHook struct {
}

func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func FmtCaller(caller runtime.Frame) string {
	return fmt.Sprintf(
		logFieldCallerFmt,
		path.Base(caller.Function),
		path.Base(caller.File),
		caller.Line,
	)
}

func (hook ContextHook) Fire(entry *logrus.Entry) error {
	const (
		minCallDepth = 6 // logrus.Logger.Log
		maxCallDepth = 8 // logrus.Logger.<Level>f
	)
	var pcs [1 + maxCallDepth - minCallDepth]uintptr
	if _, ok := entry.Data[logFieldCaller]; !ok {
		// We don't know how deep we are in the callstack since the hook can be fired
		// at different levels. Search between depth 6 -> 8.
		i := runtime.Callers(minCallDepth, pcs[:])
		frames := runtime.CallersFrames(pcs[:i])
		var caller *runtime.Frame
		for frame, _ := frames.Next(); frame.PC != 0; frame, _ = frames.Next() {
			if !strings.HasPrefix(frame.Function, pkgSirupsen) {
				caller = &frame
				break
			}
		}
		if caller != nil {
			entry.Data[logFieldCaller] = FmtCaller(*caller)
		}
	}
	return nil
}

// WithCallerContext returns a new logger with caller set to the parent caller
// context. The skipParents select how many caller contexts to skip, a value of
// 0 sets the context to the caller of this function.
func (l *Logger) WithCallerContext(skipParents int) *Logger {
	const calleeDepth = 2
	var pc [1]uintptr
	newEntry := l
	i := runtime.Callers(calleeDepth+skipParents, pc[:])
	frame, _ := runtime.CallersFrames(pc[:i]).
		Next()
	if frame.Func != nil {
		newEntry = &Logger{Entry: l.Dup()}
		newEntry.Data[logFieldCaller] = FmtCaller(frame)
	}
	return newEntry
}

// Grab an instance of Logger that may have been passed in context.Context.
// Returns the logger or creates a new instance if none was found in ctx. Since
// Logger is based on logrus.Entry, if logger instance from context is any of
// logrus.Logger, logrus.Entry, necessary adaption will be applied.
func FromContext(ctx context.Context) *Logger {
	l := ctx.Value(loggerContextKey)
	if l == nil {
		return New(Ctx{})
	}

	switch v := l.(type) {
	case *Logger:
		return v
	case *logrus.Entry:
		return NewFromEntry(v, Ctx{})
	case *logrus.Logger:
		return NewFromLogger(v, Ctx{})
	default:
		return New(Ctx{})
	}
}

// WithContext adds logger to context `ctx` and returns the resulting context.
func WithContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}
