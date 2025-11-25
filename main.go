package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type App struct {
	DB     *sql.DB
	Server *http.Server
}

func NewApp() (*App, error) {
	app := &App{}

	db, err := sql.Open("sqlite3", "./data.db")
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

	app.DB = db

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", app.hello)
	mux.HandleFunc("/login", app.login)

	server := &http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	app.Server = server
	return app, nil
}

func (a *App) hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "hello"})
	log.WithFields(log.Fields{
		"method": req.Method,
		"path":   req.URL.Path,
	}).Info("say hello")
}

func (a *App) login(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		log.WithFields(log.Fields{
			"method": req.Method,
			"path":   req.URL.Path,
		}).Error("invalid method")
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		log.WithError(err).Error("invalid json")
		return
	}

	var exists bool
	err = a.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = ? AND password = ?)`,
		body.Username, body.Password,
	).Scan(&exists)

	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		log.WithError(err).Error("database query error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if exists {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		log.WithFields(log.Fields{
			"method":   req.Method,
			"path":     req.URL.Path,
			"username": body.Username,
		}).Info("login successful")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"status": "invalid username or password"})
		log.WithFields(log.Fields{
			"method":   req.Method,
			"path":     req.URL.Path,
			"username": body.Username,
		}).Error("login failed")
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	app, err := NewApp()
	if err != nil {
		log.WithError(err).Error("canot create app")
	}

	log.Info("server started on :8090")
	app.Server.ListenAndServe()
}
