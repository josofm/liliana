package service

import (
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
)

type Service struct {
	repo     deckRepo.Repository
	importer SourceImporter
}

func NewService(repo deckRepo.Repository) *Service {
	return &Service{repo: repo, importer: NewArchidektImporter()}
}

func NewServiceWithImporter(repo deckRepo.Repository, importer SourceImporter) *Service {
	return &Service{repo: repo, importer: importer}
}

func (s *Service) Prepare(d *deckEntity.Deck) error {
	if d.SourceLink == "" {
		return nil
	}
	imported, err := s.importer.Import(d.SourceLink)
	if err != nil {
		// Keep backwards compatibility for fully specified manual decks. The
		// source is enrichment in this case; source-only requests still fail.
		if hasRequiredMetadata(d) {
			return nil
		}
		return err
	}
	imported.OwnerID = d.OwnerID
	imported.SourceLink = d.SourceLink
	*d = *imported
	return nil
}

func hasRequiredMetadata(d *deckEntity.Deck) bool {
	return d.Name != "" && d.Color != "" && d.Format != "" && (d.Format != "commander" || d.Commander != "")
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
