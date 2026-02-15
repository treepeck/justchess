package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"justchess/internal/db"

	"github.com/treepeck/chego"
)

// Declaration of error messages.
const (
	msgNotFound        string = "The requested page doesn't exist"
	msgInvalidTemplate string = "The requested page cannot be rendered"
)

// Template path prefix.
const tmplPrefix = "./_web/templates"

type Service struct {
	playerRepo db.PlayerRepo
	gameRepo   db.GameRepo
	// maps URL path to a [Page].
	pages map[string]page
}

// InitService tries to parse the template files and stores them in the [Service].
func InitService(pr db.PlayerRepo, gr db.GameRepo) (Service, error) {
	pagesData := []struct {
		url   string
		tmpls []string
		base  baseData
	}{
		{"/", []string{"/home.tmpl"}, baseData{Title: "Home"}},
		{"/queue", []string{"/queue.tmpl"}, baseData{Title: "Queue"}},
		{"/signup", []string{"/signup.tmpl"}, baseData{Title: "Sign up"}},
		{"/signin", []string{"/signin.tmpl"}, baseData{Title: "Sign in"}},
		{"/active", []string{"/active_game.tmpl", "/board.tmpl"}, baseData{}},
		{"/archive", []string{"/archive_game.tmpl", "/board.tmpl"}, baseData{}},
		{"/error", []string{"/error.tmpl"}, baseData{Title: "Error"}},
	}

	pages := make(map[string]page, len(pagesData))
	for _, data := range pagesData {
		paths := append([]string{"/base.tmpl"}, data.tmpls...)
		// Append template prefix to each path.
		for i, path := range paths {
			paths[i] = tmplPrefix + path
		}

		t, err := template.ParseFiles(paths...)
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

// Possible time controls.
var controls = [9]queueData{
	{1, 0}, {2, 1}, {3, 0}, {3, 2}, {5, 0}, {5, 2}, {10, 0}, {10, 10}, {15, 10},
}

func (s Service) serveQueue(rw http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path)
	if err != nil || id < 0 || id > 8 {
		p := s.pages["/error"]
		s.renderPage(rw, r, p)
	}

	p := s.pages["/queue"]
	p.Data = controls[id]
	s.renderPage(rw, r, p)
}

func (s Service) serveGame(rw http.ResponseWriter, r *http.Request) {
	g, err := s.gameRepo.SelectById(r.URL.Path)
	if err != nil {
		p := s.pages["/error"]
		s.renderPage(rw, r, p)
		return
	}

	page := s.pages["/active"]
	// If the game was been already terminated, serve the corresponding page.
	if g.Termination != chego.Unterminated {
		page = s.pages["/archive"]
	}

	// Fill up the template with more game data.
	page.Data = g

	page.Base.Title = g.White.Name + " vs " + g.Black.Name

	s.renderPage(rw, r, page)
}

func (s Service) serveStaticRoutePage(rw http.ResponseWriter, r *http.Request) {
	page, exists := s.pages[r.URL.Path]
	if !exists {
		page = s.pages["/error"]
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
