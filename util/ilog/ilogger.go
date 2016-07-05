package ilog

type (
	// ILogger is the interface an object can implement if it wants to be
	// compatible with the Watcher.
	ILogger interface {
		// SetLogFunc sets the log function for standard log messages from this
		// ILogger instance. If set to nil, standard logs must be ignored.
		// nil, then logs must
		SetLogFunc(LogFunc)
		// SetDebugFunc is similar to SetLogFunc except it controls debug-level
		// messages.
		SetDebugFunc(LogFunc)
	}
	// LogFunc is a function that can be called from methods on the object
	// implementing ILogger to log its activity.
	LogFunc func(string, interface{})
	logType int
)

const (
	info logType = iota
	debug
)
