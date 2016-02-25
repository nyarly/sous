package commands

import (
	"flag"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func LogsHelp() string {
	return `sous logs pipes your applications' stdout and stderr to stdout and stderr, use -f to follow`
}

var logsFlags = flag.NewFlagSet("logs", flag.ExitOnError)

var follow = logsFlags.Bool("f", false, "keep watching log output (similar to tail -f)")
var lines = logsFlags.Int("n", 0, "number of lines to print")

func Logs(sous *core.Sous, args []string) {
	logsFlags.Parse(args)
	args = logsFlags.Args()
	target := "app"
	if len(args) != 0 {
		target = args[0]
	}
	tc := sous.TargetContext(target)

	out := makeTail(tc.FilePath("stdout"), *follow, *lines, os.Stdout)
	err := makeTail(tc.FilePath("stderr"), *follow, *lines, os.Stderr)

	if err := out.Start(); err != nil {
		cli.Fatalf("Unable to begin tailing %s", tc.FilePath("stdout"))
	}
	if err := err.Start(); err != nil {
		cli.Fatalf("Unable to begin tailing %s", tc.FilePath("stderr"))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if *follow {
		<-c
		out.Process.Signal(os.Interrupt)
		err.Process.Signal(os.Interrupt)
	}

	errs := []string{}
	if err := out.Wait(); err != nil {
		errs = append(errs, err.Error())
	}
	if err := err.Wait(); err != nil {
		errs = append(errs, err.Error())
	}
	if !*follow {
		if len(errs) != 0 {
			cli.Fatalf("Done with errors: %s", strings.Join(errs, ", "))
		}
	}
	cli.Success()
}

func makeTail(file string, follow bool, lines int, out io.Writer) *exec.Cmd {
	tail := exec.Command("tail")
	if follow {
		tail.Args = append(tail.Args, "-f")
	}
	if lines > 0 {
		tail.Args = append(tail.Args, "-n"+strconv.Itoa(lines))
	}
	tail.Args = append(tail.Args, file)
	tail.Stdout = out
	return tail
}
