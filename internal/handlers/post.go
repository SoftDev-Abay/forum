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
