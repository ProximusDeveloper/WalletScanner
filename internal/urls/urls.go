package urls

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"WalletScanner/internal/logger/sl"
	"WalletScanner/internal/setup"
)

func GetURLs(log *slog.Logger, cfg *setup.Config) ([][]string, error) {
	var urlsSlice [][]string
	apiKeys, err := readAPIKeys(log, cfg.RpcProvider.ApiKeysPath)
	if err != nil {
		log.Error("Failed to read API keys", sl.Err(err))
		return nil, err
	}
	for _, apiKey := range apiKeys {
		var urls []string
		for _, chain := range cfg.Chains {
			url := fmt.Sprintf("%s%s/%s", cfg.RpcProvider.URL, chain.Endpoint, apiKey)
			urls = append(urls, url)
		}
		urlsSlice = append(urlsSlice, urls)
	}
	return urlsSlice, nil
}

func readAPIKeys(log *slog.Logger, filePath string) ([]string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Error("File %s does not exist", sl.Err(err))
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Failed to read API keys", sl.Err(err))
		return nil, err
	}
	defer file.Close()

	var apiKeys []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		apiKey := strings.TrimSpace(scanner.Text())
		if apiKey != "" {
			apiKeys = append(apiKeys, apiKey)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("Scanner error:", sl.Err(err))
		return nil, err
	}
	return apiKeys, nil
}
