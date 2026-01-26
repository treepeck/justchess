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
		{"/", "./_web/templates/home.tmpl", baseData{Title: "Home"}},
		{"/queue", "./_web/templates/queue.tmpl", baseData{Title: "Queue"}},
		{"/signup", "./_web/templates/signup.tmpl", baseData{Title: "Sign up"}},
		{"/signin", "./_web/templates/signin.tmpl", baseData{Title: "Sign in"}},
		{"/game", "./_web/templates/game.tmpl", baseData{Title: "Game"}},
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
	// Serve js and css bundles.
	bundles := http.Dir("./_web/bundles")
	mux.Handle("/bundles/", http.StripPrefix("/bundles/", http.FileServer(bundles)))

	// Serve files from the _web/fonts folder.
	fonts := http.Dir("./_web/fonts")
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(fonts)))

	// Serve files from the _web/images folder.
	images := http.Dir("./_web/images")
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(images)))

	// Serve pages with static routes.
	mux.Handle("/", http.HandlerFunc(s.serveStaticRoutePage))

	// Serve pages with dynamic routes.
	mux.Handle("/queue/", http.StripPrefix("/queue/", http.HandlerFunc(s.serveQueue)))
	mux.Handle("/game/", http.StripPrefix("/game/", http.HandlerFunc(s.serveGame)))
}

func (s Service) serveQueue(rw http.ResponseWriter, r *http.Request) {
	page := s.pages["/queue"]

	// There are 9 queues.
	switch r.URL.Path {
	case "1":
		page.Data = queueData{Control: 1, Bonus: 0}
	case "2":
		page.Data = queueData{Control: 2, Bonus: 1}
	case "3":
		page.Data = queueData{Control: 3, Bonus: 0}
	case "4":
		page.Data = queueData{Control: 3, Bonus: 2}
	case "5":
		page.Data = queueData{Control: 5, Bonus: 0}
	case "6":
		page.Data = queueData{Control: 5, Bonus: 2}
	case "7":
		page.Data = queueData{Control: 10, Bonus: 0}
	case "8":
		page.Data = queueData{Control: 10, Bonus: 10}
	case "9":
		page.Data = queueData{Control: 15, Bonus: 10}

	default:
		// TODO: render 404 template.
		http.Error(rw, msgNotFound, http.StatusNotFound)
		return
	}

	s.renderPage(rw, r, page)
}

func (s Service) serveGame(rw http.ResponseWriter, r *http.Request) {
	g, err := s.repo.SelectGameById(r.URL.Path)
	if err != nil {
		// TODO: render 404 template.
		http.Error(rw, msgNotFound, http.StatusNotFound)
		return
	}

	page := s.pages["/game"]
	// Fill up the template with more game data.
	page.Data = g

	s.renderPage(rw, r, page)
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
