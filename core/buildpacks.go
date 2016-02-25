package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/file"
)

type Buildpacks []Buildpack

type Buildpack struct {
	Name, Desc          string
	StackVersions       *StackVersions
	DefaultStackVersion string
	Scripts             struct {
		Common, Base, Command, Compile, Detect, Test, ListBaseimage string
	}
}

func (bps Buildpacks) Detect(dirPath string) Buildpacks {
	packs := Buildpacks{}
	for _, p := range bps {
		if err := p.Detect(dirPath); err == nil {
			packs = append(packs, p)
		}
	}
	return packs
}

func (bp Buildpack) Detect(dirPath string) error {
	path := "./detect.sh"
	data := []byte(bp.Scripts.Detect)
	file.Write(data, path)
	file.RemoveOnExit(path)

	c := exec.Command(path)
	if allout, err := c.CombinedOutput(); err != nil {
		return fmt.Errorf("Error: %s; output from %s:\n%s", err, path, allout)
	}
	return nil
}

func (tc *TargetContext) RunScript(name, contents, inDir string) (string, error) {
	bp := tc.Buildpack
	path := tc.FilePath(name)

	// Add common.sh and base.sh
	contents = fmt.Sprintf("# common.sh\n%s\n\n# base.sh\n%s\n\n# %s\n%s\n",
		bp.Scripts.Common, bp.Scripts.Base, name, contents)

	data := []byte(contents)
	file.Write(data, path)
	file.RemoveOnExit(path)

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	combined := &bytes.Buffer{}

	teeout := io.MultiWriter(stdout, combined)
	teeerr := io.MultiWriter(stderr, combined)

	c := exec.Command(path)
	c.Dir = inDir
	c.Stdout = teeout
	c.Stderr = teeerr

	if err := c.Start(); err != nil {
		return "", err
	}

	if err := c.Wait(); err != nil {
		return "", fmt.Errorf("Error: %s; output from %s:\n%s", err, name, combined.String())
	}

	return stdout.String(), nil
}

func (tc *TargetContext) BaseImage(dirPath, targetName string) (string, error) {
	bp := tc.Buildpack
	detected, err := tc.RunScript("detect.sh", bp.Scripts.Detect, dirPath)
	if err != nil {
		return "", err
	}
	parts := strings.Split(detected, " ")
	if len(parts) != 2 || parts[0] != bp.Name {
		return "", fmt.Errorf("detect.sh returned %s; want '%s <stackversion>' where <stackversion> is either 'default' or semver range", bp.Name)
	}
	stackVersion := parts[1]
	if stackVersion == "default" {
		stackVersion = bp.DefaultStackVersion
	}
	image, ok := bp.StackVersions.GetBaseImageTag(stackVersion, targetName)
	if !ok {
		return "", fmt.Errorf("buildpack %s does not have a base image for version %s", bp.Name, stackVersion)
	}
	return image, nil
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
