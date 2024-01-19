package core

import (
	"crypto/ecdsa"
	"fmt"
	"log/slog"

	"WalletScanner/internal/logger/sl"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

type Secret struct {
	Phrase     string
	PrivateKey string
	Address    string
	Balance    string
}

func GetSecrets(log *slog.Logger, wallets int, entropySize int) ([]Secret, error) {
	secrets := make([]Secret, 0, wallets)

	for i := 0; i < wallets; i++ {
		entropy, err := bip39.NewEntropy(entropySize) // Используйте entropySize вместо entropy
		if err != nil {
			log.Error("Failed to generate entropy", sl.Err(err))
			return secrets, err
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			log.Error("Failed to generate mnemonic", sl.Err(err))
			return secrets, err
		}

		seed := bip39.NewSeed(mnemonic, "")
		privateKey, err := crypto.ToECDSA(seed[:32])
		if err != nil {
			log.Error("Failed to generate privateKey", sl.Err(err))
			return secrets, err
		}

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			log.Error("Failed to assert type: publicKey is not of type *ecdsa.PublicKey")
			return secrets, fmt.Errorf("publicKey is not of type *ecdsa.PublicKey")
		}

		address := crypto.PubkeyToAddress(*publicKeyECDSA)

		secrets = append(secrets, Secret{
			Phrase:     mnemonic,
			PrivateKey: hexutil.Encode(crypto.FromECDSA(privateKey))[2:], // Удаление префикса '0x'
			Address:    address.Hex(),
			Balance:    "",
		})
	}

	return secrets, nil
}
