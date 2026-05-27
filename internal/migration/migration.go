package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version int64
	Name    string
	Path    string
}

func Up(db *sql.DB, dir string) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}

	migrations, err := loadMigrations(dir)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		applied, err := migrationApplied(db, m.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if err := applyMigration(db, m); err != nil {
			return err
		}
		fmt.Printf("applied migration %06d %s\n", m.Version, m.Name)
	}

	return nil
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func loadMigrations(dir string) ([]Migration, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	migrations := make([]Migration, 0, len(files))
	for _, path := range files {
		base := filepath.Base(path)
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration filename %q", base)
		}

		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %q: %w", base, err)
		}

		name := strings.TrimSuffix(parts[1], ".up.sql")
		if name == parts[1] {
			return nil, fmt.Errorf("invalid up migration filename %q", base)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			Path:    path,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version == migrations[i-1].Version {
			return nil, fmt.Errorf("duplicate migration version %06d", migrations[i].Version)
		}
	}

	return migrations, nil
}

func migrationApplied(db *sql.DB, version int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
	return exists, err
}

func applyMigration(db *sql.DB, m Migration) error {
	statement, err := os.ReadFile(m.Path)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer rollback(tx)

	if _, err := tx.Exec(string(statement)); err != nil {
		return fmt.Errorf("apply migration %06d %s: %w", m.Version, m.Name, err)
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		m.Version,
		m.Name,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func rollback(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		fmt.Fprintln(os.Stderr, err)
	}
}
