//go:build integration

package group

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	entity "github.com/josofm/liliana/internal/entity/group"
	"github.com/josofm/liliana/internal/migration"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func invitationDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("LILIANA_POSTGRES_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://liliana:liliana@localhost:5432/liliana_test?sslmode=disable"
	}
	admin, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { admin.Close() })
	if err := admin.Ping(); err != nil {
		t.Skipf("postgres test database unavailable: %v", err)
	}
	schema := fmt.Sprintf("invitations_test_%d", time.Now().UnixNano())
	_, err = admin.Exec(`CREATE SCHEMA ` + schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, err := admin.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); require.NoError(t, err) })
	cfg, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	// Each concurrent connection must use the isolated schema.
	cfg.RuntimeParams["search_path"] = schema
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(12)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, migration.Up(db, "../../../migrations"))
	_, err = db.Exec(`INSERT INTO users (name,email,password) VALUES ('Owner','owner@example.com','hash'),('Invited','invited@example.com','hash')`)
	require.NoError(t, err)
	return db
}

func TestPostgresInvitations(t *testing.T) {
	testInvitations(t, NewPostgresRepo(invitationDatabase(t)))
}
func TestPostgresConcurrentInvitations(t *testing.T) {
	testConcurrentInvitations(t, NewPostgresRepo(invitationDatabase(t)))
}

func TestPostgresInvitationRollback(t *testing.T) {
	db := invitationDatabase(t)
	r := NewPostgresRepo(db)
	g := &entity.Group{Name: "Atomic accept", OwnerID: 1}
	require.NoError(t, r.Create(g))
	i, err := r.CreateInvitation(g.ID, 2, 1)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE FUNCTION reject_invitation_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced update failure'; END $$;
CREATE TRIGGER reject_invitation_update BEFORE UPDATE ON group_invitations FOR EACH ROW EXECUTE FUNCTION reject_invitation_update()`)
	require.NoError(t, err)
	_, err = r.RespondInvitation(i.ID, entity.InvitationAccepted)
	require.Error(t, err)
	members, err := r.GetMembers(g.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	stored, err := r.GetInvitation(i.ID)
	require.NoError(t, err)
	require.Equal(t, entity.InvitationPending, stored.Status)
	require.Nil(t, stored.RespondedAt)
	_, err = db.Exec(`DROP TRIGGER reject_invitation_update ON group_invitations`)
	require.NoError(t, err)
	_, err = r.RespondInvitation(i.ID, entity.InvitationAccepted)
	require.NoError(t, err)
}
