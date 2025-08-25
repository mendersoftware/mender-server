package log

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

const MaxTraceback = 32

var (
	ErrPanic = errors.New("recovered from panic")
)

func CollectTrace() string {
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

type RecoveryOption struct {
	channel chan error
	err     error
}

func NewRecoveryOption() *RecoveryOption {
	return &RecoveryOption{}
}

func (o *RecoveryOption) WithChannel(channel chan error) *RecoveryOption {
	o.channel = channel
	return o
}

func (o *RecoveryOption) WithError(err error) *RecoveryOption {
	o.err = err
	return o
}

// SimpleRecovery recovers from panics, logs the panic and trace,
// and optionally sends the error to an error channel.
func (l *Logger) SimpleRecovery(opts ...*RecoveryOption) {
	r := recover()
	if r == nil {
		return
	}
	trace := CollectTrace()
	var err error
	var opt *RecoveryOption
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt == nil || opt.err == nil {
		err = ErrPanic
	} else {
		err = errors.Wrap(opt.err, ErrPanic.Error())
	}

	if opt != nil && opt.channel != nil {
		select {
		case opt.channel <- err:
		default:
			// avoid blocking if there is no receiver
		}
	}

	l.WithField("panic", r).
		WithField("trace", trace).
		Error(err)
}
