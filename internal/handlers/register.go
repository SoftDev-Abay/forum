package handlers

import (
	"net/http"
)

// Define a snippetCreateForm struct to represent the form data and validation
// errors for the form fields. Note that all the struct fields are deliberately
// exported (i.e. start with a capital letter). This is because struct fields
// must be exported in order to be read by the html/template package when
// rendering the template.
type registerForm struct {
	Email           string
	Username        string
	Password        string
	ConfirmPassword string
	FieldErrors     map[string]string
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	// Initialize a new createSnippetForm instance and pass it to the template.
	// Notice how this is also a great opportunity to set any default or
	// 'initial' values for the form --- here we set the initial value for the
	// snippet expiry to 365 days.
	data.Form = registerForm{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "register.html", data)
}

func (app *application) registerPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := registerForm{
		Email:       r.PostForm.Get("email"),
		Username:    r.PostForm.Get("username"),
		Password:    r.PostForm.Get("password4"),
		FieldErrors: map[string]string{},
	}
	// Update the validation checks so that they operate on the snippetCreateForm
	// instance.
	if strings.TrimSpace(form.Email) == "" {
		form.FieldErrors["email"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Email) > 50 {
		form.FieldErrors["email"] = "This field cannot be more than 50 characters long"
	}
	if strings.TrimSpace(form.Username) == "" {
		form.FieldErrors["username"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Username) > 30 {
		form.FieldErrors["username"] = "This field cannot be more than 30 characters long"
	}
	if strings.TrimSpace(form.Password) == "" {
		form.FieldErrors["password"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Password) > 30 {
		form.FieldErrors["password"] = "This field cannot be more than 30 characters long"
	}
	if strings.TrimSpace(form.ConfirmPassword) == "" {
		form.FieldErrors["confirmPassword"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.ConfirmPassword) > 30 {
		form.FieldErrors["confirmPassword"] = "Cofirm password has to be the same"
	}

	// If there are any validation errors re-display the create.html template,
	// passing in the snippetCreateForm instance as dynamic data in the Form
	// field. Note that we use the HTTP status code 422 Unprocessable Entity
	// when sending the response to indicate that there was a validation error.
	if len(form.FieldErrors) > 0 {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	hashedPassword := app.GenerateHashPassword(form.Password)

	// We also need to update this line to pass the data from the
	// snippetCreateForm instance to our Insert() method.
	id, err := app.users.Insert(form.Email, form.Username, hashedPassword)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/home"), http.StatusSeeOther)
}
