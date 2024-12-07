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
	Categories    models.CategoriesModelInterface
	Posts         models.PostsModelInterface
	PostReactions models.PostReactionModelInterface
	Comments      models.CommentsModelInterface
}

func NewApp(logger *slog.Logger, teamplateCache map[string]*template.Template, categories *models.CategoriesModel, posts *models.PostModel, users *models.UserModel, session *models.SessionModel, postReactions *models.PostReactionsModel, comments *models.CommentsModel) *Application {
	app := &Application{
		Logger:        logger,
		TemplateCache: teamplateCache,
		Categories:    categories,
		Posts:         posts,
		Users:         users,
		Session:       session,
		PostReactions: postReactions,
		Comments:      comments,
	}
	return app
}
