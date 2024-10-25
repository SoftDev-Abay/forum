package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"game-forum-abaliyev-ashirbay/internal/validator"
	"net/http"

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
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	// Initialize the validator
	v := validator.Validator{}

	// Perform validation using your validator package
	v.CheckField(validator.NotBlank(form.Email), "email", "Email cannot be blank")
	v.CheckField(validator.MaxChars(form.Email, 50), "email", "Email must not exceed 50 characters")
	v.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Invalid email address")

	v.CheckField(validator.NotBlank(form.Password), "password", "Password cannot be blank")
	v.CheckField(validator.MaxChars(form.Password, 30), "password", "Password must not exceed 30 characters")

	// If validation fails, re-render the form with errors
	if !v.Valid() {
		data := templateData{
			Form:       form,
			FormErrors: v.FieldErrors,
		}
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	// Retrieve the user from the database
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

	// Verify the password
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

	// Generate token and set cookies
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
	// Set the session token cookie
	err = setLoginCookies(r, w, userInfo, token)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// setLoginCookies sets the user info and token as cookies
func setLoginCookies(r *http.Request, w http.ResponseWriter, userInfo UserInfo, token string) error {
	// Serialize user info to JSON

	// // Set user info cookie
	// userInfoCookie := http.Cookie{
	// 	Name:     "user_info",
	// 	Value:    string(userInfoJSON),
	// 	Path:     "/",
	// 	MaxAge:   24 * 60 * 60 * 60 * 60, // 1 day
	// 	HttpOnly: true,
	// 	Secure:   true,
	// 	SameSite: http.SameSiteLaxMode,
	// }

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

	// r.AddCookie(&userInfoCookie)
	// r.AddCookie(&tokenCookie)

	// http.SetCookie(w, &userInfoCookie)
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
