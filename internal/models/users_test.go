package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUserDB struct {
	UserDB
	txCreate       func(*User)
	get            func(int64) (User, error)
	listUsersByIDs func(...int64) ([]User, error)
}

func (t *testUserDB) TxCreate(i *User) {
	if t.txCreate != nil {
		t.txCreate(i)
	}
}

func (t *testUserDB) Get(userID int64) (User, error) {
	if t.get != nil {
		return t.get(userID)
	}
	return User{}, nil
}

func (t *testUserDB) ListUsersByIDs(userIDs ...int64) ([]User, error) {
	if t.listUsersByIDs != nil {
		return t.listUsersByIDs(userIDs...)
	}
	return nil, nil
}

func TestUserService_TxCreate(t *testing.T) {
	tudb := &testUserDB{}

	db := CreateDatabase()
	usvc := NewUserService(db)

	usvc.(userService).UserService.(*userCapsule).UserDB = tudb

	var cases = []struct {
		name    string
		outuser map[int64]User
		setup   func(*testing.T)
	}{
		{
			"ok",
			map[int64]User{
				1: {ID: 1, Name: "test"},
			},
			func(t *testing.T) {
				tudb.txCreate = func(u *User) {
					db.users.Create(u)
				}
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			for _, v := range tt.outuser {
				usvc.TxCreate(&v)
			}

			assert.Equal(t, tt.outuser, db.users.data)

			*tudb = testUserDB{}
		})
	}
}
