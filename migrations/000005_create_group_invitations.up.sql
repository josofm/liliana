CREATE TABLE group_invitations (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id),
    invited_player_id BIGINT NOT NULL REFERENCES users(id),
    invited_by_player_id BIGINT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    responded_at TIMESTAMPTZ,
    CHECK ((status = 'pending' AND responded_at IS NULL)
        OR (status <> 'pending' AND responded_at IS NOT NULL))
);

CREATE UNIQUE INDEX group_invitations_pending_unique
    ON group_invitations(group_id, invited_player_id) WHERE status = 'pending';
CREATE INDEX group_invitations_inbox_idx
    ON group_invitations(invited_player_id, id) WHERE status = 'pending';
