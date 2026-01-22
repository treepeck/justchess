package web

import "html/template"

// Relative path to a base.tmpl file.
const basePath string = "./_web/base.tmpl"

// baseData is a data object used to fill up the base.tmpl file while executing
// a template.
type baseData struct {
	Title      string
	PlayerName string
	Script     string
}

// queueData is a data object used to fill up the queue.tmpl file while executing
// a template.
type queueData struct {
	Control int
	Bonus   int
}

// page combines the data objects and a parsed template.
type page struct {
	Base     baseData
	Data     any
	template *template.Template
}
