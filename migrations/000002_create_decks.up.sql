CREATE TABLE decks (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	color TEXT NOT NULL,
	commander TEXT NOT NULL,
	owner_id BIGINT NOT NULL,
	source_link TEXT NOT NULL DEFAULT ''
);

CREATE INDEX decks_owner_id_idx ON decks (owner_id);
