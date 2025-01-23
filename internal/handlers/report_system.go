package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (app *Application) ReportPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	postIDStr := r.PostForm.Get("post_id")
	reasonIDStr := r.PostForm.Get("report_reason_id")
	description := r.PostForm.Get("description")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}
	reasonID, err := strconv.Atoi(reasonIDStr)
	if err != nil || reasonID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	user, err := app.Users.GetById(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if user.Role != "moderator" && user.Role != "admin" {
		app.clientError(w, r, http.StatusForbidden)
		return
	}

	err = app.Reports.CreateReport(userID, postID, reasonID, description, time.Now())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	redirectURL := fmt.Sprintf("/post/view?id=%d&msg=report_submitted", postID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (app *Application) adminReportList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	reports, err := app.Reports.GetAllReports()
	if err != nil {
		app.serverError(w, r, err)
		return
	}


	data := templateData{
		Reports: reports,
	}

	app.render(w, r, http.StatusOK, "admin_reports.html", data)
}

func (app *Application) adminReportDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	reportIDStr := r.URL.Query().Get("id")
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil || reportID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}
	user, err := app.Users.GetById(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if user.Role != "admin" {
		app.clientError(w, r, http.StatusForbidden)
		return
	}

	report, err := app.Reports.Get(reportID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.Posts.DeletePostById(report.PostID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.Reports.DeleteReportByID(reportID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/report/list", http.StatusSeeOther)
}

func (app *Application) adminReportReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	reportIDStr := r.URL.Query().Get("id")
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil || reportID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	user, err := app.Users.GetById(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if user.Role != "admin" {
		app.clientError(w, r, http.StatusForbidden)
		return
	}

	err = app.Reports.DeleteReportByID(reportID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/post/report/list", http.StatusSeeOther)
}
