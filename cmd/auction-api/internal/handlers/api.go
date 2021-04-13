package handlers

import (
	"log"

	"github.com/noelruault/auction-bid-tracker/internal/models"
	"github.com/noelruault/auction-bid-tracker/internal/views"
)

type API struct {
	bidsvc  models.BidService
	itemsvc models.ItemService
	usersvc models.UserService

	viewErr views.Error
	log     *log.Logger
}

func NewAPI(db *models.DB, log *log.Logger) API {
	us := models.NewUserService(db)
	is := models.NewItemService(db, us)
	bs := models.NewBidService(db, is, us)

	return API{
		bidsvc:  bs,
		itemsvc: is,
		usersvc: us,
		log:     log,
	}
}
