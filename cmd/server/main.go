package main

import (
	"context"

	"github.com/RyanTrue/yandex-metrica-collector/internal/flags"
	"github.com/RyanTrue/yandex-metrica-collector/internal/runner/server"
)

func main() {
	params := flags.Init(
		flags.WithConfig(),
		flags.WithAddr(),
		flags.WithStoreInterval(),
		flags.WithFileStoragePath(),
		flags.WithRestore(),
		flags.WithDatabase(),
		flags.WithKey(),
		flags.WithTLSKeyPath(),
	)
	ctx := context.Background()
	serverRunner := server.New(params)

	serverRunner.Run(ctx)
}
