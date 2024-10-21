package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid"
)

// UserInfo struct to hold user information
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type loginForm struct {
	Email       string
	Password    string
	FieldErrors map[string]string
}

func (app *Application) login(w http.ResponseWriter, r *http.Request) {
	data := templateData{}

	data.Form = loginForm{}

	app.render(w, r, http.StatusOK, "login.html", data)
}

func (app *Application) loginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := loginForm{
		Email:       r.PostForm.Get("email"),
		Password:    r.PostForm.Get("password"),
		FieldErrors: map[string]string{},
	}

	if strings.TrimSpace(form.Email) == "" {
		form.FieldErrors["email"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Email) > 50 {
		form.FieldErrors["email"] = "This field cannot be more than 50 characters long"
	}
	if strings.TrimSpace(form.Password) == "" {
		form.FieldErrors["password"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Password) > 30 {
		form.FieldErrors["password"] = "This field cannot be more than 30 characters long"
	}

	if len(form.FieldErrors) > 0 {
		fmt.Println("log 1 ")

		data := templateData{}
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	user, err := app.Users.GetByUsernameOrEmail(form.Email)
	if err != nil {
		fmt.Println("log 2 ")

		fmt.Println(user)
		fmt.Println(err)

		if err == models.ErrNoRecord {
			fmt.Println("log 3 ")

			form.FieldErrors["general"] = "Incorrect password or email"

			data := templateData{}
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
			return
		}

		data := templateData{}
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	errHash := app.compareHashPassword(form.Password, user.Password)

	if !errHash {
		fmt.Println("log 4 ")

		form.FieldErrors["general"] = "Incorrect password or email"

		data := templateData{}
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}
	fmt.Println("log 5 ")

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
	fmt.Println("log 6")

	err = setLoginCookies(w, userInfo, token)

	if err != nil {
		fmt.Println(err)
		app.serverError(w, r, err)
		return
	}

	fmt.Println("hello")
	http.Redirect(w, r, fmt.Sprintf("/"), http.StatusSeeOther)
}

// setLoginCookies sets the user info and token as cookies
func setLoginCookies(w http.ResponseWriter, userInfo UserInfo, token string) error {
	// Serialize user info to JSON
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		fmt.Println("Error marshaling user info:", err)
		return err
	}

	// Set user info cookie
	userInfoCookie := http.Cookie{
		Name:     "user_info",
		Value:    string(userInfoJSON),
		Path:     "/",
		MaxAge:   24 * 60 * 60 * 60 * 60, // 1 day
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	// Set token cookie
	tokenCookie := http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60 * 60 * 60, // 1 day
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &userInfoCookie)
	http.SetCookie(w, &tokenCookie)

	return nil
}

// GenerateToken generates a new UUID token
func GenerateToken() (string, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil // Returns the UUID as a string
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
