package handlers

import (
	"game-forum-abaliyev-ashirbay/ui"
	"io/fs"
	"net/http"
	"os"
)

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	staticFiles, err := fs.Sub(ui.Files, "static")
	if err != nil {
		app.Logger.Error(err.Error())
		os.Exit(1)
	}

	fileServer := http.FileServer(http.FS(staticFiles))

	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Serve /imgs/<filename> from ./data/imgs
	mux.Handle("/imgs/", http.StripPrefix("/imgs/", http.FileServer(http.Dir("./data/imgs"))))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/post/view", app.postView)
	mux.HandleFunc("/post/reaction", app.handlePostReaction)

	mux.Handle("/post/create/post", app.loginMiddware(http.HandlerFunc(app.postCreatePost)))
	mux.Handle("/post/create", app.loginMiddware(http.HandlerFunc(app.postCreate)))
	mux.Handle("/comments/create", app.loginMiddware(http.HandlerFunc(app.createComment)))
	mux.HandleFunc("/comments/reaction", app.handleCommentReaction)
	mux.Handle("/personal-page", app.loginMiddware(http.HandlerFunc(app.personalPage)))

	mux.HandleFunc("/register", app.register)
	mux.HandleFunc("/register/post", app.registerPost)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/login/post", app.loginPost)
	mux.HandleFunc("/logout", app.logout)

	return mux
}
