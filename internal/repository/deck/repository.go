package deck

import "github.com/josofm/liliana/internal/entity/deck"

type Repository interface {
	Create(d *deck.Deck) error
	GetAll() ([]*deck.Deck, error)
	GetByID(id int64) (*deck.Deck, error)
	Update(id int64, d *deck.Deck) error
	PatchCards(id int64, upsert []deck.Card, remove []string) error
	Delete(id int64) error
}
