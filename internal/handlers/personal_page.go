package handlers

import "net/http"

func (app *Application) personalPage(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
	}

	user, err := app.Users.GetById(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	userPosts, err := app.Posts.GetPostsByUserID(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		User: user,
		Posts: userPosts,
	}

	app.render(w, r, http.StatusOK, "personal_page.html", data)
}
