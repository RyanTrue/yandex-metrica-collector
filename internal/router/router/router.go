package router

import (
	"github.com/RyanTrue/yandex-metrica-collector/internal/compressor"
	"github.com/RyanTrue/yandex-metrica-collector/internal/flags"
	log "github.com/RyanTrue/yandex-metrica-collector/internal/logger"
	"github.com/RyanTrue/yandex-metrica-collector/internal/router/handlers"
	"github.com/go-chi/chi/v5"
)

func New(params flags.Params) *chi.Mux {
	handler := handlers.New(params.DatabaseAddress, params.Key)

	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Use(compressor.Compress)
	r.Post("/update/", handler.SaveMetricFromJSON)
	r.Post("/value/", handler.GetMetricFromJSON)
	r.Post("/update/{type}/{name}/{value}", handler.SaveMetric)
	r.Get("/value/{type}/{name}", handler.GetMetric)
	r.Get("/", handler.ShowMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/updates/", handler.SaveListMetricsFromJSON)

	return r
}
