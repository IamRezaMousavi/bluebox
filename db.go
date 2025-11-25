package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.WithError(err).Error("failed to open database")
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE,
			password TEXT
		)
	`)
	if err != nil {
		log.WithError(err).Error("failed to create users table")
		return nil, err
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO users
		(username, password)
		VALUES
		('reza', 'r')
	`)
	if err != nil {
		log.WithError(err).Error("failed to insert sample user")
		return nil, err
	}

	return &Database{DB: db}, nil
}

func (db *Database) IsUserExists(username string, password string) (bool, error) {
	var exists bool
	err := db.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = ? AND password = ?)`,
		username, password,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
