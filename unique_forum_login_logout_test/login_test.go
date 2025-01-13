// Package unique_forum_login_logout_test ensures uniqueness for the code
package unique_forum_login_logout_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

const (
	// Use the path to your chromedriver or geckodriver
	// e.g. "/home/user/Downloads/chromedriver" or "/home/user/Downloads/geckodriver"
	driverPath = "/home/shiradi/Downloads/geckodriver-v0.35.0-linux64/geckodriver"

	seleniumPort = 9515
	baseURL      = "http://localhost:4000"
)

// setupWebDriver starts the WebDriver service and returns a selenium.WebDriver instance.
func setupWebDriver(t *testing.T) selenium.WebDriver {
	// If using Firefox + geckodriver, use:
	service, err := selenium.NewGeckoDriverService(driverPath, seleniumPort)

	if err != nil {
		t.Fatalf("Error starting WebDriver service: %v", err)
	}
	t.Cleanup(func() {
		service.Stop()
	})

	// ------------------------------------------------------
	// 2) Create browser capabilities
	// ------------------------------------------------------
	caps := selenium.Capabilities{
		"browserName": "firefox",
	}

	// ------------------------------------------------------
	// 3) Connect to the running service
	// ------------------------------------------------------
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", seleniumPort))
	if err != nil {
		t.Fatalf("Error creating new remote WebDriver: %v", err)
	}

	t.Cleanup(func() {
		wd.Quit()
	})

	return wd
}

// TestLoginLogout is an end-to-end test for login & logout.
func TestLoginLogout(t *testing.T) {
	wd := setupWebDriver(t)

	// 1) Navigate to the login page
	loginURL := baseURL + "/login"
	if err := wd.Get(loginURL); err != nil {
		t.Fatalf("Failed to load login page %s: %v", loginURL, err)
	}

	// Give it a moment to render
	time.Sleep(1 * time.Second)

	// 2) Fill in the login form
	// - Adjust these credentials to match a real user in your Forum DB
	validEmail := "test@example.com"
	validPassword := "Secret1234"

	//    2a) Email input (CSS selector example)
	emailField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='email']`)
	if err != nil {
		t.Fatalf("Could not find email input: %v", err)
	}
	if err := emailField.Clear(); err == nil {
		emailField.SendKeys(validEmail)
	}

	//    2b) Password input (CSS selector example)
	passwordField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='password']`)
	if err != nil {
		t.Fatalf("Could not find password input: %v", err)
	}
	if err := passwordField.Clear(); err == nil {
		passwordField.SendKeys(validPassword)
	}

	//    2c) Click "Login" button (XPath example)
	//        <input type='submit' value='Login'>
	loginBtn, err := wd.FindElement(selenium.ByXPATH, `//input[@type='submit' and @value='Login']`)
	if err != nil {
		t.Fatalf("Could not find the login submit button: %v", err)
	}
	if err := loginBtn.Click(); err != nil {
		t.Fatalf("Failed to click the login button: %v", err)
	}

	// 3) After login, wait for redirect or success
	time.Sleep(2 * time.Second)

	// 4) Validate success:
	//    One approach is to check if there's a "Logout" link
	logoutLink, err := wd.FindElement(selenium.ByLinkText, "Logout")
	if err != nil {
		t.Fatalf("Login might have failed; could not find 'Logout' link: %v", err)
	}

	// Alternatively, check cookies or check new URL:
	currentURL, _ := wd.CurrentURL()
	if !strings.Contains(currentURL, baseURL+"/") {
		t.Errorf("Expected to be on home page after login, got: %s", currentURL)
	}

	cookies, _ := wd.GetCookies()
	foundToken := false
	for _, c := range cookies {
		if c.Name == "token" {
			foundToken = true
			break
		}
	}
	if !foundToken {
		t.Errorf("Expected 'token' cookie after login, but none found")
	}

	// 5) Log out
	if err := logoutLink.Click(); err != nil {
		t.Fatalf("Failed to click the logout link: %v", err)
	}

	// 6) Wait for logout to complete
	time.Sleep(2 * time.Second)

	// 7) Validate we are logged out (no 'Logout' link, or check for 'Login' link, etc.)
	//    For example, look for "Login" link after logout
	_, err = wd.FindElement(selenium.ByLinkText, "Login")

	if err != nil {
		t.Errorf("Expected to find 'Login' link after logout, but got error: %v", err)
	} else {
		t.Log("Successfully found 'Login' link after logout. Logout is likely successful.")
	}
}
