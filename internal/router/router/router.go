package router

import (
	"fmt"

	"github.com/RyanTrue/yandex-metrica-collector/internal/flags"
	"github.com/RyanTrue/yandex-metrica-collector/internal/middlewares/compressor"
	log "github.com/RyanTrue/yandex-metrica-collector/internal/middlewares/logger"
	"github.com/RyanTrue/yandex-metrica-collector/internal/router/handlers"
	"github.com/go-chi/chi/v5"
)

func New(params flags.Params) (*chi.Mux, error) {
	handler, err := handlers.New(params.DatabaseAddress, params.Key, params.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error while creating handler: %w", err)
	}

	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Use(compressor.Compress)
	r.Use(handler.CheckSubscription)
	r.Post("/update/", handler.SaveMetricFromJSON)
	r.Post("/value/", handler.GetMetricFromJSON)
	r.Post("/update/{type}/{name}/{value}", handler.SaveMetric)
	r.Get("/value/{type}/{name}", handler.GetMetric)
	r.Get("/", handler.ShowMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/updates/", handler.SaveListMetricsFromJSON)

	return r, nil
}
