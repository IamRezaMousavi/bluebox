package main

import log "github.com/sirupsen/logrus"

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	app, err := NewApp(":8090", "./data.db")
	if err != nil {
		log.WithError(err).Error("canot create app")
	}

	log.Info("server started on :8090")
	app.Server.ListenAndServe()
}
