package user

import (
	"errors"
	"github.com/josofm/liliana/internal/entity/user"
	r "github.com/josofm/liliana/internal/repository/user"
	"strings"
	"unicode/utf8"
)

var ErrInvalidSearch = errors.New("query must contain between 2 and 100 characters")

func (s *Service) Search(query string) ([]*user.User, error) {
	query = strings.TrimSpace(query)
	if n := utf8.RuneCountInString(query); n < 2 || n > 100 {
		return nil, ErrInvalidSearch
	}
	return s.repo.Search(query)
}

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

func (s *Service) Update(id int64, u *user.User) error {
	return s.repo.Update(id, u)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
