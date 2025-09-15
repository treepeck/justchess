package tmpl

import (
	"html/template"
	"net/http"
)

func Exec(rw http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"./static/templates/base.html",
		"./static/templates"+r.URL.Path+".html",
	)
	if err != nil {
		http.Error(rw, "Page not found.", http.StatusNotFound)
		return
	}

	if t.Execute(rw, nil) != nil {
		http.Error(rw, "Internal server error. Please, try again later.", http.StatusInternalServerError)
		return
	}
}
