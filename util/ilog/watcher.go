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
		// Tags is a list of all tags applied to this message.
		Tags []string
		// Fields is a dictionary of keys/values which can be used for
		// structured logging. Fields may be nil, and may be empty. It is
		// populated automatically if the final argument to a LogFunc happens to
		// be a map[string]string, otherwise it is nil.
		Fields map[string]string
	}
	// Options provides an object-level set of configuration on how log messages
	// should be collected.
	Options struct {
		// Tags are applied to every log message sent by this type, and are used
		// mainly for filtering.
		Tags []string
		// AddFields allows you to add fields to each log message. These fields
		// will be overridden by any fields with matching keys set by the log
		// message itself.
		AddFields map[string]string
		// EnableFileAndLineNumber is used for performance tuning, to stop the
		// call to runtime.Caller used to identify the file and line the log
		// message came from.
		EnableFileAndLineNumber,
		// EnableStackTrace populates all log messages from this watch call with
		// the full stack trace of where the message came from.
		EnableStackTrace bool
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

// Watch begins watching all non-debug-level logs from an ILogger, configured to
// record file and line number, and not to record the stack trace.
// You may optionally pass a configure function to override these defaults.
func (w *Watcher) Watch(obj ILogger, configure ...func(*Options)) {
	defaultOpts := Options{
		EnableFileAndLineNumber: true,
	}
	w.addWatch(Info, obj, defaultOpts, configure, obj.SetLogFunc)
}

// WatchDebug begins watching debug-level logs from an ILogger, configured to
// record file, line number, and stack trace.
// You may optionally pass a configure function to override these defaults.
func (w *Watcher) WatchDebug(obj ILogger, configure ...func(*Options)) {
	defaultOpts := Options{
		EnableFileAndLineNumber: true,
		EnableStackTrace:        true,
	}
	w.addWatch(Debug, obj, defaultOpts, configure, obj.SetDebugFunc)
}

func (w *Watcher) addWatch(l Level, obj ILogger, o Options, c []func(*Options), f func(func(...interface{}))) {
	for _, cf := range c {
		cf(&o)
	}
	source := reflect.TypeOf(obj).Name()
	f(w.makeLogFunc(l, source, o))
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

func (w *Watcher) makeLogFunc(l Level, source string, opts Options) func(...interface{}) {
	return func(v ...interface{}) {
		// source filter before anything else
		if w.sourceFilter != nil && !w.sourceFilter(source) {
			return
		}

		var (
			file           string
			line           int
			gotFileAndLine bool
			fields         map[string]string
		)

		if opts.EnableFileAndLineNumber {
			// calldepth == 2 because we always care about the direct caller of this
			// func.
			if _, file, line, gotFileAndLine = runtime.Caller(2); !gotFileAndLine {
				file = "???"
				line = 0
			}
		}

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

		m := Message{
			LogType: l,
			Source:  source,
			File:    file,
			Tags:    opts.Tags,
			Text:    fmt.Sprint(v...),
			Line:    line,
			Fields:  fields,
		}
		// finally, apply any message filter
		if w.messageFilter == nil || w.messageFilter(m) {
			w.write(m)
		}
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
