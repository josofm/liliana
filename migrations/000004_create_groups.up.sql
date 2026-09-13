CREATE TABLE groups (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 100),
    description TEXT NOT NULL DEFAULT '',
    owner_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_members (
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES users(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, player_id)
);

-- Ownership has one source of truth: groups.owner_id. Role is derived on read.
-- Deferred so group creation and owner membership can share one transaction.
ALTER TABLE groups ADD CONSTRAINT groups_owner_membership_fk
    FOREIGN KEY (id, owner_id) REFERENCES group_members(group_id, player_id)
    DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX group_members_player_id_idx ON group_members(player_id, group_id);
