// Copyright 2018-2020 Celer Network

package transactor

import (
	"fmt"
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
	handler *TransactionStateHandler,
	txconfig *TxConfig,
	method TxMethod) (*types.Transaction, error) {
	return p.nextTransactor().Transact(handler, txconfig, method)
}

func (p *Pool) SubmitWaitMined(
	description string,
	txconfig *TxConfig,
	method TxMethod) (*types.Receipt, error) {
	return p.nextTransactor().TransactWaitMined(description, txconfig, method)
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
