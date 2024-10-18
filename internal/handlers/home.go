package handlers

import (
	"net/http"
)

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w, r)
		return
	}

	var data templateData

	app.render(w, r, http.StatusOK, "home.html", data)
}
