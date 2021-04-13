package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/noelruault/auction-bid-tracker/internal/models"
	"github.com/noelruault/auction-bid-tracker/internal/web"
)

func (app *App) Health(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var health struct {
		Status string `json:"status"`
	}
	health.Status = "ok"
	web.Respond(ctx, w, health, http.StatusOK)
}

func (app *App) ListItems(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	items := app.Api.itemsvc.ListItems()

	web.Respond(ctx, w, items, http.StatusOK)
}

func (app *App) CreateItem(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var ni models.Item

	if err := web.Decode(r, &ni); err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}
	app.Api.itemsvc.TxCreate(&ni)

	web.Respond(ctx, w, ni, http.StatusCreated)
}

func (app *App) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	users := app.Api.usersvc.ListUsers()

	web.Respond(ctx, w, users, http.StatusOK)
}

func (app *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var nu models.User

	if err := web.Decode(r, &nu); err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}
	app.Api.usersvc.TxCreate(&nu)

	web.Respond(ctx, w, nu, http.StatusCreated)
}

// ListBidsByItemID retrieves all the bids for a given item ID
func (app *App) ListBidsByItemID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)

	itemID, ok := vars["itemId"]
	if !ok {
		app.Api.viewErr.JSON(ctx, w, models.ValidationError{"itemId": models.ErrRequired})
		return
	}

	i, _ := strconv.ParseInt(itemID, 10, 64)

	var err error
	var bids []models.Bid // verbose declaration to let you see in a glance that we are using a list here
	bids, err = app.Api.bidsvc.ListBidsByItemID(i)
	if err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}

	web.Respond(ctx, w, bids, http.StatusCreated)
}

// GetWinningBid gets the winning bid (highest current bid) for a given item ID
func (app *App) GetWinningBid(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)

	itemID, ok := vars["itemId"]
	if !ok {
		app.Api.viewErr.JSON(ctx, w, models.ValidationError{"itemId": models.ErrRequired})
		return
	}

	i, _ := strconv.ParseInt(itemID, 10, 64)

	bid, err := app.Api.bidsvc.GetWinningBid(i)
	if err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}

	web.Respond(ctx, w, bid, http.StatusOK)
}

// ListBetItemsByUserID fetches all the items on which the user has a bid
func (app *App) ListBetItemsByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)

	userID, ok := vars["userId"]
	if !ok {
		app.Api.viewErr.JSON(ctx, w, models.ValidationError{"userId": models.ErrRequired})
		return
	}

	u, _ := strconv.ParseInt(userID, 10, 64)

	bids, err := app.Api.bidsvc.ListBidsByUserID(u)
	if err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}

	keys := make(map[int64]bool)
	var itemIDs []int64
	for _, v := range bids {
		if _, ok := keys[v.ItemID]; !ok {
			keys[v.ItemID] = true
			itemIDs = append(itemIDs, v.ItemID)
		}
	}

	items, err := app.Api.itemsvc.ListItemsByIDs(itemIDs...)
	if err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}

	web.Respond(ctx, w, items, http.StatusOK)
}

// CreateBid allows to bid. An item ID and user ID must be provided in URL path
func (app *App) CreateBid(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)

	itemID, ok := vars["itemId"]
	if !ok {
		app.Api.viewErr.JSON(ctx, w, models.ValidationError{"itemId": models.ErrRequired})
		return
	}

	userID, ok := vars["userId"]
	if !ok {
		app.Api.viewErr.JSON(ctx, w, models.ValidationError{"userId": models.ErrRequired})
		return
	}

	var nb models.Bid
	if err := web.Decode(r, &nb); err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}

	i, _ := strconv.ParseInt(itemID, 10, 64)
	u, _ := strconv.ParseInt(userID, 10, 64)

	bid := models.Bid{
		ItemID: i,
		UserID: u,
		Amount: nb.Amount,
	}
	err := app.Api.bidsvc.TxCreate(&bid)
	if err != nil {
		app.Api.viewErr.JSON(ctx, w, err)
		return
	}

	web.Respond(ctx, w, bid, http.StatusCreated)
}
