package group

import (
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	entity "github.com/josofm/liliana/internal/entity/group"
)

const invitationSelect = `SELECT i.id,i.group_id,g.name,i.invited_player_id,i.invited_by_player_id,i.status,i.created_at,i.responded_at FROM group_invitations i JOIN groups g ON g.id=i.group_id`

func scanInvitation(row scanner) (*entity.Invitation, error) {
	i := &entity.Invitation{}
	err := row.Scan(&i.ID, &i.GroupID, &i.GroupName, &i.InvitedPlayerID, &i.InvitedByPlayerID, &i.Status, &i.CreatedAt, &i.RespondedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Serialize membership checks and invitation transitions within each group.
func lockGroup(tx *sql.Tx, id int64) error {
	var lockedID int64
	err := tx.QueryRow(`SELECT id FROM groups WHERE id=$1 FOR UPDATE`, id).Scan(&lockedID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *postgresRepo) CreateInvitation(groupID, playerID, invitedByID int64) (*entity.Invitation, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := lockGroup(tx, groupID); err != nil {
		return nil, err
	}
	var member bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id=$1 AND player_id=$2)`, groupID, playerID).Scan(&member); err != nil {
		return nil, err
	}
	if member {
		return nil, ErrAlreadyMember
	}
	var id int64
	err = tx.QueryRow(`INSERT INTO group_invitations (group_id,invited_player_id,invited_by_player_id) VALUES ($1,$2,$3) RETURNING id`, groupID, playerID, invitedByID).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "group_invitations_pending_unique" {
			return nil, ErrInvitationPending
		}
		return nil, err
	}
	i, err := scanInvitation(tx.QueryRow(invitationSelect+` WHERE i.id=$1`, id))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return i, nil
}
func (r *postgresRepo) GetInvitation(id int64) (*entity.Invitation, error) {
	return scanInvitation(r.db.QueryRow(invitationSelect+` WHERE i.id=$1`, id))
}

func (r *postgresRepo) pendingInvitations(filter string, id int64) ([]entity.Invitation, error) {
	rows, err := r.db.Query(invitationSelect+` WHERE i.status='pending' AND `+filter+`=$1 ORDER BY i.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Invitation, 0)
	for rows.Next() {
		i, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *i)
	}
	return result, rows.Err()
}
func (r *postgresRepo) GetGroupInvitations(id int64) ([]entity.Invitation, error) {
	return r.pendingInvitations("i.group_id", id)
}
func (r *postgresRepo) GetPlayerInvitations(id int64) ([]entity.Invitation, error) {
	return r.pendingInvitations("i.invited_player_id", id)
}

func (r *postgresRepo) RespondInvitation(id int64, status entity.InvitationStatus) (*entity.Invitation, error) {
	if status != entity.InvitationAccepted && status != entity.InvitationDeclined && status != entity.InvitationCancelled {
		return nil, ErrInvalidInvitationStatus
	}
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	i, err := scanInvitation(tx.QueryRow(invitationSelect+` WHERE i.id=$1`, id))
	if err != nil {
		return nil, err
	}
	if err := lockGroup(tx, i.GroupID); err != nil {
		return nil, err
	}
	// Read again after acquiring the lock: a competing response may have committed.
	i, err = scanInvitation(tx.QueryRow(invitationSelect+` WHERE i.id=$1 FOR UPDATE OF i`, id))
	if err != nil {
		return nil, err
	}
	if i.Status != entity.InvitationPending {
		return nil, ErrInvitationClosed
	}
	if status == entity.InvitationAccepted {
		result, err := tx.Exec(`INSERT INTO group_members (group_id,player_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, i.GroupID, i.InvitedPlayerID)
		if err != nil {
			return nil, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, ErrAlreadyMember
		}
	}
	err = tx.QueryRow(`UPDATE group_invitations SET status=$1,responded_at=NOW() WHERE id=$2 RETURNING responded_at`, status, id).Scan(&i.RespondedAt)
	if err != nil {
		return nil, err
	}
	i.Status = status
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return i, nil
}
