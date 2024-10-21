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
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/post/view", app.postView)
	mux.HandleFunc("/post/create", app.postCreate)
	mux.HandleFunc("/register", app.register)
	mux.HandleFunc("/register/post", app.registerPost)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/login/post", app.loginPost)


	return mux
}
