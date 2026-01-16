package web

import (
	"html/template"
	"regexp"
)

// Regular expressions to allow dynamic routes.
var (
	homeEx   = regexp.MustCompile(`^\/$`)
	queueEx  = regexp.MustCompile(`^\/queue\/[123456789]$`)
	signupEx = regexp.MustCompile(`^\/signup$`)
	signinEx = regexp.MustCompile(`^\/signin$`)
)

const (
	baseTmpl = "./_web/base.tmpl"
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

func ParsePages() (map[*regexp.Regexp]Page, error) {
	pages := make(map[*regexp.Regexp]Page)

	home, err := template.ParseFiles(baseTmpl, "./_web/pages/home.tmpl")
	if err != nil {
		return nil, err
	}
	pages[homeEx] = newPage("Home", "/js/home.js", home)

	queue, err := template.ParseFiles(baseTmpl, "./_web/pages/queue.tmpl")
	if err != nil {
		return nil, err
	}
	pages[queueEx] = newPage("Queue", "/js/queue.js", queue)

	signup, err := template.ParseFiles(baseTmpl, "./_web/pages/signup.tmpl")
	if err != nil {
		return nil, err
	}
	pages[signupEx] = newPage("Sign up", "/js/signup.js", signup)

	signin, err := template.ParseFiles(baseTmpl, "./_web/pages/signin.tmpl")
	if err != nil {
		return nil, err
	}
	pages[signinEx] = newPage("Sign in", "/js/signin.js", signin)

	return pages, nil
}
