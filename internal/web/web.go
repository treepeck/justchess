package web

import (
	"net/http"

	"justchess/internal/db"
)

type Service struct {
	repo  db.Repo
	pages map[string]Page
}

func NewService(pages map[string]Page, r db.Repo) Service {
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
	p, exists := s.pages[r.URL.Path]
	if !exists {
		http.Redirect(rw, r, "/404", http.StatusNotFound)
		return
	}

	c, err := r.Cookie("Auth")
	if err == nil {
		player, err := s.repo.SelectPlayerBySessionId(c.Value)
		if err == nil {
			p.Name = player.Name
		}
	}

	p.exec(rw)
}
