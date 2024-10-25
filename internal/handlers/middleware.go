package handlers

import (
	"context"
	"net/http"
)

type contextKey string

const userContextKey contextKey = "userContextKey"

func (app *Application) loginMiddware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			app.Logger.Error(err.Error())

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := app.Users.GetByToken(tokenCookie.Value)
		if err != nil {
			app.Logger.Error(err.Error())

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

// func (app *Application) logRequest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
// 		next.ServeHTTP(w, r)
// 	})
// }

// func (app *Application) recoverPanic(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Create a deferred function (which will always be run in the event
// 		// of a panic as Go unwinds the stack).
// 		defer func() {
// 			// Use the builtin recover function to check if there has been a
// 			// panic or not. If there has...
// 			if err := recover(); err != nil {
// 				// Set a "Connection: close" header on the response.
// 				w.Header().Set("Connection", "close")
// 				// Call the app.serverError helper method to return a 500
// 				// Internal Server response.
// 				app.serverError(w, fmt.Errorf("%s", err))
// 			}
// 		}()
// 		next.ServeHTTP(w, r)
// 	})
// }
