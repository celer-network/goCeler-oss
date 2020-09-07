// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package app

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// IBooleanOutcomeABI is the input ABI used to generate the binding from.
const IBooleanOutcomeABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"getOutcome\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// IBooleanOutcome is an auto generated Go binding around an Ethereum contract.
type IBooleanOutcome struct {
	IBooleanOutcomeCaller     // Read-only binding to the contract
	IBooleanOutcomeTransactor // Write-only binding to the contract
	IBooleanOutcomeFilterer   // Log filterer for contract events
}

// IBooleanOutcomeCaller is an auto generated read-only Go binding around an Ethereum contract.
type IBooleanOutcomeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBooleanOutcomeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IBooleanOutcomeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBooleanOutcomeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IBooleanOutcomeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBooleanOutcomeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IBooleanOutcomeSession struct {
	Contract     *IBooleanOutcome  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IBooleanOutcomeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IBooleanOutcomeCallerSession struct {
	Contract *IBooleanOutcomeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// IBooleanOutcomeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IBooleanOutcomeTransactorSession struct {
	Contract     *IBooleanOutcomeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// IBooleanOutcomeRaw is an auto generated low-level Go binding around an Ethereum contract.
type IBooleanOutcomeRaw struct {
	Contract *IBooleanOutcome // Generic contract binding to access the raw methods on
}

// IBooleanOutcomeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IBooleanOutcomeCallerRaw struct {
	Contract *IBooleanOutcomeCaller // Generic read-only contract binding to access the raw methods on
}

// IBooleanOutcomeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IBooleanOutcomeTransactorRaw struct {
	Contract *IBooleanOutcomeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIBooleanOutcome creates a new instance of IBooleanOutcome, bound to a specific deployed contract.
func NewIBooleanOutcome(address common.Address, backend bind.ContractBackend) (*IBooleanOutcome, error) {
	contract, err := bindIBooleanOutcome(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IBooleanOutcome{IBooleanOutcomeCaller: IBooleanOutcomeCaller{contract: contract}, IBooleanOutcomeTransactor: IBooleanOutcomeTransactor{contract: contract}, IBooleanOutcomeFilterer: IBooleanOutcomeFilterer{contract: contract}}, nil
}

// NewIBooleanOutcomeCaller creates a new read-only instance of IBooleanOutcome, bound to a specific deployed contract.
func NewIBooleanOutcomeCaller(address common.Address, caller bind.ContractCaller) (*IBooleanOutcomeCaller, error) {
	contract, err := bindIBooleanOutcome(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IBooleanOutcomeCaller{contract: contract}, nil
}

// NewIBooleanOutcomeTransactor creates a new write-only instance of IBooleanOutcome, bound to a specific deployed contract.
func NewIBooleanOutcomeTransactor(address common.Address, transactor bind.ContractTransactor) (*IBooleanOutcomeTransactor, error) {
	contract, err := bindIBooleanOutcome(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IBooleanOutcomeTransactor{contract: contract}, nil
}

// NewIBooleanOutcomeFilterer creates a new log filterer instance of IBooleanOutcome, bound to a specific deployed contract.
func NewIBooleanOutcomeFilterer(address common.Address, filterer bind.ContractFilterer) (*IBooleanOutcomeFilterer, error) {
	contract, err := bindIBooleanOutcome(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IBooleanOutcomeFilterer{contract: contract}, nil
}

// bindIBooleanOutcome binds a generic wrapper to an already deployed contract.
func bindIBooleanOutcome(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IBooleanOutcomeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBooleanOutcome *IBooleanOutcomeRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IBooleanOutcome.Contract.IBooleanOutcomeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBooleanOutcome *IBooleanOutcomeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBooleanOutcome.Contract.IBooleanOutcomeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBooleanOutcome *IBooleanOutcomeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBooleanOutcome.Contract.IBooleanOutcomeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBooleanOutcome *IBooleanOutcomeCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IBooleanOutcome.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBooleanOutcome *IBooleanOutcomeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBooleanOutcome.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBooleanOutcome *IBooleanOutcomeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBooleanOutcome.Contract.contract.Transact(opts, method, params...)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_IBooleanOutcome *IBooleanOutcomeCaller) GetOutcome(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _IBooleanOutcome.contract.Call(opts, out, "getOutcome", _query)
	return *ret0, err
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_IBooleanOutcome *IBooleanOutcomeSession) GetOutcome(_query []byte) (bool, error) {
	return _IBooleanOutcome.Contract.GetOutcome(&_IBooleanOutcome.CallOpts, _query)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_IBooleanOutcome *IBooleanOutcomeCallerSession) GetOutcome(_query []byte) (bool, error) {
	return _IBooleanOutcome.Contract.GetOutcome(&_IBooleanOutcome.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_IBooleanOutcome *IBooleanOutcomeCaller) IsFinalized(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _IBooleanOutcome.contract.Call(opts, out, "isFinalized", _query)
	return *ret0, err
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_IBooleanOutcome *IBooleanOutcomeSession) IsFinalized(_query []byte) (bool, error) {
	return _IBooleanOutcome.Contract.IsFinalized(&_IBooleanOutcome.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_IBooleanOutcome *IBooleanOutcomeCallerSession) IsFinalized(_query []byte) (bool, error) {
	return _IBooleanOutcome.Contract.IsFinalized(&_IBooleanOutcome.CallOpts, _query)
}
