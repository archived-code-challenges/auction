package models

type Bid struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"userId"`
	ItemID int64 `json:"itemId"`
	Amount int   `json:"amount"`
}

type BidDB interface {
	TxCreate(*Bid) error
	ListBidsByItemID(int64) ([]Bid, error)
	GetWinningBid(int64) (Bid, error)
	ListBidsByUserID(int64) ([]Bid, error)
}

type BidService interface {
	BidDB
}

// bidService wraps the BidService interface to allow mocking by interfaces
type bidService struct {
	BidService
}

func NewBidService(db *DB, isvc ItemService, usvc UserService) BidService {
	return bidService{
		BidService: &bidValidator{
			BidDB: &BidStorage{
				db.bids.mu,
				db.bids.data,
				0,
			},
			itemService: isvc,
			userService: usvc,
		},
	}
}

type bidValidator struct {
	BidDB
	itemService ItemService
	userService UserService
}

func (bs *bidValidator) TxCreate(b *Bid) error {
	if err := bs.runValFuncs(b,
		bs.itemExists,
		bs.userExists,
		bs.higherItemValue,
		bs.higherBidAmount,
	); err != nil {
		return err
	}

	return bs.BidDB.TxCreate(b)
}

func (bs *bidValidator) ListBidsByItemID(itemID int64) ([]Bid, error) {
	b := &Bid{ItemID: itemID}

	if err := bs.runValFuncs(b,
		bs.itemExists,
	); err != nil {
		return nil, err
	}

	return bs.BidDB.ListBidsByItemID(itemID)
}

func (bs *bidValidator) GetWinningBid(itemID int64) (Bid, error) {
	b := &Bid{ItemID: itemID}

	if err := bs.runValFuncs(b,
		bs.itemExists,
	); err != nil {
		return Bid{}, err
	}

	return bs.BidDB.GetWinningBid(itemID)
}

func (bs *bidValidator) ListBidsByUserID(userID int64) ([]Bid, error) {
	b := &Bid{UserID: userID}

	if err := bs.runValFuncs(b,
		bs.userExists,
	); err != nil {
		return nil, err
	}

	return bs.BidDB.ListBidsByUserID(userID)
}

type bidValFn func(b *Bid) error

func (bv *bidValidator) runValFuncs(b *Bid, fns ...func() (string, bidValFn)) error {
	return runValidationFunctions(b, fns)
}

func (bv *bidValidator) itemExists() (string, bidValFn) {
	return "item", func(b *Bid) error {
		if _, err := bv.itemService.Get(b.ItemID); err != nil {
			return ErrNotFound
		}
		return nil
	}
}

func (bv *bidValidator) userExists() (string, bidValFn) {
	return "user", func(b *Bid) error {
		if _, err := bv.userService.Get(b.UserID); err != nil {
			return ErrNotFound
		}
		return nil
	}
}

func (bv *bidValidator) higherItemValue() (string, bidValFn) {
	return "item", func(b *Bid) error {
		item, err := bv.itemService.Get(b.ItemID)
		if err != nil {
			return err
		}

		if b.Amount <= item.Value {
			return ErrLowValue
		}

		return nil
	}
}

func (bv *bidValidator) higherBidAmount() (string, bidValFn) {
	return "bid", func(b *Bid) error {
		winning, err := bv.BidDB.GetWinningBid(b.ItemID)
		if err != nil && err != ErrNotFound {
			return err
		}

		if b.Amount <= winning.Amount {
			return ErrLowValue
		}

		return nil
	}
}
