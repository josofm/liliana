package user

import (
	"github.com/josofm/liliana/internal/entity/user"
	r "github.com/josofm/liliana/internal/repository/user"
)

type Service struct {
	repo r.Repository
}

func NewService(r r.Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Create(u *user.User) error {
	return s.repo.Create(u)
}

func (s *Service) GetAll() ([]*user.User, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByID(id int64) (*user.User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
