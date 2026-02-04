package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("failed to select dialect: %v", err)
	}

	curr, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %v", err)
	}

	log.Printf("Current migration version: %d", curr)

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	newVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return err
	}

	if newVersion > curr {
		log.Printf("Successfully migrated from version %d to %d", curr, newVersion)
	} else {
		log.Printf("Database is already up to date at version %d", curr)
	}

	return nil
}
