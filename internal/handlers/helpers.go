package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (app *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.Logger.Error(err.Error(), "method", method, "uri", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) clientError(w http.ResponseWriter, r *http.Request, status int) {
	data := templateData{
		ErrorCode: status,
		ErrorMsg:  http.StatusText(status),
	}

	app.render(w, r, status, "error.html", data)
}

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		ErrorCode: http.StatusNotFound,
		ErrorMsg:  http.StatusText(http.StatusNotFound),
	}

	app.render(w, r, http.StatusNotFound, "error.html", data)
}

func (app *Application) notAuthenticated(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		ErrorCode: http.StatusUnauthorized,
		ErrorMsg:  "You must be logged in to access this page.",
	}

	app.render(w, r, http.StatusUnauthorized, "error.html", data)
}

func (app *Application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	userID, err := app.getAuthenticatedUserID(r)
	if err == nil && userID > 0 {
		data.IsAuthenticated = true
		data.User, err = app.Users.GetById(userID)
		if err != nil {
			app.serverError(w, r, err)
		}
		data.NotificationsCount, err = app.Notifications.GetUsersNorificationsCount(data.User.ID)
		if err != nil {
			app.serverError(w, r, err)
		}

	} else {
		data.IsAuthenticated = false
	}

	ts, ok := app.TemplateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	if page == "error.html" {
		err = ts.ExecuteTemplate(buf, "error_base", data)
	} else {
		err = ts.ExecuteTemplate(buf, "base", data)
	}

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *Application) generateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (app *Application) compareHashPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (app *Application) getAuthenticatedUserID(r *http.Request) (int, error) {
	tokenCookie, err := r.Cookie("token")
	if err != nil || tokenCookie.Value == "" {
		return 0, errors.New("user not authenticated")
	}

	token := tokenCookie.Value

	userID, err := app.Session.GetUserIDByToken(token)
	if err != nil {
		return 0, errors.New("invalid session token")
	}

	return userID, nil
}

func isAllowedImageExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".svg":
		return true
	}
	return false
}

func generateUniqueFileName(originalFilename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalFilename))

	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return u.String() + ext, nil
}

func contains(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
