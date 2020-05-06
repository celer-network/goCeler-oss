// Copyright 2018-2020 Celer Network

package chain

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract is the generic interface used by BoundContract
// and mock contracts in unit tests.
type Contract interface {
	GetAddr() ctype.Addr
	GetABI() string
	GetETHClient() *ethclient.Client
	SendTransaction(*bind.TransactOpts, string, ...interface{}) (*ethtypes.Transaction, error)
	CallFunc(interface{}, string, ...interface{}) error
	WatchEvent(string, *bind.WatchOpts, <-chan bool) (ethtypes.Log, error)
	FilterEvent(string, *bind.FilterOpts, interface{}) (*EventIterator, error)
	ParseEvent(string, ethtypes.Log, interface{}) error
}

// BoundContract is a binding object for Ethereum smart contract
// It contains *bind.BoundContract (in go-ethereum) as an embedding
type BoundContract struct {
	*bind.BoundContract
	addr ctype.Addr
	abi  string
	conn *ethclient.Client
}

// NewBoundContract creates a new contract binding
func NewBoundContract(
	conn *ethclient.Client,
	addr ctype.Addr,
	rawABI string) (*BoundContract, error) {
	parsedABI, err := abi.JSON(strings.NewReader((rawABI)))
	return &BoundContract{
		bind.NewBoundContract(addr, parsedABI, conn, conn, conn),
		addr,
		rawABI,
		conn,
	}, err
}

// GetAddr returns contract addr
func (c *BoundContract) GetAddr() ctype.Addr {
	return c.addr
}

// GetABI returns contract abi
func (c *BoundContract) GetABI() string {
	return c.abi
}

// GetETHClient return ethereum client
func (c *BoundContract) GetETHClient() *ethclient.Client {
	return c.conn
}

// SendTransaction sends transactions to smart contract via bound contract
func (c *BoundContract) SendTransaction(
	auth *bind.TransactOpts,
	method string,
	params ...interface{}) (*ethtypes.Transaction, error) {
	return c.Transact(auth, method, params...)
}

// CallFunc invokes a view-only contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (c *BoundContract) CallFunc(
	result interface{},
	method string,
	params ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	return c.Call(&bind.CallOpts{Context: ctx}, result, method, params...)
}

// WatchEvent subscribes to future events
// This function blocks until an event is catched or done signal is received
func (c *BoundContract) WatchEvent(
	name string,
	opts *bind.WatchOpts,
	done <-chan bool) (ethtypes.Log, error) {

	logs, sub, err := c.WatchLogs(opts, name)
	if err != nil {
		return ethtypes.Log{}, err
	}
	defer sub.Unsubscribe()
	select {
	case log := <-logs:
		return log, nil
	case <-done:
	}
	return ethtypes.Log{}, nil
}

// FilterEvent gets historical events
// This function returns an iterator over historical events
func (c *BoundContract) FilterEvent(
	name string,
	opts *bind.FilterOpts,
	event interface{}) (*EventIterator, error) {
	logs, sub, err := c.FilterLogs(opts, name)
	if err != nil {
		return nil, err
	}
	return &EventIterator{
		Contract: c,
		Event:    event,
		Name:     name,
		Logs:     logs,
		Sub:      sub,
	}, nil
}

// ParseEvent parses the catched event according to the event template
func (c *BoundContract) ParseEvent(
	name string,
	log ethtypes.Log,
	event interface{}) error {
	err := c.UnpackLog(event, name, log)
	return err
}

type BlockNumber interface {
	GetCurrentBlockNumber() (*big.Int, error)
}
