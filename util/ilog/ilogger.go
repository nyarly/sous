package ilog

type (
	// ILogger is the interface an object can implement if it wants to be
	// compatible with the Watcher.
	ILogger interface {
		// SetLogFunc sets the log function for standard log messages from this
		// ILogger instance. The Watcher guarantees that this will never be nil.
		//
		// The file and line location of the reported debug call will be the
		// file and line of the direct caller of the provided function.
		// Therefore, in order to get meaningful file and line data, the
		// implementor of this interface must ensure they use this function
		// directly, with no wrappers, in their code.
		SetLogFunc(func(...interface{}))
		// SetDebugFunc is similar to SetLogFunc except it controls debug-level
		// messages. SetDebugFunc will never be passed nil by Watcher.
		//
		// The file and line location of the reported debug call will be the
		// file and line of the direct caller of the provided function.
		// Therefore, in order to get meaningful file and line data, the
		// implementor of this interface must ensure they use this function
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
