package handlers

import (
	"game-forum-abaliyev-ashirbay/internal/models"
	"game-forum-abaliyev-ashirbay/ui"
	"html/template"
	"io/fs"
	"path/filepath"
	"time"
)

type templateData struct {
	Form                interface{}
	FormErrors          map[string]string
	Categories          []*models.Categories
	Category            *models.Categories
	Posts               []*models.Post
	LikedPosts          []*models.Post
	IsAuthenticated     bool
	Post                *models.Post
	Comments            []*models.CommentReaction
	CommentsNum         int
	User                *models.User
	PostReaction        *models.PostReaction
	CurrentPage         int
	TotalPages          int
	CurrentCatID        int
	VisiblePages        []int
	PageSize            int
	PromotionRequests   []*models.PromotionRequests
	PromotionRequest    *models.PromotionRequests
	Users               []*models.User
	PostsByUser         []*models.PostByUser
	PostByUser          *models.PostByUser
	ReportReasons       []*models.ReportReasons
	Reports             []*models.Reports 
	CommentPostAddition []*models.CommentPostAddition

	// ERROR FIELDS:
	ErrorCode int
	ErrorMsg  string

	// advanced features
	NotificationsCount int
	Notifications      []*models.Notifications
	UserNotifications  []NotificationView
}

type NotificationView struct {
	ID            int
	Type          string
	ActorUsername string
	PostID        int
	CommentText   string
	CreatedAt     string
	IsRead        bool
}

var functions = template.FuncMap{
	"humanDate": humanDate,
	"add":       add,
	"sub":       sub,
	"slice":     slice,
	"or": func(a, b bool) bool {
		return a || b
	},
}

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.html",
			"html/error_base.html",
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func slice(nums ...int) []int {
	return nums
}
