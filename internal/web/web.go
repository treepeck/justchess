package web

import (
	"encoding/json"
	"html/template"
	"justchess/internal/auth"
	"justchess/internal/db"
	"log"
	"math"
	"net/http"
	"time"
)

const (
	msgRenderingError = "An error occurred while rendering the page"
	msgGamesNotFound  = "Start playing and games will be displayed here"
	msgBadRequest     = "Malformed request body"
	msgCannotEncode   = "Please, try again later"
)

type Service struct {
	playerRepo db.PlayerRepo
	gameRepo   db.GameRepo
	// Maps URL to the parsed template of the page.
	static map[string]page
	// Maps static part of URL to the parsed template of the page.
	// Templates are stored separately so that user cannot visit page with
	// dynamic route via static part of the route.
	dynamic map[string]page
}

func NewService(pr db.PlayerRepo, gr db.GameRepo) Service {
	return Service{
		playerRepo: pr,
		gameRepo:   gr,
		static:     make(map[string]page),
		dynamic:    make(map[string]page),
	}
}

// templateFile defines the template to be parsed and stored in a map.
type templateFile struct {
	key      string
	title    string
	files    []string
	isStatic bool
}

func (s Service) ParsePages(folder string) error {
	// Define paths to template files.
	base := folder + "base.tmpl"
	home := folder + "home.tmpl"
	signin := folder + "signin.tmpl"
	signup := folder + "signup.tmpl"
	leaderboard := folder + "leaderboard.tmpl"
	reset := folder + "reset_password.tmpl"
	about := folder + "about.tmpl"
	notFound := folder + "404.tmpl"
	queue := folder + "queue.tmpl"
	ratedGame := folder + "rated_game.tmpl"
	engineGame := folder + "engine_game.tmpl"
	player := folder + "player.tmpl"

	// Templates of each page.
	templates := []templateFile{
		{key: "/", title: "Home", files: []string{home}, isStatic: true},
		{key: "/signup", title: "Sign up", files: []string{signup}, isStatic: true},
		{key: "/signin", title: "Sign in", files: []string{signin}, isStatic: true},
		{key: "/leaderboard", title: "Leaderboard", files: []string{leaderboard}, isStatic: true},
		{key: "/reset", title: "Reset", files: []string{reset}, isStatic: true},
		{key: "/about", title: "About", files: []string{about}, isStatic: true},
		{key: "/404", title: "Not found", files: []string{notFound}, isStatic: true},
		{key: "/queue", title: "Queue", files: []string{queue}, isStatic: false},
		{key: "/rated", files: []string{ratedGame}, isStatic: false},
		{key: "/engine", files: []string{engineGame}, isStatic: false},
		{key: "/player", files: []string{player}, isStatic: false},
	}
	for _, t := range templates {
		tmpl := template.New("base.tmpl")

		tmpl.Funcs(template.FuncMap{
			"round": func(x float64) float64 { return math.Round(x) },
		})

		if t.key == "/engine" || t.key == "/rated" {
			tmpl.Funcs(template.FuncMap{
				"div": func(x, y int) int { return x / y },
				"mod": func(x, y int) int { return x % y },
				"add": func(x, y int) int { return x + y },
				"sub": func(x, y int) int { return y - x },
			})
		}

		// Prepend base template before each page.
		t.files = append([]string{base}, t.files...)
		// Parse template.
		var err error
		tmpl, err = tmpl.ParseFiles(t.files...)
		if err != nil {
			return err
		}

		if t.isStatic {
			p := page{tmpl: tmpl, Title: t.title}

			switch t.key {
			case "/":
				p.Data = [9]string{"1+0", "2+1", "3+0", "3+2", "5+0", "5+2", "10+0", "10+10", "15+10"}
			}

			s.static[t.key] = p
		} else {
			p := page{tmpl: tmpl, Title: t.title}
			s.dynamic[t.key] = p
		}
	}
	return nil
}

func (s Service) RegisterRoutes(authService auth.Service, mux *http.ServeMux) {
	// Serve pages with static routes.
	mux.HandleFunc("GET /", authService.Authorize(s.staticRoutePage))
	// Serve pages with path values.
	mux.HandleFunc("GET /queue/{id}", authService.Authorize(s.queuePage))
	mux.HandleFunc("GET /rated/{id}", authService.Authorize(s.ratedGamePage))
	mux.HandleFunc("GET /engine/{id}", authService.Authorize(s.engineGamePage))
	mux.HandleFunc("GET /player/{id}", authService.Authorize(s.playerPage))
	// API endpoints.
	mux.HandleFunc("GET /api/rated", s.ratedGamesBrief)
	mux.HandleFunc("GET /api/engine", s.engineGamesBrief)

	// Serve static bundles.
	bundlesHandler := http.FileServer(http.Dir("./_web/bundles"))
	mux.Handle("GET /bundles/", http.StripPrefix("/bundles/", bundlesHandler))

	// Serve fonts.
	fontsHandler := http.FileServer(http.Dir("./_web/fonts"))
	mux.Handle("GET /fonts/", http.StripPrefix("/fonts/", fontsHandler))

	imagesHandler := http.FileServer(http.Dir("./_web/images"))
	mux.Handle("GET /images/", http.StripPrefix("/images/", imagesHandler))

	soundsHandler := http.FileServer(http.Dir("./_web/sounds"))
	mux.Handle("GET /sounds/", http.StripPrefix("/sounds/", soundsHandler))

	stockfishHandler := http.FileServer(http.Dir("./_web/stockfish"))
	mux.Handle("GET /stockfish/", http.StripPrefix("/stockfish/", stockfishHandler))
}

