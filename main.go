package main

import (
	"WalletScanner/internal/checker"
	"WalletScanner/internal/core"
	"WalletScanner/internal/logger/sl"
	"WalletScanner/internal/logger/slogpretty"
	"WalletScanner/internal/save"
	"WalletScanner/internal/setup"
	"WalletScanner/internal/urls"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

const configPath = "./config/setup.yaml"

func main() {

	cfg := setup.MustLoad(configPath)

	log := setupLogger()

	log.Info(
		"starting wallet scanner",
		slog.String("version", cfg.Version),
	)

	urls, err := urls.GetURLs(log, cfg)
	if err != nil {
		log.Error("Error: GetURLs", sl.Err(err))
	}

	var walletsPerCycle = cfg.RpcProvider.BatchSize * cfg.RpcProvider.RateLimit * len(urls)

	log.Info("wallets:", slog.Int("PerCycle", walletsPerCycle))

	var counter int = 0

	for {
		secrets, err := core.GetSecrets(log, walletsPerCycle, cfg.Entropy)
		if err != nil {
			log.Error("Error: GetSecret", sl.Err(err))
		}

		walletsPerRoutine := walletsPerCycle / (len(urls) * len(urls[0]))

		secretChunks := chunkSecrets(log, secrets, len(urls)*len(urls[0]))

		var wg sync.WaitGroup

		successCh := make(chan []core.Secret, walletsPerCycle)
		emptyCh := make(chan []core.Secret, walletsPerCycle)

		totalURLs := len(urls) * len(urls[0])
		for i, secretChunk := range secretChunks {
			wg.Add(1)
			urlIndex := i % totalURLs
			go func(secretChunk []core.Secret, url string) {
				defer wg.Done()
				ticker := time.NewTicker(time.Second / time.Duration(walletsPerRoutine))
				defer ticker.Stop()

				<-ticker.C
				success, empty, err := checker.CheckBalance(log, secretChunk, url, cfg.Logging.SaveEmpty)
				if err != nil {
					log.Error("Error: CheckBalances", sl.Err(err))
					return
				}

				if len(success) > 0 {
					successCh <- success
				}

				if len(empty) > 0 && cfg.Logging.SaveEmpty {
					emptyCh <- empty
				}
			}(secretChunk, urls[urlIndex/len(urls[0])][urlIndex%len(urls[0])])
		}

		wg.Wait()

		close(successCh)
		close(emptyCh)

		for success := range successCh {
			err := save.Logs(log, cfg.Logging.Success, success)
			if err != nil {
				log.Error("Failed to save success wallets", sl.Err(err))
			}
		}
		if cfg.Logging.SaveEmpty {
			for empty := range emptyCh {
				err := save.Logs(log, cfg.Logging.Empty, empty)
				if err != nil {
					log.Error("Failed to save empty wallets", sl.Err(err))
				}
			}
		}
		counter++
		log.Info(fmt.Sprintf("Cycle №%d is complete", counter))
	}
}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	log = slog.New(handler)

	return log
}

func chunkSecrets(log *slog.Logger, secrets []core.Secret, chunkSize int) [][]core.Secret {
	var chunks [][]core.Secret
	for chunkSize < len(secrets) {
		secrets, chunks = secrets[chunkSize:], append(chunks, secrets[0:chunkSize:chunkSize])
	}
	chunks = append(chunks, secrets)
	return chunks
}
