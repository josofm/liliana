package user

import "github.com/josofm/liliana/internal/entity/user"

type Repository interface {
	Create(u *user.User) error
	GetAll() ([]*user.User, error)
	GetByID(id int64) (*user.User, error)
	Update(id int64, u *user.User) error
	Delete(id int64) error
}
