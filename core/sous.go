package core

import "github.com/opentable/sous/build"

type Sous struct {
	Version, Revision, OS, Arch string
	Packs                       []*build.Pack
	Commands                    map[string]*Command
	cleanupTasks                []func() error
}

type Command struct {
	Func      func(*Sous, []string)
	HelpFunc  func() string
	ShortDesc string
}

func NewSous(version, revision, os, arch string, commands map[string]*Command, packs []*build.Pack) *Sous {
	return &Sous{
		Version:      version,
		Revision:     revision,
		OS:           os,
		Arch:         arch,
		Packs:        packs,
		Commands:     commands,
		cleanupTasks: []func() error{},
	}
}

func (s *Sous) AddCleanupTask(f func() error) {
	s.cleanupTasks = append(s.cleanupTasks, f)
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
