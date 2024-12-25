package handlers

import (
	"errors"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"game-forum-abaliyev-ashirbay/internal/validator"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type User struct {
	ID       int
	Username string
	Password string
	Email    string
	Enabled  bool
}

type PostForm struct {
	Title      string
	CategoryID int
	Content    string
}

func (app *Application) postView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil ||  id < 1 {
		app.clientError(w, r, http.StatusBadRequest)
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
	category, err := app.Categories.Get(post.CategoryID)
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
	postReaction, err := app.PostReactions.GetReaction(userID, id)
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

	comments, err := app.Comments.GetAllByPostIdAndUserId(userID, id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Render the post with its reactions
	data := templateData{
		Category: category,
		Post:     post,
		Comments: comments,
	}

	app.render(w, r, http.StatusOK, "view.html", data)
}

func (app *Application) postCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
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
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form (allow up to ~20 MB in memory; adjust if needed).
	err := r.ParseMultipartForm(20 << 20) // 20 MB
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	categoryIDStr := r.FormValue("category_id")
	content := r.FormValue("content")

	// Retrieve the file from the form. The field name is "image".
	file, header, imgErr := r.FormFile("image")
	if imgErr != nil && imgErr != http.ErrMissingFile {
		app.serverError(w, r, imgErr)
		return
	}
	// We may have no file uploaded (http.ErrMissingFile), so handle that possibility below.
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	// Convert categoryID
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		categoryID = 0
	}

	// Validation checks for title & content
	form := PostForm{
		Title:      title,
		CategoryID: categoryID,
		Content:    content,
	}
	v := validator.Validator{}

	// 1) If user provided a file (not missing)
	if imgErr != http.ErrMissingFile {
		// 2) Validate file size (<= 20 MB)
		const maxFileSize = 20 << 20 // 20 MB in bytes

		v.CheckField(header.Size < maxFileSize, "image", "File too large: must be <= 20MB")

		v.CheckField(isAllowedImageExt(header.Filename), "image", "Only .jpg, .png, or .gif files are allowed")

	}

	v.CheckField(validator.NotBlank(form.Title), "title", "Title must not be blank")
	v.CheckField(validator.MaxChars(form.Title, 100), "title", "Title must not be more than 100 characters long")
	v.CheckField(validator.MinChars(form.Title, 5), "title", "Title must be at least 5 characters long")

	v.CheckField(form.CategoryID > 0, "category_id", "You must select a category")

	v.CheckField(validator.NotBlank(form.Content), "content", "Content must not be blank")
	v.CheckField(validator.MinChars(form.Content, 10), "content", "Content must be at least 10 characters long")

	if !v.Valid() {
		categories, err := app.Categories.GetAll()
		if err != nil {
			app.serverError(w, r, err)
			return
		}

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
		return
	}

	// Initialize imgUrl as empty string for the DB if user doesn't upload an image
	imgUrl := ""

	if imgErr != http.ErrMissingFile {

		// 4) Generate random filename
		newFileName, err := generateUniqueFileName(header.Filename)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// 5) Save the file to ./data/imgs/<randomname>
		dst, err := os.Create("./data/imgs/" + newFileName)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// Assign new file name to store in DB
		imgUrl = newFileName
	}
		defer dst.Close()

	// 6) Insert post in the DB using `imgUrl`
	postID, err := app.Posts.Insert(title, content, imgUrl, time.Now(), categoryID, userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect to the newly created post
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}

func (app *Application) handlePostReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}
	// Get the post ID from the query parameters
	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// Get the reaction type (like or dislike) from the form
	reaction := r.FormValue("reaction")
	if reaction != "like" && reaction != "dislike" {
		app.clientError(w, r, http.StatusBadRequest)
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

	// lastly I need to update the count like dislike in post itself

	// Get the user ID from the session or context
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		// If not authenticated, return an error or handle gracefully
		app.notAuthenticated(w, r)
		return
	}

	// Check current user reaction to decide if they are changing their reaction
	existingReaction, err := app.PostReactions.GetReaction(userID, postID)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	post, err := app.Posts.Get(postID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	newLikeCount := post.LikeCount
	newDislikeCount := post.DislikeCount

	if existingReaction != nil {
		// User already reacted, handle toggling reactions
		if existingReaction.Type == reaction {
			// If they click on the same reaction, it will be removed (toggle)
			err = app.PostReactions.DeleteReaction(userID, postID)
			if err != nil {
				app.serverError(w, r, err)
				return
			}

			if reaction == "like" {
				newLikeCount -= 1
			} else {
				newDislikeCount -= 1
			}

		} else {
			// If they switch reactions, update accordingly
			err = app.PostReactions.UpdateReaction(userID, postID, reaction)
			if err != nil {
				app.serverError(w, r, err)
				return
			}

			if reaction == "like" {
				newLikeCount += 1
				newDislikeCount -= 1
			} else {
				newLikeCount -= 1
				newDislikeCount += 1
			}
		}
	} else {
		// No existing reaction, so we add the new one
		err = app.PostReactions.AddReaction(userID, postID, reaction)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		if reaction == "like" {
			newLikeCount += 1
		} else {
			newDislikeCount += 1
		}

	}
	err = app.Posts.UpdatePostLikeDislikeCounts(postID, newLikeCount, newDislikeCount)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// After updating, redirect to the post view to update the UI
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}
