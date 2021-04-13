package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/noelruault/auction-bid-tracker/cmd/auction-api/internal/handlers"
	"github.com/noelruault/auction-bid-tracker/internal/models"
)

func main() {
	if err := run(); err != nil {
		log.Println("shutting down", "error:", err)
		os.Exit(1)
	}
}

func run() error {
	log := log.New(os.Stdout, "AUCTION-BID-TRACKER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	log.Printf("main : Started")
	defer log.Println("main : Completed")

	database := models.CreateDatabase()
	app := &handlers.App{
		Router: mux.NewRouter().StrictSlash(true),
		Api:    handlers.NewAPI(database, log),
	}

	app.SetupRouter()

	return http.ListenAndServe(":8080", app.Router)
}
