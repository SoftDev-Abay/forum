package handlers

import (
	"net/http"
)

func (router *Router) Routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("/", router.home)
	mux.HandleFunc("/post/view", router.postView)
	mux.HandleFunc("/post/create", router.postCreate)


	return mux
}
