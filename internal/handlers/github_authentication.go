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

func (app *Application) githubLogin(w http.ResponseWriter, r *http.Request) {
    redirectURI := "http://localhost:4000/auth/github/callback"
    state := "someRandomState"

    githubURL := fmt.Sprintf(
        "https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
        app.GitHubClientID,
        url.QueryEscape(redirectURI),
        state,
    )
    http.Redirect(w, r, githubURL, http.StatusTemporaryRedirect)
}

func (app *Application) githubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}
	// Exchange code for token
	accessToken, err := exchangeGitHubCodeForToken(app.GitHubClientID, app.GitHubClientSecret, code)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Now fetch user info
	githubUser, err := getGitHubUser(accessToken)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Check if user is in DB by email or login name
	// If not found, create a new user
	user, err := app.Users.GetByUsernameOrEmail(githubUser.Login)
	if err == models.ErrNoRecord {
		hashedPass, _ := app.generateHashPassword("github_oauth_auto")
		fmt.Println(githubUser)
		newID, insertErr := app.Users.Insert(githubUser.Email, githubUser.Login, hashedPass, true)
		if insertErr != nil {
			app.serverError(w, r, insertErr)
			return
		}
		user, _ = app.Users.GetById(newID)
	} else if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Create session token, set cookie
	token, _ := GenerateToken()
	_, _ = app.Session.Insert(token, user.ID)
	setLoginCookies(r, w, UserInfo{ID: user.ID, Username: user.Username, Email: user.Email}, token)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Exchange code for token
func exchangeGitHubCodeForToken(clientID, clientSecret, code string) (string, error) {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("client_secret", clientSecret)
	params.Set("code", code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token: %s", string(body))
	}

	var data struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	
	return data.AccessToken, nil
}

// Fetch user info
func getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user: %s", string(b))
	}

	var ghUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, err
	}
	return &ghUser, nil
}

// Just an example struct
type GitHubUser struct {
	Login string `json:"login"`
	Email string `json:"email"`
	// more fields if needed
}
