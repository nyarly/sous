package docker

import "fmt"

type Run struct {
	Image, Name string
	Env         []string
	Net         string
}

func NewRun(image string) *Run {
	return &Run{
		Image: image,
		Net:   "host",
		Env:   []string{},
	}
}

func (r *Run) AddEnv(key, value string) {
	r.Env = append(r.Env, fmt.Sprintf("%s=%s", key, value))
}

func (r *Run) ExitCode() int {
	args := []string{"run"}
	if r.Name != "" {
		args = append(args, "--name", r.Name)
	}
	for _, e := range r.Env {
		args = append(args, "-e", e)
	}
	if r.Net != "" {
		args = append(args, "--net="+r.Net)
	}
	args = append(args, r.Image)
	return dockerCmd(args...).ExitCode()
}
