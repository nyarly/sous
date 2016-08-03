package main

import (
	"fmt"
	"log"
	"os"

	"github.com/opentable/sous/cli"
)

func main() {
	defer handlePanic()
	// Eventually, these should become flags on the top level application
	//sous.Log.Info.SetOutput(os.Stderr)
	//sous.Log.Debug.SetOutput(os.Stderr)
	log.SetFlags(log.Flags() | log.Lshortfile)

	c, err := cli.NewSousCLI(Version, os.Stdout, os.Stderr)
	if err != nil {
		die(err)
	}

	result := c.Invoke(os.Args)
	exitCode := result.ExitCode()

	os.Exit(exitCode)
}

// die is only used to exit during very early initialisation
func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(70)
}

// handlePanic gives us one last chance to send a message to the user in case a
// panic leaks right up to the top of the program. You can disable this message
// for brevity of output by setting DEBUG=YES
func handlePanic() {
	if os.Getenv("DEBUG") == "YES" {
		return
	}
	fmt.Println(panicMessage)
	fmt.Printf("Sous Version: %s\n\n", Version)
}

const panicMessage = `
################################################################################
#                                                                              #
#                                       OOPS                                   #
#                                                                              #
#        Sous has panicked, due to programmer error. Please report this        #
#        to the project maintainers at:                                        #
#                                                                              #
#                https://github.com/opentable/sous/issues                      #
#                                                                              #
#        Please include this entire message and the stack trace below          #
#        and we will investigate and fix it as soon as possible.               #
#                                                                              #
#        Thanks for your help in improving Sous for all!                       #
#                                                                              #
#        - The OpenTable DevTools Team                                         #
#                                                                              #
################################################################################
`
