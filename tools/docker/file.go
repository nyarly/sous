package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

type Dockerfile struct {
	From, Maintainer, Workdir string
	Labels                    map[string]string
	Add                       []Add
	Run, Entrypoint           []string
	CMD                       StringList
	LabelPrefix               string
}

type StringList []string

func (sl StringList) String() string {
	if len(sl) == 1 {
		return sl[0]
	}
	j, err := json.Marshal(sl)
	if err != nil {
		panic("Unable to marshal json array")
	}
	return string(j)
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
	if d.LabelPrefix != "" {
		name = fmt.Sprintf("%s.%s", d.LabelPrefix, name)
	}
	if d.Labels == nil {
		d.Labels = map[string]string{}
	}
	d.Labels[name] = value
}

func (d *Dockerfile) AddRun(format string, a ...interface{}) {
	d.Run = append(d.Run, fmt.Sprintf(format, a...))
}

// AddAdd adds an Add line to the Dockerfile. The last argument
// you pass becomes the destination, all previous arguments
// are the source files.
func (d *Dockerfile) AddAdd(f1, f2 string, fn ...string) {
	sources := []string{f1}
	dest := f2
	if len(fn) != 0 {
		dest = fn[len(fn)-1]
		sources = append(sources, f2)
		for _, f := range fn[:len(fn)-1] {
			sources = append(sources, f)
		}
	}
	add := Add{
		Files: sources,
		Dest:  dest,
	}
	if d.Add == nil {
		d.Add = []Add{}
	}
	d.Add = append(d.Add, add)
}

var dockerfileTemplate = `FROM {{.From}}
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
{{if .CMD}}CMD {{.CMD}}{{end}}
`
