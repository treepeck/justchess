package web

import (
	"html/template"
	"log"
	"net/http"

	"justchess/internal/db"
)

// Declaration of error messages.
const (
	msgNotFound        string = "The requested page doesn't exist"
	msgInvalidTemplate string = "The requested page cannot be rendered"
)

type Service struct {
	repo db.Repo
	// maps URL path to a [Page].
	pages map[string]page
}

// InitService tries to parse the template files and stores them in the [Service].
func InitService(r db.Repo) (Service, error) {
	pagesData := []struct {
		url      string
		tmplPath string
		base     baseData
	}{
		{"/", "./_web/pages/home.tmpl", baseData{Title: "Home", Script: "/js/home.js"}},
		{"/queue", "./_web/pages/queue.tmpl", baseData{Title: "Queue", Script: "/js/queue.js"}},
		{"/signup", "./_web/pages/signup.tmpl", baseData{Title: "Sign up", Script: "/js/signup.js"}},
		{"/signin", "./_web/pages/signin.tmpl", baseData{Title: "Sign in", Script: "/js/signin.js"}},
		{"/game", "./_web/pages/game.tmpl", baseData{Title: "Game", Script: "/js/game.js"}},
	}

	pages := make(map[string]page, len(pagesData))
	for _, data := range pagesData {
		t, err := template.ParseFiles(basePath, data.tmplPath)
		if err != nil {
			return Service{}, err
		}
		pages[data.url] = page{Base: data.base, template: t}
	}

	return Service{pages: pages, repo: r}, nil
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

	// Serve pages with static routes.
	mux.Handle("/", http.HandlerFunc(s.serveStaticRoutePage))

	// Serve pages with dynamic routes.
	mux.Handle("/queue/", http.StripPrefix("/queue/", http.HandlerFunc(s.serveQueue)))
}

func (s Service) serveQueue(rw http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/queue/"):]

	// There are 9 queues.
	switch id {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		s.renderPage(rw, r, s.pages["/queue"])

	default:
		// TODO: render 404 template.
		http.Error(rw, msgNotFound, http.StatusNotFound)
		return
	}
}

func (s Service) serveStaticRoutePage(rw http.ResponseWriter, r *http.Request) {
	page, exists := s.pages[r.URL.Path]
	if !exists {
		// TODO: render 404 template.
		http.Error(rw, msgNotFound, http.StatusNotFound)
		return
	}

	s.renderPage(rw, r, page)
}

func (s Service) renderPage(rw http.ResponseWriter, r *http.Request, p page) {
	c, err := r.Cookie("Auth")
	if err == nil {
		player, err := s.repo.SelectPlayerBySessionId(c.Value)
		if err == nil {
			p.Base.PlayerName = player.Name
		} else {
			p.Base.PlayerName = "signup"
		}
	}

	if err := p.template.Execute(rw, p); err != nil {
		log.Print(err)
		http.Error(rw, msgInvalidTemplate, http.StatusInternalServerError)
	}
}
