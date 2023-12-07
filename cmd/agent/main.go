package main

import (
	"context"
	"fmt"
	agent2 "github.com/RyanTrue/yandex-metrica-collector/internal/agent"
	"github.com/RyanTrue/yandex-metrica-collector/internal/collector"
	"github.com/RyanTrue/yandex-metrica-collector/internal/flags"
	log "github.com/RyanTrue/yandex-metrica-collector/internal/logger"
	aggregator "github.com/RyanTrue/yandex-metrica-collector/internal/metrics"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"os"
)

func main() {
	params := flags.Init(
		flags.WithPollInterval(),
		flags.WithReportInterval(),
		flags.WithAddr(),
		flags.WithKey(),
		flags.WithRateLimit(),
	)

	errs, ctx := errgroup.WithContext(context.Background())

	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()
	log.SugarLogger = *logger.Sugar()

	agent := agent2.New(params, aggregator.New(&collector.Collector), log.SugarLogger)
	errs.Go(func() error {
		return agent.CollectMetrics(ctx)
	})
	errs.Go(func() error {
		return agent.SendMetrics(ctx)
	})
	if err := errs.Wait(); err != nil {
		log.SugarLogger.Errorf(fmt.Sprintf("error while running agent: %s", err.Error()))
	}
}
