package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate applies every embedded migration that isn't yet recorded in
// schema_migrations, in filename order, each in its own transaction. It's
// safe to call on every process start: already-applied migrations are
// skipped, so this is how the application both bootstraps a fresh database
// and keeps an existing one current.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		if err := applyMigration(db, name); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}

func applyMigration(db *sql.DB, name string) error {
	var ignored int
	err := db.QueryRow(`SELECT 1 FROM schema_migrations WHERE filename = ?`, name).Scan(&ignored)
	if err == nil {
		return nil // already applied
	}
	if err != sql.ErrNoRows {
		return err
	}

	contents, err := migrationFiles.ReadFile("migrations/" + name)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(contents)); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (filename) VALUES (?)`, name); err != nil {
		return err
	}
	return tx.Commit()
}
