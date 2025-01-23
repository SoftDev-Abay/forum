package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"game-forum-abaliyev-ashirbay/internal/handlers"
	"game-forum-abaliyev-ashirbay/internal/models"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	addr := flag.String("addr", ":8433", "HTTP network address")
	dbPath := flag.String("db", "./data/app.db", "Path to SQLite database file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := loadEnvFile(".env")
	if err != nil {
		logger.Error("could not load .env file: %v\n", err)
		os.Exit(1)
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		logger.Error("Unable to connect to the database: " + err.Error())
		os.Exit(1)
	}

	templateCache, err := handlers.NewTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	users := &models.UserModel{DB: db}
	session := &models.SessionModel{DB: db}
	categoriesModel := &models.CategoriesModel{DB: db}
	postRecactions := &models.PostReactionsModel{DB: db}
	commentReactions := &models.CommentsReactionsModel{DB: db}
	postsModel := &models.PostModel{DB: db, PostReactionsModel: postRecactions}
	commentsModel := &models.CommentsModel{DB: db}
	promotionRequestsModel := &models.PromotionRequestsModel{DB: db}
	reports := &models.ReportsModel{DB: db}
	reportReasons := &models.ReportReasonsModel{DB: db}
	notifications := &models.NotificationsModel{DB: db}

	app := handlers.NewApp(
		addr,
		logger,
		templateCache,
		categoriesModel,
		postsModel,
		users,
		session,
		postRecactions,
		commentsModel,
		commentReactions,
		promotionRequestsModel,
		reports,
		reportReasons,

		// authentication
		googleClientID,
		googleClientSecret,
		githubClientID,
		githubClientSecret,

		// notifications
		notifications,
	)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		Handler:  app.Routes(),
	}

	logger.Info("Starting server", "addr", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Server forced to shutdown: " + err.Error())
		}

		db.Close()

		os.Exit(0)
	}()

	err = srv.ListenAndServeTLS("app/certs/server.crt", "app/certs/server.key")
	logger.Error(err.Error())
}

func loadEnvFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Trim whitespace
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			// Skip blank lines or comments
			continue
		}

		// Split on the first '=' only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Not a valid KEY=VALUE line; skip or return an error
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Set it as an environment variable
		os.Setenv(key, val)
	}

	return scanner.Err()
}
