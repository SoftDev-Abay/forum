package handlers

import (
	"net/http"
)

func (app *Application) adminPanel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Fetch all users
	users, err := app.Users.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Fetch all promotion requests
	promotionRequests, err := app.PromotionRequests.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Fetch all reports
	reports, err := app.Reports.GetAllReports()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// fetch categories

	categories, err := app.Categories.GetAll()

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		Users:             users,
		PromotionRequests: promotionRequests,
		Reports:           reports,
		Categories:        categories,
	}

	app.render(w, r, http.StatusOK, "admin_panel.html", data)
}
