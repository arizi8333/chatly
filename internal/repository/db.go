package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func ConnectDB(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Use an advisory lock to prevent concurrent migrations from multiple instances.
	// 54321 is just a random lock ID.
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("SELECT pg_advisory_xact_lock(54321)"); err != nil {
		return fmt.Errorf("could not acquire migration lock: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".sql" {
			path := filepath.Join(migrationsDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			fmt.Printf("Running migration: %s\n", entry.Name())
			if _, err := tx.Exec(string(content)); err != nil {
				return fmt.Errorf("error in %s: %w", entry.Name(), err)
			}
		}
	}

	return tx.Commit()
}
