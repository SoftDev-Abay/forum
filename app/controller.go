package app

import (
	"log"
)

type Application struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

func NewApp(errorLog *log.Logger, infoLog *log.Logger) *Application {
	app := &Application{
		ErrorLog: errorLog,
		InfoLog:  infoLog,
	}

	return app
}
