package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/RyanTrue/yandex-metrica-collector/internal/collector"
	"github.com/RyanTrue/yandex-metrica-collector/internal/flags"
	aggregator "github.com/RyanTrue/yandex-metrica-collector/internal/metrics"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// CollectMetrics - is a method of the Agent struct for capturing runtime and gopsutil metrics.
func (a *Agent) CollectMetrics(ctx context.Context) {
	aggTicker := time.NewTicker(time.Duration(a.params.PollInterval) * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-aggTicker.C:
				a.aggregator.AggregateRuntimeMetrics()
				a.aggregator.AggregateGopsutilMetrics()
			}
		}
	}()
}

// SendMetricsLoop - a method for sending metrics to the server by timer
func (a *Agent) SendMetricsLoop(ctx context.Context) (err error) {
	numRequests := make(chan struct{}, a.params.RateLimit)
	reportTicker := time.NewTicker(time.Duration(a.params.ReportInterval) * time.Second)
	// send metrics by timer
	for {
		select {
		case <-ctx.Done():
			return
		// check if time to send metrics on server
		case <-reportTicker.C:
			select {
			// check if the rate limit is exceeded
			case numRequests <- struct{}{}:
				if err = a.SendMetrics(ctx); err != nil {
					return err
				}
			default:
				a.log.Info("rate limit is exceeded")
			}
		}
		// release the request pool for one
		<-numRequests
	}
}

// SendMetrics - a method that encapsulates the logic for sending a http request to the server.
func (a *Agent) SendMetrics(ctx context.Context) error {
	req := a.client.SetRetryCount(3).R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip")

	for _, v := range collector.Collector().Metrics {
		jsonInput, _ := json.Marshal(collector.MetricRequest{
			ID:    v.ID,
			MType: v.MType,
			Delta: v.CounterValue,
			Value: v.GaugeValue,
		})
		if a.params.Key != "" {
			//h := hmac.New(sha256.New, []byte(key))
			//h.Write(jsonInput)
			//dst := h.Sum(nil)
			dst := sha256.Sum256(jsonInput)
			req.SetHeader("HashSHA256", fmt.Sprintf("%x", dst))
		}
		message := string(jsonInput)
		if a.cryptoKey != nil {
			encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, a.cryptoKey, jsonInput)
			if err != nil {
				return fmt.Errorf("error encrypting message with public key: %w", err)
			}
			message = string(encryptedData)
		}
		if err := a.sendRequestsWithRetries(req, message); err != nil {
			return fmt.Errorf("error while sending agent request for counter metric: %w", err)
		}
	}
	a.log.Info("metrics sent on server")
	return nil
}

// SendMetrics - a method that implements the logic for sending a request with retries.
func (a *Agent) sendRequestsWithRetries(req *resty.Request, jsonInput string) error {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	if _, err := zb.Write([]byte(jsonInput)); err != nil {
		return fmt.Errorf("error while write json input: %w", err)
	}
	if err := zb.Close(); err != nil {
		return fmt.Errorf("error while trying to close writer: %w", err)
	}
	if _, err := req.
		SetBody(buf).
		Post(fmt.Sprintf("http://%s/update/", a.params.FlagRunAddr)); err != nil {
		return fmt.Errorf("error while trying to create post request: %w", err)
	}
	return nil
}

// New is a method for creating Agent object.
func New(params *flags.Params, aggregator *aggregator.Aggregator, log *zap.SugaredLogger) (*Agent, error) {
	agent := &Agent{
		params:     params,
		aggregator: aggregator,
		log:        log,
		client:     resty.New(),
	}
	if params.CryptoKeyPath != "" {
		b, err := os.ReadFile(params.CryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error while reading file with crypto public key: %w", err)
		}
		block, _ := pem.Decode(b)
		publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing public key: %w", err)
		}
		agent.cryptoKey = publicKey.(*rsa.PublicKey)
	}
	return agent, nil
}

// Agent is a struct for capturing and sending metrics to the server.
type Agent struct {
	params     *flags.Params
	aggregator *aggregator.Aggregator
	cryptoKey  *rsa.PublicKey
	log        *zap.SugaredLogger
	client     *resty.Client
}
