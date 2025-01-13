// Package forum_advanced_test ensures uniqueness
package forum_advanced_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

// Adjust these constants to match your environment
const (
	driverPath   = "/home/shiradi/Downloads/geckodriver-v0.35.0-linux64/geckodriver"
	seleniumPort = 9515
	baseURL      = "http://localhost:4000"
)

// setupWebDriver sets up the WebDriver service and returns a selenium.WebDriver
func setupWebDriver(t *testing.T) selenium.WebDriver {
	// 1) Start GeckoDriver
	service, err := selenium.NewGeckoDriverService(driverPath, seleniumPort)
	if err != nil {
		t.Fatalf("Cannot start geckodriver service: %v", err)
	}
	t.Cleanup(func() {
		service.Stop()
	})

	// 2) Desired Capabilities
	caps := selenium.Capabilities{
		"browserName": "firefox",
		"moz:firefoxOptions": map[string]interface{}{
			// Remove/comment out "args" if you want to see the browser run in normal mode
			"args": []string{"-headless"},
		},
	}

	// 3) Create a new Remote WebDriver
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", seleniumPort))
	if err != nil {
		t.Fatalf("Error creating WebDriver session: %v", err)
	}

	// Ensure WebDriver quits at test end
	t.Cleanup(func() {
		wd.Quit()
	})

	return wd
}

// TestAdvancedSelenium demonstrates all required features.
func TestAdvancedSelenium(t *testing.T) {
	wd := setupWebDriver(t)

	// ----------------------------------------------------------------------------
	// 1) Implement Implicit, Explicit, and Fluent Wait
	// ----------------------------------------------------------------------------

	// a) Implicit Wait: WebDriver will poll for elements up to 5s
	if err := wd.SetImplicitWaitTimeout(5 * time.Second); err != nil {
		t.Fatalf("Failed to set implicit wait: %v", err)
	}

	// Navigate to /login
	if err := wd.Get(baseURL + "/login"); err != nil {
		t.Fatalf("Failed to load /login page: %v", err)
	}

	// b) Explicit Wait for "email" input
	err := wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		_, e := wd.FindElement(selenium.ByCSSSelector, `input[name="email"]`)
		return e == nil, nil
	}, 10*time.Second, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Explicit wait: email input not found: %v", err)
	}

	// c) Fluent Wait for "password" input
	err = wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		_, findErr := wd.FindElement(selenium.ByCSSSelector, `input[name="password"]`)
		return findErr == nil, nil
	}, 15*time.Second, 1*time.Second)
	if err != nil {
		t.Fatalf("Fluent wait: password input not found: %v", err)
	}

	t.Log("Waits (implicit, explicit, fluent) validated successfully.")

	// ----------------------------------------------------------------------------
	// 2) Implement Action Class (Hover)
	// ----------------------------------------------------------------------------

	// Fill in login fields
	emailElem, _ := wd.FindElement(selenium.ByCSSSelector, `input[name="email"]`)
	emailElem.Clear()
	emailElem.SendKeys("test@example.com") // Adjust to a valid user
	passwordElem, _ := wd.FindElement(selenium.ByCSSSelector, `input[name="password"]`)
	passwordElem.Clear()
	passwordElem.SendKeys("Secret1234") // Adjust to correct password

	loginBtn, _ := wd.FindElement(selenium.ByXPATH, `//input[@type="submit" and @value="Login"]`)
	loginBtn.Click()

	// Wait for the redirect
	time.Sleep(2 * time.Second)

	// Try to hover over the "Create" link
	navCreate, err := wd.FindElement(selenium.ByCSSSelector, `nav a[href="/post/create"]`)
	if err != nil {
		t.Log("Could not find 'Create' link, skipping hover example.")
	} else {
		// "Hover" by moving mouse over that element
		if moveErr := navCreate.MoveTo(0, 0); moveErr != nil {
			t.Errorf("Failed to hover over 'Create' link: %v", moveErr)
		} else {
			t.Log("Successfully hovered over 'Create' link (Action Class).")
		}
	}

	// ----------------------------------------------------------------------------
	// 3) Implement Select Class
	// ----------------------------------------------------------------------------

	// Navigate to /post/create
	if err := wd.Get(baseURL + "/post/create"); err != nil {
		t.Fatalf("Failed to load /post/create page: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Find the category dropdown
	categorySelect, err := wd.FindElement(selenium.ByCSSSelector, `select[name="category_id"]`)
	if err != nil {
		t.Log("No category select found. Possibly no categories exist.")
	} else {
		// We can pick an option by text or value with .SendKeys("1")
		if sendErr := categorySelect.SendKeys("1"); sendErr != nil {
			t.Errorf("Failed to select category '1': %v", sendErr)
		} else {
			t.Log("Select Class: Chose category with visible text or value '1'.")
		}
	}

	// ----------------------------------------------------------------------------
	// 4) Make a test report
	// ----------------------------------------------------------------------------
	// "go test -v" outputs logs to console. You can do: go test -v > test-report.txt
	t.Log("Test complete. Check the terminal output (or test-report.txt) for results.")
}

// TestLoginLogout re-uses your existing login/logout test as an example
func TestLoginLogout(t *testing.T) {
	wd := setupWebDriver(t)

	// 1) Navigate to the login page
	loginURL := baseURL + "/login"
	if err := wd.Get(loginURL); err != nil {
		t.Fatalf("Failed to load login page %s: %v", loginURL, err)
	}

	time.Sleep(1 * time.Second)

	validEmail := "test@example.com"
	validPassword := "Secret1234"

	// Fill email
	emailField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='email']`)
	if err != nil {
		t.Fatalf("Could not find email input: %v", err)
	}
	emailField.Clear()
	emailField.SendKeys(validEmail)

	// Fill password
	passwordField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='password']`)
	if err != nil {
		t.Fatalf("Could not find password input: %v", err)
	}
	passwordField.Clear()
	passwordField.SendKeys(validPassword)

	// Click login
	loginBtn, err := wd.FindElement(selenium.ByXPATH, `//input[@type='submit' and @value='Login']`)
	if err != nil {
		t.Fatalf("Could not find the login button: %v", err)
	}
	loginBtn.Click()

	// Wait
	time.Sleep(2 * time.Second)

	// Validate success: look for "Logout" link
	logoutLink, err := wd.FindElement(selenium.ByLinkText, "Logout")
	if err != nil {
		t.Fatalf("Login might have failed; no 'Logout' link found: %v", err)
	}

	// Check URL
	currentURL, _ := wd.CurrentURL()
	if !strings.Contains(currentURL, baseURL+"/") {
		t.Errorf("Expected home page after login, got: %s", currentURL)
	}

	// Check for token cookie
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

	// Logout
	if err := logoutLink.Click(); err != nil {
		t.Fatalf("Failed to click logout link: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Validate we see "Login" link after logout
	_, err = wd.FindElement(selenium.ByLinkText, "Login")
	if err != nil {
		t.Errorf("Expected 'Login' link after logout, got error: %v", err)
	} else {
		t.Log("Logout successful: Found 'Login' link post-logout.")
	}
}
