You said:
I need to get the user from the context in postCreatePost

type User struct {
	ID       uint
	Username string
	Password string
	Email    string
	Enabled  bool
}
func (app *Application) postCreatePost(w http.ResponseWriter, r *http.Request) {
	// Ensure the user is authenticated
	

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

	user := r.Context().Value(userContextKey) // Assuming User is your user type

	fmt.Println(user)

	// Insert the new post into the database
	postID, err := app.Posts.Insert(form.Title, form.Content, "", time.Now(), form.CategoryID, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect to the newly created post
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}


type contextKey string

const userContextKey contextKey = "userContextKey"

func (app *Application) loginMiddware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			fmt.Println(err)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := app.Users.GetByToken(tokenCookie.Value)
		if err != nil {
			fmt.Println(err)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Attach userInfo to the request context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
ChatGPT said:
ChatGPT

To get the user from the context in your postCreatePost function, you are almost there. You just need to ensure that you properly extract the user information from the context and handle any potential type assertion. Here’s how you can modify your postCreatePost method to retrieve the user correctly:

go

func (app *Application) postCreatePost(w http.ResponseWriter, r *http.Request) {
	// Ensure the user is authenticated
	user, ok := r.Context().Value(userContextKey).(*User) // Assuming *User is your user type
	if !ok || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
			Categories: categories,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	// Use the user's ID for inserting the post
	postID, err := app.Posts.Insert(form.Title, form.Content, "", time.Now(), form.CategoryID, user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect to the newly created post
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postID), http.StatusSeeOther)
}

Key Changes:

    User Retrieval: We extract the user from the context and assert its type. If the user is not found or is of an unexpected type, we redirect to the login page.
    Use user.ID: When inserting the post, we use the ID from the authenticated user.

This will ensure that your postCreatePost function correctly handles user authentication and retrieves user data as needed.
You said:
	user, ok := r.Context().Value(userContextKey).(*User) // Assuming *User is your user type
	if !ok || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
ChatGPT said:
ChatGPT

Your code snippet for retrieving the user from the context looks good! It checks if the user exists and whether the type assertion is successful. Here’s a brief breakdown of what you have and a slight enhancement to clarify things:
Breakdown

go

user, ok := r.Context().Value(userContextKey).(*User) // Attempt to retrieve user from context
if !ok || user == nil { // Check if the user was found and if the type assertion was successful
	http.Redirect(w, r, "/login", http.StatusSeeOther) // Redirect to login if not authenticated
	return
}

Enhancement

To make it even clearer, you might consider adding a comment explaining what userContextKey is, for future reference:

go

// Retrieve the user from the context
user, ok := r.Context().Value(userContextKey).(*User) // userContextKey is used to store user information in the context
if !ok || user == nil {
	http.Redirect(w, r, "/login", http.StatusSeeOther) // Redirect to login if not authenticated
	return
}

This small addition helps anyone reading the code understand that userContextKey is crucial for accessing the user data stored in the request context.

If your application has extensive logging or error handling, you might also want to log the situation when a user is not found or is of the wrong type, but that would depend on your application's logging strategy.

Everything else in your implementation looks solid! If you have any other questions or need further assistance, feel free to ask!


ChatGPT can make mistakes. Check important info.
