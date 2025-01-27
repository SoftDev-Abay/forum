package handlers

import (
	"game-forum-abaliyev-ashirbay/ui"
	"io/fs"
	"net/http"
	"os"
)

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	staticFiles, err := fs.Sub(ui.Files, "static")
	if err != nil {
		app.Logger.Error(err.Error())
		os.Exit(1)
	}

	fileServer := http.FileServer(http.FS(staticFiles))

	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.Handle("/imgs/", http.StripPrefix("/imgs/", http.FileServer(http.Dir("./data/imgs"))))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/post/view", app.postView)
	mux.Handle("/post/reaction", app.loginMiddware(http.HandlerFunc(app.handlePostReaction)))

	mux.Handle("/post/create/post", app.loginMiddware(http.HandlerFunc(app.postCreatePost)))
	mux.Handle("/post/create", app.loginMiddware(http.HandlerFunc(app.postCreate)))
	mux.Handle("/post/delete", app.loginMiddware(http.HandlerFunc(app.postDelete)))
	mux.Handle("/post/edit", app.loginMiddware(http.HandlerFunc(app.postEdit)))
	mux.Handle("/post/edit/post", app.loginMiddware(http.HandlerFunc(app.postEditPost)))

	// Report system routes
	mux.Handle("/post/report", app.loginMiddware(http.HandlerFunc(app.ReportPost), "moderator", "admin"))
	mux.Handle("/post/report/list", app.loginMiddware(http.HandlerFunc(app.adminReportList), "admin"))

	mux.Handle("/comments/create", app.loginMiddware(http.HandlerFunc(app.createCommentPost)))
	mux.Handle("/comments/reaction", app.loginMiddware(http.HandlerFunc(app.handleCommentReaction)))
	mux.Handle("/comments/delete", app.loginMiddware(http.HandlerFunc(app.commentDelete)))

	mux.Handle("/user/personal-page", app.loginMiddware(http.HandlerFunc(app.personalPage)))
	mux.Handle("/user/notifications", app.loginMiddware(http.HandlerFunc(app.notificationsPage)))

	mux.HandleFunc("/register", app.register)
	mux.HandleFunc("/register/post", app.RegisterPost)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/login/post", app.LoginPost)
	mux.HandleFunc("/logout", app.logout)
	// external authentication
	mux.HandleFunc("/auth/google", app.googleLogin)
	mux.HandleFunc("/auth/google/callback", app.googleCallback)
	mux.HandleFunc("/auth/github", app.githubLogin)
	mux.HandleFunc("/auth/github/callback", app.githubCallback)

	// Promotion request routes
	mux.Handle("/promotion_requests", app.loginMiddware(http.HandlerFunc(app.getAllPromotionRequests)))
	mux.Handle("/promotion_requests/view", app.loginMiddware(http.HandlerFunc(app.getPromotionRequest)))
	mux.Handle("/promotion_requests/create", app.loginMiddware(http.HandlerFunc(app.promotionRequestCreate)))
	mux.Handle("/promotion_requests/create/post", app.loginMiddware(http.HandlerFunc(app.promotionRequestCreatePost)))
	mux.Handle("/promotion_requests/change_status", app.loginMiddware(http.HandlerFunc(app.changePromotionRequestStatus), "admin"))

	mux.Handle("/admin/report/delete-post", app.loginMiddware(http.HandlerFunc(app.adminReportDeletePost), "admin"))
	mux.Handle("/admin/report/delete", app.loginMiddware(http.HandlerFunc(app.adminReportReject), "admin"))

	// Admin panel routes

	mux.Handle("/admin/users/change_role", app.loginMiddware(http.HandlerFunc(app.changeUserRole), "admin"))

	mux.Handle("/admin", app.loginMiddware(http.HandlerFunc(app.adminPanel), "admin"))
	mux.Handle("/admin/categories/create", app.loginMiddware(http.HandlerFunc(app.categoryCreate), "admin"))
	mux.Handle("/admin/categories/create/post", app.loginMiddware(http.HandlerFunc(app.categoryCreatePost), "admin"))
	mux.Handle("/admin/categories/delete", app.loginMiddware(http.HandlerFunc(app.DeleteCategory), "admin"))

	return app.rateLimitMiddleware(app.secureHeaders(mux))
}
