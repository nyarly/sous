package commands

import "github.com/opentable/sous/core"

func TestHelp() string {
	return `sous test is an alias for sous run test`
}

// Test is simply an alias for `sous run test`
func Test(sous *core.Sous, args []string) {
	Run(sous, []string{"test"})
}
