package forum_advanced_test

import (
	"github.com/tebeka/selenium"
	"testing"
	"time"
)

func TestLoginViaBrowserStack(t *testing.T) {
	// 1) read Excel
	users, err := readUsersFromExcel("users.xlsx")
	if err != nil {
		t.Fatalf("failed to read excel: %v", err)
	}
	testUser := users[0]

	// 2) define remote capabilities
	caps := selenium.Capabilities{
		"browserName":         "chrome",
		"browserVersion":      "latest",
		"acceptInsecureCerts": true,
		"bstack:options": map[string]interface{}{
			"os":        "Windows",
			"osVersion": "10",
			"userName":  "shiradi_yvD1hr",
			"accessKey": "j8qspEsAK2CJdbeaNPjb",
			"local":     "false",
		},
	}
	remoteURL := "https://hub.browserstack.com/wd/hub"

	// 3) connect
	wd, err := selenium.NewRemote(caps, remoteURL)
	if err != nil {
		t.Fatalf("failed to start remote driver: %v", err)
	}
	defer wd.Quit()

	// 4) open forum site (public or local with tunnel)
	if err := wd.Get("https://188.227.35.5/login"); err != nil {
		t.Fatalf("could not load login page: %v", err)
	}

	// 5) fill form from Excel data
	emailElem, err := wd.FindElement(selenium.ByCSSSelector, `input[name="email"]`)
	if err != nil {
		t.Fatalf("no email input: %v", err)
	}
	emailElem.SendKeys(testUser.Email)

	passElem, err := wd.FindElement(selenium.ByCSSSelector, `input[name="password"]`)
	if err != nil {
		t.Fatalf("no password input: %v", err)
	}
	passElem.SendKeys(testUser.Password)

	submit, err := wd.FindElement(selenium.ByCSSSelector, `input[type="submit"]`)
	if err != nil {
		t.Fatalf("no submit button: %v", err)
	}
	submit.Click()

	// 6) check result
	time.Sleep(3 * time.Second)
	_, err = wd.FindElement(selenium.ByLinkText, "Logout")
	if err != nil {
		t.Errorf("Login may have failed, no logout link: %v", err)
	}
}
