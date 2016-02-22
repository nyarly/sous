package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/file"
)

type Buildpacks []Buildpack

type Buildpack struct {
	Name, Desc string
	Scripts    struct {
		Common, Base, Command, Compile, Detect, Test string
	}
}

func ParseBuildpacks(baseDir string) (Buildpacks, error) {
	if !dir.Exists(baseDir) {
		return nil, fmt.Errorf("buildpack dir not found: %s", baseDir)
	}

	common, _ := file.ReadString(filepath.Join(baseDir, "common.sh"))

	packs := Buildpacks{}
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || path == baseDir {
			return nil
		}
		pack, err := ParseBuildpack(path)
		if err != nil {
			return fmt.Errorf("error parsing buildpack at %s: %s", path, err)
		}
		pack.Name = info.Name()
		pack.Scripts.Common = common
		packs = append(packs, pack)
		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}
	return packs, nil
}

func ParseBuildpack(baseDir string) (Buildpack, error) {
	p := Buildpack{}
	var err error
	read := func(filename string) string {
		path := filepath.Join(baseDir, filename)
		s, ok := file.ReadString(path)
		if !ok {
			err = fmt.Errorf("unable to read file %s", path)
		}
		return s
	}
	p.Scripts.Base = read("base.sh")
	p.Scripts.Command = read("command.sh")
	p.Scripts.Compile = read("compile.sh")
	p.Scripts.Detect = read("detect.sh")
	p.Scripts.Test = read("test.sh")
	return p, err
}
