package service

import (
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
)

type Service struct {
	repo deckRepo.Repository
}

func NewService(repo deckRepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(deck *deckEntity.Deck) error {
	return s.repo.Create(deck)
}

func (s *Service) GetAll() ([]*deckEntity.Deck, error) {
	decks, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	// Convert []*deckEntity.Deck to []deckEntity.Deck
	result := make([]*deckEntity.Deck, len(decks))
	for i, d := range decks {
		if d != nil {
			result[i] = d
		}
	}
	return result, nil
}

func (s *Service) GetByID(id int64) (*deckEntity.Deck, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Update(id int64, d *deckEntity.Deck) error {
	return s.repo.Update(id, d)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
