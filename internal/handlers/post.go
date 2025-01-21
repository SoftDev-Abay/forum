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
	if err != nil || id < 1 {
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

	fullPost := &models.PostByUser{
		PostUserAdditionals: models.PostUserAdditionals{
			IsLiked:    false,
			IsDisliked: false,
		},
		Post: *post,
	}

	// Check if user is authenticated
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		// If not authenticated, they can view the post but can't interact with reactions

		// TODO: here it is not showing commnets need to check guest user view later
		app.render(w, r, http.StatusOK, "view.html", templateData{
			Category:   category,
			PostByUser: fullPost,
		})
		return
	}

	var reportReasons []*models.ReportReasons
	user, err := app.Users.GetById(userID)
	if err == nil && (user.Role == "moderator" || user.Role == "admin") {
		reportReasons, _ = app.ReportReasons.GetAllReasons()
		// if error, just ignore or handle
	} else if err != nil {
		app.serverError(w, r, err)
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
			fullPost.IsLiked = true
		} else if postReaction.Type == "dislike" {
			fullPost.IsDisliked = true
		}
	}

	comments, err := app.Comments.GetAllCommentsReactionsByPostID(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Render the post with its reactions
	data := templateData{
		Category:      category,
		PostByUser:    fullPost,
		Comments:      comments,
		ReportReasons: reportReasons,
		User:          user,
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

	app.render(w, r, http.StatusOK, "create.html", data)
}

func (app *Application) postCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(15 << 20)
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
		const maxFileSize = 15 << 20 // 20 MB in bytes

		v.CheckField(header.Size < maxFileSize, "image", "File too large: must be <= 15MB")

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

	// 1) Parse post ID from the query
	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// 2) Get reaction type from form
	reaction := r.FormValue("reaction")
	if reaction != "like" && reaction != "dislike" {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// 3) Get user ID from session
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	// 4) Check existing reaction
	existingReaction, err := app.PostReactions.GetReaction(userID, postID)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	// 5) Fetch the post to update counts and see the owner
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

	// 6) Reaction logic
	if existingReaction != nil {
		// (A) Reaction already exists
		if existingReaction.Type == reaction {
			// (A1) Same reaction => remove (toggle off)
			err = app.PostReactions.DeleteReaction(userID, postID)
			if err != nil {
				app.serverError(w, r, err)
				return
			}
			if reaction == "like" {
				newLikeCount--
			} else {
				newDislikeCount--
			}
		} else {
			// (A2) Different reaction => update
			err = app.PostReactions.UpdateReaction(userID, postID, reaction)
			if err != nil {
				app.serverError(w, r, err)
				return
			}
			if reaction == "like" {
				newLikeCount++
				newDislikeCount--
			} else {
				newLikeCount--
				newDislikeCount++
			}
		}
	} else {
		// (B) No existing reaction => add
		err = app.PostReactions.AddReaction(userID, postID, reaction)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if reaction == "like" {
			newLikeCount++
		} else {
			newDislikeCount++
		}
	}

	// 7) Update the post's like/dislike counts in DB
	err = app.Posts.UpdatePostLikeDislikeCounts(postID, newLikeCount, newDislikeCount)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 8) Add notification if the post belongs to a different user
	//    (so you don't notify someone about their own reaction)
	if post.OwnerID != userID {
		var notifType string
		if reaction == "like" {
			notifType = "post_like"
		} else {
			notifType = "post_dislike"
		}

		// Insert a new notification for the post owner
		_, notifErr := app.Notifications.Insert(
			notifType,
			userID,       // actor
			post.OwnerID, // recipient
			postID,
			nil, // no comment_id for a post reaction
		)
		if notifErr != nil {
			app.serverError(w, r, notifErr)
			return
		}
	}

	// 9) Finally, redirect to refresh the UI
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}

func (app *Application) postDelete(w http.ResponseWriter, r *http.Request) {
	// Only allow POST (or possibly DELETE, if you are using REST conventions).
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Parse the `id` from the query parameters (e.g. /post/delete?id=123).
	idStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// (Optional) Check if the user is authenticated and/or is allowed to delete the post.
	_, err = app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	// Retrieve the post to get the image URL
	post, err := app.Posts.Get(postID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.PostReactions.DeleteReactionsByPostId(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 1) Get all comments for the post
	comments, err := app.Comments.GetAllByPostId(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 2) For each comment, delete associated comment reactions
	for _, comment := range comments {
		err = app.CommentsReactions.DeleteReactioByCommentId(comment.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	// 3) Delete all comments for the post
	err = app.Comments.DeleteCommentsByPostId(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 4) Finally, delete the post itself
	err = app.Posts.DeletePostById(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// If the post has an image, delete the image file
	if post.ImgUrl != "" {
		imagePath := "./data/imgs/" + post.ImgUrl
		err = os.Remove(imagePath)
		if err != nil && !os.IsNotExist(err) {
			// Log the error but don't fail the request
			app.Logger.Error("Error deleting image file: %s, error: %v\n", imagePath, err)
		}
	}

	// Redirect or respond with success
	// e.g., redirect to homepage or back to some list:
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
