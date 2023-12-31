package flags

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

const (
	defaultAddr            string = "127.0.0.1:8080"
	defaultReportInterval  int    = 5
	defaultPollInterval    int    = 1
	defaultStoreInterval   int    = 15
	defaultFileStoragePath string = "/tmp/short-url-db.json"
	defaultRestore         bool   = true
)

type Option func(params *Params)

func WithRateLimit() Option {
	return func(p *Params) {
		flag.IntVar(&p.RateLimit, "l", p.RateLimit, "max requests to send on server")
		if envKey := os.Getenv("RATE_LIMIT"); envKey != "" {
			p.Key = envKey
		}
	}
}

func WithKey() Option {
	return func(p *Params) {
		flag.StringVar(&p.Key, "k", p.Key, "key for using hash subscription")
		if envKey := os.Getenv("KEY"); envKey != "" {
			p.Key = envKey
		}
	}
}

func WithDatabase() Option {
	return func(p *Params) {
		result := ""
		flag.StringVar(&result, "d", p.DatabaseAddress, "connection string for db")
		if envDBAddr := os.Getenv("DATABASE_DSN"); envDBAddr != "" {
			result = envDBAddr
		}
		p.DatabaseAddress = result
	}
}

func WithAddr() Option {
	return func(p *Params) {
		flag.StringVar(&p.FlagRunAddr, "a", p.FlagRunAddr, "address and port to run server")
		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			p.FlagRunAddr = envRunAddr
		}
	}
}

func WithReportInterval() Option {
	return func(p *Params) {
		flag.IntVar(&p.ReportInterval, "r", p.ReportInterval, "report interval")
		if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
			reportIntervalEnv, err := strconv.Atoi(envReportInterval)
			if err == nil {
				p.ReportInterval = reportIntervalEnv
			}
		}
	}
}

func WithPollInterval() Option {
	return func(p *Params) {
		flag.IntVar(&p.PollInterval, "p", p.PollInterval, "poll interval")
		if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
			pollIntervalEnv, err := strconv.Atoi(envPollInterval)
			if err == nil {
				p.PollInterval = pollIntervalEnv
			}
		}
	}
}

func WithStoreInterval() Option {
	return func(p *Params) {
		flag.IntVar(&p.StoreInterval, "i", p.StoreInterval, "store interval in seconds")
		if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
			storeIntervalEnv, err := strconv.Atoi(envStoreInterval)
			if err == nil {
				p.StoreInterval = storeIntervalEnv
			}
		}
	}
}

func WithFileStoragePath() Option {
	return func(p *Params) {
		flag.StringVar(&p.FileStoragePath, "f", p.FileStoragePath, "file name for metrics collection")
		if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
			fileStoragePath, err := strconv.Atoi(envFileStoragePath)
			if err == nil {
				p.StoreInterval = fileStoragePath
			}
		}
	}
}

func WithRestore() Option {
	return func(p *Params) {
		flag.BoolVar(&p.Restore, "r", p.Restore, "restore data from file")
		if envRestore := os.Getenv("RESTORE"); envRestore != "" {
			restore, err := strconv.Atoi(envRestore)
			if err == nil {
				p.StoreInterval = restore
			}
		}
	}
}

func WithTLSKeyPath() Option {
	return func(p *Params) {
		flag.StringVar(&p.CryptoKeyPath, "crypto-key", p.CryptoKeyPath, "crypto key path")
		if envCryptoKeyPath := os.Getenv("CRYPTO_KEY"); envCryptoKeyPath != "" {
			p.CryptoKeyPath = envCryptoKeyPath
		}
	}
}

func WithConfig() Option {
	return func(p *Params) {
		var configPath string
		flag.StringVar(&configPath, "c", "", "config path")
		for i, arg := range os.Args {
			if arg == "-c" || arg == "-config" {
				configPath = os.Args[i+1]
			}
		}
		// priority for the env variables
		if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
			configPath = envConfigPath
		}
		if configPath != "" {
			config, err := os.ReadFile(configPath)
			if err != nil {
				log.Printf("config path was provided, but an error ocurred while opening: %s\n", err.Error())
				log.Println("using default values, values from command line and from env variables...")
				return
			}
			if err = json.Unmarshal(config, p); err != nil {
				log.Printf("error while parsing config: %s\n", err.Error())
			}
		}
	}
}

func Init(opts ...Option) *Params {
	p := &Params{
		RateLimit:       1,
		FlagRunAddr:     defaultAddr,
		ReportInterval:  defaultReportInterval,
		PollInterval:    defaultPollInterval,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         defaultRestore,
	}

	for _, opt := range opts {
		opt(p)
	}
	flag.Parse()
	return p
}

// Params is a struct for storing run parameters
type Params struct {
	FlagRunAddr     string `json:"address"`         // address and port to run server
	DatabaseAddress string `json:"database_dsn"`    // database address
	ReportInterval  int    `json:"report_interval"` // time interval for sending metrics to the server
	PollInterval    int    `json:"poll_interval"`   // time interval for capturing metrics
	StoreInterval   int    `json:"store_interval"`  // time interval for saving metrics in the db/file
	FileStoragePath string `json:"store_file"`      // path for file to store metrics
	Restore         bool   `json:"restore"`         // is need to restore metrics from db/file
	Key             string `json:"hash_key"`        // key for using hash subscription
	RateLimit       int    `json:"rate_limit"`      // rate limit for querying server
	CryptoKeyPath   string `json:"crypto_key"`      // tls key path
}
