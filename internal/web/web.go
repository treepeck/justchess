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
	playerRepo db.PlayerRepo
	gameRepo   db.GameRepo
	// maps URL path to a [Page].
	pages map[string]page
}

// InitService tries to parse the template files and stores them in the [Service].
func InitService(pr db.PlayerRepo, gr db.GameRepo) (Service, error) {
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
		{"/404", "./_web/templates/404.tmpl", baseData{Title: "Not found"}},
	}

	pages := make(map[string]page, len(pagesData))
	for _, data := range pagesData {
		t, err := template.ParseFiles(basePath, data.tmplPath)
		if err != nil {
			return Service{}, err
		}
		pages[data.url] = page{Base: data.base, template: t}
	}

	return Service{
		pages:      pages,
		playerRepo: pr,
		gameRepo:   gr,
	}, nil
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
	var data queueData

	// There are 9 queues.
	switch r.URL.Path {
	case "1":
		data = queueData{Control: 1, Bonus: 0}
	case "2":
		data = queueData{Control: 2, Bonus: 1}
	case "3":
		data = queueData{Control: 3, Bonus: 0}
	case "4":
		data = queueData{Control: 3, Bonus: 2}
	case "5":
		data = queueData{Control: 5, Bonus: 0}
	case "6":
		data = queueData{Control: 5, Bonus: 2}
	case "7":
		data = queueData{Control: 10, Bonus: 0}
	case "8":
		data = queueData{Control: 10, Bonus: 10}
	case "9":
		data = queueData{Control: 15, Bonus: 10}

	default:
		p := s.pages["/404"]
		s.renderPage(rw, r, p)
		return
	}

	p := s.pages["/queue"]
	p.Data = data
	s.renderPage(rw, r, p)
}

func (s Service) serveGame(rw http.ResponseWriter, r *http.Request) {
	g, err := s.gameRepo.SelectById(r.URL.Path)
	if err != nil {
		p := s.pages["/404"]
		s.renderPage(rw, r, p)
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
		page = s.pages["/404"]
	}

	s.renderPage(rw, r, page)
}

func (s Service) renderPage(rw http.ResponseWriter, r *http.Request, p page) {
	p.Base.Player.Name = "signup"
	c, err := r.Cookie("Auth")
	if err == nil {
		player, err := s.playerRepo.SelectBySessionId(c.Value)
		if err == nil {
			p.Base.Player = player
		}
	}

	if err := p.template.Execute(rw, p); err != nil {
		log.Print(err)
		http.Error(rw, msgInvalidTemplate, http.StatusInternalServerError)
	}
}
