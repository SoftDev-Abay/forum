package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"game-forum-abaliyev-ashirbay/internal/models"
	"game-forum-abaliyev-ashirbay/internal/validator"
)

type PromotionRequestForm struct {
	Description string
}

func (app *Application) sendPromotionRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	description := r.FormValue("description")

	form := PromotionRequestForm{
		Description: description,
	}
	v := validator.Validator{}

	v.CheckField(validator.NotBlank(form.Description), "description", "Description must not be blank")
	v.CheckField(validator.MinChars(form.Description, 10), "description", "Description must be at least 10 characters long")

	if !v.Valid() {
		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "create_promotion_request.html", data)
		return
	}

	_, err = app.PromotionRequests.Insert(userID, description, false)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/promotion_requests", http.StatusSeeOther)
}

func (app *Application) changePromotionRequestStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	statusStr := r.FormValue("status")
	status, err := strconv.ParseBool(statusStr)
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	err = app.PromotionRequests.UpdateStatus(id, status)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/promotion_requests", http.StatusSeeOther)
}

func (app *Application) getAllPromotionRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	requests, err := app.PromotionRequests.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		PromotionRequests: requests,
	}

	app.render(w, r, http.StatusOK, "promotion_requests.html", data)
}

func (app *Application) getPromotionRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	request, err := app.PromotionRequests.GetByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := templateData{
		PromotionRequest: request,
	}

	app.render(w, r, http.StatusOK, "promotion_request.html", data)
}
