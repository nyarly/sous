package singularity

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	// Log collects various loggers to use for different levels of logging
	// TODO: Copied from sous/lib; replace with externally provided logger.
	Log = struct {
		Debug  *log.Logger
		Info   *log.Logger
		Warn   *log.Logger
		Notice *log.Logger
		Vomit  *log.Logger
	}{
		// Debug is a logger - use log.SetOutput to get output from
		Vomit:  log.New(ioutil.Discard, "vomit: ", log.Lshortfile),
		Debug:  log.New(ioutil.Discard, "debug: ", log.Lshortfile),
		Info:   log.New(ioutil.Discard, "info: ", 0),
		Warn:   log.New(os.Stderr, "warn: ", 0),
		Notice: log.New(ioutil.Discard, "notice: ", 0),
	}
)
