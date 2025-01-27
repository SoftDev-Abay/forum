package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"game-forum-abaliyev-ashirbay/internal/validator"
	"net/http"
)

type registerForm struct {
	Email           string
	Username        string
	Password        string
	ConfirmPassword string
	FieldErrors     map[string]string
}

func (app *Application) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	data := templateData{}

	data.Form = registerForm{}

	app.render(w, r, http.StatusOK, "register.html", data)
}

func (app *Application) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	form := registerForm{
		Email:           r.PostForm.Get("email"),
		Username:        r.PostForm.Get("username"),
		Password:        r.PostForm.Get("password"),
		ConfirmPassword: r.PostForm.Get("confirmPassword"),
	}

	v := validator.Validator{}

	v.CheckField(validator.NotBlank(form.Email), "email", "Email cannot be blank")
	v.CheckField(validator.MaxChars(form.Email, 50), "email", "Email must not exceed 50 characters")
	v.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Invalid email address")

	v.CheckField(validator.NotBlank(form.Username), "username", "Username cannot be blank")
	v.CheckField(validator.MaxChars(form.Username, 30), "username", "Username must not exceed 30 characters")

	v.CheckField(validator.NotBlank(form.Password), "password", "Password cannot be blank")
	v.CheckField(validator.MinChars(form.Password, 8), "password", "Password must be at least 8 characters long")
	v.CheckField(validator.MaxChars(form.Password, 30), "password", "Password must not exceed 30 characters")

	v.CheckField(validator.NotBlank(form.ConfirmPassword), "confirmPassword", "Confirm Password cannot be blank")
	v.CheckField(form.Password == form.ConfirmPassword, "confirmPassword", "Passwords do not match")

	if !v.Valid() {
		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	hashedPassword, err := app.generateHashPassword(form.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_, err = app.Users.Insert(form.Email, form.Username, hashedPassword, false)
	if err != nil {
		if err == models.ErrDuplicateEmail {
			v.AddFieldError("email", "Email is already in use")
		} else if err == models.ErrDuplicateUsername {
			v.AddFieldError("username", "Username is already taken")
		} else {
			app.serverError(w, r, err)
			return
		}

		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
		}

		app.render(w, r, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
