package handlers

import (
	"net/http"
)

// notificationsPage shows all notifications for the current user.
func (app *Application) notificationsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	// 1) Get the currently logged-in user
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	// 2) Fetch all notifications for this user
	notifs, err := app.Notifications.GetAllByRecipient(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var notificationViews []NotificationView

	for _, n := range notifs {
		// Fetch the actor's username
		actorUser, err := app.Users.GetById(n.Actor_ID)
		var actorName string
		if err == nil && actorUser != nil {
			actorName = actorUser.Username
		} else {
			actorName = "UnknownUser"
		}

		// If comment_id is set, fetch the comment to get its text
		var commentText string
		if n.Comment_ID.Valid {
			commentID := int(n.Comment_ID.Int64)
			c, err := app.Comments.Get(commentID)
			if err == nil && c != nil {
				commentText = c.Text
			}
		}

		notificationViews = append(notificationViews, NotificationView{
			ID:            n.ID,
			Type:          n.Type,
			ActorUsername: actorName,
			PostID:        n.Post_ID,
			CommentText:   commentText,
			CreatedAt:     n.Created_at.Format("2006-01-02 15:04"),
			IsRead:        n.Is_read,
		})
	}

	// 4) Mark them all as read in one go (or individually).
	err = app.Notifications.MarkAllAsReadByUser(userID)
	if err != nil {
		// If this fails, we won't block displaying them
	}

	// 5) Prepare template data
	data := templateData{
		UserNotifications: notificationViews, // We'll store in .Notifications
	}

	// 6) Render
	app.render(w, r, http.StatusOK, "notifications.html", data)
}
