package handlers

import (
	"net/http"
	"strconv"
)

func (app *Application) changeUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	role := r.FormValue("role")
	if role != "user" && role != "admin" && role != "moderator" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	err = app.Users.UpdateRole(id, role)
	if err != nil {
		http.Error(w, "Unable to update user role", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
