package handlers

import (
	"errors"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"game-forum-abaliyev-ashirbay/internal/validator"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	ID       uint
	Username string
	Password string
	Email    string
	Enabled  bool
}

type PostForm struct {
	Title      string
	CategoryID uint
	Content    string
}

func (app *Application) postView(w http.ResponseWriter, r *http.Request) {
	fmt.Println("post view accessed ")

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w, r)
		return
	}

	post, err := app.Posts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	category, err := app.Categories.Get(int(post.CategoryID))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Default reaction is none
	post.IsLiked = false
	post.IsDisliked = false

	// Check if user is authenticated
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		// If not authenticated, they can view the post but can't interact with reactions
		app.render(w, r, http.StatusOK, "view.html", templateData{
			Category: category,
			Post:     post,
		})
		return
	}

	// Get the user's reaction on this post (if any)
	postReaction, err := app.PostReactions.GetReaction(userID, uint(id))
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	// Set flags based on the reaction type
	if postReaction != nil {
		if postReaction.Type == "like" {
			post.IsLiked = true
		} else if postReaction.Type == "dislike" {
			post.IsDisliked = true
		}
	}

	// Render the post with its reactions
	data := templateData{
		Category: category,
		Post:     post,
	}

	app.render(w, r, http.StatusOK, "view.html", data)
}

func (app *Application) postCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	categories, err := app.Categories.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		Categories: categories,
	}
	fmt.Println(categories)

	app.render(w, r, http.StatusOK, "create.html", data)
}

func (app *Application) postCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	categoryIDStr := r.PostForm.Get("category_id")
	content := r.PostForm.Get("content")

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		categoryID = 0
	}

	form := PostForm{
		Title:      title,
		CategoryID: uint(categoryID),
		Content:    content,
	}

	v := validator.Validator{}

	v.CheckField(validator.NotBlank(form.Title), "title", "Title must not be blank")
	v.CheckField(validator.MaxChars(form.Title, 100), "title", "Title must not be more than 100 characters long")
	v.CheckField(validator.MinChars(form.Title, 5), "title", "Title must be at least 5 characters long")

	v.CheckField(form.CategoryID > 0, "category_id", "You must select a category")

	v.CheckField(validator.NotBlank(form.Content), "content", "Content must not be blank")
	v.CheckField(validator.MinChars(form.Content, 10), "content", "Content must be at least 10 characters long")

	fmt.Println(form)

	if !v.Valid() {
		categories, err := app.Categories.GetAll()
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		fmt.Println(categories)
		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
			Categories: categories,
		}

		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	userId, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
	}

	postID, err := app.Posts.Insert(form.Title, form.Content, "", time.Now(), form.CategoryID, userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}

func (app *Application) handlePostReaction(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the query parameters
	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID < 1 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Get the reaction type (like or dislike) from the form
	reaction := r.FormValue("reaction")
	if reaction != "like" && reaction != "dislike" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// reaction param can be either a like or a dislike
	// post id is always that post id

	// first I need to get the post reaction if exists

	// three ways this can go:
	// 1) it doesnt exist
	// I need to create a post reaction with respectfull type: like/dislike

	// 2) it exists and its the same - the reaction with same type was already made
	// suppose its a like
	// then I need to delete this reaction completely

	// 3) it exists but its a different reaction type

	// I need to update the found reaction to a different type

	// Get the user ID from the session or context
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		// If not authenticated, return an error or handle gracefully
		app.notAuthenticated(w, r)
		return
	}

	// Check current user reaction to decide if they are changing their reaction
	existingReaction, err := app.PostReactions.GetReaction(userID, uint(postID))
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	if existingReaction != nil {
		// User already reacted, handle toggling reactions
		if existingReaction.Type == reaction {
			// If they click on the same reaction, it will be removed (toggle)
			err = app.PostReactions.DeleteReaction(userID, uint(postID))
			if err != nil {
				app.serverError(w, r, err)
				return
			}
		} else {
			// If they switch reactions, update accordingly
			err = app.PostReactions.UpdateReaction(userID, uint(postID), reaction)
			if err != nil {
				app.serverError(w, r, err)
				return
			}
		}
	} else {
		// No existing reaction, so we add the new one
		err = app.PostReactions.AddReaction(userID, uint(postID), reaction)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	// After updating, redirect to the post view to update the UI
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}
