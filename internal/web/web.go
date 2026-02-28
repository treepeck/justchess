package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"justchess/internal/db"

	"github.com/treepeck/chego"
)

// Declaration of error messages.
const (
	msgNotFound        string = "The requested page doesn't exist"
	msgCannotEncode    string = "Cannot encode the response"
	msgBadRequest      string = "The request body is malformed"
	msgInvalidTemplate string = "The requested page cannot be rendered"
)

// Template path prefix.
const tmplPrefix = "./_web/templates/"

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
		{"/", []string{"home.tmpl"}, baseData{Title: "Home"}},
		{"/queue", []string{"queue.tmpl"}, baseData{Title: "Queue"}},
		{"/signup", []string{"signup.tmpl"}, baseData{Title: "Sign up"}},
		{"/signin", []string{"signin.tmpl"}, baseData{Title: "Sign in"}},
		{"/active", []string{"active_game.tmpl", "board.tmpl"}, baseData{}},
		{"/archive", []string{"archive_game.tmpl", "board.tmpl"}, baseData{}},
		{"/player", []string{"player.tmpl"}, baseData{}},
		{"/error", []string{"error.tmpl"}, baseData{Title: "Error"}},
	}

	// Parse and store templates to avoid reparsing them on each HTTP request.
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
	mux.HandleFunc("/api/profile-games", s.getProfileGames)

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
	mux.HandleFunc("/queue/{id}", s.serveQueue)
	mux.HandleFunc("/game/{id}", s.serveGame)
	mux.HandleFunc("/player/{name}", s.servePlayer)
}

// Possible time controls.
var Controls = [9]QueueData{
	{1, 0}, {2, 1}, {3, 0}, {3, 2}, {5, 0}, {5, 2}, {10, 0}, {10, 10}, {15, 10},
}

func (s Service) serveQueue(rw http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 || id > 8 {
		s.renderPage(rw, r, s.pages["/error"])
	}

	p := s.pages["/queue"]
	p.Data = Controls[id-1]
	s.renderPage(rw, r, p)
}

func (s Service) serveGame(rw http.ResponseWriter, r *http.Request) {
	g, err := s.gameRepo.SelectById(r.PathValue("id"))
	if err != nil {
		s.renderPage(rw, r, s.pages["/error"])
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

func (s Service) servePlayer(rw http.ResponseWriter, r *http.Request) {
	p, err := s.playerRepo.SelectProfileData(r.PathValue("name"))
	if err != nil {
		s.renderPage(rw, r, s.pages["/error"])
		return
	}

	// Fill up the template with player data.
	page := s.pages["/player"]
	page.Data = p
	page.Base.Title = p.Name

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

// getProfileGames fetches up to 100 games from database, encodes them and
// sends the response.
func (s Service) getProfileGames(rw http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cat"))

	var games []db.ProfileGame
	if cursorId != "" && err == nil {
		// Apply pagination if cursors are defined.
		games, err = s.gameRepo.SelectOlderProfileGames(
			name,
			cursorId,
			cursorCreatedAt,
		)
	} else {
		games, err = s.gameRepo.SelectNewestProfileGames(name)
	}

	if err != nil || len(games) == 0 {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}

	if err = json.NewEncoder(rw).Encode(games); err != nil {
		log.Print(err)
		http.Error(rw, msgCannotEncode, http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
}
