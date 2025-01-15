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

	// Serve /imgs/<filename> from ./data/imgs
	mux.Handle("/imgs/", http.StripPrefix("/imgs/", http.FileServer(http.Dir("./data/imgs"))))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/post/view", app.postView)
	mux.Handle("/post/reaction", app.loginMiddware(http.HandlerFunc(app.handlePostReaction)))

	mux.Handle("/post/create/post", app.loginMiddware(http.HandlerFunc(app.postCreatePost)))
	mux.Handle("/post/create", app.loginMiddware(http.HandlerFunc(app.postCreate)))
	mux.Handle("/post/delete", app.loginMiddware(http.HandlerFunc(app.postDelete), "moderator", "admin"))
	// Report system routes
	mux.Handle("/post/report", app.loginMiddware(http.HandlerFunc(app.ReportPost), "moderator", "admin"))
	mux.Handle("/post/report/list", app.loginMiddware(http.HandlerFunc(app.adminReportList), "admin"))

	mux.Handle("/comments/create", app.loginMiddware(http.HandlerFunc(app.createCommentPost)))
	mux.Handle("/comments/reaction", app.loginMiddware(http.HandlerFunc(app.handleCommentReaction)))
	mux.Handle("/comments/delete", app.loginMiddware(http.HandlerFunc(app.commentDelete), "moderator", "admin"))

	mux.Handle("/personal-page", app.loginMiddware(http.HandlerFunc(app.personalPage)))

	mux.HandleFunc("/register", app.register)
	mux.HandleFunc("/register/post", app.registerPost)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/login/post", app.loginPost)
	mux.HandleFunc("/logout", app.logout)

	// Promotion request routes
	mux.Handle("/promotion_requests", app.loginMiddware(http.HandlerFunc(app.getAllPromotionRequests)))
	mux.Handle("/promotion_requests/view", app.loginMiddware(http.HandlerFunc(app.getPromotionRequest)))
	mux.Handle("/promotion_requests/create", app.loginMiddware(http.HandlerFunc(app.promotionRequestCreate)))
	mux.Handle("/promotion_requests/create/post", app.loginMiddware(http.HandlerFunc(app.promotionRequestCreatePost)))
	mux.Handle("/promotion_requests/change_status", app.loginMiddware(http.HandlerFunc(app.changePromotionRequestStatus), "admin"))

	return mux
}
