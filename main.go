package main

import (
	"log/slog"
	"servicesubs/internal/config"
	"servicesubs/internal/database/pgsql"
	"servicesubs/internal/server"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		return
	}

	err = pgsql.Init(cfg)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer pgsql.CloseDB()

	server.Start(cfg)
}
