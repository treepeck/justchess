package web

import (
	"html/template"
	"log"
	"math"
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

// InitService parses the template files and stores them in the [Service].
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
		{"/leaderboard", []string{"leaderboard.tmpl"}, baseData{Title: "Leaderboard"}},
		{"/reset-password", []string{"reset_password.tmpl"}, baseData{Title: "Reset Password"}},
		{"/error", []string{"error.tmpl"}, baseData{Title: "Error"}},
	}

	// Parse and store templates to avoid reparsing them on each HTTP request.
	pages := make(map[string]page, len(pagesData))
	for _, data := range pagesData {
		t := template.New("base.tmpl")

		// Add default functions.
		t.Funcs(template.FuncMap{
			"eq":    func(s1, s2 string) bool { return s1 == s2 },
			"eqC":   func(c1, c2 chego.Color) bool { return c1 == c2 },
			"round": func(n float64) float64 { return math.Round(n) },
		})

		if data.url == "/leaderboard" {
			t.Funcs(template.FuncMap{
				"formatDate": func(date time.Time) string {
					return date.Format("Jan 02, 2006")
				}})
		}

		paths := append([]string{"/base.tmpl"}, data.tmpls...)
		// Append template prefix to each path.
		for i, path := range paths {
			paths[i] = tmplPrefix + path
		}

		t, err := t.ParseFiles(paths...)
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
	mux.Handle("GET /bundles/", http.StripPrefix("/bundles/", http.FileServer(bundles)))

	// Serve files from the _web/fonts folder.
	fonts := http.Dir("./_web/fonts")
	mux.Handle("GET /fonts/", http.StripPrefix("/fonts/", http.FileServer(fonts)))

	// Serve files from the _web/images folder.
	images := http.Dir("./_web/images")
	mux.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(images)))

	// Serve pages with static routes.
	mux.Handle("GET /", http.HandlerFunc(s.serveStaticRoutePage))

	// Serve pages with dynamic routes.
	mux.HandleFunc("GET /queue/{id}", s.serveQueue)
	mux.HandleFunc("GET /game/{kind}/{id}", s.serveGame)
	mux.HandleFunc("GET /player/{id}", s.servePlayer)
}

// Time controls. First number is in minutes. Second number is in seconds.
var controls = [9]QueueData{
	{1, 0}, {2, 1}, {3, 0}, {3, 2}, {5, 0}, {5, 2}, {10, 0}, {10, 10}, {15, 10},
}

func (s Service) serveQueue(rw http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 || id > 8 {
		s.renderPage(rw, r, s.pages["/error"])
		return
	}

	p := s.pages["/queue"]
	p.Data = controls[id]
	s.renderPage(rw, r, p)
}

func (s Service) serveGame(rw http.ResponseWriter, r *http.Request) {
	switch r.PathValue("kind") {
	case "rated":
		g, err := s.gameRepo.SelectRated(r.PathValue("id"))
		if err != nil {
			s.renderPage(rw, r, s.pages["/error"])
			return
		}
		page := s.pages["/active"]
		// If the game was been already terminated, serve the corresponding page.
		if g.Termination != chego.Unterminated {
			page = s.pages["/archive"]
		}
		// Fill up the template with game
		page.Data = GameData{
			Game:       g,
			IsVsEngine: false,
		}
		page.Base.Title = g.White.Name + " vs " + g.Black.Name
		s.renderPage(rw, r, page)
	case "engine":
		g, err := s.gameRepo.SelectEngine(r.PathValue("id"))
		if err != nil {
			s.renderPage(rw, r, s.pages["/error"])
			return
		}
		page := s.pages["/active"]
		// If the game was been already terminated, serve the corresponding page.
		if g.Termination != chego.Unterminated {
			page = s.pages["/archive"]
		}
		// Fill up the template with game
		page.Data = GameData{
			Game:       g,
			IsVsEngine: true,
		}
		page.Base.Title = g.Player.Name + " vs Engine"
		s.renderPage(rw, r, page)

	default:
		s.renderPage(rw, r, s.pages["/error"])
	}
}

func (s Service) servePlayer(rw http.ResponseWriter, r *http.Request) {
	p, err := s.playerRepo.SelectProfileData(r.PathValue("id"))
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
		return
	}

	if r.URL.Path == "/leaderboard" {
		leaders, err := s.playerRepo.SelectLeaderboard()
		if err != nil {
			log.Print(err)
			page = s.pages["/error"]
			return
		}
		page.Data = leaders
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
