package core

type Sous struct {
	Version, Revision, OS, Arch string
	Packs                       []Pack
	Commands                    map[string]*Command
	cleanupTasks                []func() error
}

type Command struct {
	Func      func(*Sous, []string)
	HelpFunc  func() string
	ShortDesc string
}

var sous *Sous

func NewSous(version, revision, os, arch string, commands map[string]*Command, packs []Pack) *Sous {
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
