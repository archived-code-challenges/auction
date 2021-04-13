package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testBidDB struct {
	BidDB
	txCreate         func(*Bid) error
	getWinningBid    func(int64) (Bid, error)
	listItemBids     func(int64) ([]Bid, error)
	listBidsByUserID func(int64) ([]Bid, error)
}

func (t *testBidDB) TxCreate(b *Bid) error {
	if t.txCreate != nil {
		return t.txCreate(b)
	}

	return nil
}

func (t *testBidDB) ListBidsByItemID(itemID int64) ([]Bid, error) {
	if t.listItemBids != nil {
		return t.listItemBids(itemID)
	}

	return nil, nil
}

func (t *testBidDB) GetWinningBid(itemID int64) (Bid, error) {
	if t.getWinningBid != nil {
		return t.getWinningBid(itemID)
	}

	return Bid{}, nil
}

func (t *testBidDB) ListBidsByUserID(userID int64) ([]Bid, error) {
	if t.listBidsByUserID != nil {
		return t.listBidsByUserID(userID)
	}

	return nil, nil
}

func TestBidService_TxCreate(t *testing.T) {
	tudb := &testUserDB{}
	tidb := &testItemDB{}
	tbdb := &testBidDB{}

	db := CreateDatabase()
	usvc := NewUserService(db)
	isvc := NewItemService(db, usvc)
	bsvc := NewBidService(db, isvc, usvc)

	usvc.(userService).UserService.(*userCapsule).UserDB = tudb
	isvc.(itemService).ItemService.(*itemCapsule).ItemDB = tidb
	bsvc.(bidService).BidService.(*bidValidator).BidDB = tbdb

	var cases = []struct {
		name   string
		bid    *Bid
		outbid *Bid
		outerr error
		setup  func(*testing.T)
	}{
		{
			"ok",
			&Bid{ID: 0, UserID: 1, ItemID: 1, Amount: 1},
			&Bid{ID: 1, UserID: 1, ItemID: 1, Amount: 1},
			nil,
			func(t *testing.T) {
				tbdb.txCreate = func(b *Bid) error {
					b.ID = 1
					return nil
				}
				tidb.get = func(int64) (Item, error) {
					return Item{ID: 1, Name: "test", Value: 0}, nil
				}
				tudb.get = func(int64) (User, error) {
					return User{ID: 1, Name: "test"}, nil
				}
			},
		},
		{
			"winning_bid_not_found_but_ok",
			&Bid{ID: 0, UserID: 1, ItemID: 1, Amount: 1},
			&Bid{ID: 1, UserID: 1, ItemID: 1, Amount: 1},
			nil,
			func(t *testing.T) {
				tbdb.txCreate = func(b *Bid) error {
					b.ID = 1
					return nil
				}
				tidb.get = func(int64) (Item, error) {
					return Item{ID: 1, Name: "test", Value: 0}, nil
				}
				tudb.get = func(int64) (User, error) {
					return User{ID: 1, Name: "test"}, nil
				}
				tbdb.getWinningBid = func(int64) (Bid, error) {
					return Bid{}, ErrNotFound
				}
			},
		},
		{
			"item_not_found",
			&Bid{ID: 0, UserID: 1, ItemID: 1, Amount: 1},
			nil,
			ValidationError{"item": ErrNotFound},
			func(t *testing.T) {
				tbdb.txCreate = func(b *Bid) error {
					b.ID = 1
					return nil
				}
				tidb.get = func(int64) (Item, error) {
					return Item{}, ErrNotFound
				}
			},
		},
		{
			"user_not_found",
			&Bid{ID: 0, UserID: 1, ItemID: 1, Amount: 1},
			nil,
			ValidationError{"user": ErrNotFound},
			func(t *testing.T) {
				tbdb.txCreate = func(b *Bid) error {
					b.ID = 1
					return nil
				}
				tudb.get = func(int64) (User, error) {
					return User{}, ErrNotFound
				}
			},
		},
		{
			"winning_bid_higher_value",
			&Bid{ID: 0, UserID: 1, ItemID: 1, Amount: 1},
			nil,
			ValidationError{"bid": ErrLowValue},
			func(t *testing.T) {
				tbdb.txCreate = func(b *Bid) error {
					b.ID = 1
					return nil
				}
				tidb.get = func(int64) (Item, error) {
					return Item{ID: 1, Name: "test", Value: 0}, nil
				}
				tudb.get = func(int64) (User, error) {
					return User{ID: 1, Name: "test"}, nil
				}
				tbdb.getWinningBid = func(int64) (Bid, error) {
					return Bid{ID: 1, UserID: 1, ItemID: 1, Amount: 999}, nil
				}
			},
		},
		{
			"got_item_higher_value",
			&Bid{ID: 0, UserID: 1, ItemID: 1, Amount: 1},
			nil,
			ValidationError{"item": ErrLowValue},
			func(t *testing.T) {
				tbdb.txCreate = func(b *Bid) error {
					b.ID = 1
					return nil
				}
				tidb.get = func(int64) (Item, error) {
					return Item{ID: 1, Name: "test", Value: 999}, nil
				}
				tudb.get = func(int64) (User, error) {
					return User{ID: 1, Name: "test"}, nil
				}
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			err := bsvc.TxCreate(tt.bid)

			if tt.outerr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.outerr), "errors must match, expected %v, got %v", tt.outerr, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.outbid, tt.bid)
			}

			*tudb = testUserDB{}
			*tidb = testItemDB{}
			*tbdb = testBidDB{}
		})
	}
}

