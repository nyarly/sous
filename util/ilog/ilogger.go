// Package ilog (for "interface log") provides an interface and a filtered queue
// for getting log messages out of your structs. It's IoC for logging.
//
// ilog is designed for large projects made of co-operating types spanning
// multiple packages, where fine-grained log and debug output is desired.
// Compliant types need only implement one or more of the interfaces, and need
// not reference this package.  In fact, any type already using the standard
// library log package can easily add support for ilog by adding a few tiny
// methods.
//
// ilog follows standard library log conventions, and specifies a new convention
// for adding structured data to your logs. It provides simple output options,
// but its real power is in gathering and filtering log messages, rather than
// processing/displaying/shipping them. For that purpose, we recommend you use
// implement a LogWriter using something like sirupsen/logrus, or
// inconshreveable/log15 as a log processor.
package ilog

type (
	// ILogger is the union of InfoLogger and DebugLogger.
	ILogger interface {
		InfoLogger
		DebugLogger
	}
	// InfoLogger is the interface an object can implement if it wants to be
	// compatible with being watched for regular log messages.
	InfoLogger interface {
		// SetLogFunc sets the log function for standard log messages from this
		// instance. The Watcher guarantees that this will never be nil.
		//
		// The file and line location of the reported debug call will be the
		// file and line of the direct caller of the provided function.
		// Therefore, in order to get meaningful file and line data, the
		// implementer of this interface must ensure they use this function
		// directly, with no wrappers, in their code.
		SetLogFunc(func(...interface{}))
	}
	// DebugLogger is the interface an object can implement if it wants to be
	// compatible with being watched for debug log messages.
	DebugLogger interface {
		// SetDebugFunc is similar to SetLogFunc except it controls debug-level
		// messages. SetDebugFunc will never be passed nil by Watcher.
		//
		// The file and line location of the reported debug call will be the
		// file and line of the direct caller of the provided function.
		// Therefore, in order to get meaningful file and line data, the
		// implementer of this interface must ensure they use this function
		// directly, with no wrappers, in their code.
		SetDebugFunc(func(...interface{}))
	}
	// Level is the log level.
	Level int
)

const (
	// Info is the log level added to non-debug messages.
	Info Level = iota
	// Debug is the log level added to messages created by calling debug funcs.
	Debug
)
