package deck

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
)

type postgresRepo struct{ db *sql.DB }

func NewPostgresRepo(db *sql.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) Create(d *deckEntity.Deck) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	const query = `INSERT INTO decks (name, color, format, commander, commander_image_uri, owner_id, source_link) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	if err := tx.QueryRow(query, d.Name, d.Color, d.Format, d.Commander, d.CommanderImageURI, d.OwnerID, d.SourceLink).Scan(&d.ID); err != nil {
		return err
	}
	if err := saveCards(tx, d); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *postgresRepo) GetAll() ([]*deckEntity.Deck, error) {
	rows, err := r.db.Query(`SELECT id, name, color, format, commander, commander_image_uri, owner_id, source_link FROM decks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	decks := make([]*deckEntity.Deck, 0)
	for rows.Next() {
		d := &deckEntity.Deck{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Color, &d.Format, &d.Commander, &d.CommanderImageURI, &d.OwnerID, &d.SourceLink); err != nil {
			rows.Close()
			return nil, err
		}
		decks = append(decks, d)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for _, d := range decks {
		if err := loadCards(r.db, d); err != nil {
			return nil, err
		}
	}
	return decks, nil
}

func (r *postgresRepo) GetByID(id int64) (*deckEntity.Deck, error) {
	d := &deckEntity.Deck{}
	err := r.db.QueryRow(`SELECT id, name, color, format, commander, commander_image_uri, owner_id, source_link FROM decks WHERE id=$1`, id).Scan(&d.ID, &d.Name, &d.Color, &d.Format, &d.Commander, &d.CommanderImageURI, &d.OwnerID, &d.SourceLink)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("deck not found")
	}
	if err != nil {
		return nil, err
	}
	if err := loadCards(r.db, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (r *postgresRepo) Update(id int64, d *deckEntity.Deck) error {
	result, err := r.db.Exec(`UPDATE decks SET name=$1,color=$2,format=$3,commander=$4,commander_image_uri=$5,owner_id=$6,source_link=$7 WHERE id=$8`, d.Name, d.Color, d.Format, d.Commander, d.CommanderImageURI, d.OwnerID, d.SourceLink, id)
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

func (r *postgresRepo) PatchCards(id int64, upsert []deckEntity.Card, remove []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRow(`SELECT 1 FROM decks WHERE id=$1 FOR UPDATE`, id).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return errors.New("deck not found")
	} else if err != nil {
		return err
	}
	for _, oracleID := range remove {
		if _, err := tx.Exec(`DELETE FROM deck_cards WHERE deck_id=$1 AND oracle_id=$2`, id, oracleID); err != nil {
			return err
		}
	}
	for _, card := range upsert {
		if err := upsertCard(tx, card); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO deck_cards (deck_id,oracle_id,quantity) VALUES ($1,$2,$3) ON CONFLICT (deck_id,oracle_id) DO UPDATE SET quantity=EXCLUDED.quantity`, id, card.OracleID, card.Quantity); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *postgresRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM decks WHERE id=$1`, id)
	return err
}

func saveCards(tx *sql.Tx, d *deckEntity.Deck) error {
	if _, err := tx.Exec(`DELETE FROM deck_cards WHERE deck_id=$1`, d.ID); err != nil {
		return err
	}
	for _, card := range d.Cards {
		if err := upsertCard(tx, card); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO deck_cards (deck_id,oracle_id,quantity) VALUES ($1,$2,$3) ON CONFLICT (deck_id,oracle_id) DO UPDATE SET quantity=EXCLUDED.quantity`, d.ID, card.OracleID, card.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func upsertCard(tx *sql.Tx, card deckEntity.Card) error {
	if card.OracleID == "" {
		return fmt.Errorf("card %q has no oracle id", card.Name)
	}
	colorIdentity := card.ColorIdentity
	if colorIdentity == nil {
		colorIdentity = []string{}
	}
	colors, err := json.Marshal(colorIdentity)
	if err != nil {
		return err
	}
	cardFaces := card.CardFaces
	if cardFaces == nil {
		cardFaces = []deckEntity.CardFace{}
	}
	faces, err := json.Marshal(cardFaces)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
			INSERT INTO cards (oracle_id,name,mana_cost,type_line,color_identity,image_uri,card_faces,updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
			ON CONFLICT (oracle_id) DO UPDATE SET
				name=EXCLUDED.name,
				mana_cost=CASE WHEN EXCLUDED.mana_cost<>'' THEN EXCLUDED.mana_cost ELSE cards.mana_cost END,
				type_line=CASE WHEN EXCLUDED.type_line<>'' THEN EXCLUDED.type_line ELSE cards.type_line END,
				color_identity=CASE WHEN jsonb_array_length(EXCLUDED.color_identity)>0 THEN EXCLUDED.color_identity ELSE cards.color_identity END,
				image_uri=CASE WHEN EXCLUDED.image_uri<>'' THEN EXCLUDED.image_uri ELSE cards.image_uri END,
				card_faces=CASE WHEN jsonb_array_length(EXCLUDED.card_faces)>0 THEN EXCLUDED.card_faces ELSE cards.card_faces END,
				updated_at=NOW()`,
		card.OracleID, card.Name, card.ManaCost, card.TypeLine, colors, card.ImageURI, faces)
	return err
}

type cardQueryer interface {
	Query(string, ...any) (*sql.Rows, error)
}

func loadCards(queryer cardQueryer, d *deckEntity.Deck) error {
	rows, err := queryer.Query(`SELECT c.oracle_id,c.name,dc.quantity,c.mana_cost,c.type_line,c.color_identity,c.image_uri,c.card_faces FROM deck_cards dc JOIN cards c ON c.oracle_id=dc.oracle_id WHERE dc.deck_id=$1 ORDER BY c.name`, d.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	d.Cards = make([]deckEntity.Card, 0)
	for rows.Next() {
		var card deckEntity.Card
		var colors []byte
		var faces []byte
		if err := rows.Scan(&card.OracleID, &card.Name, &card.Quantity, &card.ManaCost, &card.TypeLine, &colors, &card.ImageURI, &faces); err != nil {
			return err
		}
		if err := json.Unmarshal(colors, &card.ColorIdentity); err != nil {
			return err
		}
		if err := json.Unmarshal(faces, &card.CardFaces); err != nil {
			return err
		}
		d.Cards = append(d.Cards, card)
	}
	return rows.Err()
}
