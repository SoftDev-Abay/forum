package unique_forum_login_logout_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

const (
	driverPath = "/home/student/forum/geckodriver"

	seleniumPort = 9515
	baseURL      = "https://localhost:8433"
)

func setupWebDriver(t *testing.T) selenium.WebDriver {
	service, err := selenium.NewGeckoDriverService(driverPath, seleniumPort)
	if err != nil {
		t.Fatalf("Error starting WebDriver service: %v", err)
	}
	t.Cleanup(func() {
		service.Stop()
	})

	caps := selenium.Capabilities{
		"acceptInsecureCerts": true,
		"browserName":         "firefox",
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", seleniumPort))
	if err != nil {
		t.Fatalf("Error creating new remote WebDriver: %v", err)
	}

	t.Cleanup(func() {
		wd.Quit()
	})

	return wd
}

func TestLoginLogout(t *testing.T) {
	wd := setupWebDriver(t)

	loginURL := baseURL + "/login"
	if err := wd.Get(loginURL); err != nil {
		t.Fatalf("Failed to load login page %s: %v", loginURL, err)
	}

	time.Sleep(1 * time.Second)

	validEmail := "admin@gmail.com"
	validPassword := "12345678"

	emailField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='email']`)
	if err != nil {
		t.Fatalf("Could not find email input: %v", err)
	}
	if err := emailField.Clear(); err == nil {
		emailField.SendKeys(validEmail)
	}

	passwordField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='password']`)
	if err != nil {
		t.Fatalf("Could not find password input: %v", err)
	}
	if err := passwordField.Clear(); err == nil {
		passwordField.SendKeys(validPassword)
	}

	loginBtn, err := wd.FindElement(selenium.ByXPATH, `//input[@type='submit' and @value='Login']`)
	if err != nil {
		t.Fatalf("Could not find the login submit button: %v", err)
	}
	if err := loginBtn.Click(); err != nil {
		t.Fatalf("Failed to click the login button: %v", err)
	}

	time.Sleep(2 * time.Second)

	logoutLink, err := wd.FindElement(selenium.ByLinkText, "Logout")
	if err != nil {
		t.Fatalf("Login might have failed; could not find 'Logout' link: %v", err)
	}

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

	if err := logoutLink.Click(); err != nil {
		t.Fatalf("Failed to click the logout link: %v", err)
	}

	time.Sleep(2 * time.Second)

	_, err = wd.FindElement(selenium.ByLinkText, "Login")

	if err != nil {
		t.Errorf("Expected to find 'Login' link after logout, but got error: %v", err)
	} else {
		t.Log("Successfully found 'Login' link after logout. Logout is likely successful.")
	}
}
