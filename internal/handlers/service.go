package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"html/template"
	"log/slog"
)

type Application struct {
	Logger        *slog.Logger
	TemplateCache map[string]*template.Template
	Categories    models.CategoriesModelInterface
	Posts         models.PostsModelInterface
}

func NewApp(logger *slog.Logger, teamplateCache map[string]*template.Template, categories *models.CategoriesModel, posts *models.PostModel) *Application {
	app := &Application{
		Logger:        logger,
		TemplateCache: teamplateCache,
		Categories:    categories,
		Posts:         posts,
	}

	return app
}
