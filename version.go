package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"text/template"
)

var (
	appName   = "app"
	version   = "major.minor.patch"
	branch    = "git/branch"
	revision  = "git/revision"
	buildDate = "yyyy-mm-dd hh:mm:ss"
	goVersion = runtime.Version()
)

const versionInfoTmpl = `
{{.program}}, version: {{.version}} 
(branch: {{.branch}}; revision: {{.revision}})
  build date:   {{.buildDate}}
  go version:   {{.goVersion}}
`

func PrintVersion() {
	m := map[string]string{
		"program":   appName,
		"version":   version,
		"branch":    branch,
		"revision":  revision,
		"buildDate": buildDate,
		"goVersion": goVersion,
	}

	tmpl, err := template.Must(template.New("version"), nil).Parse(versionInfoTmpl)
	if err != nil {
		panic(err)
	}

	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, m); err != nil {
		panic(err)
	}

	fmt.Println(strings.TrimSpace(buf.String()))
}
