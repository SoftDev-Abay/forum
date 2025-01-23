package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type contextKey string

const userContextKey contextKey = "userContextKey"

var defaultRoles = []string{"user", "moderator", "admin"}

var (
	// requests stores for each IP a slice of timestamps (recent request times).
	requests = make(map[string][]time.Time)
	mu       sync.Mutex

	// MaxRequests defines how many requests an IP can make in TimeWindow.
	MaxRequests = 100

	// TimeWindow is how far back we count requests for rate-limiting.
	TimeWindow = time.Minute
)

func (app *Application) loginMiddware(next http.Handler, roles ...string) http.Handler {
	if len(roles) == 0 {
		roles = defaultRoles
	}

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

		if !contains(roles, user.Role) {
			app.clientError(w, r, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) secureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Security-Policy",
    	// main directives
		"default-src 'self'; " +
		// allow inline styles if Font Awesome injects CSS
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://use.fontawesome.com; " +
		// allow fonts from google & fontawesome
		"font-src https://fonts.gstatic.com https://use.fontawesome.com data:; " +
		// allow scripts from your domain & fontawesome.com
		"script-src 'self' 'unsafe-inline' https://use.fontawesome.com https://ka-f.fontawesome.com; " +
		// allow images only from your site (adjust if you need external images)
		"img-src 'self' http://i9.photobucket.com/albums/a88/creaticode/avatar_2_zps7de12f8b.jpg; " +
		// no objects
		"object-src 'none'; " +
		// limit connections to same-origin
		"connect-src 'self';",
)
        w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "deny")
        w.Header().Set("X-XSS-Protection", "0")
        
        next.ServeHTTP(w, r)
    })
}

func (app *Application) rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := clientIP(r)
        now := time.Now()

        mu.Lock()
        defer mu.Unlock() 

        timestamps := requests[ip]
        filtered := make([]time.Time, 0, len(timestamps))
        for _, t := range timestamps {
            if now.Sub(t) < TimeWindow {
                filtered = append(filtered, t)
            }
        }

        if len(filtered) >= MaxRequests {
			app.clientError(w, r, http.StatusTooManyRequests)
            return
        }

        filtered = append(filtered, now)
        requests[ip] = filtered

        next.ServeHTTP(w, r)
    })
}

func clientIP(r *http.Request) string {
    return r.RemoteAddr 
}