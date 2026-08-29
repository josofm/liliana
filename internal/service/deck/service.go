package service

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
)

type Service struct {
	repo      deckRepo.Repository
	importer  SourceImporter
	validator CardValidator
}

func NewService(repo deckRepo.Repository) *Service {
	return &Service{repo: repo, importer: NewArchidektImporter(), validator: NewScryfallValidator()}
}

func NewServiceWithImporter(repo deckRepo.Repository, importer SourceImporter) *Service {
	return &Service{repo: repo, importer: importer, validator: NewScryfallValidator()}
}

func NewServiceWithDependencies(repo deckRepo.Repository, importer SourceImporter, validator CardValidator) *Service {
	return &Service{repo: repo, importer: importer, validator: validator}
}

func (s *Service) Prepare(d *deckEntity.Deck) error {
	if d.SourceLink == "" {
		return s.prepareManual(d)
	}
	imported, err := s.importer.Import(d.SourceLink)
	if err != nil {
		// Keep backwards compatibility for fully specified manual decks. The
		// source is enrichment in this case; source-only requests still fail.
		if hasRequiredMetadata(d) {
			return s.validateCards(d)
		}
		return err
	}
	imported.OwnerID = d.OwnerID
	imported.SourceLink = d.SourceLink
	if err := s.prepareCommander(imported, false); err != nil {
		return err
	}
	if err := s.validateCards(imported); err != nil {
		return err
	}
	*d = *imported
	return nil
}

func (s *Service) prepareManual(d *deckEntity.Deck) error {
	if d.Cards == nil {
		d.Cards = make([]deckEntity.Card, 0)
	}
	if err := s.prepareCommander(d, true); err != nil {
		return err
	}
	return s.validateCards(d)
}

func (s *Service) prepareCommander(d *deckEntity.Deck, deriveColor bool) error {
	if d.Name == "" || d.Format != "commander" || d.Commander == "" {
		return nil
	}
	commanderName := strings.SplitN(d.Commander, " / ", 2)[0]
	commander, err := s.validator.ResolveCommander(commanderName)
	if err != nil {
		return err
	}
	if commanderName == d.Commander {
		d.Commander = commander.Name
	}
	d.CommanderImageURI = commander.ImageURI
	if deriveColor {
		d.Color = colorIdentityCode(commander.ColorIdentity)
	}
	return nil
}

func colorIdentityCode(colors []string) string {
	present := make(map[string]bool, len(colors))
	for _, color := range colors {
		present[strings.ToUpper(color)] = true
	}
	var result strings.Builder
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		if present[color] {
			result.WriteString(color)
		}
	}
	if result.Len() == 0 {
		return "C"
	}
	return result.String()
}

func (s *Service) validateCards(d *deckEntity.Deck) error {
	if len(d.Cards) == 0 {
		return nil
	}
	cards, err := s.validator.Validate(d.Cards)
	if err != nil {
		return err
	}
	d.Cards = cards
	return nil
}

func hasRequiredMetadata(d *deckEntity.Deck) bool {
	return d.Name != "" && d.Color != "" && d.Format != "" && (d.Format != "commander" || d.Commander != "")
}

func (s *Service) Create(deck *deckEntity.Deck) error {
	return s.repo.Create(deck)
}

func (s *Service) SearchCommanders(query string) ([]CommanderSuggestion, error) {
	return s.validator.SearchCommanders(query)
}

func (s *Service) SearchCards(query string) ([]deckEntity.Card, error) {
	return s.validator.SearchCards(query)
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

func (s *Service) PatchCards(id int64, upsert []deckEntity.Card, remove []string) (*deckEntity.Deck, error) {
	if len(upsert) == 0 && len(remove) == 0 {
		return nil, errors.New("at least one card change is required")
	}
	for _, card := range upsert {
		if strings.TrimSpace(card.Name) == "" || card.Quantity <= 0 {
			return nil, errors.New("upsert cards require name and quantity greater than zero")
		}
	}
	if len(upsert) > 0 {
		validated, err := s.validator.Validate(upsert)
		if err != nil {
			return nil, err
		}
		upsert = validated
	}
	uniqueRemove := make([]string, 0, len(remove))
	seen := make(map[string]bool, len(remove))
	for _, oracleID := range remove {
		oracleID = strings.TrimSpace(oracleID)
		if oracleID == "" {
			return nil, errors.New("remove requires non-empty oracle ids")
		}
		if !seen[oracleID] {
			seen[oracleID] = true
			uniqueRemove = append(uniqueRemove, oracleID)
		}
	}
	if err := s.repo.PatchCards(id, upsert, uniqueRemove); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}

func ParseCardList(value string) ([]deckEntity.Card, error) {
	cards := make([]deckEntity.Card, 0)
	positions := make(map[string]int)
	scanner := bufio.NewScanner(strings.NewReader(value))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid card at line %d: expected '<quantity> <card name>'", lineNumber)
		}
		quantityText := fields[0]
		name := strings.TrimSpace(strings.TrimPrefix(line, quantityText))
		quantity, err := strconv.Atoi(quantityText)
		if err != nil || quantity <= 0 || name == "" {
			return nil, fmt.Errorf("invalid card at line %d: expected '<quantity> <card name>'", lineNumber)
		}
		key := strings.ToLower(name)
		if position, exists := positions[key]; exists {
			cards[position].Quantity += quantity
			continue
		}
		positions[key] = len(cards)
		cards = append(cards, deckEntity.Card{Name: name, Quantity: quantity})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read card list: %w", err)
	}
	return cards, nil
}

func (s *Service) AddCards(id int64, cards []deckEntity.Card) (*deckEntity.Deck, error) {
	if len(cards) == 0 {
		return nil, errors.New("card list cannot be empty")
	}
	cards, err := s.validator.Validate(cards)
	if err != nil {
		return nil, err
	}
	d, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	existingQuantities := make(map[string]int, len(d.Cards))
	for _, card := range d.Cards {
		existingQuantities[strings.ToLower(card.Name)] = card.Quantity
	}
	for index := range cards {
		card := &cards[index]
		key := strings.ToLower(card.Name)
		card.Quantity += existingQuantities[key]
	}
	if err := s.repo.PatchCards(id, cards, nil); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}
