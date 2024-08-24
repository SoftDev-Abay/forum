package handlers

import (
	"forum/app"
)

type Router struct {
	app *app.Application
}



func NewRouter(app *app.Application) *Router {
	

	router := &Router{app}

	return router


}