CREATE TABLE cards (
	oracle_id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	mana_cost TEXT NOT NULL DEFAULT '',
	type_line TEXT NOT NULL DEFAULT '',
	color_identity JSONB NOT NULL DEFAULT '[]'::jsonb,
	image_uri TEXT NOT NULL DEFAULT '',
	card_faces JSONB NOT NULL DEFAULT '[]'::jsonb,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE deck_cards (
	deck_id BIGINT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
	oracle_id TEXT NOT NULL REFERENCES cards(oracle_id),
	quantity INTEGER NOT NULL CHECK (quantity > 0),
	PRIMARY KEY (deck_id, oracle_id)
);

CREATE INDEX deck_cards_deck_id_idx ON deck_cards (deck_id);
