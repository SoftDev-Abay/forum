package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

type registerForm struct {
	Email           string
	Username        string
	Password        string
	ConfirmPassword string
	FieldErrors     map[string]string
}

func (app *Application) register(w http.ResponseWriter, r *http.Request) {
	data := templateData{}

	data.Form = registerForm{}

	app.render(w, r, http.StatusOK, "register.html", data)
}

func (app *Application) registerPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := registerForm{
		Email:           r.PostForm.Get("email"),
		Username:        r.PostForm.Get("username"),
		Password:        r.PostForm.Get("password"),
		ConfirmPassword: r.PostForm.Get("confirmPassword"),
		FieldErrors:     map[string]string{},
	}

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

	if len(form.FieldErrors) > 0 {
		data := templateData{}
		data.Form = form
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
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/"), http.StatusSeeOther)
}
