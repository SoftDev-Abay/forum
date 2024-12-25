package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"net/http"
)

func (app *Application) personalPage(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	// 1) Get the user's own posts
	userPosts, err := app.Posts.GetPostsByUserID(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 2) Get all postIDs the user has liked
	likedPostIDs, err := app.PostReactions.GetLikedPostIDsByUserID(userID)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	// 3) Fetch the actual liked posts (if any)
	likedPosts, err := app.Posts.GetPostsByIDs(likedPostIDs)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 4) Prepare template data
	data := templateData{
		Posts:      userPosts,  // The user’s own posts
		LikedPosts: likedPosts, // The user’s liked posts
	}

	app.render(w, r, http.StatusOK, "personal_page.html", data)
}
