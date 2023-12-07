package main

import (
	"context"

	"github.com/RyanTrue/yandex-metrica-collector/internal/flags"
	"github.com/RyanTrue/yandex-metrica-collector/internal/runner/agent"
)

func main() {
	params := flags.Init(
		flags.WithConfig(),
		flags.WithPollInterval(),
		flags.WithReportInterval(),
		flags.WithAddr(),
		flags.WithKey(),
		flags.WithRateLimit(),
		flags.WithTLSKeyPath(),
	)
	ctx := context.Background()
	runner := agent.New(params)

	runner.Run(ctx)
}
