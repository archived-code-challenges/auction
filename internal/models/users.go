package models

type UserService interface {
	UserDB
}

type UserDB interface {
	TxCreate(*User)
	Get(int64) (User, error)
	ListUsers() []User
}

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// userService wraps the UserService interface to allow mocking by interfaces
type userService struct {
	UserService
}

type userCapsule struct {
	UserDB
}

func NewUserService(db *DB) UserService {
	return userService{
		UserService: &userCapsule{
			UserDB: &UserStorage{
				db.users.mu,
				db.users.data,
				0,
			},
		},
	}
}
