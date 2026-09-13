//go:build integration

package group

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	entity "github.com/josofm/liliana/internal/entity/group"
	"github.com/josofm/liliana/internal/migration"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestPostgresGroupConstraints(t *testing.T) {
	dsn := os.Getenv("LILIANA_POSTGRES_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://liliana:liliana@localhost:5432/liliana_test?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skipf("postgres test database unavailable: %v", err)
	}
	// Isolate this suite from other packages that truncate users.
	schema := fmt.Sprintf("groups_test_%d", time.Now().UnixNano())
	_, err = db.Exec(`CREATE SCHEMA ` + schema)
	require.NoError(t, err)
	defer db.Exec(`DROP SCHEMA ` + schema + ` CASCADE`)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	_, err = db.Exec(`SET search_path TO ` + schema)
	require.NoError(t, err)
	require.NoError(t, migration.Up(db, "../../../migrations"))
	_, err = db.Exec(`INSERT INTO users (name,email,password) VALUES ('Owner','owner@example.com','hash'),('Member','member@example.com','hash')`)
	require.NoError(t, err)
	r := NewPostgresRepo(db)
	g := &entity.Group{Name: "Commander Night", OwnerID: 1}
	require.NoError(t, r.Create(g))
	members, err := r.GetMembers(g.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, entity.RoleOwner, members[0].Role)
	require.NoError(t, r.AddMember(g.ID, 2))
	require.ErrorIs(t, r.AddMember(g.ID, 2), ErrAlreadyMember)
	other := &entity.Group{Name: "Other", OwnerID: 2}
	require.NoError(t, r.Create(other))
	groups, err := r.GetByPlayerID(2)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	g.Name = "Updated"
	require.NoError(t, r.Update(g))
	stored, err := r.GetByID(g.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated", stored.Name)
	require.ErrorIs(t, r.RemoveMember(g.ID, 1), ErrOwnerCannotLeave)
	_, err = db.Exec(`DELETE FROM group_members WHERE group_id=$1 AND player_id=1`, g.ID)
	require.Error(t, err)
	_, err = db.Exec(`INSERT INTO groups (name,owner_id) VALUES ('Missing membership',1)`)
	require.Error(t, err)
	require.Error(t, r.Create(&entity.Group{Name: "Invalid owner", OwnerID: 999}))
	require.Error(t, r.AddMember(g.ID, 999))
	var count int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM groups`).Scan(&count))
	require.Equal(t, 2, count)
	require.NoError(t, r.RemoveMember(g.ID, 2))
	require.ErrorIs(t, r.RemoveMember(g.ID, 2), ErrMemberNotFound)
}
