package main

import (
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type App struct {
	DB     *Database
	C      *Cache
	Server *http.Server
}

func NewApp(addr string, dbPath string, cacheAddr string) (*App, error) {
	app := &App{}

	db, err := NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	app.DB = db

	cache, err := NewCache(cacheAddr)
	if err != nil {
		return nil, err
	}
	app.C = cache

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", app.hello)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/validate", app.validate)
	mux.HandleFunc("/protected", app.protected)

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
	if err != nil {
		return err
	}

	err = app.C.Close()
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
		otp := generateOTP()
		err := app.C.SetOTP(body.Username, otp)
		if err != nil {
			http.Error(w, "cache error", http.StatusInternalServerError)
			log.WithError(err).Error("failed to store user otp")
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"otp":    otp,
		})
		log.WithFields(log.Fields{
			"method":   req.Method,
			"path":     req.URL.Path,
			"username": body.Username,
		}).Info("auth successful")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"status": "invalid username or password"})
		log.WithFields(log.Fields{
			"method":   req.Method,
			"path":     req.URL.Path,
			"username": body.Username,
		}).Error("auth failed")
	}
}

func (app *App) validate(w http.ResponseWriter, req *http.Request) {
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
		Otp      string `json:"otp"`
	}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		log.WithError(err).Error("invalid json")
		return
	}

	storedOtp, err := app.C.GetOTP(body.Username)
	if err != nil {
		http.Error(w, "otp expired or not found", http.StatusUnauthorized)
		log.WithField("username", body.Username).WithError(err).Error("user otp expired or not found")
		return
	}

	if storedOtp != body.Otp {
		http.Error(w, "invalid otp", http.StatusUnauthorized)
		log.WithField("username", body.Username).WithError(err).Error("invalid otp")
		return
	}

	tokenString, err := createToken(body.Username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		log.WithField("username", body.Username).WithError(err).Error("create token error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"token":  tokenString,
	})
	log.WithFields(log.Fields{
		"method":   req.Method,
		"path":     req.URL.Path,
		"username": body.Username,
	}).Info("validate successful")
}

func (app *App) protected(w http.ResponseWriter, req *http.Request) {
	tokenString := req.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		log.WithFields(log.Fields{
			"method": req.Method,
			"path":   req.URL.Path,
		}).Error("missing auth header")
		return
	}

	tokenString = tokenString[len("Bearer "):]
	err := verifyToken(tokenString)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		log.WithFields(log.Fields{
			"method": req.Method,
			"path":   req.URL.Path,
		}).Error("invalid token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	log.WithFields(log.Fields{
		"method": req.Method,
		"path":   req.URL.Path,
	}).Info("protected called")
}
