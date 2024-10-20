package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"html/template"
	"log/slog"
)

type Application struct {
	Logger        *slog.Logger
	TemplateCache map[string]*template.Template
	Users         models.UserModelInterface
	Session       models.SessionModelInterface
}

func NewApp(logger *slog.Logger, teamplateCache map[string]*template.Template, users *models.UserModel, session *models.SessionModel) *Application {
	app := &Application{
		Logger:        logger,
		TemplateCache: teamplateCache,
		Users:         users,
		Session:       session,
	}

	return app
}
