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

	author, err := app.Users.GetById(post.OwnerID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	fullPost := &models.PostByUser{
		PostUserAdditionals: models.PostUserAdditionals{
			IsLiked:    false,
			IsDisliked: false,
		},
		Post: *post,
		PostAdditionals: models.PostAdditionals{
			OwnerName:    author.Username,
			CategoryName: category.Name,
		},
	}

	userId, err := app.getAuthenticatedUserID(r)

	comments, err := app.Comments.GetAllCommentsReactionsByPostID(id, userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.render(w, r, http.StatusOK, "view.html", templateData{
			Category:   category,
			PostByUser: fullPost,
			Comments:   comments,
		})
		return
	}

	var reportReasons []*models.ReportReasons
	user, err := app.Users.GetById(userID)
	if err == nil && (user.Role == "moderator" || user.Role == "admin") {
		reportReasons, _ = app.ReportReasons.GetAllReasons()
	} else if err != nil {
		app.serverError(w, r, err)
	}

	postReaction, err := app.PostReactions.GetReaction(userID, id)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	if postReaction != nil {
		if postReaction.Type == "like" {
			fullPost.IsLiked = true
		} else if postReaction.Type == "dislike" {
			fullPost.IsDisliked = true
		}
	}

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

	file, header, imgErr := r.FormFile("image")
	if imgErr != nil && imgErr != http.ErrMissingFile {
		app.serverError(w, r, imgErr)
		return
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		categoryID = 0
	}

	form := PostForm{
		Title:      title,
		CategoryID: categoryID,
		Content:    content,
	}
	v := validator.Validator{}

	if imgErr != http.ErrMissingFile {
		const maxFileSize = 15 << 20

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

	imgUrl := ""

	if imgErr != http.ErrMissingFile {
		newFileName, err := generateUniqueFileName(header.Filename)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

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

		imgUrl = newFileName
	}

	postID, err := app.Posts.Insert(title, content, imgUrl, time.Now(), categoryID, userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}

func (app *Application) handlePostReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	reaction := r.FormValue("reaction")
	if reaction != "like" && reaction != "dislike" {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

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
		if existingReaction.Type == reaction {
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

	err = app.Posts.UpdatePostLikeDislikeCounts(postID, newLikeCount, newDislikeCount)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if post.OwnerID != userID {
		var notifType string
		if reaction == "like" {
			notifType = "post_like"
		} else {
			notifType = "post_dislike"
		}

		_, notifErr := app.Notifications.Insert(
			notifType,
			userID,
			post.OwnerID,
			postID,
			nil,
		)
		if notifErr != nil {
			app.serverError(w, r, notifErr)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}

func (app *Application) postDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
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

	user, err := app.Users.GetById(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if post.OwnerID != userID && user.Role != "moderator" && user.Role != "admin" {
		app.clientError(w, r, http.StatusForbidden)
		return
	}

	err = app.PostReactions.DeleteReactionsByPostId(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	comments, err := app.Comments.GetAllByPostId(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	for _, comment := range comments {
		err = app.CommentsReactions.DeleteReactioByCommentId(comment.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	err = app.Comments.DeleteCommentsByPostId(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.Posts.DeletePostById(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if post.ImgUrl != "" {
		imagePath := "./data/imgs/" + post.ImgUrl
		err = os.Remove(imagePath)
		if err != nil && !os.IsNotExist(err) {
			app.Logger.Error("Error deleting image file: %s, error: %v\n", imagePath, err)
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) postEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
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

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	if post.OwnerID != userID {
		app.clientError(w, r, http.StatusForbidden)
		return
	}

	categories, err := app.Categories.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		Post:       post,
		Categories: categories,
	}

	app.render(w, r, http.StatusOK, "edit_post.html", data)
}

func (app *Application) postEditPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil || postID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
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

	if post.OwnerID != userID {
		app.clientError(w, r, http.StatusForbidden)
		return
	}

	err = r.ParseMultipartForm(15 << 20)
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	categoryIDStr := r.FormValue("category_id")
	content := r.FormValue("content")

	file, header, imgErr := r.FormFile("image")
	if imgErr != nil && imgErr != http.ErrMissingFile {
		app.serverError(w, r, imgErr)
		return
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		categoryID = 0
	}

	form := PostForm{
		Title:      title,
		CategoryID: categoryID,
		Content:    content,
	}
	v := validator.Validator{}

	if imgErr != http.ErrMissingFile {
		const maxFileSize = 15 << 20
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

		app.render(w, r, http.StatusUnprocessableEntity, "edit_post.html", data)
		return
	}

	imgUrl := post.ImgUrl
	if imgErr != http.ErrMissingFile {
		newFileName, err := generateUniqueFileName(header.Filename)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

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

		imgUrl = newFileName

		if post.ImgUrl != "" {
			imagePath := "./data/imgs/" + post.ImgUrl
			err = os.Remove(imagePath)
			if err != nil && !os.IsNotExist(err) {
				app.Logger.Error("Error deleting image file: %s, error: %v\n", imagePath, err)
			}
		}
	}

	err = app.Posts.UpdatePost(postID, title, content, imgUrl, categoryID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}
