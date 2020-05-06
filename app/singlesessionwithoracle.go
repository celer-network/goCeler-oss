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

// ISingleSessionWithOracleABI is the input ABI used to generate the binding from.
const ISingleSessionWithOracleABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"}],\"name\":\"settleBySigTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"}],\"name\":\"settleByMoveTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"},{\"name\":\"_cosignedStateProof\",\"type\":\"bytes\"}],\"name\":\"settleByInvalidTurn\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"},{\"name\":\"_cosignedStateProof\",\"type\":\"bytes\"}],\"name\":\"settleByInvalidState\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ISingleSessionWithOracleBin is the compiled bytecode used for deploying new contracts.
const ISingleSessionWithOracleBin = `0x`

// DeployISingleSessionWithOracle deploys a new Ethereum contract, binding an instance of ISingleSessionWithOracle to it.
func DeployISingleSessionWithOracle(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ISingleSessionWithOracle, error) {
	parsed, err := abi.JSON(strings.NewReader(ISingleSessionWithOracleABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ISingleSessionWithOracleBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ISingleSessionWithOracle{ISingleSessionWithOracleCaller: ISingleSessionWithOracleCaller{contract: contract}, ISingleSessionWithOracleTransactor: ISingleSessionWithOracleTransactor{contract: contract}, ISingleSessionWithOracleFilterer: ISingleSessionWithOracleFilterer{contract: contract}}, nil
}

// ISingleSessionWithOracle is an auto generated Go binding around an Ethereum contract.
type ISingleSessionWithOracle struct {
	ISingleSessionWithOracleCaller     // Read-only binding to the contract
	ISingleSessionWithOracleTransactor // Write-only binding to the contract
	ISingleSessionWithOracleFilterer   // Log filterer for contract events
}

// ISingleSessionWithOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type ISingleSessionWithOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISingleSessionWithOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ISingleSessionWithOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISingleSessionWithOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ISingleSessionWithOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISingleSessionWithOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ISingleSessionWithOracleSession struct {
	Contract     *ISingleSessionWithOracle // Generic contract binding to set the session for
	CallOpts     bind.CallOpts             // Call options to use throughout this session
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ISingleSessionWithOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ISingleSessionWithOracleCallerSession struct {
	Contract *ISingleSessionWithOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                   // Call options to use throughout this session
}

// ISingleSessionWithOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ISingleSessionWithOracleTransactorSession struct {
	Contract     *ISingleSessionWithOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                   // Transaction auth options to use throughout this session
}

// ISingleSessionWithOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type ISingleSessionWithOracleRaw struct {
	Contract *ISingleSessionWithOracle // Generic contract binding to access the raw methods on
}

// ISingleSessionWithOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ISingleSessionWithOracleCallerRaw struct {
	Contract *ISingleSessionWithOracleCaller // Generic read-only contract binding to access the raw methods on
}

// ISingleSessionWithOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ISingleSessionWithOracleTransactorRaw struct {
	Contract *ISingleSessionWithOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewISingleSessionWithOracle creates a new instance of ISingleSessionWithOracle, bound to a specific deployed contract.
func NewISingleSessionWithOracle(address common.Address, backend bind.ContractBackend) (*ISingleSessionWithOracle, error) {
	contract, err := bindISingleSessionWithOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ISingleSessionWithOracle{ISingleSessionWithOracleCaller: ISingleSessionWithOracleCaller{contract: contract}, ISingleSessionWithOracleTransactor: ISingleSessionWithOracleTransactor{contract: contract}, ISingleSessionWithOracleFilterer: ISingleSessionWithOracleFilterer{contract: contract}}, nil
}

// NewISingleSessionWithOracleCaller creates a new read-only instance of ISingleSessionWithOracle, bound to a specific deployed contract.
func NewISingleSessionWithOracleCaller(address common.Address, caller bind.ContractCaller) (*ISingleSessionWithOracleCaller, error) {
	contract, err := bindISingleSessionWithOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ISingleSessionWithOracleCaller{contract: contract}, nil
}

// NewISingleSessionWithOracleTransactor creates a new write-only instance of ISingleSessionWithOracle, bound to a specific deployed contract.
func NewISingleSessionWithOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*ISingleSessionWithOracleTransactor, error) {
	contract, err := bindISingleSessionWithOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ISingleSessionWithOracleTransactor{contract: contract}, nil
}

