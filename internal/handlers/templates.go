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
	Form            interface{}
	FormErrors      map[string]string
	Categories      []*models.Categories
	Category        *models.Categories
	Posts           []*models.Posts
	LikedPosts      []*models.Posts
	IsAuthenticated bool
	Post            *models.Posts
	Comments        []*models.CommentReaction
	CommentsNum     int
	User            *models.User
	PostReaction    *models.PostReaction
	CurrentPage     int
	TotalPages      int
	CurrentCatID    int
	VisiblePages    []int
	PageSize        int

	// ERROR FIELDS:
	ErrorCode int
	ErrorMsg  string
}

var functions = template.FuncMap{
	"humanDate": humanDate,
	"add":       add,
	"sub":       sub,
	"slice":     slice,
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
