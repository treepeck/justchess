package web

import (
	"encoding/json"
	"html/template"
	"justchess/internal/auth"
	"justchess/internal/db"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

const (
	msgRenderingError = "An error occurred while rendering the page"
	msgGamesNotFound  = "Start playing and games will be displayed here"
	msgBadRequest     = "Malformed request body"
	msgCannotEncode   = "Please, try again later"
	// Path to the base template file.
	baseTmpl = "./_web/templates/base.tmpl"
)

// ParsePages parses the specific [page] for each template in the named folder.
func ParsePages(folder string) (map[string]page, error) {
	// Parse pages with static routes.
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	pages := make(map[string]page, len(files))
	for i := range files {
		tmplFile := files[i].Name()

		// Create an empty template.
		t := template.New("base.tmpl")
		// Add common funcs to empty template.
		t.Funcs(template.FuncMap{
			"round": func(x float64) float64 { return math.Round(x) },
		})

		// Add specific funcs to specific templates.
		if tmplFile == "engine.tmpl" || tmplFile == "rated.tmpl" {
			t.Funcs(template.FuncMap{
				"div": func(x, y int) int { return x / y },
				"mod": func(x, y int) int { return x % y },
				"add": func(x, y int) int { return x + y },
				"sub": func(x, y int) int { return y - x },
			})
		}

		// Parse template.
		var err error
		t, err = t.ParseFiles([]string{baseTmpl, folder + tmplFile}...)
		if err != nil {
			return nil, err
		}
		// The page key in map must be equal to the page URL. Therefore the
		// file extension (.tmpl) is truncated.
		key := "/" + tmplFile[:len(tmplFile)-5]
		// Special case: truncate /home to /.
		if key == "/home" {
			key = "/"
		}
		pages[key] = page{tmpl: t}
	}
	return pages, nil
}

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

func NewService(pr db.PlayerRepo, gr db.GameRepo, static, dynamic map[string]page) Service {
	return Service{
		playerRepo: pr,
		gameRepo:   gr,
		static:     static,
		dynamic:    dynamic,
	}
}

func (s Service) RegisterRoutes(authService auth.Service, mux *http.ServeMux) {
	// Serve pages with static routes.
	mux.HandleFunc("GET /", authService.Authorize(s.staticRoutePage))
	// Serve pages with dynamic routes.
	mux.HandleFunc("GET /queue/{id}", authService.Authorize(s.queuePage))
	mux.HandleFunc("GET /rated/{id}", authService.Authorize(s.ratedGamePage))
	mux.HandleFunc("GET /engine/{id}", authService.Authorize(s.engineGamePage))
	mux.HandleFunc("GET /player/{id}", authService.Authorize(s.playerPage))
	// API endpoints.
	mux.HandleFunc("GET /api/rated", s.ratedGamesBrief)
	mux.HandleFunc("GET /api/engine", s.engineGamesBrief)

	bundlesHandler := http.FileServer(http.Dir("./_web/bundles"))
	mux.Handle("GET /bundles/", http.StripPrefix("/bundles/", bundlesHandler))

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
	case "/":
		page.Data = [9]string{"1+0", "2+1", "3+0", "3+2", "5+0", "5+2", "10+0", "10+10", "15+10"}

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
