package web

import (
	"html/template"
	"justchess/internal/db"
)

// page defines the parsed template and data used to fill it.
type page struct {
	// Player who requested the page.
	Player db.Player
	Data   any
	Title  string
	tmpl   *template.Template
}

type QueueData struct {
	Control int
	Bonus   int
}
