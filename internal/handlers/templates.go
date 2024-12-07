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
	IsAuthenticated bool
	Post            *models.Posts
	Comments        []*models.Comments
	CommentsNum     int
	User 			*models.User
}

var functions = template.FuncMap{
	"humanDate": humanDate,
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
