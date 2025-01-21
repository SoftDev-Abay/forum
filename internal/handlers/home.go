package handlers

import (
	"fmt"
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

	// 1) Parse the `page` query param (default = 1)
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// 2) Parse the `pageSize` query param (default = 10 posts per page)
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 3) Parse the `category` query param (default = 0 means "all categories")
	categoryStr := r.URL.Query().Get("category")
	categoryID, err := strconv.Atoi(categoryStr)
	if err != nil {
		categoryID = 0
	}

	// Check if user is authenticated
	userID, _ := app.getAuthenticatedUserID(r)

	fmt.Println("userID", userID)

	// 4) Get the posts for this page & category
	posts, err := app.Posts.GetFilteredPosts(userID, categoryID, page, pageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	fmt.Println("posts", posts)

	// 5) Count how many posts in total (for pagination)
	totalPosts, err := app.Posts.CountPosts(categoryID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 6) Calculate total pages and visible page range
	totalPages := int(math.Ceil(float64(totalPosts) / float64(pageSize)))

	// Determine the range of pages to show
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

	// 7) Also get all categories for a dropdown (optional)
	categories, err := app.Categories.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 8) Prepare your template data
	data := templateData{
		PostsByUser:  posts,
		Categories:   categories,
		CurrentPage:  page,
		TotalPages:   totalPages,
		CurrentCatID: categoryID,
		VisiblePages: visiblePages,
		PageSize:     pageSize, // Current page size
	}

	app.render(w, r, http.StatusOK, "home.html", data)
}

// pagination logic
// http request  -  page, limit
// to the sql request - offset = page * limit, limit = page * limit + 1
// http responce - page, limit, array of pages, posts

// for the category filter logic
// http request  -  categoryValue
// to the sql request - where category  = categoryValue
// http responce posts
