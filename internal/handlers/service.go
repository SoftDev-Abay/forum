package handlers

import (
	"database/sql"
	"html/template"
	"log/slog"
)

type Application struct {
	Logger        *slog.Logger
	TemplateCache map[string]*template.Template
	db            *sql.DB
}

func NewApp(logger *slog.Logger, teamplateCache map[string]*template.Template, db *sql.DB) *Application {
	app := &Application{
		Logger:        logger,
		TemplateCache: teamplateCache,
		db:            db,
	}

	return app
}
