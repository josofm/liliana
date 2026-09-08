package deck

import (
	"errors"

	"github.com/josofm/liliana/internal/entity/deck"
)

var ErrIdempotencyKeyNotFound = errors.New("idempotency key not found")

type Repository interface {
	Create(d *deck.Deck) error
	GetByIdempotencyKey(ownerID int64, key string) (*deck.Deck, error)
	GetAll() ([]*deck.Deck, error)
	GetByID(id int64) (*deck.Deck, error)
	Update(id int64, d *deck.Deck) error
	PatchCards(id int64, upsert []deck.Card, remove []string) error
	Delete(id int64) error
}
