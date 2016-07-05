package ilog

import (
	"reflect"
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
	}
	// LogWriter is a special type of LogFunc that is used to output messages.
	LogWriter LogFunc
	// Message is a log message, which is the type passed to LogHandler.
	Message struct {
		typ logType
		Source,
		Text string
		Data interface{}
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
	obj.SetLogFunc(w.makeLogFunc(info, source))
	w.once.Do(func() { w.watch() })
}

// WatchDebug begins watching debug-level logs from an ILogger.
func (w *Watcher) WatchDebug(obj ILogger) {
	source := reflect.TypeOf(obj).Name()
	obj.SetDebugFunc(w.makeLogFunc(debug, source))
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

func (w *Watcher) makeLogFunc(typ logType, source string) LogFunc {
	return func(m string, data interface{}) {
		w.write(Message{typ, source, m, data})
	}
}

func (w *Watcher) write(m Message) {
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
				w.infoWriter(m.Text, m.Data)
			} else if m.typ == debug && w.debugWriter != nil {
				w.debugWriter(m.Text, m.Data)
			}
		}
		close(w.done)
	}()
}
