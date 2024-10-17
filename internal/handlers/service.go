package handlers

import (
	"html/template"
	"log/slog"
)

type Application struct {
	Logger        *slog.Logger
	TemplateCache map[string]*template.Template
}

func NewApp(logger *slog.Logger, teamplateCache map[string]*template.Template) *Application {
	app := &Application{
		Logger:        logger,
		TemplateCache: teamplateCache,
	}

	return app
}
