// Copyright 2018-2020 Celer Network

package transactor

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Pool struct {
	transactors []*Transactor
	current     int
	mu          sync.Mutex
}

type TransactorConfig struct {
	KeyStore   string
	PassPhrase string
}

func NewTransactorConfig(keyStore string, passPhrase string) *TransactorConfig {
	return &TransactorConfig{KeyStore: keyStore, PassPhrase: passPhrase}
}

func NewPool(transactors []*Transactor) (*Pool, error) {
	if len(transactors) == 0 {
		return nil, fmt.Errorf("Empty transactor pool")
	}
	return &Pool{transactors: transactors, current: 0}, nil
}

func NewPoolFromConfig(
	client *ethclient.Client,
	configs []*TransactorConfig) (*Pool, error) {
	transactors := []*Transactor{}
	for _, config := range configs {
		transactor, err := NewTransactor(config.KeyStore, config.PassPhrase, client)
		if err != nil {
			log.Errorln(err)
		} else {
			transactors = append(transactors, transactor)
		}
	}
	return NewPool(transactors)
}

func (p *Pool) Submit(
	handler *TransactionMinedHandler,
	value *big.Int,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Transaction, error) {
	return p.nextTransactor().Transact(handler, value, method)
}

func (p *Pool) SubmitWithGasLimit(
	handler *TransactionMinedHandler,
	value *big.Int,
	gasLimit uint64,
	quickCatch bool,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Transaction, error) {
	return p.nextTransactor().TransactWithGasLimit(handler, value, gasLimit, quickCatch, method)
}

func (p *Pool) SubmitAndWaitMinedWithGenericHandler(
	description string,
	value *big.Int,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Receipt, error) {
	return p.submitAndWaitMinedWithGenericHandler(description, value, 0 /* gasLimit */, false /* quickCatch */, method)
}

func (p *Pool) SubmitAndQuickWaitMinedWithGenericHandler(
	description string,
	value *big.Int,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Receipt, error) {
	return p.submitAndWaitMinedWithGenericHandler(description, value, 0 /* gasLimit */, true /* quickCatch */, method)
}

func (p *Pool) SubmitAndWaitMinedWithGasLimitAndGenericHandler(
	description string,
	value *big.Int,
	gasLimit uint64,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Receipt, error) {
	return p.submitAndWaitMinedWithGenericHandler(description, value, gasLimit, false /* quickCatch */, method)
}

func (p *Pool) SubmitAndQuickWaitMinedWithGasLimitAndGenericHandler(
	description string,
	value *big.Int,
	gasLimit uint64,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Receipt, error) {
	return p.submitAndWaitMinedWithGenericHandler(description, value, gasLimit, true /* quickCatch */, method)
}

func (p *Pool) submitAndWaitMinedWithGenericHandler(
	description string,
	value *big.Int,
	gasLimit uint64,
	quickCatch bool,
	method func(
		transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error),
) (*types.Receipt, error) {
	receiptChan := make(chan *types.Receipt, 1)
	_, err := p.SubmitWithGasLimit(
		newGenericTransactionHandler(description, receiptChan),
		value,
		gasLimit,
		quickCatch,
		method)
	if err != nil {
		return nil, err
	}
	res := <-receiptChan
	return res, nil
}

func (p *Pool) nextTransactor() *Transactor {
	p.mu.Lock()
	defer p.mu.Unlock()
	current := p.current
	p.current = (p.current + 1) % len(p.transactors)
	return p.transactors[current]
}

func (p *Pool) ContractCaller() bind.ContractCaller {
	return p.nextTransactor().client
}

func (p *Pool) WaitMined(txHash string) (*types.Receipt, error) {
	return p.nextTransactor().WaitMined(txHash)
}

func newGenericTransactionHandler(
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
