package main

import (
	"flag"
	"game-forum-abaliyev-ashirbay/internal/handlers"
	"log/slog"
	"net/http"
	"os"
	// New import
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	templateCache, err := handlers.NewTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := handlers.NewApp(logger, templateCache)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		Handler:  app.Routes(),
	}

	logger.Info("Starting server on %s", *addr)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
}
