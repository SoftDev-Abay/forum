package app

import (
	"log/slog"
)

type Application struct {
	Logger *slog.Logger
}

func NewApp(logger *slog.Logger) *Application {
	app := &Application{
		Logger: logger,
	}

	return app
}
