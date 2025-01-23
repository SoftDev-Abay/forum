package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/validator"
	"net/http"
	"strconv"
)

type CategoryForm struct {
	Name string
}

// delete a category
func (app *Application) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	if id == 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	err = app.Categories.Delete(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// create a category
func (app *Application) categoryCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	data := templateData{}
	app.render(w, r, http.StatusOK, "create_category.html", data)
}

func (app *Application) categoryCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")

	form := CategoryForm{
		Name: name,
	}
	v := validator.Validator{}

	v.CheckField(validator.NotBlank(form.Name), "name", "Name must not be blank")
	v.CheckField(validator.MinChars(form.Name, 3), "name", "Name must be at least 3 characters long")

	if !v.Valid() {
		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "create_category.html", data)
		return
	}

	_, err := app.Categories.Insert(name)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
