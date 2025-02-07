package forum_advanced_test

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

const (
	driverPath   = "../geckodriver"
	seleniumPort = 9515
	baseURL      = "https://localhost:8433"
)

func setupWebDriver(t *testing.T) selenium.WebDriver {
	service, err := selenium.NewGeckoDriverService(driverPath, seleniumPort)
	if err != nil {
		t.Fatalf("Cannot start geckodriver service: %v", err)
	}
	t.Cleanup(func() {
		service.Stop()
	})

	caps := selenium.Capabilities{
		"browserName":         "firefox",
		"acceptInsecureCerts": true,
		"moz:firefoxOptions":  map[string]interface{}{},
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", seleniumPort))
	if err != nil {
		t.Fatalf("Error creating WebDriver session: %v", err)
	}

	t.Cleanup(func() {
		wd.Quit()
	})

	return wd
}

func TestAdvancedSelenium(t *testing.T) {
	wd := setupWebDriver(t)

	if err := wd.SetImplicitWaitTimeout(5 * time.Second); err != nil {
		t.Fatalf("Failed to set implicit wait: %v", err)
	}

	if err := wd.Get(baseURL + "/login"); err != nil {
		t.Fatalf("Failed to load /login page: %v", err)
	}

	err := wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		_, e := wd.FindElement(selenium.ByCSSSelector, `input[name="email"]`)
		return e == nil, nil
	}, 10*time.Second, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Explicit wait: email input not found: %v", err)
	}

	err = wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		_, findErr := wd.FindElement(selenium.ByCSSSelector, `input[name="password"]`)
		return findErr == nil, nil
	}, 15*time.Second, 1*time.Second)
	if err != nil {
		t.Fatalf("Fluent wait: password input not found: %v", err)
	}

	t.Log("Waits (implicit, explicit, fluent) validated successfully.")

	emailElem, _ := wd.FindElement(selenium.ByCSSSelector, `input[name="email"]`)
	emailElem.Clear()
	emailElem.SendKeys("admin@gmail.com")
	passwordElem, _ := wd.FindElement(selenium.ByCSSSelector, `input[name="password"]`)
	passwordElem.Clear()
	passwordElem.SendKeys("12345678")

	loginBtn, _ := wd.FindElement(selenium.ByXPATH, `//input[@type="submit" and @value="Login"]`)
	loginBtn.Click()

	time.Sleep(2 * time.Second)

	navCreate, err := wd.FindElement(selenium.ByCSSSelector, `nav a[href="/post/create"]`)
	if err != nil {
		t.Log("Could not find 'Create' link, skipping hover example.")
	} else {
		if moveErr := navCreate.MoveTo(0, 0); moveErr != nil {
			t.Errorf("Failed to hover over 'Create' link: %v", moveErr)
		} else {
			t.Log("Successfully hovered over 'Create' link (Action Class).")
		}
	}

	if err := wd.Get(baseURL + "/post/create"); err != nil {
		t.Fatalf("Failed to load /post/create page: %v", err)
	}
	time.Sleep(1 * time.Second)

	categorySelect, err := wd.FindElement(selenium.ByCSSSelector, `select[name="category_id"]`)
	if err != nil {
		t.Log("No category select found. Possibly no categories exist.")
	} else {
		if sendErr := categorySelect.SendKeys("1"); sendErr != nil {
			t.Errorf("Failed to select category '1': %v", sendErr)
		} else {
			t.Log("Select Class: Chose category with visible text or value '1'.")
		}
	}

	t.Log("Test complete. Check the terminal output (or test-report.txt) for results.")
}

func TestLoginLogout(t *testing.T) {
	wd := setupWebDriver(t)

	loginURL := baseURL + "/login"
	if err := wd.Get(loginURL); err != nil {
		t.Fatalf("Failed to load login page %s: %v", loginURL, err)
	}

	time.Sleep(1 * time.Second)

	validEmail := "test@example.com"
	validPassword := "Secret1234"

	emailField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='email']`)
	if err != nil {
		t.Fatalf("Could not find email input: %v", err)
	}
	emailField.Clear()
	emailField.SendKeys(validEmail)

	passwordField, err := wd.FindElement(selenium.ByCSSSelector, `input[name='password']`)
	if err != nil {
		t.Fatalf("Could not find password input: %v", err)
	}
	passwordField.Clear()
	passwordField.SendKeys(validPassword)

	loginBtn, err := wd.FindElement(selenium.ByXPATH, `//input[@type='submit' and @value='Login']`)
	if err != nil {
		t.Fatalf("Could not find the login button: %v", err)
	}
	loginBtn.Click()

	time.Sleep(2 * time.Second)

	logoutLink, err := wd.FindElement(selenium.ByLinkText, "Logout")
	if err != nil {
		t.Fatalf("Login might have failed; no 'Logout' link found: %v", err)
	}

	currentURL, _ := wd.CurrentURL()
	if !strings.Contains(currentURL, baseURL+"/") {
		t.Errorf("Expected home page after login, got: %s", currentURL)
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
		t.Fatalf("Failed to click logout link: %v", err)
	}
	time.Sleep(2 * time.Second)

	_, err = wd.FindElement(selenium.ByLinkText, "Login")
	if err != nil {
		t.Errorf("Expected 'Login' link after logout, got error: %v", err)
	} else {
		t.Log("Logout successful: Found 'Login' link post-logout.")
	}
}

type UserData struct {
	Email    string
	Username string
	Password string
}

func readUsersFromExcel(filename string) ([]UserData, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Users")
	if err != nil {
		return nil, err
	}

	var users []UserData
	for i, row := range rows {
		if i == 0 { // skip header row
			continue
		}
		if len(row) < 3 {
			continue
		}
		user := UserData{
			Email:    row[0],
			Username: row[1],
			Password: row[2],
		}
		users = append(users, user)
	}
	return users, nil
}
