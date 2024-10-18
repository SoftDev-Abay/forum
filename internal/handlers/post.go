package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

func (app Application) postView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w, r)
		return
	}
	fmt.Fprintf(w, "Display a specific post with ID %d...", id)
}

func (app *Application) postCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed) // Use the clientError() helper.
		return
	}
	w.Write([]byte("Create a new post..."))
}
