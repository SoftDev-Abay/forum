package handlers

import (
	"math"
	"net/http"
	"strconv"
)

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	categoryStr := r.URL.Query().Get("category")
	categoryID, err := strconv.Atoi(categoryStr)
	if err != nil {
		categoryID = 0
	}

	userID, _ := app.getAuthenticatedUserID(r)

	posts, err := app.Posts.GetFilteredPosts(userID, categoryID, page, pageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	totalPosts, err := app.Posts.CountPosts(categoryID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	totalPages := int(math.Ceil(float64(totalPosts) / float64(pageSize)))

	startPage := page - 3
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 6
	if endPage > totalPages {
		endPage = totalPages
	}
	visiblePages := make([]int, 0, endPage-startPage+1)
	for i := startPage; i <= endPage; i++ {
		visiblePages = append(visiblePages, i)
	}

	categories, err := app.Categories.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		PostsByUser:  posts,
		Categories:   categories,
		CurrentPage:  page,
		TotalPages:   totalPages,
		CurrentCatID: categoryID,
		VisiblePages: visiblePages,
		PageSize:     pageSize,
	}

	app.render(w, r, http.StatusOK, "home.html", data)
}
