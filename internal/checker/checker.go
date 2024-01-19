package checker

import (
	"context"
	"log/slog"
	"time"

	"WalletScanner/internal/core"
	"WalletScanner/internal/logger/sl"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

func CheckBalance(log *slog.Logger, secrets []core.Secret, url string, saveEmpty bool) ([]core.Secret, []core.Secret, error) {
	rpcClient, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		log.Error("Failed to get rpcClient", sl.Err(err))
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var balanceRequests []rpc.BatchElem
	for _, secret := range secrets {
		address := common.HexToAddress(secret.Address)
		balanceRequests = append(balanceRequests, rpc.BatchElem{
			Method: "eth_getBalance",
			Args:   []interface{}{address, "latest"},
			Result: new(string),
		})
	}

	err = rpcClient.BatchCallContext(ctx, balanceRequests)
	if err != nil {
		return nil, nil, nil
	}

	var successSecrets []core.Secret
	var otherSecrets []core.Secret
	for i, req := range balanceRequests {
		if req.Error != nil {
			log.Error("Request error", slog.String("error", req.Error.Error()))
			otherSecrets = append(otherSecrets, secrets[i])
			continue
		}
		balanceStr := *(req.Result.(*string))
		secrets[i].Balance = balanceStr
		if balanceStr != "0x0" {
			successSecrets = append(successSecrets, secrets[i])
		} else if saveEmpty {
			otherSecrets = append(otherSecrets, secrets[i])
		}
	}

	return successSecrets, otherSecrets, nil
}
