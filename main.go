package main

import (
	"log/slog"
	"net/http"
	"os"
	"servicesubs/internal/config"
	"servicesubs/internal/database/pgsql"
	"servicesubs/internal/server"

	_ "servicesubs/docs"

	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title			Service Subs
// @version		1.0
// @description	API Server for Subs users
func main() {
	cfg, err := config.New()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(cfg.LOG.LogLevel),
	})))

	err = pgsql.Init(cfg)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer pgsql.CloseDB()

	go func() {
		slog.Debug("Start swagger")
		r := chi.NewRouter()
		//r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:1323/swagger/doc.json")))
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:1323/swagger/doc.json"), // или относительный путь
			httpSwagger.DeepLinking(true),                             // разрешить глубокие ссылки
			httpSwagger.DocExpansion("list"),                          // 'list', 'full' или 'none'
			httpSwagger.PersistAuthorization(true),                    // сохранять авторизацию при обновлении
		))
		if err := http.ListenAndServe(":1323", r); err != nil {
			slog.Error("Swagger server failed", "error", err)
		}
	}()

	server.Start(cfg)
}
