package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (app *Application) ReportPost(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Parse form inputs
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// Extract form values
	postIDStr := r.PostForm.Get("post_id")
	reasonIDStr := r.PostForm.Get("report_reason_id")
	description := r.PostForm.Get("description")

	// Convert postID and reasonID to int
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

	// Get current user (moderator/admin)
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
	// Ensure role is moderator or admin
	if user.Role != "moderator" && user.Role != "admin" {
		app.clientError(w, r, http.StatusForbidden)
		return
	}

	// Insert new report into the DB
	err = app.Reports.CreateReport(userID, postID, reasonID, description, time.Now())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect user back to the post page with a success message
	redirectURL := fmt.Sprintf("/post/view?id=%d&msg=report_submitted", postID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (app *Application) adminReportList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Fetch all reports (or only "open" ones, depending on your logic)
	reports, err := app.Reports.GetAllReports()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Optionally, fetch additional details about moderators, reasons, etc.
	// For now, assume your `GetAllReports()` returns a slice of something like:
	//    type Report struct {
	//       ID            int
	//       ModeratorID   int
	//       PostID        int
	//       ReportReasonID int
	//       Description   string
	//       DateCreated   time.Time
	//       AdminID       *int
	//       AdminResponse *string
	//       ...
	//    }

	// Pass them to the template
	data := templateData{
		Reports: reports, // Make sure you have Reports []*models.Reports in templateData
	}

	// Render your admin reports page
	app.render(w, r, http.StatusOK, "admin_reports.html", data)
}
