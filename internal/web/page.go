package web

import (
	"html/template"
	"justchess/internal/db"
)

// baseData is a data object used to fill up the base.tmpl file while executing
// a template.
type baseData struct {
	Title  string
	Player db.Player
}

// QueueData is a data object used to fill up the queue.tmpl file while executing
// a template.
type QueueData struct {
	Control int
	Bonus   int
}

// page combines the data objects and a parsed template.
type page struct {
	Base     baseData
	Data     any
	template *template.Template
}
