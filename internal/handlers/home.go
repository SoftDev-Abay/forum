package handlers

import (
	"html/template" // New import
	"net/http"
)

func (router *Router) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		router.notFound(w, r)
		return
	}
	// Include the navigation partial in the template files.
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/home.tmpl",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		router.serverError(w, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		router.serverError(w, err)
	}
}
