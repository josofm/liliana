package deck

import (
	"errors"
	"fmt"
	"sync"

	"github.com/josofm/liliana/internal/entity/deck"
)

type inMemoryRepo struct {
	mu     sync.RWMutex
	decks  map[int64]*deck.Deck
	nextID int64
	keys   map[string]int64
}

func NewInMemoryRepo() Repository {
	return &inMemoryRepo{
		decks:  make(map[int64]*deck.Deck),
		nextID: 1,
		keys:   make(map[string]int64),
	}
}

func (r *inMemoryRepo) Create(d *deck.Deck) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%d:%s", d.OwnerID, d.IdempotencyKey)
	if _, exists := r.keys[key]; d.IdempotencyKey != "" && exists {
		return errors.New("idempotency key already exists")
	}
	d.ID = r.nextID
	r.decks[d.ID] = d
	if d.IdempotencyKey != "" {
		r.keys[key] = d.ID
	}
	r.nextID++
	return nil
}

func (r *inMemoryRepo) GetByIdempotencyKey(ownerID int64, key string) (*deck.Deck, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.keys[fmt.Sprintf("%d:%s", ownerID, key)]
	if !exists {
		return nil, ErrIdempotencyKeyNotFound
	}
	return r.decks[id], nil
}

func (r *inMemoryRepo) GetAll() ([]*deck.Deck, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*deck.Deck
	for _, d := range r.decks {
		result = append(result, d)
	}
	return result, nil
}

func (r *inMemoryRepo) GetByID(id int64) (*deck.Deck, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.decks[id]
	if !ok {
		return nil, errors.New("deck not found")
	}
	return d, nil
}

func (r *inMemoryRepo) Update(id int64, d *deck.Deck) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.decks[id]; !exists {
		return errors.New("deck not found")
	}
	d.Cards = r.decks[id].Cards
	d.ID = id
	r.decks[id] = d
	return nil
}

func (r *inMemoryRepo) PatchCards(id int64, upsert []deck.Card, remove []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, exists := r.decks[id]
	if !exists {
		return errors.New("deck not found")
	}
	removed := make(map[string]bool, len(remove))
	for _, oracleID := range remove {
		removed[oracleID] = true
	}
	cards := make([]deck.Card, 0, len(d.Cards)+len(upsert))
	positions := make(map[string]int)
	for _, card := range d.Cards {
		if removed[card.OracleID] {
			continue
		}
		positions[card.OracleID] = len(cards)
		cards = append(cards, card)
	}
	for _, card := range upsert {
		if position, ok := positions[card.OracleID]; ok {
			cards[position] = card
			continue
		}
		positions[card.OracleID] = len(cards)
		cards = append(cards, card)
	}
	d.Cards = cards
	return nil
}

func (r *inMemoryRepo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.decks, id)
	return nil
}
