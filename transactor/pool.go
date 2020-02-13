// Copyright 2018-2019 Celer Network

package transactor

import (
	"fmt"
	"math/big"
	"sync"

	log "github.com/celer-network/goCeler-oss/clog"
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
	keyStore   string
	passPhrase string
}

func NewTransactorConfig(keyStore string, passPhrase string) *TransactorConfig {
	return &TransactorConfig{keyStore: keyStore, passPhrase: passPhrase}
}

func NewPool(
	client *ethclient.Client,
	blockDelay uint64,
	chainId *big.Int,
	configs []*TransactorConfig) (*Pool, error) {
	transactors := []*Transactor{}
	for _, config := range configs {
		transactor, err :=
			NewTransactor(config.keyStore, config.passPhrase, chainId, client, blockDelay)
		if err != nil {
			log.Errorln(err)
		} else {
			transactors = append(transactors, transactor)
		}
	}
	if len(transactors) == 0 {
		return nil, fmt.Errorf("Empty transactor pool")
	}
	return &Pool{transactors: transactors, current: 0}, nil
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
	_, err :=
		p.SubmitWithGasLimit(
			NewGenericTransactionHandler(description, receiptChan),
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
