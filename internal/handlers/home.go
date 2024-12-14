package handlers

import (
	"net/http"
)

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w, r)
		return
	}

	posts, err := app.Posts.Latest()
	if err != nil {
		app.serverError(w, r, err)
	}

	var data templateData

	data.Posts = posts
	
	app.render(w, r, http.StatusOK, "home.html", data)
}
