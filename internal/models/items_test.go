package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testItemDB struct {
	ItemDB
	txCreate       func(*Item)
	get            func(int64) (Item, error)
	listItemsByIDs func(...int64) ([]Item, error)
}

func (t *testItemDB) TxCreate(i *Item) {
	if t.txCreate != nil {
		t.txCreate(i)
	}
}

func (t *testItemDB) Get(itemID int64) (Item, error) {
	if t.get != nil {
		return t.get(itemID)
	}
	return Item{}, nil
}

func (t *testItemDB) ListItemsByIDs(itemIDs ...int64) ([]Item, error) {
	if t.listItemsByIDs != nil {
		return t.listItemsByIDs(itemIDs...)
	}
	return nil, nil
}

func TestItemService_TxCreate(t *testing.T) {
	tudb := &testItemDB{}

	db := CreateDatabase()
	usvc := NewItemService(db, nil)

	usvc.(itemService).ItemService.(*itemCapsule).ItemDB = tudb

	var cases = []struct {
		name    string
		outitem map[int64]Item
		setup   func(*testing.T)
	}{
		{
			"ok",
			map[int64]Item{
				1: {ID: 1, Name: "test", Value: 10},
			},
			func(t *testing.T) {
				tudb.txCreate = func(u *Item) {
					db.items.Create(u)
				}
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			for _, v := range tt.outitem {
				usvc.TxCreate(&v)
			}

			assert.Equal(t, tt.outitem, db.items.data)

			*tudb = testItemDB{}
		})
	}
}
