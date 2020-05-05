// Copyright 2020 Celer Network

package utils

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func SigIsValid(signer ctype.Addr, data []byte, sig []byte) bool {
	recoveredAddr := RecoverSigner(data, sig)
	return recoveredAddr == signer
}

func RecoverSigner(data []byte, sig []byte) ctype.Addr {
	if len(sig) == 65 { // we could return zeroAddr if len not 65
		if sig[64] == 27 || sig[64] == 28 {
			// SigToPub only expect v to be 0 or 1, see https://github.com/ethereum/go-ethereum/blob/v1.8.23/internal/ethapi/api.go#L468
			// we've been ok as our own code only has v 0 or 1, but using external signer may cause issue
			// we also fix v in celersdk.PublishSignedResult to be extra safe
			sig[64] -= 27
		}
	}
	pubKey, err := crypto.SigToPub(GeneratePrefixedHash(data), sig)
	if err != nil {
		log.Errorf("sig error: %v, sig: %x", err, sig)
		return ctype.Addr{}
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr
}

func GeneratePrefixedHash(data []byte) []byte {
	return crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n32"), crypto.Keccak256(data))
}

func GetAddrAndPrivKey(keyStore string, passPhrase string) (ctype.Addr, string, error) {
	key, err := keystore.DecryptKey([]byte(keyStore), passPhrase)
	if err != nil {
		return ctype.ZeroAddr, "", err
	}
	privKey := ctype.Bytes2Hex(crypto.FromECDSA(key.PrivateKey))
	return key.Address, privKey, nil
}

func WaitMined(ctx context.Context, ec *ethclient.Client,
	tx *ethtypes.Transaction, blockDelay uint64) (*ethtypes.Receipt, error) {
	return WaitMinedWithTxHash(ctx, ec, tx.Hash().Hex(), blockDelay)
}

// WaitMined waits for tx to be mined on the blockchain
// It returns tx receipt when the tx has been mined and enough block confirmations have passed
func WaitMinedWithTxHash(ctx context.Context, ec *ethclient.Client,
	txHash string, blockDelay uint64) (*ethtypes.Receipt, error) {
	// an error possibly returned when a transaction is pending
	const missingFieldErr = "missing required field 'transactionHash' for Log"

	if ec == nil {
		return nil, errors.New("nil ethclient")
	}
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()
	// wait tx to be mined
	txHashBytes := ethcommon.HexToHash(txHash)
	for {
		receipt, rerr := ec.TransactionReceipt(ctx, txHashBytes)
		if rerr == nil {
			log.Debugf("Transaction mined. Waiting for %d block confirmations", blockDelay)
			if blockDelay == 0 {
				return receipt, rerr
			}
			break
		} else if rerr == ethereum.NotFound || rerr.Error() == missingFieldErr {
			// Wait for the next round
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-queryTicker.C:
			}
		} else {
			return receipt, rerr
		}
	}
	// wait for enough block confirmations
	ddl := big.NewInt(0)
	latestBlockHeader, err := ec.HeaderByNumber(ctx, nil)
	if err == nil {
		ddl.Add(new(big.Int).SetUint64(blockDelay), latestBlockHeader.Number)
	}
	for {
		latestBlockHeader, err := ec.HeaderByNumber(ctx, nil)
		if err == nil && ddl.Cmp(latestBlockHeader.Number) < 0 {
			receipt, rerr := ec.TransactionReceipt(ctx, txHashBytes)
			if rerr == nil {
				log.Debugln("tx confirmed!")
				return receipt, rerr
			} else if rerr == ethereum.NotFound || rerr.Error() == missingFieldErr {
				return nil, errors.New("tx is dropped due to chain re-org")
			} else {
				return receipt, rerr
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}
