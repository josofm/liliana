package group

import (
	"database/sql"
	"errors"
	entity "github.com/josofm/liliana/internal/entity/group"
)

type postgresRepo struct{ db *sql.DB }

func NewPostgresRepo(db *sql.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) Create(g *entity.Group) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = tx.QueryRow(`INSERT INTO groups (name,description,owner_id) VALUES ($1,$2,$3) RETURNING id,created_at,updated_at`, g.Name, g.Description, g.OwnerID).Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO group_members (group_id,player_id) VALUES ($1,$2)`, g.ID, g.OwnerID); err != nil {
		return err
	}
	return tx.Commit()
}

type scanner interface{ Scan(...any) error }

func scanGroup(row scanner) (*entity.Group, error) {
	g := &entity.Group{}
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.OwnerID, &g.CreatedAt, &g.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return g, nil
}
func (r *postgresRepo) GetByID(id int64) (*entity.Group, error) {
	return scanGroup(r.db.QueryRow(`SELECT id,name,description,owner_id,created_at,updated_at FROM groups WHERE id=$1`, id))
}
func (r *postgresRepo) GetByPlayerID(id int64) ([]*entity.Group, error) {
	rows, err := r.db.Query(`SELECT g.id,g.name,g.description,g.owner_id,g.created_at,g.updated_at FROM groups g JOIN group_members m ON m.group_id=g.id WHERE m.player_id=$1 ORDER BY g.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*entity.Group, 0)
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, rows.Err()
}
func (r *postgresRepo) Update(g *entity.Group) error {
	stored, err := scanGroup(r.db.QueryRow(`UPDATE groups SET name=$1,description=$2,updated_at=NOW() WHERE id=$3 RETURNING id,name,description,owner_id,created_at,updated_at`, g.Name, g.Description, g.ID))
	if err != nil {
		return err
	}
	*g = *stored
	return nil
}
func (r *postgresRepo) GetMembers(id int64) ([]entity.Member, error) {
	if _, err := r.GetByID(id); err != nil {
		return nil, err
	}
	rows, err := r.db.Query(`SELECT m.group_id,m.player_id,CASE WHEN m.player_id=g.owner_id THEN 'owner' ELSE 'member' END,m.joined_at FROM group_members m JOIN groups g ON g.id=m.group_id WHERE m.group_id=$1 ORDER BY m.player_id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Member, 0)
	for rows.Next() {
		var m entity.Member
		if err := rows.Scan(&m.GroupID, &m.PlayerID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}
func (r *postgresRepo) AddMember(id, playerID int64) error {
	if _, err := r.GetByID(id); err != nil {
		return err
	}
	result, err := r.db.Exec(`INSERT INTO group_members (group_id,player_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, playerID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAlreadyMember
	}
	return nil
}
func (r *postgresRepo) RemoveMember(id, playerID int64) error {
	g, err := r.GetByID(id)
	if err != nil {
		return err
	}
	if g.OwnerID == playerID {
		return ErrOwnerCannotLeave
	}
	result, err := r.db.Exec(`DELETE FROM group_members WHERE group_id=$1 AND player_id=$2`, id, playerID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrMemberNotFound
	}
	return nil
}
