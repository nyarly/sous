package ilog

import (
	"fmt"
	"sync"
)

type (
	// LogFunc is a function that can be called from methods on the object
	// implementing ILogger to log its activity.
	LogFunc func(string, interface{})
	// ILogger is the interface an object can implement if it wants to be
	// compatible with the Watcher.
	ILogger interface {
		SetLogFunc(LogFunc)
		SetDebugFunc(LogFunc)
	}
	// Mixin is a pre-baked implementation of ILogger that you can optionally
	// use to make any struct type instantly compatible with Watcher. It
	// provides some nice functionality like the Debugf and Logf formatting
	// methods, as well as adding file, line and method data to debug output.
	Mixin struct {
		logFunc, debugFunc LogFunc
	}
	// Watcher is a centralised log message processor, which uses a single
	// channel to ingest and process all log and debug messages. By using a
	// single channel, all messages should be processed in the order they are
	// received. When the buffer is full, Watcher simply discards messages, so
	// you never get blocked waiting for logs to be written.
	Watcher struct {
		buf                     chan message
		bufSize                 int
		infoWriter, debugWriter LogWriter
		once                    sync.Once
		lock                    sync.RWMutex
		closed                  bool
		done                    chan struct{}
	}
	// LogWriter is a special type of LogFunc that is used to output messages.
	LogWriter LogFunc
	logType   int
	message   struct {
		typ  logType
		text string
		data interface{}
	}
)

const (
	info logType = iota
	debug
)

// SetLogFunc implements ILogger.SetLogFunc
func (m *Mixin) SetLogFunc(f LogFunc) { m.logFunc = f }

// SetDebugFunc implements ILogger.SetDebugFunc
func (m *Mixin) SetDebugFunc(f LogFunc) { m.debugFunc = f }

func (m *Mixin) log(message string, data interface{}) {
	if m.logFunc == nil {
		return
	}
	m.logFunc(message, data)
}

func (m *Mixin) debug(message string, data interface{}) {
	if m.debugFunc == nil {
		return
	}
	m.debugFunc(message, data)
}

// Log uses fmt.Sprint to construct a textual log message from the parameters.
func (m *Mixin) Log(a ...interface{}) { m.Log(fmt.Sprint(a...), nil) }

// Debug uses fmt.Sprint to construct a textual log message from the parameters.
func (m *Mixin) Debug(a ...interface{}) { m.Debug(fmt.Sprint(a...), nil) }

// Logf uses fmt.Sprintf to construct a textual log message.
func (m *Mixin) Logf(format string, a ...interface{}) {
	m.log(fmt.Sprintf(format, a...), nil)
}

// Logf uses fmt.Sprintf to construct a textual log message.
func (m *Mixin) Debugf(format string, a ...interface{}) {
	m.debug(fmt.Sprintf(format, a...), nil)
}

// LogData logs a textual message along with some free-form data.
func (m *Mixin) LogData(message string, data interface{}) {
	m.log(message, data)
}

// DebugData logs a textual message along with some free-form data.
func (m *Mixin) DebugData(message string, data interface{}) {
	m.debug(message, data)
}

// Watch begins watching logs from an ILogger.
func (w *Watcher) Watch(obj ILogger) {
	obj.SetLogFunc(w.makeLogFunc(info))
	obj.SetDebugFunc(w.makeLogFunc(debug))
	w.once.Do(func() { w.watch() })
}

// Close causes all further incoming messages to be discarded. Messages which
// are already buffered will still be processed.
func (w *Watcher) Close() {
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.closed {
		return
	}
	close(w.buf)
	w.closed = true
}

// Wait waits until all buffered log messages have been processed. You must
// always call Close before calling Wait, otherwise you will wait forever. Or,
// better, use CloseWait instead.
func (w *Watcher) Wait() {
	if w.done == nil {
		panic("Watcher.Wait called before Watcher.Watch")
	}
	<-w.done
}

// CloseWait calls Close, and then Wait.
func (w *Watcher) CloseWait() {
	w.Close()
	w.Wait()
}

func (w *Watcher) makeLogFunc(typ logType) LogFunc {
	return func(m string, data interface{}) {
		w.write(message{typ, m, data})
	}
}

func (w *Watcher) write(m message) {
	// "write" uses the read lock, because here we are considering writes to
	// the state of watcher itself.
	w.lock.RLock()
	defer w.lock.RUnlock()
	if w.closed || len(w.buf) == int(w.bufSize) {
		return // Drop, don't block.
	}
	w.buf <- m
}

func (w *Watcher) watch() {
	w.done = make(chan struct{})
	go func() {
		for m := range w.buf {
			if m.typ == info && w.infoWriter != nil {
				w.infoWriter(m.text, m.data)
			} else if m.typ == debug && w.debugWriter != nil {
				w.debugWriter(m.text, m.data)
			}
		}
		close(w.done)
	}()
}

func NewWatcher(bufSize int, infoWriter, debugWriter LogWriter) *Watcher {
	return &Watcher{
		buf:         make(chan message, bufSize),
		bufSize:     bufSize,
		infoWriter:  infoWriter,
		debugWriter: debugWriter,
	}
}
