package handlers_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"game-forum-abaliyev-ashirbay/internal/handlers"
	"game-forum-abaliyev-ashirbay/internal/models"
	_ "github.com/mattn/go-sqlite3" // for SQLite
)

func TestRegisterLogin(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Register & Login Suite")
}

var _ = ginkgo.Describe("Register and Login Handlers", func() {
	var (
		db  *sql.DB
		app *handlers.Application
		// Possibly store test server if you want to do full integration
	)

	// Create an in-memory SQLite DB and run minimal migrations
	ginkgo.BeforeEach(func() {
		var err error
		db, err = sql.Open("sqlite3", "file::memory:?cache=shared")
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		// In a real scenario, run your entire migrations or at least the needed tables:
		err = createTestSchema(db)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		// Initialize your models
		userModel := &models.UserModel{DB: db}
		sessionModel := &models.SessionModel{DB: db}

		// Create your Application
		// If needed, pass logger, templateCache, etc.
		// For now we just fill in these 2
		app = &handlers.Application{
			Users:   userModel,
			Session: sessionModel,
		}
	})

	// Cleanup after each test
	ginkgo.AfterEach(func() {
		// close DB
		db.Close()
	})

	ginkgo.Describe("POST /register", func() {
		ginkgo.It("should fail if form data is invalid", func() {
			formData := "email=&username=john&password=abc&confirmPassword=abc"
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()

			// call your registerPost
			app.RegisterPost(rr, req)

			// Expect Unprocessable Entity
			gomega.Expect(rr.Code).To(gomega.Equal(http.StatusUnprocessableEntity))
			gomega.Expect(rr.Body.String()).To(gomega.ContainSubstring("Email cannot be blank"))
		})

		ginkgo.It("should insert user and redirect if valid", func() {
			formData := "email=john@example.com&username=johnny&password=abcdefgh&confirmPassword=abcdefgh"
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()

			app.RegisterPost(rr, req)

			// Should see a 303 or 302 redirect to /login
			gomega.Expect(rr.Code).To(gomega.BeNumerically("==", http.StatusSeeOther))
			gomega.Expect(rr.Header().Get("Location")).To(gomega.Equal("/login"))

			// Now check DB to ensure user was inserted
			row := db.QueryRow("SELECT email, username FROM users WHERE email = ?", "john@example.com")
			var gotEmail, gotUsername string
			err := row.Scan(&gotEmail, &gotUsername)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(gotEmail).To(gomega.Equal("john@example.com"))
			gomega.Expect(gotUsername).To(gomega.Equal("johnny"))
		})
	})

	ginkgo.Describe("POST /login", func() {
		ginkgo.BeforeEach(func() {
			// Insert a user to login with
			_, err := db.Exec(`INSERT INTO users (email, username, password, enabled) 
			                   VALUES (?, ?, ?, 1)`,
				"valid@example.com", "validuser", // hashed password must match compareHashPassword
				// For simplicity, store a known "hash" or skip hashing in test
				// We'll store plain "abcdefgh" if your compareHashPassword just compares strings
				"abcdefgh",
			)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.It("should fail for blank form", func() {
			formData := "email=&password="
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			app.LoginPost(rr, req)

			gomega.Expect(rr.Code).To(gomega.Equal(http.StatusUnprocessableEntity))
			gomega.Expect(rr.Body.String()).To(gomega.ContainSubstring("Email cannot be blank"))
		})

		ginkgo.It("should fail for incorrect credentials", func() {
			// wrong password
			formData := "email=valid@example.com&password=wrongpass"
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			app.LoginPost(rr, req)

			gomega.Expect(rr.Code).To(gomega.Equal(http.StatusUnprocessableEntity))
			gomega.Expect(rr.Body.String()).To(gomega.ContainSubstring("Incorrect email or password"))
		})

		ginkgo.It("should login with correct credentials", func() {
			formData := "email=valid@example.com&password=abcdefgh"
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			app.LoginPost(rr, req)

			gomega.Expect(rr.Code).To(gomega.Equal(http.StatusSeeOther))
			gomega.Expect(rr.Header().Get("Location")).To(gomega.Equal("/"))

			// check for a "token" cookie
			cookies := rr.Result().Cookies()
			var tokenCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "token" {
					tokenCookie = c
					break
				}
			}
			gomega.Expect(tokenCookie).ToNot(gomega.BeNil())
			gomega.Expect(tokenCookie.Value).ToNot(gomega.BeEmpty())

			// also check session is inserted in DB
			row := db.QueryRow("SELECT token FROM sessions WHERE user_id = (SELECT id FROM users WHERE email=?)", "valid@example.com")
			var dbToken string
			err := row.Scan(&dbToken)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(dbToken).To(gomega.Equal(tokenCookie.Value))
		})
	})

})

func createTestSchema(db *sql.DB) error {
	// Minimal for login & register:
	// create Users table
	userTable := `
	CREATE TABLE Users (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(100) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		role VARCHAR(20) NOT NULL DEFAULT 'user',
		enabled BOOLEAN NOT NULL DEFAULT 1
	);
    `
	_, err := db.Exec(userTable)
	if err != nil {
		return fmt.Errorf("creating Users table: %w", err)
	}

	// create Sessions table
	sessionsTable := `
	CREATE TABLE Sessions (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		token TEXT NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		createdAt DATETIME NOT NULL,
		expiresAt DATETIME NOT NULL
	);
    `
	_, err = db.Exec(sessionsTable)
	if err != nil {
		return fmt.Errorf("creating Sessions table: %w", err)
	}

	return nil
}
