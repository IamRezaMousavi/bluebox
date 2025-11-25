package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	app, err := NewApp(":8090", "./data.db")
	if err != nil {
		log.WithError(err).Error("canot create app")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server started on :8090")
		err = app.RunServer()
		if err != nil {
			log.WithError(err).Error("server error")
		}
	}()

	sig := <-quit
	log.WithField("signal", sig.String()).Info("received signal")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = app.Close(ctx)
	if err != nil {
		log.WithError(err).Error("error while closing the app")
	}

	log.Info("shutdown!")
}
