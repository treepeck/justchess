package web

import (
	"html/template"
	"io/fs"
	"justchess/internal/db"
	"log"
	"net/http"
)

const (
	msgCannotRender = "The requested page cannot be rendered"
)

type Service struct {
	repo      db.Repo
	templates fs.FS
	public    fs.FS
}

func NewService(templates, public fs.FS, r db.Repo) Service {
	return Service{
		repo:      r,
		templates: templates,
		public:    public,
	}
}

func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.servePages)

	// Server static files.
	mux.Handle("/public/", http.StripPrefix("/public/",
		http.FileServerFS(s.public)))
}

func (s Service) servePages(rw http.ResponseWriter, r *http.Request) {
	p := make(Page)

	session, err := r.Cookie("Auth")
	if err == nil {
		p["Player"], _ = s.repo.SelectPlayerBySessionId(session.Value)
	}

	switch r.URL.Path {
	case "/":
		s.renderPage(rw, p, "home.tmpl")

	case "/signin":
		p["Form"] = Form{IsSignUp: false}
		s.renderPage(rw, p, "form.tmpl", "signin.tmpl")

	case "/signup":
		p["Form"] = Form{IsSignUp: true}
		s.renderPage(rw, p, "form.tmpl", "signup.tmpl")

	case "/game":
		s.renderPage(rw, p, "game.tmpl")

		// default:
		// 	s.redirect(rw, r)
	}
}

func (s Service) renderPage(rw http.ResponseWriter, p Page, tmpls ...string) {
	// Load base template.
	t, err := template.ParseFS(s.templates, append([]string{"layout.tmpl"},
		tmpls...)...)

	if err != nil {
		http.Error(rw, msgCannotRender, http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if err := t.Execute(rw, p); err != nil {
		http.Error(rw, msgCannotRender, http.StatusInternalServerError)
		log.Print(err)
		return
	}
}

func (s Service) redirect(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/404", http.StatusFound)
}
