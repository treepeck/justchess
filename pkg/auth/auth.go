package auth

import (
	"justchess/pkg/db"
	"net/http"
	"time"
)

func Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/signup", createSession)
	mux.HandleFunc("GET /auth/signin", signBySession)
	return mux
}

func createSession(rw http.ResponseWriter, r *http.Request) {
	var sid, uid string

	row := db.InsertSession()
	if err := row.Scan(&sid, &uid); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	setAutorizationCookie(rw, sid)
}

func signBySession(rw http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	var uid string
	var expiresAt time.Time

	row := db.SelectSessionById(c.Value)
	if err = row.Scan(&uid, &expiresAt); err != nil {
		http.Error(rw, "Session doesn't exist", http.StatusUnauthorized)
		return
	} else if expiresAt.Compare(time.Now()) != 1 {
		db.DeleteSession(c.Value)

		http.Error(rw, "Session expired", http.StatusUnauthorized)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	rw.Write([]byte(uid))
}

func setAutorizationCookie(rw http.ResponseWriter, sid string) {
	c := http.Cookie{
		Name:     "Authorization",
		Value:    sid,
		Path:     "/",
		MaxAge:   86400, // Session will last 1 day.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &c)
}
