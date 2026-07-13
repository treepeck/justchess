// Package web implements the HTTP API which serves HTML pages and assets.
// Inspired by https://pkg.go.dev/golang.org/x/website/internal/web
package web

import (
	"justchess/internal/auth"
	"justchess/internal/db"
	"justchess/internal/response"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Service serves [page]s and assets from the file system.
type Service struct {
	gameRepo   db.GameRepo
	playerRepo db.PlayerRepo
	// Maps filename with leading slash to parsed [page].
	// Special case: "/home" is shortened to "/" to follow the URL scheme.
	pages map[string]page
}

// InitService parses the [page]s from the specified folder and initializes [Service].
func InitService(gr db.GameRepo, pr db.PlayerRepo, folder string) (Service, error) {
	tmpls, err := os.ReadDir(folder)
	if err != nil {
		return Service{}, err
	}

	pages := make(map[string]page, len(tmpls))
	for _, t := range tmpls {
		// Skip nested directories.
		if t.IsDir() {
			continue
		}

		path := folder + t.Name()
		// Add leading slash and exclude the file extension to follow the URL scheme.
		key := "/" + t.Name()[:len(t.Name())-5]
		// Special case: "/home" URL is shortened to "/"
		if key == "/home" {
			key = "/"
		}

		file, err := os.ReadFile(path)
		if err != nil {
			return Service{}, err
		}

		p, err := parsePage(path, file)
		if err != nil {
			return Service{}, err
		}
		pages[key] = p
	}

	return Service{
		gameRepo:   gr,
		playerRepo: pr,
		pages:      pages,
	}, nil
}

func (s Service) RegisterRoutes(authService auth.Service, mux *http.ServeMux) {
	mux.HandleFunc("GET /", authService.Authorize(s.static))

	// Serve pages with dynamic content.
	mux.HandleFunc("GET /leaderboard", authService.Authorize(s.leaderboard))
	mux.HandleFunc("GET /player/{id}", authService.Authorize(s.profile))
	/*
		mux.HandleFunc("GET /engine/{id}", authService.Authorize(s.engineGame))
		mux.HandleFunc("GET /rated/{id}", authService.Authorize(s.ratedGame))
	*/

	mux.HandleFunc("GET /queue/{id}", authService.Authorize(s.queue))

	// Serve assets.
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("_web/assets"))))
}

// static serves [page]s with static content.
func (s Service) static(rw http.ResponseWriter, r *http.Request) {
	p, exists := s.pages[r.URL.Path]
	if !exists {
		// Render 404 error page if page does not exists.
		p = s.pages["/error"]
		p.Data = response.NotFound
		p.Title = response.NotFound
		if err := p.tmpl.Execute(rw, p); err != nil {
			log.Printf("%s: %s page key: %s", response.InternalError, err.Error(), r.URL.Path)
		}
		return
	}
	s.renderPage(rw, r, p, nil, r.URL.Path)
}

func (s Service) leaderboard(rw http.ResponseWriter, r *http.Request) {
	leaderboard, err := s.playerRepo.SelectLeaderboard()
	if err != nil {
		s.renderPage(rw, r, s.pages["/error"], response.DatabaseError, "/error")
		return
	}
	s.renderPage(rw, r, s.pages["/leaderboard"], leaderboard, "/leaderboard")
}

func (s Service) profile(rw http.ResponseWriter, r *http.Request) {
	profile, err := s.playerRepo.SelectProfile(r.PathValue("id"))
	if err != nil {
		s.renderPage(rw, r, s.pages["/error"], response.NotFound, "/error")
		return
	}
	p := s.pages["/player"]
	p.Title = profile.Name
	s.renderPage(rw, r, p, profile, "/player")
}

/*
func (s Service) engineGame(rw http.ResponseWriter, r *http.Request) {
	game, err := s.gameRepo.SelectEngine(r.PathValue("id"))
	if err != nil {
		s.renderPage(rw, r, s.pages["/error"], response.NotFound, "/error")
		return
	}
	p := s.pages["/engine"]
	p.Title = game.Player.Name + " vs Engine"
	s.renderPage(rw, r, p, game, "/engine")
}

func (s Service) ratedGame(rw http.ResponseWriter, r *http.Request) {
	game, err := s.gameRepo.SelectRated(r.PathValue("id"))
	if err != nil {
		s.renderPage(rw, r, s.pages["/error"], response.NotFound, "/error")
		return
	}
	p := s.pages["/rated"]
	p.Title = game.White.Name + " vs " + game.Black.Name
	s.renderPage(rw, r, p, game, "/rated")
}
*/

// Used to fill up the queue template.
var controls = []struct {
	Control int // In minutes.
	Bonus   int // In seconds.
}{
	{1, 0}, {2, 1}, {3, 0},
	{3, 2}, {5, 0}, {5, 2},
	{10, 0}, {10, 10}, {15, 10},
}

func (s Service) queue(rw http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 || id > 8 {
		s.renderPage(rw, r, s.pages["/error"], response.NotFound, "/error")
		return
	}

	session, ok := r.Context().Value(auth.SessionKey).(auth.Session)
	if !ok {
		http.Redirect(rw, r, "/signup", http.StatusTemporaryRedirect)
		return
	}
	log.Println(session.Id)
	p := s.pages["/queue"]

	// Construct the page title.
	var title strings.Builder
	title.WriteString("Queue ")
	title.WriteString(strconv.Itoa(controls[id].Control))
	title.WriteString("m + ")
	title.WriteString(strconv.Itoa(controls[id].Bonus))
	title.WriteByte('s')
	p.Title = title.String()

	s.renderPage(rw, r, p, controls[id], "/queue")
}

// renderPage renders named [page] passing given data to the parsed template.
func (s Service) renderPage(rw http.ResponseWriter, r *http.Request, p page,
	data any, key string) {
	session, ok := r.Context().Value(auth.SessionKey).(auth.Session)
	if !ok {
		log.Print("web: request context does not contain session")
		http.Error(rw, response.InternalError, http.StatusInternalServerError)
		return
	}
	player, err := s.playerRepo.SelectById(session.Id)
	if err != nil {
		player.Name = "Guest"
	}
	p.Player = player

	// Pass optional data.
	if data != nil {
		p.Data = data
	}
	if err := p.tmpl.Execute(rw, p); err != nil {
		log.Printf("web: %s on page %s", err.Error(), key)
	}
}
