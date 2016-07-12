package ilog

import (
	"fmt"
	"runtime"
)

type (
	// Mixin is a pre-baked implementation of ILogger that you can optionally
	// use to make any struct type instantly compatible with Watcher. It
	// provides some nice functionality like the Debugf and Logf formatting
	// methods, as well as adding file, line and method data to debug output.
	Mixin struct {
		logFunc, debugFunc func(...interface{})
	}
)

// SetLogFunc implements ILogger.SetLogFunc
func (m *Mixin) SetLogFunc(f func(...interface{})) { m.logFunc = f }

// SetDebugFunc implements ILogger.SetDebugFunc
func (m *Mixin) SetDebugFunc(f func(...interface{})) { m.debugFunc = f }

func (m *Mixin) log(message string, data interface{}) {
	if m.logFunc == nil {
		return
	}
	m.logFunc(message, data)
}

func (m *Mixin) debug(message string, data interface{}, calldepth int) {
	if m.debugFunc == nil {
		return
	}
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	m.debugFunc(fmt.Sprintf("%s:%d ", file, line)+message, data)
}

// Log uses fmt.Sprint to construct a textual log message from the parameters.
func (m *Mixin) Log(a ...interface{}) { m.Log(fmt.Sprint(a...), nil) }

// Debug uses fmt.Sprint to construct a textual log message from the parameters.
func (m *Mixin) Debug(a ...interface{}) { m.Debug(fmt.Sprint(a...), nil, 2) }

// Logf uses fmt.Sprintf to construct a textual log message.
func (m *Mixin) Logf(format string, a ...interface{}) {
	m.log(fmt.Sprintf(format, a...), nil)
}

// Logf uses fmt.Sprintf to construct a textual log message.
func (m *Mixin) Debugf(format string, a ...interface{}) {
	m.debug(fmt.Sprintf(format, a...), nil, 2)
}

// LogData logs a textual message along with some free-form data.
func (m *Mixin) LogData(message string, data interface{}) {
	m.log(message, data)
}

// DebugData logs a textual message along with some free-form data.
func (m *Mixin) DebugData(message string, data interface{}) {
	m.debug(message, data, 2)
}
