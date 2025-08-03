package deck

import (
	"errors"
	"sync"

	"github.com/josofm/liliana/internal/entity/deck"
)

type inMemoryRepo struct {
	mu     sync.RWMutex
	decks  map[int64]*deck.Deck
	nextID int64
}

func NewInMemoryRepo() Repository {
	return &inMemoryRepo{
		decks:  make(map[int64]*deck.Deck),
		nextID: 1,
	}
}

func (r *inMemoryRepo) Create(d *deck.Deck) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d.ID = r.nextID
	r.decks[d.ID] = d
	r.nextID++
	return nil
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
	d.ID = id
	r.decks[id] = d
	return nil
}

func (r *inMemoryRepo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.decks, id)
	return nil
}
