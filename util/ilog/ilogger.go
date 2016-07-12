package ilog

type (
	// ILogger is the interface an object can implement if it wants to be
	// compatible with the Watcher.
	ILogger interface {
		// SetLogFunc sets the log function for standard log messages from this
		// ILogger instance. If set to nil, standard logs must be ignored.
		// nil, then logs must
		SetLogFunc(func(...interface{}))
		// SetDebugFunc is similar to SetLogFunc except it controls debug-level
		// messages.
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
