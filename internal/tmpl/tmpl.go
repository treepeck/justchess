package tmpl

import (
	"html/template"
	"net/http"
	"strings"
)

type GameData struct {
	WhiteId string
	BlackId string
}

func Exec(rw http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	if name == "/" {
		name = "/home"
	} else if strings.Contains(name, "/game") {
		name = "/game"
	}

	// TODO: cache templates, do not parse for each request.
	t, err := template.ParseFiles(
		"./static/templates/base.html",
		"./static/templates"+name+".html",
	)
	if err != nil {
		http.Error(rw, "Page not found.", http.StatusNotFound)
		return
	}

	http.Error(rw, "Internal server error. Please, try again later.", http.StatusInternalServerError)
	if t.Execute(rw, nil) != nil {
	}
}
