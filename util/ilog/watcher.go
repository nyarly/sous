package ilog

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

type (
	// Watcher is a centralised log message processor, which uses a single
	// channel to ingest and process all log and debug messages. By using a
	// single channel, all messages should be processed in the order they are
	// received. When the buffer is full, Watcher simply discards messages, so
	// you never get blocked waiting for logs to be written.
	Watcher struct {
		buf                     chan Message
		bufSize                 int
		infoWriter, debugWriter LogWriter
		once                    sync.Once
		lock                    sync.RWMutex
		closed                  bool
		done                    chan struct{}
		sourceFilter            func(string) bool
		messageFilter           func(Message) bool
	}
	// LogWriter is used to process log Message.
	LogWriter func(Message)
	// Message is a log message, which is the type passed to LogHandler.
	Message struct {
		// LogType is either Debug or Info
		LogType Level
		// Source is the name of the type writing this log.
		Source,
		// File is the absolute path of the source file writing this message.
		File,
		// Text is the text of this log entry (may be empty)
		Text string
		// Line is the line number of this log entry.
		Line int
		// Fields is a dictionary of keys/values which can be used for
		// structured logging. Fields may be nil, and may be empty. It is
		// populated automatically if the final argument to a LogFunc happens to
		// be a map[string]string, otherwise it is nil.
		Fields map[string]string
	}
)

// NewWatcher creates a ready-to-use watcher. You may pass in nil for either the
// infoWriter or debugWriter to ignore that kind of log output entirely, but this
// is better controlled by calling the respective SetLogFunc and SetDebugFunc on
// the logging object (implementing ILogger) that you want to control log output
// for.
//
// bufSize determines the total size of the buffer for all messages. When the
// buffer is full, log messages are silently dropped.
func NewWatcher(bufSize int, infoWriter, debugWriter LogWriter) *Watcher {
	return &Watcher{
		buf:         make(chan Message, bufSize),
		bufSize:     bufSize,
		infoWriter:  infoWriter,
		debugWriter: debugWriter,
	}
}

// Watch begins watching standard logs from an ILogger.
func (w *Watcher) Watch(obj ILogger) {
	source := reflect.TypeOf(obj).Name()
	obj.SetLogFunc(w.makeLogFunc(Info, source))
	w.once.Do(func() { w.watch() })
}

// WatchDebug begins watching debug-level logs from an ILogger.
func (w *Watcher) WatchDebug(obj ILogger) {
	source := reflect.TypeOf(obj).Name()
	obj.SetDebugFunc(w.makeLogFunc(Debug, source))
	w.once.Do(func() { w.watch() })
}

// UnwatchInfo stops watching standard logs from an ILogger.
func (w *Watcher) Unwatch(obj ILogger) {
	obj.SetLogFunc(nil)
}

// UnwatchDebug stops watching debug-level logs from an ILogger.
func (w *Watcher) UnwatchDebug(obj ILogger) {
	obj.SetDebugFunc(nil)
}

// SetSourceFilter allows you to set a filter for all log messages based on
// their source. It is more efficient than SetMessageFilter, when applicable.
// Messages from sources for which the filter returns false are discarded and
// not placed in the buffer for processing.
// You may set the source filter to nil to remove it and accept all messages
// which are not filtered by the message filter (see SetMessageFilter)
func (w *Watcher) SetSourceFilter(f func(string) bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.sourceFilter = f
}

// SetMessageFilter allows you to set a filter based on all message fields.
// Messages for which the filter returns false are discarded and not placed in
// the buffer for processing.
// You may set the message filter to nil to remove it and accept all messages
// which are not already filtered by the source filter (see SetSourceFilter)
func (w *Watcher) SetMessageFilter(f func(Message) bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.messageFilter = f
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
		return
	}
	<-w.done
}

// CloseWait calls Close, and then Wait.
func (w *Watcher) CloseWait() {
	w.Close()
	w.Wait()
}

func (w *Watcher) makeLogFunc(typ Level, source string) func(...interface{}) {
	return func(v ...interface{}) {
		// calldepth == 2 because we always care about the direct caller of this
		// func.
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 0
		}

		var fields map[string]string

		if len(v) != 0 {
			// If the last arg is a map[string]string then we have fields! Otherwise
			// we have nil, as per spec.
			var hasFields bool
			if fields, hasFields = v[len(v)-1].(map[string]string); hasFields {
				// Remove fields from the list to avoid the risk of printing it
				// twice.
				v = v[:len(v)-2]
			}
		}

		w.write(Message{
			LogType: typ,
			Source:  source,
			File:    file,
			Text:    fmt.Sprint(v...),
			Line:    line,
			Fields:  fields,
		})
	}
}

func (w *Watcher) write(m Message) {
	// "write" uses the read lock, because here we are considering writes to
	// the state of watcher itself, not to the buffered channel.
	w.lock.RLock()
	defer w.lock.RUnlock()
	if w.closed || len(w.buf) == int(w.bufSize) {
		return // Drop, don't block if the buffer is full.
	}
	w.buf <- m
}

func (w *Watcher) watch() {
	w.done = make(chan struct{})
	go func() {
		for m := range w.buf {
			if m.LogType == Info && w.infoWriter != nil {
				w.infoWriter(m)
			} else if m.LogType == Debug && w.debugWriter != nil {
				w.debugWriter(m)
			}
		}
		close(w.done)
	}()
}
