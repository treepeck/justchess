package web

import (
	"html/template"
)

const (
	baseTmpl   = "./_web/base.tmpl"
	homeTmpl   = "./_web/pages/home.tmpl"
	queueTmpl  = "./_web/pages/queue.tmpl"
	signupTmpl = "./_web/pages/signup.tmpl"
	signinTmpl = "./_web/pages/signin.tmpl"
)

type Page struct {
	Title    string
	Name     string
	Script   string
	template *template.Template
}

func newPage(title, script string, t *template.Template) Page {
	return Page{
		Title:  title,
		Script: script,
		// By default sign up since user can be unauthorized.
		Name:     "Sign up",
		template: t,
	}
}

func ParsePages() (map[string]Page, error) {
	pages := make(map[string]Page)

	home, err := template.ParseFiles(baseTmpl, homeTmpl)
	if err != nil {
		return nil, err
	}
	pages["/"] = newPage("Home", "/js/home.js", home)

	queue, err := template.ParseFiles(baseTmpl, queueTmpl)
	if err != nil {
		return nil, err
	}
	pages["/queue"] = newPage("Queue", "/js/queue.js", queue)

	signup, err := template.ParseFiles(baseTmpl, signupTmpl)
	if err != nil {
		return nil, err
	}
	pages["/signup"] = newPage("Sign up", "/js/signup.js", signup)

	signin, err := template.ParseFiles(baseTmpl, signinTmpl)
	if err != nil {
		return nil, err
	}
	pages["/signin"] = newPage("Sign in", "/js/signin.js", signin)

	return pages, nil
}
