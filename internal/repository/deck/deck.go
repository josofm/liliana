package deck

import "github.com/josofm/liliana/internal/entity/deck"

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Create(d *deck.Deck) error {
	return s.repo.Create(d)
}

func (s *Service) GetAll() ([]*deck.Deck, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByID(id int64) (*deck.Deck, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
