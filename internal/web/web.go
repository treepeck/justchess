package web

import (
	"log"
	"net/http"
	"regexp"

	"justchess/internal/db"
)

// Declaration of error messages.
const (
	msgNotFound        string = "The requested page doesn't exist."
	msgInvalidTemplate string = "The requested page cannot be rendered."
)

type Service struct {
	repo  db.Repo
	pages map[*regexp.Regexp]Page
}

func NewService(pages map[*regexp.Regexp]Page, r db.Repo) Service {
	return Service{pages: pages, repo: r}
}

func (s Service) RegisterRoutes(mux *http.ServeMux) {
	// Serve files from the _web/css folder.
	css := http.Dir("./_web/css")
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(css)))

	// Serve files from the _web/fonts folder.
	fonts := http.Dir("./_web/fonts")
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(fonts)))

	// Serve files from the _web/images folder.
	images := http.Dir("./_web/images")
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(images)))

	// Serve files from the _web/js folder.
	js := http.Dir("./_web/js")
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(js)))

	// Serve pages.
	mux.Handle("/", http.HandlerFunc(s.servePage))
}

func (s Service) servePage(rw http.ResponseWriter, r *http.Request) {
	// Find page using regular expressions.
	exists := false
	page := Page{}
	for ex, p := range s.pages {
		if ex.MatchString(r.URL.Path) {
			exists = true
			page = p
			break
		}
	}

	if !exists {
		http.Error(rw, msgNotFound, http.StatusNotFound)
		return
	}

	c, err := r.Cookie("Auth")
	if err == nil {
		player, err := s.repo.SelectPlayerBySessionId(c.Value)
		if err == nil {
			page.Name = player.Name
		}
	}

	if err := page.template.Execute(rw, page); err != nil {
		log.Print(err)
		http.Error(rw, msgInvalidTemplate, http.StatusInternalServerError)
	}
}