func (s Service) staticRoutePage(rw http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(auth.PlayerKey).(db.Player)
	if !ok {
		http.Redirect(rw, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	page, exists := s.static[r.URL.Path]
	if !exists {
		http.Redirect(rw, r, "/404", http.StatusFound)
		return
	}

	switch r.URL.Path {
	case "/leaderboard":
		var err error
		page.Data, err = s.playerRepo.SelectLeaderboard()
		if err != nil {
			log.Print(err)
			http.Error(rw, msgRenderingError, http.StatusInternalServerError)
			return
		}
	}

	page.Player = p
	if err := page.tmpl.Execute(rw, page); err != nil {
		log.Print(err)
		http.Error(rw, msgRenderingError, http.StatusInternalServerError)
	}
}

func (s Service) queuePage(rw http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(auth.PlayerKey).(db.Player)
	if !ok {
		http.Redirect(rw, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	var d QueueData
	switch r.PathValue("id") {
	case "0":
		d = QueueData{Control: 1, Bonus: 0}
	case "1":
		d = QueueData{Control: 2, Bonus: 1}
	case "2":
		d = QueueData{Control: 3, Bonus: 0}
	case "3":
		d = QueueData{Control: 3, Bonus: 2}
	case "4":
		d = QueueData{Control: 5, Bonus: 0}
	case "5":
		d = QueueData{Control: 5, Bonus: 2}
	case "6":
		d = QueueData{Control: 10, Bonus: 0}
	case "7":
		d = QueueData{Control: 10, Bonus: 10}
	case "8":
		d = QueueData{Control: 15, Bonus: 10}
	default:
		http.Redirect(rw, r, "/404", http.StatusFound)
		return
	}

	page := s.dynamic["/queue"]
	page.Player = p
	page.Data = d
	if err := page.tmpl.Execute(rw, page); err != nil {
		log.Print(err)
		http.Error(rw, msgRenderingError, http.StatusInternalServerError)
	}
}

func (s Service) ratedGamePage(rw http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(auth.PlayerKey).(db.Player)
	if !ok {
		http.Redirect(rw, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	g, err := s.gameRepo.SelectRated(r.PathValue("id"))
	if err != nil {
		http.Redirect(rw, r, "/404", http.StatusFound)
		return
	}

	page := s.dynamic["/rated"]
	page.Title = g.White.Name + " vs " + g.Black.Name
	page.Player = p
	page.Data = g
	if err = page.tmpl.Execute(rw, page); err != nil {
		log.Print(err)
		http.Error(rw, msgRenderingError, http.StatusInternalServerError)
	}
}

func (s Service) engineGamePage(rw http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(auth.PlayerKey).(db.Player)
	if !ok {
		http.Redirect(rw, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	g, err := s.gameRepo.SelectEngine(r.PathValue("id"))
	if err != nil {
		http.Redirect(rw, r, "/404", http.StatusFound)
		return
	}

	page := s.dynamic["/engine"]
	page.Title = p.Name + " vs Engine"
	page.Player = p
	page.Data = g
	if err = page.tmpl.Execute(rw, page); err != nil {
		log.Print(err)
		http.Error(rw, msgRenderingError, http.StatusInternalServerError)
	}
}

func (s Service) playerPage(rw http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(auth.PlayerKey).(db.Player)
	if !ok {
		http.Redirect(rw, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	profile, err := s.playerRepo.SelectProfileData(r.PathValue("id"))
	if err != nil {
		http.Redirect(rw, r, "/404", http.StatusFound)
		return
	}

	page := s.dynamic["/player"]
	page.Title = profile.Name
	page.Player = p
	page.Data = profile
	if err = page.tmpl.Execute(rw, page); err != nil {
		log.Print(err)
		http.Error(rw, msgRenderingError, http.StatusInternalServerError)
	}
}

func (s Service) ratedGamesBrief(rw http.ResponseWriter, r *http.Request) {
	playerId := r.URL.Query().Get("pid")
	if len(playerId) != 12 {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}
	// Optional parameters.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.RatedGameBrief
	if err == nil && len(cursorId) == 12 {
		// If optional pagination parameters are defined.
		games, err = s.gameRepo.SelectOlderRated(playerId, db.Pagination{
			CursorId:        cursorId,
			CursorCreatedAt: cursorCreatedAt,
		})
	} else {
		games, err = s.gameRepo.SelectNewestRated(playerId)
	}

	if err != nil {
		http.Error(rw, msgGamesNotFound, http.StatusNotFound)
		return
	}

	if err = json.NewEncoder(rw).Encode(games); err != nil {
		log.Print(err)
		http.Error(rw, msgCannotEncode, http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
}

func (s Service) engineGamesBrief(rw http.ResponseWriter, r *http.Request) {
	playerId := r.URL.Query().Get("pid")
	if len(playerId) != 12 {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}
	// Optional parameters.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.EngineGameBrief
	if err == nil && len(cursorId) == 12 {
		// If optional pagination parameters are defined.
		games, err = s.gameRepo.SelectOlderEngine(playerId, db.Pagination{
			CursorId:        cursorId,
			CursorCreatedAt: cursorCreatedAt,
		})
	} else {
		games, err = s.gameRepo.SelectNewestEngine(playerId)
	}

	if err != nil {
		http.Error(rw, msgGamesNotFound, http.StatusNotFound)
		return
	}

	if err = json.NewEncoder(rw).Encode(games); err != nil {
		log.Print(err)
		http.Error(rw, msgCannotEncode, http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
}