func TestBidService_GetWinningBid(t *testing.T) {
	tidb := &testItemDB{}
	tbdb := &testBidDB{}

	db := CreateDatabase()
	isvc := NewItemService(db, nil)
	bsvc := NewBidService(db, isvc, nil)

	isvc.(itemService).ItemService.(*itemCapsule).ItemDB = tidb
	bsvc.(bidService).BidService.(*bidValidator).BidDB = tbdb

	var cases = []struct {
		name   string
		itemID int64
		outbid Bid
		outerr error
		setup  func(*testing.T)
	}{
		{
			"ok",
			1,
			Bid{ID: 1, UserID: 1, ItemID: 1, Amount: 777},
			nil,
			func(t *testing.T) {
				tbdb.getWinningBid = func(int64) (Bid, error) {
					return Bid{ID: 1, UserID: 1, ItemID: 1, Amount: 777}, nil
				}
				tidb.get = func(int64) (Item, error) {
					return Item{ID: 1, Name: "test", Value: 0}, nil
				}
			},
		},
		{
			"item_not_exists",
			1,
			Bid{},
			ValidationError{"item": ErrNotFound},
			func(t *testing.T) {
				tidb.get = func(int64) (Item, error) {
					return Item{}, ErrNotFound
				}
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			bid, err := bsvc.GetWinningBid(tt.itemID)

			if tt.outerr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.outerr), "errors must match, expected %v, got %v", tt.outerr, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.outbid, bid)
			}

			*tidb = testItemDB{}
			*tbdb = testBidDB{}
		})
	}
}

func TestBidService_ListBidsByItemID(t *testing.T) {
	tidb := &testItemDB{}
	tbdb := &testBidDB{}

	db := CreateDatabase()
	isvc := NewItemService(db, nil)
	bsvc := NewBidService(db, tidb, nil)

	isvc.(itemService).ItemService.(*itemCapsule).ItemDB = tidb
	bsvc.(bidService).BidService.(*bidValidator).BidDB = tbdb

	var cases = []struct {
		name    string
		itemID  int64
		outbids []Bid
		outerr  error
		setup   func(*testing.T)
	}{
		{
			"ok",
			1,
			[]Bid{
				{ID: 1, UserID: 1, ItemID: 1, Amount: 1},
				{ID: 2, UserID: 2, ItemID: 1, Amount: 2},
			},
			nil,
			func(t *testing.T) {
				tbdb.listItemBids = func(int64) ([]Bid, error) {
					return []Bid{
						{ID: 1, UserID: 1, ItemID: 1, Amount: 1},
						{ID: 2, UserID: 2, ItemID: 1, Amount: 2},
					}, nil
				}
				tidb.get = func(int64) (Item, error) {
					return Item{ID: 1, Name: "test", Value: 0}, nil
				}
			},
		},
		{
			"item_not_found",
			1,
			nil,
			ValidationError{"item": ErrNotFound},
			func(t *testing.T) {
				tidb.get = func(int64) (Item, error) {
					return Item{}, ErrNotFound
				}
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			bids, err := bsvc.ListBidsByItemID(tt.itemID)

			if tt.outerr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.outerr), "errors must match, expected %v, got %v", tt.outerr, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.outbids, bids)
			}

			*tidb = testItemDB{}
			*tbdb = testBidDB{}
		})
	}
}

func TestBidService_ListBidsByUserID(t *testing.T) {
	tudb := &testUserDB{}
	tbdb := &testBidDB{}

	db := CreateDatabase()
	usvc := NewUserService(db)
	bsvc := NewBidService(db, nil, usvc)

	bsvc.(bidService).BidService.(*bidValidator).BidDB = tbdb
	usvc.(userService).UserService.(*userCapsule).UserDB = tudb

	var cases = []struct {
		name    string
		userID  int64
		outbids []Bid
		outerr  error
		setup   func(*testing.T)
	}{
		{
			"ok",
			1,
			[]Bid{
				{ID: 1, UserID: 1, ItemID: 1, Amount: 1},
				{ID: 2, UserID: 2, ItemID: 1, Amount: 2},
			},
			nil,
			func(t *testing.T) {
				tbdb.listBidsByUserID = func(int64) ([]Bid, error) {
					return []Bid{
						{ID: 1, UserID: 1, ItemID: 1, Amount: 1},
						{ID: 2, UserID: 2, ItemID: 1, Amount: 2},
					}, nil
				}
				tudb.get = func(int64) (User, error) {
					return User{ID: 1, Name: "test"}, nil
				}
			},
		},
		{
			"user_not_exists",
			1,
			nil,
			ValidationError{"user": ErrNotFound},
			func(t *testing.T) {
				tudb.get = func(int64) (User, error) {
					return User{}, ErrNotFound
				}
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			bids, err := bsvc.ListBidsByUserID(tt.userID)

			if tt.outerr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.outerr), "errors must match, expected %v, got %v", tt.outerr, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.outbids, bids)
			}

			*tbdb = testBidDB{}
			*tudb = testUserDB{}
		})
	}
}
