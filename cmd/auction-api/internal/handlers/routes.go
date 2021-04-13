package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	Api    API
}

func (app *App) SetupRouter() {
	app.Router.
		Methods(http.MethodGet).
		Path("/").
		HandlerFunc(app.Health)

	// BID
	app.Router.
		Methods(http.MethodPost).
		Path("/users/{userId}/items/{itemId}/bids/").
		HandlerFunc(app.CreateBid)

	app.Router.
		Methods(http.MethodGet).
		Path("/users/{userId}/bids/items/").
		HandlerFunc(app.ListBetItemsByUserID)

	app.Router.
		Methods(http.MethodGet).
		Path("/items/{itemId}/bids/highest/").
		HandlerFunc(app.GetWinningBid)

	app.Router.
		Methods(http.MethodGet).
		Path("/items/{itemId}/bids/").
		HandlerFunc(app.ListBidsByItemID)

	// Items
	app.Router.
		Methods(http.MethodGet).
		Path("/items/").
		HandlerFunc(app.ListItems)

	app.Router.
		Methods(http.MethodPost).
		Path("/items/").
		HandlerFunc(app.CreateItem)

	// Users
	app.Router.
		Methods(http.MethodGet).
		Path("/users/").
		HandlerFunc(app.ListUsers)

	app.Router.
		Methods(http.MethodPost).
		Path("/users/").
		HandlerFunc(app.CreateUser)
}
