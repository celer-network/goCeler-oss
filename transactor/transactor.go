// Copyright 2018-2020 Celer Network

package transactor

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"sync"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/cobj"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	parityErrIncrementNonce = "incrementing the nonce"
)

type Transactor struct {
	address ctype.Addr
	signer  common.Signer
	client  *ethclient.Client
	nonce   uint64
	sentTx  bool
	lock    sync.Mutex
}

type TransactionMinedHandler struct {
	OnMined func(receipt *types.Receipt)
}

func NewTransactor(
	keyStore string,
	passPhrase string,
	client *ethclient.Client) (*Transactor, error) {
	address, privKey, err := utils.GetAddrAndPrivKey(keyStore, passPhrase)
	if err != nil {
		return nil, err
	}
	signer, err := cobj.NewCelerSigner(privKey)
	if err != nil {
		return nil, err
	}
	return &Transactor{
		address: address,
		signer:  signer,
		client:  client,
	}, nil
}

func NewTransactorByExternalSigner(
	address ctype.Addr,
	signer common.Signer,
	client *ethclient.Client) *Transactor {
	return &Transactor{
		address: address,
		signer:  signer,
		client:  client,
	}
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
	signer := t.newTransactOpts()
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
			log.Warnf("suggested gas price %s larger than cap %s, set to cap", signer.GasPrice, capPrice)
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
					blockDelay := config.BlockDelay
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

func (t *Transactor) Address() ctype.Addr {
	return t.address
}

func (t *Transactor) WaitMined(txHash string) (*types.Receipt, error) {
	return utils.WaitMinedWithTxHash(context.Background(), t.client, txHash, config.BlockDelay)
}

func (t *Transactor) newTransactOpts() *bind.TransactOpts {
	return &bind.TransactOpts{
		From: t.address,
		// Ignore the passed in Signer to enforce EIP-155
		Signer: func(
			signer types.Signer,
			address ctype.Addr,
			tx *types.Transaction) (*types.Transaction, error) {
			if address != t.address {
				return nil, errors.New("not authorized to sign this account")
			}
			rawTx, err := rlp.EncodeToBytes(tx)
			if err != nil {
				return nil, err
			}
			rawTx, err = t.signer.SignEthTransaction(rawTx)
			if err != nil {
				return nil, err
			}
			err = rlp.DecodeBytes(rawTx, tx)
			if err != nil {
				return nil, err
			}
			return tx, nil
		},
	}
}
