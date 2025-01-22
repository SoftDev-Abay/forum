package handlers

import (
	"encoding/json"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (app *Application) googleLogin(w http.ResponseWriter, r *http.Request) {
	// 1) The "endpoint" to start Google OAuth2
	// This is the authorization URL with required query params:
	// - client_id, redirect_uri, scope, response_type=code, state, etc.

	redirectURI := "https://localhost" + *app.Addr + "/auth/google/callback"

	// Typically you want a random "state" token to prevent CSRF, but keep it simple for now
	state := "xyz123"

	googleAuthURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile&state=%s&access_type=offline",
		app.GoogleClientID,
		url.QueryEscape(redirectURI),
		state,
	)

	// 2) Redirect the user to Google's consent screen
	http.Redirect(w, r, googleAuthURL, http.StatusTemporaryRedirect)
}

func (app *Application) googleCallback(w http.ResponseWriter, r *http.Request) {
	// 1) Parse the query params
	code := r.URL.Query().Get("code")
	if code == "" {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// 2) Exchange the code for an access token
	tokenResp, err := app.exchangeGoogleCodeForToken(app.GoogleClientID, app.GoogleClientSecret, code)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 3) Use the access token to get the user's profile
	googleUser, err := getGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 4) googleUser has e.g. {Email, Name, Sub (ID), ...}

	// 5) In your DB, check if email already exists
	user, err := app.Users.GetByUsernameOrEmail(googleUser.Email)
	if err == models.ErrNoRecord {
		// Not in DB => create a new user
		randomPass := "google_oauth_auto" // or some random string
		hashedPass, _ := app.generateHashPassword(randomPass)

		// Insert the new user
		newID, insertErr := app.Users.Insert(googleUser.Email, googleUser.Name, hashedPass, true)
		if insertErr != nil {
			app.serverError(w, r, insertErr)
			return
		}

		user, _ = app.Users.GetById(newID) // now we have the user object
	} else if err != nil {
		// Another error
		app.serverError(w, r, err)
		return
	}

	// 6) Now create a session token & cookie (just like normal login)
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
	// Set the cookie with "token"
	setLoginCookies(r, w, UserInfo{ID: user.ID, Username: user.Username, Email: user.Email}, token)

	// 7) Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Helper function: exchange code for token
type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	IdToken     string `json:"id_token"`
}

func (app *Application) exchangeGoogleCodeForToken(clientID, clientSecret, code string) (*GoogleTokenResponse, error) {
	redirectURI := "https://localhost" + *app.Addr + "/auth/google/callback"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google token exchange failed: %s", string(bodyBytes))
	}

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

// Helper function: get Google user info
type GoogleUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
	// Possibly more fields
}

func getGoogleUserInfo(accessToken string) (*GoogleUser, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get google user info: %s", string(bodyBytes))
	}

	var user GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
