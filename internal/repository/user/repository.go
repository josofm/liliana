package user

import (
	"errors"
	"github.com/josofm/liliana/internal/entity/user"
)

var ErrNotFound = errors.New("user not found")

type Repository interface {
	Create(u *user.User) error
	GetAll() ([]*user.User, error)
	Search(query string) ([]*user.User, error)
	GetByID(id int64) (*user.User, error)
	GetByEmail(email string) (*user.User, error)
	Update(id int64, u *user.User) error
	Delete(id int64) error
}
