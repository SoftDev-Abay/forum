package main

import (
	"flag"
	"forum/app"
	"forum/internal/handlers"
	"log"
	"net/http"

	"os" // New import
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := app.NewApp(errorLog, infoLog)

	router := handlers.NewRouter(app)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  router.Routes(),
	}
	infoLog.Printf("Starting server on %s", *addr)
	// Call the ListenAndServe() method on our new http.Server struct.
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