// NewISingleSessionWithOracleFilterer creates a new log filterer instance of ISingleSessionWithOracle, bound to a specific deployed contract.
func NewISingleSessionWithOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*ISingleSessionWithOracleFilterer, error) {
	contract, err := bindISingleSessionWithOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ISingleSessionWithOracleFilterer{contract: contract}, nil
}

// bindISingleSessionWithOracle binds a generic wrapper to an already deployed contract.
func bindISingleSessionWithOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ISingleSessionWithOracleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISingleSessionWithOracle *ISingleSessionWithOracleRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ISingleSessionWithOracle.Contract.ISingleSessionWithOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISingleSessionWithOracle *ISingleSessionWithOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.ISingleSessionWithOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISingleSessionWithOracle *ISingleSessionWithOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.ISingleSessionWithOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISingleSessionWithOracle *ISingleSessionWithOracleCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ISingleSessionWithOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.contract.Transact(opts, method, params...)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleCaller) GetState(opts *bind.CallOpts, _key *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _ISingleSessionWithOracle.contract.Call(opts, out, "getState", _key)
	return *ret0, err
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleSession) GetState(_key *big.Int) ([]byte, error) {
	return _ISingleSessionWithOracle.Contract.GetState(&_ISingleSessionWithOracle.CallOpts, _key)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleCallerSession) GetState(_key *big.Int) ([]byte, error) {
	return _ISingleSessionWithOracle.Contract.GetState(&_ISingleSessionWithOracle.CallOpts, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleCaller) GetStatus(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _ISingleSessionWithOracle.contract.Call(opts, out, "getStatus")
	return *ret0, err
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleSession) GetStatus() (uint8, error) {
	return _ISingleSessionWithOracle.Contract.GetStatus(&_ISingleSessionWithOracle.CallOpts)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleCallerSession) GetStatus() (uint8, error) {
	return _ISingleSessionWithOracle.Contract.GetStatus(&_ISingleSessionWithOracle.CallOpts)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleCaller) IsFinalized(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ISingleSessionWithOracle.contract.Call(opts, out, "isFinalized", _query)
	return *ret0, err
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleSession) IsFinalized(_query []byte) (bool, error) {
	return _ISingleSessionWithOracle.Contract.IsFinalized(&_ISingleSessionWithOracle.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_ISingleSessionWithOracle *ISingleSessionWithOracleCallerSession) IsFinalized(_query []byte) (bool, error) {
	return _ISingleSessionWithOracle.Contract.IsFinalized(&_ISingleSessionWithOracle.CallOpts, _query)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactor) SettleByInvalidState(opts *bind.TransactOpts, _oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.contract.Transact(opts, "settleByInvalidState", _oracleProof, _cosignedStateProof)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleSession) SettleByInvalidState(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleByInvalidState(&_ISingleSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactorSession) SettleByInvalidState(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleByInvalidState(&_ISingleSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactor) SettleByInvalidTurn(opts *bind.TransactOpts, _oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.contract.Transact(opts, "settleByInvalidTurn", _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleSession) SettleByInvalidTurn(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleByInvalidTurn(&_ISingleSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactorSession) SettleByInvalidTurn(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleByInvalidTurn(&_ISingleSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactor) SettleByMoveTimeout(opts *bind.TransactOpts, _oracleProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.contract.Transact(opts, "settleByMoveTimeout", _oracleProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleSession) SettleByMoveTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleByMoveTimeout(&_ISingleSessionWithOracle.TransactOpts, _oracleProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactorSession) SettleByMoveTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleByMoveTimeout(&_ISingleSessionWithOracle.TransactOpts, _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactor) SettleBySigTimeout(opts *bind.TransactOpts, _oracleProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.contract.Transact(opts, "settleBySigTimeout", _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleSession) SettleBySigTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleBySigTimeout(&_ISingleSessionWithOracle.TransactOpts, _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_ISingleSessionWithOracle *ISingleSessionWithOracleTransactorSession) SettleBySigTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _ISingleSessionWithOracle.Contract.SettleBySigTimeout(&_ISingleSessionWithOracle.TransactOpts, _oracleProof)
}
