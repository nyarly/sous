package build

import (
	"bytes"
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

var dockerfileTemplate = `
FROM {{.From}}
MAINTAINER {{.Maintainer}}

{{if .Labels}}
LABEL {{range $name, $value := .Labels}}{{$name}}={{$value}} \
      {{end}}
{{end}}
{{range .Add}}{{if .Files}}
ADD [{{range .Files}}"{{.}}",{{end}} ".Dest"]
{{end}}
{{if .Workdir}}
WORKDIR {{.Workdir}}
{{end}}
{{range .Run}}
RUN {{.}}
{{end}}
{{if .Entrypoint}}
ENTRYPOINT [{{range $i, $e := .Entrypoint}}{{if $i}},{{end}}"{{$e}}"{{end}}]
{{end}}
{{if .CMD}}
CMD [{{range $i, $e := .CMD}}{{if $i}},{{end}}"{{$e}}"{{end}}]
{{end}}
`
