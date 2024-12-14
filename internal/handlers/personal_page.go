package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"net/http"
)

func (app *Application) personalPage(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
	}

	userPosts, err := app.Posts.GetPostsByUserID(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	postReaction, err := app.PostReactions.GetReactionByUserID(userID)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		Posts:        userPosts,
		PostReaction: postReaction,
	}

	app.render(w, r, http.StatusOK, "personal_page.html", data)
}
