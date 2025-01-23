package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"html/template"
	"log/slog"
)

const (
	comment = "comment"
	like    = "like"
	dislike = "dislike"
)

type Application struct {
	Addr              *string
	Logger            *slog.Logger
	TemplateCache     map[string]*template.Template
	Users             models.UserModelInterface
	Session           models.SessionModelInterface
	Categories        models.CategoriesModelInterface
	Posts             models.PostsModelInterface
	PostReactions     models.PostReactionModelInterface
	Comments          models.CommentsModelInterface
	CommentsReactions models.CommentsReactionsModelInterface
	PromotionRequests models.PromotionRequestsModelInterface
	Reports           models.ReportsModelInterface
	ReportReasons     models.ReportsReasonsModelInterface

	// authentication optional
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	// Notifications optional
	Notifications models.NotificationsModelInterface
}

func NewApp(
	addr *string,
	logger *slog.Logger,
	teamplateCache map[string]*template.Template,
	categories *models.CategoriesModel,
	posts *models.PostModel,
	users *models.UserModel,
	session *models.SessionModel,
	postReactions *models.PostReactionsModel,
	comments *models.CommentsModel,
	commentsReactions *models.CommentsReactionsModel,
	promotionRequests *models.PromotionRequestsModel,
	reports *models.ReportsModel, 
	reportReasons *models.ReportReasonsModel, 
	googleClientID string,
	googleClientSecret string,
	gitHubClientID string,
	gitHubClientSecret string,
	notifications *models.NotificationsModel,
) *Application {
	app := &Application{
		Addr:              addr,
		Logger:            logger,
		TemplateCache:     teamplateCache,
		Categories:        categories,
		Posts:             posts,
		Users:             users,
		Session:           session,
		PostReactions:     postReactions,
		Comments:          comments,
		CommentsReactions: commentsReactions,
		PromotionRequests: promotionRequests,
		Reports:           reports,       
		ReportReasons:     reportReasons, 

		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		GitHubClientID:     gitHubClientID,
		GitHubClientSecret: gitHubClientSecret,

		Notifications: notifications,
	}
	return app
}
