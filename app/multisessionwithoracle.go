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

// IMultiSessionWithOracleABI is the input ABI used to generate the binding from.
const IMultiSessionWithOracleABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"},{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"}],\"name\":\"settleBySigTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"}],\"name\":\"settleByMoveTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"},{\"name\":\"_cosignedStateProof\",\"type\":\"bytes\"}],\"name\":\"settleByInvalidTurn\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"},{\"name\":\"_cosignedStateProof\",\"type\":\"bytes\"}],\"name\":\"settleByInvalidState\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nonce\",\"type\":\"uint256\"},{\"name\":\"_signers\",\"type\":\"address[]\"}],\"name\":\"getSessionID\",\"outputs\":[{\"name\":\"session\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// IMultiSessionWithOracleBin is the compiled bytecode used for deploying new contracts.
const IMultiSessionWithOracleBin = `0x`

// DeployIMultiSessionWithOracle deploys a new Ethereum contract, binding an instance of IMultiSessionWithOracle to it.
func DeployIMultiSessionWithOracle(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *IMultiSessionWithOracle, error) {
	parsed, err := abi.JSON(strings.NewReader(IMultiSessionWithOracleABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(IMultiSessionWithOracleBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &IMultiSessionWithOracle{IMultiSessionWithOracleCaller: IMultiSessionWithOracleCaller{contract: contract}, IMultiSessionWithOracleTransactor: IMultiSessionWithOracleTransactor{contract: contract}, IMultiSessionWithOracleFilterer: IMultiSessionWithOracleFilterer{contract: contract}}, nil
}

// IMultiSessionWithOracle is an auto generated Go binding around an Ethereum contract.
type IMultiSessionWithOracle struct {
	IMultiSessionWithOracleCaller     // Read-only binding to the contract
	IMultiSessionWithOracleTransactor // Write-only binding to the contract
	IMultiSessionWithOracleFilterer   // Log filterer for contract events
}

// IMultiSessionWithOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type IMultiSessionWithOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMultiSessionWithOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IMultiSessionWithOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMultiSessionWithOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IMultiSessionWithOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMultiSessionWithOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IMultiSessionWithOracleSession struct {
	Contract     *IMultiSessionWithOracle // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// IMultiSessionWithOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IMultiSessionWithOracleCallerSession struct {
	Contract *IMultiSessionWithOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// IMultiSessionWithOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IMultiSessionWithOracleTransactorSession struct {
	Contract     *IMultiSessionWithOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// IMultiSessionWithOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type IMultiSessionWithOracleRaw struct {
	Contract *IMultiSessionWithOracle // Generic contract binding to access the raw methods on
}

// IMultiSessionWithOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IMultiSessionWithOracleCallerRaw struct {
	Contract *IMultiSessionWithOracleCaller // Generic read-only contract binding to access the raw methods on
}

// IMultiSessionWithOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IMultiSessionWithOracleTransactorRaw struct {
	Contract *IMultiSessionWithOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIMultiSessionWithOracle creates a new instance of IMultiSessionWithOracle, bound to a specific deployed contract.
func NewIMultiSessionWithOracle(address common.Address, backend bind.ContractBackend) (*IMultiSessionWithOracle, error) {
	contract, err := bindIMultiSessionWithOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionWithOracle{IMultiSessionWithOracleCaller: IMultiSessionWithOracleCaller{contract: contract}, IMultiSessionWithOracleTransactor: IMultiSessionWithOracleTransactor{contract: contract}, IMultiSessionWithOracleFilterer: IMultiSessionWithOracleFilterer{contract: contract}}, nil
}

// NewIMultiSessionWithOracleCaller creates a new read-only instance of IMultiSessionWithOracle, bound to a specific deployed contract.
func NewIMultiSessionWithOracleCaller(address common.Address, caller bind.ContractCaller) (*IMultiSessionWithOracleCaller, error) {
	contract, err := bindIMultiSessionWithOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionWithOracleCaller{contract: contract}, nil
}

// NewIMultiSessionWithOracleTransactor creates a new write-only instance of IMultiSessionWithOracle, bound to a specific deployed contract.
func NewIMultiSessionWithOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*IMultiSessionWithOracleTransactor, error) {
	contract, err := bindIMultiSessionWithOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionWithOracleTransactor{contract: contract}, nil
}

// NewIMultiSessionWithOracleFilterer creates a new log filterer instance of IMultiSessionWithOracle, bound to a specific deployed contract.
func NewIMultiSessionWithOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*IMultiSessionWithOracleFilterer, error) {
	contract, err := bindIMultiSessionWithOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionWithOracleFilterer{contract: contract}, nil
}

