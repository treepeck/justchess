package web

import (
	"html/template"
	"log"
	"net/http"
)

const (
	baseTmpl   = "./_web/base.tmpl"
	homeTmpl   = "./_web/pages/home.tmpl"
	signupTmpl = "./_web/pages/signup.tmpl"
	signinTmpl = "./_web/pages/signin.tmpl"

	msgInvalidTemplate = "The requested page cannot be rendered."
)

type Page struct {
	Title    string
	Name     string
	Script   string
	template *template.Template
}

func NewPage(title, script string, t *template.Template) Page {
	return Page{
		Title:  title,
		Script: script,
		// By default sign up since user can be unauthorized.
		Name:     "Sign up",
		template: t,
	}
}

func (p Page) exec(rw http.ResponseWriter) {
	if err := p.template.Execute(rw, p); err != nil {
		log.Print(err)
		http.Error(rw, msgInvalidTemplate, http.StatusInternalServerError)
	}
}

func ParsePages() (map[string]Page, error) {
	pages := make(map[string]Page)

	home, err := template.ParseFiles(baseTmpl, homeTmpl)
	if err != nil {
		return nil, err
	}
	pages["/"] = NewPage("Home", "/js/home.js", home)

	signup, err := template.ParseFiles(baseTmpl, signupTmpl)
	if err != nil {
		return nil, err
	}
	pages["/signup"] = NewPage("Sign up", "/js/signup.js", signup)

	signin, err := template.ParseFiles(baseTmpl, signinTmpl)
	if err != nil {
		return nil, err
	}
	pages["/signin"] = NewPage("Sign in", "/js/signin.js", signin)

	return pages, nil
}
