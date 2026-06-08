// Package web implements the HTTP API which serves HTML pages and assets.
// Inspired by https://pkg.go.dev/golang.org/x/website/internal/web
package web

import (
	"justchess/internal/db"
	"log"
	"net/http"
	"os"
)

// Declaration of error messages.
const (
	msgNotFound    = "The requested page wasn't found"
	msgRenderError = "The requested page wasn't rendered successfully"
	msgDBError     = "Database cannot be accessed. Please, try again later"
)

// Service serves [page]s and assets from the file system.
type Service struct {
	gameRepo   db.GameRepo
	playerRepo db.PlayerRepo
	// Maps filename with leading slash to parsed [page].
	// Special case: "/home" is shortened to "/" to follow the URL scheme.
	pages map[string]page
}

// InitService parses the [page]s from the specified folder and initialized [Service].
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

func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.static)

	// Serve pages with dynamic content.
	mux.HandleFunc("GET /leaderboard", s.leaderboard)
	mux.HandleFunc("GET /player/{id}", s.profile)
	mux.HandleFunc("GET /engine/{id}", s.engineGame)
	mux.HandleFunc("GET /rated/{id}", s.ratedGame)

	// Serve assets.
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("_web/assets"))))
}

// static serves [page]s with static content.
func (s Service) static(rw http.ResponseWriter, r *http.Request) {
	s.renderPage(rw, r.URL.Path, nil)
}

func (s Service) leaderboard(rw http.ResponseWriter, r *http.Request) {
	leaderboard, err := s.playerRepo.SelectLeaderboard()
	if err != nil {
		s.renderPage(rw, "/error", msgDBError)
		return
	}
	s.renderPage(rw, "/leaderboard", leaderboard)
}

func (s Service) profile(rw http.ResponseWriter, r *http.Request) {
	profile, err := s.playerRepo.SelectProfile(r.PathValue("id"))
	if err != nil {
		s.renderPage(rw, "/error", msgNotFound)
		return
	}
	s.renderPage(rw, "/player", profile)
}

func (s Service) engineGame(rw http.ResponseWriter, r *http.Request) {
	game, err := s.gameRepo.SelectEngine(r.PathValue("id"))
	if err != nil {
		s.renderPage(rw, "/error", msgNotFound)
		return
	}
	s.renderPage(rw, "/engine", game)
}

func (s Service) ratedGame(rw http.ResponseWriter, r *http.Request) {
	game, err := s.gameRepo.SelectEngine(r.PathValue("id"))
	if err != nil {
		s.renderPage(rw, "/error", msgNotFound)
		return
	}
	s.renderPage(rw, "/rated", game)
}

func (s Service) queue(rw http.ResponseWriter, r *http.Request) {
	// Store engine game data to fill up the template.
	// f, err := s.readPage("queue.tmpl")
	// f["timeControl"] = game
	// s.servePage(rw, r, f)
}

// renderPage renders named [page] passing given data to the parsed template.
func (s Service) renderPage(rw http.ResponseWriter, key string, data any) {
	p, exists := s.pages[key]
	if !exists {
		// Render 404 error page.
		p = s.pages["/error"]
		p.Data = msgNotFound
		p.Title = msgNotFound
		if err := p.tmpl.Execute(rw, p); err != nil {
			log.Printf("%s: %s page key: %s", msgRenderError, err.Error(), key)
		}
		return
	}

	// Pass optional data.
	if data != nil {
		p.Data = data
	}
	if err := p.tmpl.Execute(rw, p); err != nil {
		log.Printf("%s: %s page key: %s", msgRenderError, err.Error(), key)
	}
}
