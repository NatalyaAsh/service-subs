package server

import (
	"log/slog"
	"net/http"
	"os"
	"servicesubs/internal/api"
	"servicesubs/internal/config"
)

func Start(cfg *config.Config) {

	mux := http.NewServeMux()

	api.Init(mux, cfg)

	slog.Debug("Start server")
	err := http.ListenAndServe(":"+cfg.HTTP.Port, mux)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

}
