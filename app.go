package main

import (
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type App struct {
	DB     *Database
	Server *http.Server
}

func NewApp(addr string, dbPath string) (*App, error) {
	app := &App{}

	db, err := NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	app.DB = db

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", app.hello)
	mux.HandleFunc("/login", app.login)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	app.Server = server
	return app, nil
}

func (app *App) RunServer() error {
	err := app.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (app *App) Close(ctx context.Context) error {
	var err error

	err = app.Server.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = app.DB.Close()
	return nil
}

func (app *App) hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "hello"})
	log.WithFields(log.Fields{
		"method": req.Method,
		"path":   req.URL.Path,
	}).Info("say hello")
}

func (app *App) login(w http.ResponseWriter, req *http.Request) {
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

	exists, err := app.DB.IsUserExists(body.Username, body.Password)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		log.WithError(err).Error("failed to check user exists")
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
