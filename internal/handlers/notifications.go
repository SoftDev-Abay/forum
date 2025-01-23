package handlers

import (
	"net/http"
)

func (app *Application) notificationsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	notifs, err := app.Notifications.GetAllByRecipient(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var notificationViews []NotificationView

	for _, n := range notifs {
		actorUser, err := app.Users.GetById(n.Actor_ID)
		var actorName string
		if err == nil && actorUser != nil {
			actorName = actorUser.Username
		} else {
			actorName = "UnknownUser"
		}

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

	err = app.Notifications.MarkAllAsReadByUser(userID)
	if err != nil {
		app.Logger.Warn(err.Error())
	}

	data := templateData{
		UserNotifications: notificationViews, 
	}

	app.render(w, r, http.StatusOK, "notifications.html", data)
}
