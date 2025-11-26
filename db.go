package main

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
)

type Database struct {
	*sql.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.WithError(err).Error("failed to open database")
		return nil, err
	}

	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Warn("waiting for database...")
		time.Sleep(time.Second)
	}

	err = db.Ping()
	if err != nil {
		log.WithError(err).Error("failed to ping database")
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT UNIQUE,
			password TEXT
		)
	`)
	if err != nil {
		log.WithError(err).Error("failed to create users table")
		return nil, err
	}

	_, err = db.Exec(`
		INSERT INTO users
		(username, password)
		VALUES
		('reza', 'r')
		ON CONFLICT (username) DO NOTHING
	`) // we should use hashing algs for password
	if err != nil {
		log.WithError(err).Error("failed to insert sample user")
		return nil, err
	}

	return &Database{DB: db}, nil
}

func (db *Database) IsUserExists(username string, password string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND password = $2)`,
		username, password,
	).Scan(&exists)
	return exists, err
}
