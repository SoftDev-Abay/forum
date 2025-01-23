package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"net/http"
)

func (app *Application) personalPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	userPosts, err := app.Posts.GetPostsByUserID(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	likedPostIDs, err := app.PostReactions.GetLikedPostIDsByUserID(userID)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	likedPosts, err := app.Posts.GetPostsByIDs(likedPostIDs)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	comments, err := app.Comments.GetAllByUserId(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		Posts:               userPosts,  // The user’s own posts
		LikedPosts:          likedPosts, // The user’s liked posts
		CommentPostAddition: comments,
	}

	app.render(w, r, http.StatusOK, "personal_page.html", data)
}
