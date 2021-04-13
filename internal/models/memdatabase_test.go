package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func (udb *UserStorage) testTxCreate(u *User, td time.Duration) {
	udb.mu.Lock()

	time.Sleep(td)
	udb.Create(u)

	udb.mu.Unlock()
}

func (bdb *BidStorage) testTxCreate(b *Bid, td time.Duration) error {
	if bdb.mu.isIDLocked(b.ItemID) {
		return ErrConflict
	}

	bdb.mu.Lock(b.ItemID)
	time.Sleep(td)
	bdb.Create(b)

	bdb.mu.Unlock(b.ItemID)
	return nil
}

func TestBidStorage_TxCreate(t *testing.T) {
	tests := []struct {
		name        string
		wanterror   error
		blockingBid *Bid
		manyBids    []Bid
		want        map[int64]Bid
	}{
		{
			name:        "conflict",
			wanterror:   ErrConflict,
			blockingBid: &Bid{UserID: 1, ItemID: 1, Amount: 10},
			manyBids: []Bid{
				{UserID: 1, ItemID: 1, Amount: 10},
			},
			want: map[int64]Bid{
				1: {ID: 1, UserID: 1, ItemID: 1, Amount: 10},
			},
		},
		{
			name:        "ok",
			wanterror:   nil,
			blockingBid: &Bid{UserID: 1, ItemID: 1, Amount: 10},
			manyBids: []Bid{
				{UserID: 1, ItemID: 2, Amount: 10},
			},
			want: map[int64]Bid{
				1: {ID: 1, UserID: 1, ItemID: 1, Amount: 10},
				2: {ID: 2, UserID: 1, ItemID: 2, Amount: 10},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := CreateDatabase()

			go func() {
				err := database.bids.testTxCreate(tt.blockingBid, time.Duration(3*time.Second))
				assert.NoError(t, err)
			}()

			time.Sleep(1 * time.Second)                // Ensure previous creation goroutine is being executed
			assert.Equal(t, 1, database.bids.mu.state) // Check if mutex is locked
			assert.True(t, database.bids.mu.isIDLocked(tt.blockingBid.ItemID))

			for _, v := range tt.manyBids {
				err := database.bids.testTxCreate(&v, 0)

				if tt.wanterror != nil {
					assert.Equal(t, tt.wanterror, err)
					fmt.Println(err.Error())
				} else {
					assert.NoError(t, err)
				}
			}

			time.Sleep(5 * time.Second) // Wait until everything finishes
			assert.Equal(t, tt.want, database.bids.data)
		})
	}
}

func TestUserStorage_TxCreate(t *testing.T) {
	tests := []struct {
		name         string
		blockingUser *User
		manyUsers    []User
		want         map[int64]User
	}{
		{
			name:         "ok",
			blockingUser: &User{ID: 1, Name: "Morty"},
			manyUsers: []User{
				{ID: 2, Name: "Rick"},
			},
			want: map[int64]User{
				1: {ID: 1, Name: "Morty"},
				2: {ID: 2, Name: "Rick"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := CreateDatabase()

			go database.users.testTxCreate(tt.blockingUser, time.Duration(3*time.Second))

			time.Sleep(1 * time.Second)                 // Ensure previous creation goroutine is being executed
			assert.Equal(t, 1, database.users.mu.state) // Check if mutex is locked
			assert.True(t, database.users.mu.isLocked())

			for _, v := range tt.manyUsers {
				database.users.testTxCreate(&v, 0)
			}

			time.Sleep(5 * time.Second) // Wait until everything finishes
			assert.Equal(t, tt.want, database.users.data)
		})
	}
}
