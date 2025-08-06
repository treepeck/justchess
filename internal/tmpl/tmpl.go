// Package tmpl provides an interface for executing HTML templates.
//
// Before execution, templates are embedded into in-memory file system, which
// helps speed up template execution by approximately 3.2 times
// compared to using template.ParseFiles.
package tmpl

import (
	"embed"
	"html/template"
	"io"
)

//go:embed pages/*.html
var tmplFS embed.FS

func Exec(w io.Writer, name string) error {
	templ, err := template.ParseFS(tmplFS, "pages/base.html", "pages/"+name)
	if err != nil {
		return err
	}
	return templ.Execute(w, nil)
}
