package save

import (
	"WalletScanner/internal/core"
	"WalletScanner/internal/logger/sl"
	"fmt"
	"log/slog"
	"os"
)

func Logs(log *slog.Logger, path string, success []core.Secret) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error("Failed to open file", sl.Err(err))
		return err
	}
	defer file.Close()

	for _, secret := range success {
		_, err := file.WriteString(fmt.Sprintf("Mnemonic: %s,\n PrivateKey: %s,\n Address: %s,\n Balance: %s\n\n", secret.Phrase, secret.PrivateKey, secret.Address, secret.Balance))
		if err != nil {
			log.Error("Failed to write file", sl.Err(err))
			return err
		}
	}

	return nil
}
