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
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// INumericOutcomeABI is the input ABI used to generate the binding from.
const INumericOutcomeABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"getOutcome\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// INumericOutcomeBin is the compiled bytecode used for deploying new contracts.
const INumericOutcomeBin = `0x`

// DeployINumericOutcome deploys a new Ethereum contract, binding an instance of INumericOutcome to it.
func DeployINumericOutcome(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *INumericOutcome, error) {
	parsed, err := abi.JSON(strings.NewReader(INumericOutcomeABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(INumericOutcomeBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &INumericOutcome{INumericOutcomeCaller: INumericOutcomeCaller{contract: contract}, INumericOutcomeTransactor: INumericOutcomeTransactor{contract: contract}, INumericOutcomeFilterer: INumericOutcomeFilterer{contract: contract}}, nil
}

// INumericOutcome is an auto generated Go binding around an Ethereum contract.
type INumericOutcome struct {
	INumericOutcomeCaller     // Read-only binding to the contract
	INumericOutcomeTransactor // Write-only binding to the contract
	INumericOutcomeFilterer   // Log filterer for contract events
}

// INumericOutcomeCaller is an auto generated read-only Go binding around an Ethereum contract.
type INumericOutcomeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// INumericOutcomeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type INumericOutcomeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// INumericOutcomeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type INumericOutcomeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// INumericOutcomeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type INumericOutcomeSession struct {
	Contract     *INumericOutcome  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// INumericOutcomeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type INumericOutcomeCallerSession struct {
	Contract *INumericOutcomeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// INumericOutcomeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type INumericOutcomeTransactorSession struct {
	Contract     *INumericOutcomeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// INumericOutcomeRaw is an auto generated low-level Go binding around an Ethereum contract.
type INumericOutcomeRaw struct {
	Contract *INumericOutcome // Generic contract binding to access the raw methods on
}

// INumericOutcomeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type INumericOutcomeCallerRaw struct {
	Contract *INumericOutcomeCaller // Generic read-only contract binding to access the raw methods on
}

// INumericOutcomeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type INumericOutcomeTransactorRaw struct {
	Contract *INumericOutcomeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewINumericOutcome creates a new instance of INumericOutcome, bound to a specific deployed contract.
func NewINumericOutcome(address common.Address, backend bind.ContractBackend) (*INumericOutcome, error) {
	contract, err := bindINumericOutcome(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &INumericOutcome{INumericOutcomeCaller: INumericOutcomeCaller{contract: contract}, INumericOutcomeTransactor: INumericOutcomeTransactor{contract: contract}, INumericOutcomeFilterer: INumericOutcomeFilterer{contract: contract}}, nil
}

// NewINumericOutcomeCaller creates a new read-only instance of INumericOutcome, bound to a specific deployed contract.
func NewINumericOutcomeCaller(address common.Address, caller bind.ContractCaller) (*INumericOutcomeCaller, error) {
	contract, err := bindINumericOutcome(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &INumericOutcomeCaller{contract: contract}, nil
}

// NewINumericOutcomeTransactor creates a new write-only instance of INumericOutcome, bound to a specific deployed contract.
func NewINumericOutcomeTransactor(address common.Address, transactor bind.ContractTransactor) (*INumericOutcomeTransactor, error) {
	contract, err := bindINumericOutcome(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &INumericOutcomeTransactor{contract: contract}, nil
}

// NewINumericOutcomeFilterer creates a new log filterer instance of INumericOutcome, bound to a specific deployed contract.
func NewINumericOutcomeFilterer(address common.Address, filterer bind.ContractFilterer) (*INumericOutcomeFilterer, error) {
	contract, err := bindINumericOutcome(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &INumericOutcomeFilterer{contract: contract}, nil
}

// bindINumericOutcome binds a generic wrapper to an already deployed contract.
func bindINumericOutcome(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(INumericOutcomeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_INumericOutcome *INumericOutcomeRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _INumericOutcome.Contract.INumericOutcomeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_INumericOutcome *INumericOutcomeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _INumericOutcome.Contract.INumericOutcomeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_INumericOutcome *INumericOutcomeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _INumericOutcome.Contract.INumericOutcomeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_INumericOutcome *INumericOutcomeCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _INumericOutcome.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_INumericOutcome *INumericOutcomeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _INumericOutcome.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_INumericOutcome *INumericOutcomeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _INumericOutcome.Contract.contract.Transact(opts, method, params...)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) constant returns(uint256)
func (_INumericOutcome *INumericOutcomeCaller) GetOutcome(opts *bind.CallOpts, _query []byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _INumericOutcome.contract.Call(opts, out, "getOutcome", _query)
	return *ret0, err
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) constant returns(uint256)
func (_INumericOutcome *INumericOutcomeSession) GetOutcome(_query []byte) (*big.Int, error) {
	return _INumericOutcome.Contract.GetOutcome(&_INumericOutcome.CallOpts, _query)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) constant returns(uint256)
func (_INumericOutcome *INumericOutcomeCallerSession) GetOutcome(_query []byte) (*big.Int, error) {
	return _INumericOutcome.Contract.GetOutcome(&_INumericOutcome.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_INumericOutcome *INumericOutcomeCaller) IsFinalized(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _INumericOutcome.contract.Call(opts, out, "isFinalized", _query)
	return *ret0, err
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_INumericOutcome *INumericOutcomeSession) IsFinalized(_query []byte) (bool, error) {
	return _INumericOutcome.Contract.IsFinalized(&_INumericOutcome.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_INumericOutcome *INumericOutcomeCallerSession) IsFinalized(_query []byte) (bool, error) {
	return _INumericOutcome.Contract.IsFinalized(&_INumericOutcome.CallOpts, _query)
}
