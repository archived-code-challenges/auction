package models

type ItemService interface {
	ItemDB
}

type ItemDB interface {
	TxCreate(*Item)
	Get(int64) (Item, error)
	ListItems() []Item
	ListItemsByIDs(...int64) ([]Item, error)
}

type Item struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"initialValue"`
}

// itemService wraps the ItemService interface to allow mocking by interfaces
type itemService struct {
	ItemService
	userService UserService
}

type itemCapsule struct {
	ItemDB
}

func NewItemService(db *DB, usvc UserService) ItemService {
	return itemService{
		ItemService: &itemCapsule{
			ItemDB: &ItemStorage{
				db.items.mu,
				db.items.data,
				0,
			},
		},
		userService: usvc,
	}
}
