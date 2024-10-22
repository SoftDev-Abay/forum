package main

import (
	"context"
	"database/sql"
	"flag"
	"game-forum-abaliyev-ashirbay/internal/handlers"
	"game-forum-abaliyev-ashirbay/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dbPath := flag.String("db", "./data/app.db", "Path to SQLite database file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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
	postsModel := &models.PostModel{DB: db}

	app := handlers.NewApp(logger, templateCache, categoriesModel, postsModel, users, session)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		Handler:  app.Routes(),
	}

	logger.Info("Starting server", "addr", srv.Addr)

	// Обработка сигналов системы для корректного завершения к бд
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

	err = srv.ListenAndServe()
	logger.Error(err.Error())
}
