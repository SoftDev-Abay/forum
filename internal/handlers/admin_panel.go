package handlers

import (
	"net/http"
	"strconv"
)

func (app *Application) usersPanel(w http.ResponseWriter, r *http.Request) {
	users, err := app.Users.GetAll()
	if err != nil {
		http.Error(w, "Unable to retrieve users", http.StatusInternalServerError)
		return
	}

	data := templateData{
		Users: users,
	}

	// app.render(w, r, http.StatusOK, "home.html", data)

	app.render(w, r, http.StatusOK, "users_control_panel.html", data)
}

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

	http.Redirect(w, r, "/users_control_panel", http.StatusSeeOther)
}
