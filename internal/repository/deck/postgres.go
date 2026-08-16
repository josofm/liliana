package deck

import (
	"database/sql"
	"encoding/json"
	"errors"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
)

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(d *deckEntity.Deck) error {
	cards, err := json.Marshal(d.Cards)
	if err != nil {
		return err
	}
	const query = `
		INSERT INTO decks (name, color, format, commander, owner_id, source_link, cards)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	return r.db.QueryRow(
		query,
		d.Name,
		d.Color,
		d.Format,
		d.Commander,
		d.OwnerID,
		d.SourceLink,
		cards,
	).Scan(&d.ID)
}

func (r *postgresRepo) GetAll() ([]*deckEntity.Deck, error) {
	const query = `
		SELECT id, name, color, format, commander, owner_id, source_link, cards
		FROM decks
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decks := make([]*deckEntity.Deck, 0)
	for rows.Next() {
		d := &deckEntity.Deck{}
		var cards []byte
		if err := rows.Scan(&d.ID, &d.Name, &d.Color, &d.Format, &d.Commander, &d.OwnerID, &d.SourceLink, &cards); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(cards, &d.Cards); err != nil {
			return nil, err
		}
		decks = append(decks, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return decks, nil
}

func (r *postgresRepo) GetByID(id int64) (*deckEntity.Deck, error) {
	const query = `
		SELECT id, name, color, format, commander, owner_id, source_link, cards
		FROM decks
		WHERE id = $1
	`

	d := &deckEntity.Deck{}
	var cards []byte
	err := r.db.QueryRow(query, id).Scan(&d.ID, &d.Name, &d.Color, &d.Format, &d.Commander, &d.OwnerID, &d.SourceLink, &cards)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("deck not found")
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cards, &d.Cards); err != nil {
		return nil, err
	}

	return d, nil
}

func (r *postgresRepo) Update(id int64, d *deckEntity.Deck) error {
	cards, err := json.Marshal(d.Cards)
	if err != nil {
		return err
	}
	const query = `
		UPDATE decks
		SET name = $1, color = $2, format = $3, commander = $4, owner_id = $5, source_link = $6, cards = $7
		WHERE id = $8
	`

	result, err := r.db.Exec(query, d.Name, d.Color, d.Format, d.Commander, d.OwnerID, d.SourceLink, cards, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("deck not found")
	}

	d.ID = id
	return nil
}

func (r *postgresRepo) Delete(id int64) error {
	const query = `DELETE FROM decks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
