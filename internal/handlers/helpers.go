package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

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

func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (app *Application) notAuthenticated(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
}

func (app *Application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	userID, err := app.getAuthenticatedUserID(r)
	if err == nil && userID > 0 {
		data.IsAuthenticated = true
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

	err = ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *Application) generateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	fmt.Println("errorka tut")
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
