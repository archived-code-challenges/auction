package models

import (
	"sync"
)

// Mutex keeps track of a sync.RWMutex and a state (0=false, 1=true).
type Mutex struct {
	rw sync.RWMutex

	// state 1 means locked, 0 unlocked and is used to avoid reflection
	// when checking if a mutex is locked
	state int
}

func (m *Mutex) Lock() {
	m.rw.Lock()
	m.state = 1 // Set state to locked
}

func (m *Mutex) Unlock() {
	m.rw.Unlock()
	m.state = 0 // Set state to unlocked
}

func (m *Mutex) isLocked() bool {
	return m.state == 1
}

// DedicatedMutex keeps track of a sync.RWMutex, a state (0=false, 1=true)
// and points to a specific (int64 identifier) element to be more strict with the blocking policy.
type DedicatedMutex struct {
	rw sync.RWMutex

	state   int
	element int64
}

func (m *DedicatedMutex) Lock(element int64) {
	m.rw.Lock()
	m.state = 1
	m.element = element
}

func (m *DedicatedMutex) Unlock(element int64) {
	m.rw.Unlock()
	m.state = 0
	m.element = 0
}

func (m *DedicatedMutex) isIDLocked(id int64) bool {
	return m.state == 1 && m.element == id
}

// BidStorage contains a data structure that stores the Bids and allows for data consistency.
type BidStorage struct {
	mu   DedicatedMutex
	data map[int64]Bid

	incrementalID int64
}

// BidStorage contains a data structure that stores the Items and allows for data consistency.
type ItemStorage struct {
	mu   Mutex
	data map[int64]Item

	incrementalID int64
}

// BidStorage contains a data structure that stores the Users and allows for data consistency.
type UserStorage struct {
	mu   Mutex
	data map[int64]User

	incrementalID int64
}

// DB contains all the data structures used by the service
type DB struct {
	bids  BidStorage
	items ItemStorage
	users UserStorage
}

func CreateDatabase() *DB {
	db := &DB{
		bids:  BidStorage{data: make(map[int64]Bid)},
		items: ItemStorage{data: make(map[int64]Item)},
		users: UserStorage{data: make(map[int64]User)},
	}
	return db
}

// Create an entity Bid in the in-memory database
func (bdb *BidStorage) Create(b *Bid) {
	bdb.incrementalID = bdb.incrementalID + 1

	bdb.data[bdb.incrementalID] = Bid{
		ID:     bdb.incrementalID,
		ItemID: b.ItemID,
		UserID: b.UserID,
		Amount: b.Amount,
	}

	b.ID = bdb.incrementalID
}

// Create a Bid entity in the in-memory database ensuring that the creation of an entity is transactional.
// Locking and unlocking the mutex attached to the data structure.
// Will raise an error if the itemID pointed is already being used by another thread.
func (bdb *BidStorage) TxCreate(b *Bid) error {
	if bdb.mu.isIDLocked(b.ItemID) {
		return ErrConflict
	}

	bdb.mu.Lock(b.ItemID)
	bdb.Create(b)
	bdb.mu.Unlock(b.ItemID)
	return nil
}

// Lists the existing Items in the in-memory database
func (idb *ItemStorage) ListItems() []Item {
	items := []Item{}
	for _, v := range idb.data {
		items = append(items, v)
	}
	return items
}

// Get an Item by its identification number
func (idb *ItemStorage) Get(id int64) (Item, error) {
	if v, found := idb.data[id]; found {
		return v, nil
	}
	return Item{}, ErrNotFound
}

// Create an Item entity in the in-memory database
func (idb *ItemStorage) Create(i *Item) {
	idb.incrementalID = idb.incrementalID + 1

	idb.data[idb.incrementalID] = Item{
		ID:    idb.incrementalID,
		Name:  i.Name,
		Value: i.Value,
	}

	i.ID = idb.incrementalID
}

// Create an Item entity in the in-memory database ensuring that the creation of an entity is transactional.
// Locking and unlocking the mutex attached to the data structure.
func (idb *ItemStorage) TxCreate(i *Item) {
	idb.mu.Lock()
	idb.Create(i)
	idb.mu.Unlock()
}

// List the existing Users in the in-memory database
func (idb *UserStorage) ListUsers() []User {
	users := []User{}
	for _, v := range idb.data {
		users = append(users, v)
	}
	return users
}

func (idb *UserStorage) Get(id int64) (User, error) {
	if v, found := idb.data[id]; found {
		return v, nil
	}
	return User{}, ErrNotFound
}

// Create a User entity in the in-memory database
func (udb *UserStorage) Create(u *User) {
	udb.incrementalID = udb.incrementalID + 1

	udb.data[udb.incrementalID] = User{
		ID:   udb.incrementalID,
		Name: u.Name,
	}

	u.ID = udb.incrementalID
}

// Create a User entity in the in-memory database ensuring that the creation of an entity is transactional.
// Locking and unlocking the mutex attached to the data structure.
func (udb *UserStorage) TxCreate(u *User) {
	udb.mu.Lock()
	udb.Create(u)
	udb.mu.Unlock()
}

// ListBidsByItemID gets all the bids for a specific item
func (bdb *BidStorage) ListBidsByItemID(itemID int64) ([]Bid, error) {
	var bids []Bid

	for _, v := range bdb.data {
		if itemID == v.ItemID {
			bids = append(bids, v)
		}
	}

	if len(bids) == 0 {
		return nil, ErrNotFound
	}

	return bids, nil
}

// GetWinningBid gets the current winning bid for an item
func (bdb *BidStorage) GetWinningBid(itemID int64) (Bid, error) {
	var winningBid Bid

	bids, err := bdb.ListBidsByItemID(itemID)
	if err != nil {
		return Bid{}, err
	}

	for _, v := range bids {
		if v.Amount > winningBid.Amount {
			winningBid = v
		}
	}

	return winningBid, nil
}

// ListBidsByUserID gets all the bids on which a specific user has a bid
func (bdb *BidStorage) ListBidsByUserID(userID int64) ([]Bid, error) {
	var bids []Bid

	for _, v := range bdb.data {
		if userID == v.UserID {
			bids = append(bids, v)
		}
	}

	return bids, nil
}

// ListItemsByIDs fetches all items given a set of IDs
func (idb *ItemStorage) ListItemsByIDs(itemIDs ...int64) ([]Item, error) {
	var items []Item

	for _, itemID := range itemIDs {
		if val, ok := idb.data[int64(itemID)]; ok {
			items = append(items, val)
		}
	}

	return items, nil
}