// bindIMultiSessionWithOracle binds a generic wrapper to an already deployed contract.
func bindIMultiSessionWithOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IMultiSessionWithOracleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMultiSessionWithOracle *IMultiSessionWithOracleRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IMultiSessionWithOracle.Contract.IMultiSessionWithOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMultiSessionWithOracle *IMultiSessionWithOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.IMultiSessionWithOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMultiSessionWithOracle *IMultiSessionWithOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.IMultiSessionWithOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMultiSessionWithOracle *IMultiSessionWithOracleCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IMultiSessionWithOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.contract.Transact(opts, method, params...)
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) constant returns(bytes)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleCaller) GetState(opts *bind.CallOpts, _session [32]byte, _key *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _IMultiSessionWithOracle.contract.Call(opts, out, "getState", _session, _key)
	return *ret0, err
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) constant returns(bytes)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) GetState(_session [32]byte, _key *big.Int) ([]byte, error) {
	return _IMultiSessionWithOracle.Contract.GetState(&_IMultiSessionWithOracle.CallOpts, _session, _key)
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) constant returns(bytes)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleCallerSession) GetState(_session [32]byte, _key *big.Int) ([]byte, error) {
	return _IMultiSessionWithOracle.Contract.GetState(&_IMultiSessionWithOracle.CallOpts, _session, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) constant returns(uint8)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleCaller) GetStatus(opts *bind.CallOpts, _session [32]byte) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _IMultiSessionWithOracle.contract.Call(opts, out, "getStatus", _session)
	return *ret0, err
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) constant returns(uint8)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) GetStatus(_session [32]byte) (uint8, error) {
	return _IMultiSessionWithOracle.Contract.GetStatus(&_IMultiSessionWithOracle.CallOpts, _session)
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) constant returns(uint8)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleCallerSession) GetStatus(_session [32]byte) (uint8, error) {
	return _IMultiSessionWithOracle.Contract.GetStatus(&_IMultiSessionWithOracle.CallOpts, _session)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleCaller) IsFinalized(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _IMultiSessionWithOracle.contract.Call(opts, out, "isFinalized", _query)
	return *ret0, err
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) IsFinalized(_query []byte) (bool, error) {
	return _IMultiSessionWithOracle.Contract.IsFinalized(&_IMultiSessionWithOracle.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleCallerSession) IsFinalized(_query []byte) (bool, error) {
	return _IMultiSessionWithOracle.Contract.IsFinalized(&_IMultiSessionWithOracle.CallOpts, _query)
}

// GetSessionID is a paid mutator transaction binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) returns(bytes32 session)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactor) GetSessionID(opts *bind.TransactOpts, _nonce *big.Int, _signers []common.Address) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.contract.Transact(opts, "getSessionID", _nonce, _signers)
}

// GetSessionID is a paid mutator transaction binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) returns(bytes32 session)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) GetSessionID(_nonce *big.Int, _signers []common.Address) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.GetSessionID(&_IMultiSessionWithOracle.TransactOpts, _nonce, _signers)
}

// GetSessionID is a paid mutator transaction binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) returns(bytes32 session)
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactorSession) GetSessionID(_nonce *big.Int, _signers []common.Address) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.GetSessionID(&_IMultiSessionWithOracle.TransactOpts, _nonce, _signers)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactor) SettleByInvalidState(opts *bind.TransactOpts, _oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.contract.Transact(opts, "settleByInvalidState", _oracleProof, _cosignedStateProof)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) SettleByInvalidState(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleByInvalidState(&_IMultiSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactorSession) SettleByInvalidState(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleByInvalidState(&_IMultiSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactor) SettleByInvalidTurn(opts *bind.TransactOpts, _oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.contract.Transact(opts, "settleByInvalidTurn", _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) SettleByInvalidTurn(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleByInvalidTurn(&_IMultiSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactorSession) SettleByInvalidTurn(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleByInvalidTurn(&_IMultiSessionWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactor) SettleByMoveTimeout(opts *bind.TransactOpts, _oracleProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.contract.Transact(opts, "settleByMoveTimeout", _oracleProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) SettleByMoveTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleByMoveTimeout(&_IMultiSessionWithOracle.TransactOpts, _oracleProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactorSession) SettleByMoveTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleByMoveTimeout(&_IMultiSessionWithOracle.TransactOpts, _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactor) SettleBySigTimeout(opts *bind.TransactOpts, _oracleProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.contract.Transact(opts, "settleBySigTimeout", _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleSession) SettleBySigTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleBySigTimeout(&_IMultiSessionWithOracle.TransactOpts, _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_IMultiSessionWithOracle *IMultiSessionWithOracleTransactorSession) SettleBySigTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _IMultiSessionWithOracle.Contract.SettleBySigTimeout(&_IMultiSessionWithOracle.TransactOpts, _oracleProof)
}
