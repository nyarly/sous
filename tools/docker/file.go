package docker

import (
	"bytes"
	"fmt"
	"text/template"
)

type Dockerfile struct {
	From, Maintainer, Workdir string
	Labels                    map[string]string
	Add                       []Add
	Run, Entrypoint, CMD      []string
}

type Add struct {
	Files []string
	Dest  string
}

func (d *Dockerfile) Render() string {
	t := template.Must(template.New("Dockerfile").Parse(dockerfileTemplate))
	buf := &bytes.Buffer{}
	t.Execute(buf, d)
	return buf.String()
}

func (d *Dockerfile) AddLabel(name, value string) {
	if d.Labels == nil {
		d.Labels = map[string]string{}
	}
	d.Labels[name] = value
}

func (d *Dockerfile) AddRun(format string, a ...interface{}) {
	d.Run = append(d.Run, fmt.Sprintf(format, a...))
}

var dockerfileTemplate = `
FROM {{.From}}
MAINTAINER {{.Maintainer}}
{{if .Labels}}{{$first:=true}}
LABEL {{range $name, $value := .Labels}} \
      {{$name}}={{$value}}{{end}}{{end}}
{{range .Add}}{{if .Files}}ADD [{{range .Files}}"{{.}}", {{end}}"{{.Dest}}"]
{{end}}{{end}}{{if .Workdir}}
WORKDIR {{.Workdir}}{{end}}
{{range .Run}}RUN {{.}}
{{end}}
{{if .Entrypoint}}ENTRYPOINT [{{range $i, $e := .Entrypoint}}{{if $i}},{{end}}"{{$e}}"{{end}}]{{end}}
{{if .CMD}}CMD [{{range $i, $e := .CMD}}{{if $i}},{{end}}"{{$e}}"{{end}}]{{end}}

`
