package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"game-forum-abaliyev-ashirbay/internal/validator"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
)

type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type loginForm struct {
	Email       string
	Password    string
	FieldErrors map[string]string
}

func (app *Application) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	data := templateData{}

	data.Form = loginForm{}

	app.render(w, r, http.StatusOK, "login.html", data)
}

func (app *Application) LoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	form := loginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.Email = strings.ToLower(form.Email)
	v := validator.Validator{}
	v.CheckField(validator.NotBlank(form.Email), "email", "Email cannot be blank")
	v.CheckField(validator.MaxChars(form.Email, 50), "email", "Email must not exceed 50 characters")
	v.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Invalid email address")

	v.CheckField(validator.NotBlank(form.Password), "password", "Password cannot be blank")
	v.CheckField(validator.MaxChars(form.Password, 30), "password", "Password must not exceed 30 characters")

	if !v.Valid() {
		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	user, err := app.Users.GetByUsernameOrEmail(form.Email)
	if err != nil {
		if err == models.ErrNoRecord {
			v.AddFieldError("general", "Incorrect email or password")
			data := templateData{
				Form:       form,
				FormErrors: v.FieldErrors,
			}
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
			return
		}
		app.serverError(w, r, err)
		return
	}

	err = app.compareHashPassword(form.Password, user.Password)
	if err != nil {
		v.AddFieldError("general", "Incorrect email or password")
		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	existingToken, err := app.Session.GetByUserId(user.ID)
	if err != nil {
		// If there's a real error
		if errors.Is(err, models.ErrNoRecord) {
			// No session found, do nothing
		} else {
			// Some unexpected error
			app.serverError(w, r, err)
			return
		}
	} else {
		// err == nil
		if existingToken != nil {
			// There's an existing session for this user
			// remove it so we create a fresh one
			app.Session.DeleteByUserId(user.ID)
			// handle any error from DeleteByUserId if needed
		}
	}

	token, err := GenerateToken()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_, err = app.Session.Insert(token, user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	err = setLoginCookies(r, w, userInfo, token)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func setLoginCookies(r *http.Request, w http.ResponseWriter, userInfo UserInfo, token string) error {
	tokenCookie := http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60 * 60 * 60, // 1 day
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &tokenCookie)

	return nil
}

func GenerateToken() (string, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}

func getUserInfoFromCookie(r *http.Request) (*UserInfo, error) {
	cookie, err := r.Cookie("user_info")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, fmt.Errorf("no user info cookie found")
		}
		return nil, err
	}

	var userInfo UserInfo
	err = json.Unmarshal([]byte(cookie.Value), &userInfo)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user info: %w", err)
	}

	return &userInfo, nil
}

func (app *Application) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	userId, err := app.getAuthenticatedUserID(r)
	if err != nil {

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = app.Session.DeleteByUserId(userId)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	deleteCookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &deleteCookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
