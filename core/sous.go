package core

import "github.com/opentable/sous/tools/cli"

type Sous struct {
	Version, Revision, OS, Arch string
	Packs                       []*Pack
	Commands                    map[string]*Command
	cleanupTasks                []func() error
}

type Command struct {
	Func      func(*Sous, []string)
	HelpFunc  func() string
	ShortDesc string
}

var sous *Sous

func NewSous(version, revision, os, arch string, commands map[string]*Command, packs []*Pack) *Sous {
	if sous == nil {
		sous = &Sous{
			Version:      version,
			Revision:     revision,
			OS:           os,
			Arch:         arch,
			Packs:        packs,
			Commands:     commands,
			cleanupTasks: []func() error{},
		}
	}
	return sous
}

func (s *Sous) AddCleanupTask(f func() error) {
	s.cleanupTasks = append(s.cleanupTasks, f)
}

func AttemptCleanup() {
	if sous == nil {
		return
	}
	if sous.NeedsCleanup() {
		cli.Logf("\nCleaning up...")
		errs := sous.Cleanup()
		if len(errs) == 0 {
			cli.Success()
		}
		for _, e := range errs {
			cli.Logf("=> %s", e)
		}
		cli.Fatalf("Errors cleaning up, see above.")
	}
}

func (s *Sous) Cleanup() []error {
	errors := []error{}
	for _, c := range s.cleanupTasks {
		if err := c(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

func (s *Sous) NeedsCleanup() bool {
	return len(s.cleanupTasks) != 0
}
