package handlers

import (
	"fmt"
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
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w, r)
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
	// Ensure the user is authenticated

	user := r.Context().Value(userContextKey) // Assuming User is your user type

	fmt.Println("user")
	fmt.Println(user)

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Extract form values
	title := r.PostForm.Get("title")
	categoryIDStr := r.PostForm.Get("category_id")
	content := r.PostForm.Get("content")

	// Convert categoryID to integer
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		categoryID = 0
	}

	// Create a PostForm instance with the extracted data
	form := PostForm{
		Title:      title,
		CategoryID: uint(categoryID),
		Content:    content,
	}

	// Initialize the validator
	v := validator.Validator{}

	// Perform validation checks
	v.CheckField(validator.NotBlank(form.Title), "title", "Title must not be blank")
	v.CheckField(validator.MaxChars(form.Title, 100), "title", "Title must not be more than 100 characters long")

	v.CheckField(form.CategoryID > 0, "category_id", "You must select a category")

	v.CheckField(validator.NotBlank(form.Content), "content", "Content must not be blank")
	v.CheckField(validator.MinChars(form.Content, 10), "content", "Content must be at least 10 characters long")

	// If validation fails, re-display the form with errors
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

	// Insert the new post into the database
	postID, err := app.Posts.Insert(form.Title, form.Content, "", time.Now(), form.CategoryID, 1)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect to the newly created post
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}
