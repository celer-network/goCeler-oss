// Copyright 2018-2019 Celer Network

package transactor

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/rtconfig"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	parityErrIncrementNonce = "incrementing the nonce"
)

type Transactor struct {
	address    common.Address
	privKey    string
	keyStore   string
	passPhrase string
	chainId    *big.Int
	client     *ethclient.Client
	blockDelay uint64
	nonce      uint64
	sentTx     bool
	lock       sync.Mutex
}

type TransactionMinedHandler struct {
	OnMined func(receipt *types.Receipt)
}

func NewTransactor(
	keyStore string,
	passPhrase string,
	chainId *big.Int,
	client *ethclient.Client,
	blockDelay uint64,
) (*Transactor, error) {
	key, err := keystore.DecryptKey([]byte(keyStore), passPhrase)
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%x", key.Address)
	privKey := ctype.Bytes2Hex(crypto.FromECDSA(key.PrivateKey))
	return &Transactor{
		address:    common.HexToAddress(addr),
		privKey:    privKey,
		keyStore:   keyStore,
		passPhrase: passPhrase,
		chainId:    chainId,
		client:     client,
		blockDelay: blockDelay,
	}, nil
}

func (t *Transactor) Transact(
	handler *TransactionMinedHandler,
	value *big.Int,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Transaction, error) {
	return t.transact(handler, value, 0 /* gasLimit */, false /* quickCatch */, method)
}

func (t *Transactor) TransactWithQuickCatch(
	handler *TransactionMinedHandler,
	value *big.Int,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Transaction, error) {
	return t.transact(handler, value, 0 /* gasLimit */, true /* quickCatch */, method)
}

func (t *Transactor) TransactWithGasLimit(
	handler *TransactionMinedHandler,
	value *big.Int,
	gasLimit uint64,
	quickCatch bool,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Transaction, error) {
	return t.transact(handler, value, gasLimit, quickCatch, method)
}

func (t *Transactor) transact(
	handler *TransactionMinedHandler,
	value *big.Int,
	gasLimit uint64,
	quickCatch bool,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Transaction, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	signer, err := newTransactOpts(strings.NewReader(t.keyStore), t.passPhrase, t.chainId)
	if err != nil {
		return nil, err
	}
	client := t.client
	suggestedPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	signer.GasPrice = suggestedPrice
	minGas := rtconfig.GetMinGasGwei()
	maxGas := rtconfig.GetMaxGasGwei()
	if minGas > 0 { // gas can't be lower than minGas
		minPrice := new(big.Int).SetUint64(minGas * 1e9) // 1e9 is 1G
		// minPrice is larger than suggested, use minPrice
		if minPrice.Cmp(suggestedPrice) > 0 {
			signer.GasPrice = minPrice
		} else {
			signer.GasPrice = suggestedPrice
		}
	}
	if maxGas > 0 { // maxGas 0 means no cap on gas price, otherwise won't set bigger than it
		capPrice := new(big.Int).SetUint64(maxGas * 1e9) // 1e9 is 1G
		// GasPrice is larger than allowed cap, set to cap
		if capPrice.Cmp(signer.GasPrice) < 0 {
			signer.GasPrice = capPrice
		}
	}
	signer.GasLimit = gasLimit
	signer.Value = value
	pendingNonce, err := t.client.PendingNonceAt(context.Background(), t.address)
	if err != nil {
		return nil, err
	}
	if pendingNonce > t.nonce || !t.sentTx {
		t.nonce = pendingNonce
	} else {
		t.nonce++
	}
	for {
		nonceInt := big.NewInt(0)
		nonceInt.SetUint64(t.nonce)
		signer.Nonce = nonceInt
		tx, err := method(client, signer)
		if err != nil {
			errStr := err.Error()
			if errStr == core.ErrNonceTooLow.Error() ||
				errStr == core.ErrReplaceUnderpriced.Error() ||
				strings.Contains(errStr, parityErrIncrementNonce) {
				t.nonce++
			} else {
				return nil, err
			}
		} else {
			t.sentTx = true
			if handler != nil {
				go func() {
					txHash := tx.Hash().Hex()
					log.Debugf("Waiting for tx %s to be mined", txHash)
					blockDelay := t.blockDelay
					if quickCatch {
						blockDelay = config.QuickCatchBlockDelay
					}
					receipt, err := utils.WaitMined(context.Background(), client, tx, blockDelay)
					if err == nil {
						log.Debugf(
							"Tx %s mined, status: %d, gas estimate: %d, gas used: %d",
							txHash,
							receipt.Status,
							tx.Gas(),
							receipt.GasUsed)
						handler.OnMined(receipt)
					} else {
						log.Error(err)
					}
				}()
			}
			return tx, nil
		}
	}
}

func (t *Transactor) ContractCaller() bind.ContractCaller {
	return t.client
}

func (t *Transactor) Address() common.Address {
	return t.address
}

func (t *Transactor) WaitMined(txHash string) (*types.Receipt, error) {
	return utils.WaitMinedWithTxHash(context.Background(), t.client, txHash, t.blockDelay)
}

func NewGenericTransactionHandler(
	description string, receiptChan chan *types.Receipt) *TransactionMinedHandler {
	return &TransactionMinedHandler{
		OnMined: func(receipt *types.Receipt) {
			if receipt.Status == types.ReceiptStatusSuccessful {
				log.Debugf("%s transaction %s succeeded", description, receipt.TxHash.String())
			} else {
				log.Errorf("%s transaction %s failed", description, receipt.TxHash.String())
			}
			receiptChan <- receipt
		},
	}
}
