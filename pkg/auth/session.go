package auth

import (
	"net/http"
)

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
