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

// ISingleSessionABI is the input ABI used to generate the binding from.
const ISingleSessionABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"}],\"name\":\"IntendSettle\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_stateProof\",\"type\":\"bytes\"}],\"name\":\"intendSettle\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getSettleFinalizedTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_action\",\"type\":\"bytes\"}],\"name\":\"applyAction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getActionDeadline\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeOnActionTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getSeqNum\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ISingleSessionBin is the compiled bytecode used for deploying new contracts.
const ISingleSessionBin = `0x`

// DeployISingleSession deploys a new Ethereum contract, binding an instance of ISingleSession to it.
func DeployISingleSession(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ISingleSession, error) {
	parsed, err := abi.JSON(strings.NewReader(ISingleSessionABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ISingleSessionBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ISingleSession{ISingleSessionCaller: ISingleSessionCaller{contract: contract}, ISingleSessionTransactor: ISingleSessionTransactor{contract: contract}, ISingleSessionFilterer: ISingleSessionFilterer{contract: contract}}, nil
}

// ISingleSession is an auto generated Go binding around an Ethereum contract.
type ISingleSession struct {
	ISingleSessionCaller     // Read-only binding to the contract
	ISingleSessionTransactor // Write-only binding to the contract
	ISingleSessionFilterer   // Log filterer for contract events
}

// ISingleSessionCaller is an auto generated read-only Go binding around an Ethereum contract.
type ISingleSessionCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISingleSessionTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ISingleSessionTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISingleSessionFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ISingleSessionFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ISingleSessionSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ISingleSessionSession struct {
	Contract     *ISingleSession   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ISingleSessionCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ISingleSessionCallerSession struct {
	Contract *ISingleSessionCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// ISingleSessionTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ISingleSessionTransactorSession struct {
	Contract     *ISingleSessionTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ISingleSessionRaw is an auto generated low-level Go binding around an Ethereum contract.
type ISingleSessionRaw struct {
	Contract *ISingleSession // Generic contract binding to access the raw methods on
}

// ISingleSessionCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ISingleSessionCallerRaw struct {
	Contract *ISingleSessionCaller // Generic read-only contract binding to access the raw methods on
}

// ISingleSessionTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ISingleSessionTransactorRaw struct {
	Contract *ISingleSessionTransactor // Generic write-only contract binding to access the raw methods on
}

// NewISingleSession creates a new instance of ISingleSession, bound to a specific deployed contract.
func NewISingleSession(address common.Address, backend bind.ContractBackend) (*ISingleSession, error) {
	contract, err := bindISingleSession(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ISingleSession{ISingleSessionCaller: ISingleSessionCaller{contract: contract}, ISingleSessionTransactor: ISingleSessionTransactor{contract: contract}, ISingleSessionFilterer: ISingleSessionFilterer{contract: contract}}, nil
}

// NewISingleSessionCaller creates a new read-only instance of ISingleSession, bound to a specific deployed contract.
func NewISingleSessionCaller(address common.Address, caller bind.ContractCaller) (*ISingleSessionCaller, error) {
	contract, err := bindISingleSession(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ISingleSessionCaller{contract: contract}, nil
}

// NewISingleSessionTransactor creates a new write-only instance of ISingleSession, bound to a specific deployed contract.
func NewISingleSessionTransactor(address common.Address, transactor bind.ContractTransactor) (*ISingleSessionTransactor, error) {
	contract, err := bindISingleSession(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ISingleSessionTransactor{contract: contract}, nil
}

// NewISingleSessionFilterer creates a new log filterer instance of ISingleSession, bound to a specific deployed contract.
func NewISingleSessionFilterer(address common.Address, filterer bind.ContractFilterer) (*ISingleSessionFilterer, error) {
	contract, err := bindISingleSession(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ISingleSessionFilterer{contract: contract}, nil
}

// bindISingleSession binds a generic wrapper to an already deployed contract.
func bindISingleSession(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ISingleSessionABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISingleSession *ISingleSessionRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ISingleSession.Contract.ISingleSessionCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISingleSession *ISingleSessionRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISingleSession.Contract.ISingleSessionTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISingleSession *ISingleSessionRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISingleSession.Contract.ISingleSessionTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ISingleSession *ISingleSessionCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ISingleSession.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ISingleSession *ISingleSessionTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISingleSession.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ISingleSession *ISingleSessionTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ISingleSession.Contract.contract.Transact(opts, method, params...)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xbbc35280.
//
// Solidity: function getActionDeadline() constant returns(uint256)
func (_ISingleSession *ISingleSessionCaller) GetActionDeadline(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ISingleSession.contract.Call(opts, out, "getActionDeadline")
	return *ret0, err
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xbbc35280.
//
// Solidity: function getActionDeadline() constant returns(uint256)
func (_ISingleSession *ISingleSessionSession) GetActionDeadline() (*big.Int, error) {
	return _ISingleSession.Contract.GetActionDeadline(&_ISingleSession.CallOpts)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xbbc35280.
//
// Solidity: function getActionDeadline() constant returns(uint256)
func (_ISingleSession *ISingleSessionCallerSession) GetActionDeadline() (*big.Int, error) {
	return _ISingleSession.Contract.GetActionDeadline(&_ISingleSession.CallOpts)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() constant returns(uint256)
func (_ISingleSession *ISingleSessionCaller) GetSeqNum(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ISingleSession.contract.Call(opts, out, "getSeqNum")
	return *ret0, err
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() constant returns(uint256)
func (_ISingleSession *ISingleSessionSession) GetSeqNum() (*big.Int, error) {
	return _ISingleSession.Contract.GetSeqNum(&_ISingleSession.CallOpts)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() constant returns(uint256)
func (_ISingleSession *ISingleSessionCallerSession) GetSeqNum() (*big.Int, error) {
	return _ISingleSession.Contract.GetSeqNum(&_ISingleSession.CallOpts)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() constant returns(uint256)
func (_ISingleSession *ISingleSessionCaller) GetSettleFinalizedTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ISingleSession.contract.Call(opts, out, "getSettleFinalizedTime")
	return *ret0, err
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() constant returns(uint256)
func (_ISingleSession *ISingleSessionSession) GetSettleFinalizedTime() (*big.Int, error) {
	return _ISingleSession.Contract.GetSettleFinalizedTime(&_ISingleSession.CallOpts)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() constant returns(uint256)
func (_ISingleSession *ISingleSessionCallerSession) GetSettleFinalizedTime() (*big.Int, error) {
	return _ISingleSession.Contract.GetSettleFinalizedTime(&_ISingleSession.CallOpts)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_ISingleSession *ISingleSessionCaller) GetState(opts *bind.CallOpts, _key *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _ISingleSession.contract.Call(opts, out, "getState", _key)
	return *ret0, err
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_ISingleSession *ISingleSessionSession) GetState(_key *big.Int) ([]byte, error) {
	return _ISingleSession.Contract.GetState(&_ISingleSession.CallOpts, _key)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_ISingleSession *ISingleSessionCallerSession) GetState(_key *big.Int) ([]byte, error) {
	return _ISingleSession.Contract.GetState(&_ISingleSession.CallOpts, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
func (_ISingleSession *ISingleSessionCaller) GetStatus(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _ISingleSession.contract.Call(opts, out, "getStatus")
	return *ret0, err
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
func (_ISingleSession *ISingleSessionSession) GetStatus() (uint8, error) {
	return _ISingleSession.Contract.GetStatus(&_ISingleSession.CallOpts)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
func (_ISingleSession *ISingleSessionCallerSession) GetStatus() (uint8, error) {
	return _ISingleSession.Contract.GetStatus(&_ISingleSession.CallOpts)
}

// ApplyAction is a paid mutator transaction binding the contract method 0x1f2b71e5.
//
// Solidity: function applyAction(bytes _action) returns()
func (_ISingleSession *ISingleSessionTransactor) ApplyAction(opts *bind.TransactOpts, _action []byte) (*types.Transaction, error) {
	return _ISingleSession.contract.Transact(opts, "applyAction", _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0x1f2b71e5.
//
// Solidity: function applyAction(bytes _action) returns()
func (_ISingleSession *ISingleSessionSession) ApplyAction(_action []byte) (*types.Transaction, error) {
	return _ISingleSession.Contract.ApplyAction(&_ISingleSession.TransactOpts, _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0x1f2b71e5.
//
// Solidity: function applyAction(bytes _action) returns()
func (_ISingleSession *ISingleSessionTransactorSession) ApplyAction(_action []byte) (*types.Transaction, error) {
	return _ISingleSession.Contract.ApplyAction(&_ISingleSession.TransactOpts, _action)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xfa5e7ff5.
//
// Solidity: function finalizeOnActionTimeout() returns()
func (_ISingleSession *ISingleSessionTransactor) FinalizeOnActionTimeout(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ISingleSession.contract.Transact(opts, "finalizeOnActionTimeout")
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xfa5e7ff5.
//
// Solidity: function finalizeOnActionTimeout() returns()
func (_ISingleSession *ISingleSessionSession) FinalizeOnActionTimeout() (*types.Transaction, error) {
	return _ISingleSession.Contract.FinalizeOnActionTimeout(&_ISingleSession.TransactOpts)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xfa5e7ff5.
//
// Solidity: function finalizeOnActionTimeout() returns()
func (_ISingleSession *ISingleSessionTransactorSession) FinalizeOnActionTimeout() (*types.Transaction, error) {
	return _ISingleSession.Contract.FinalizeOnActionTimeout(&_ISingleSession.TransactOpts)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_ISingleSession *ISingleSessionTransactor) IntendSettle(opts *bind.TransactOpts, _stateProof []byte) (*types.Transaction, error) {
	return _ISingleSession.contract.Transact(opts, "intendSettle", _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_ISingleSession *ISingleSessionSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _ISingleSession.Contract.IntendSettle(&_ISingleSession.TransactOpts, _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_ISingleSession *ISingleSessionTransactorSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _ISingleSession.Contract.IntendSettle(&_ISingleSession.TransactOpts, _stateProof)
}

// ISingleSessionIntendSettleIterator is returned from FilterIntendSettle and is used to iterate over the raw logs and unpacked data for IntendSettle events raised by the ISingleSession contract.
type ISingleSessionIntendSettleIterator struct {
	Event *ISingleSessionIntendSettle // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ISingleSessionIntendSettleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ISingleSessionIntendSettle)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ISingleSessionIntendSettle)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ISingleSessionIntendSettleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ISingleSessionIntendSettleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ISingleSessionIntendSettle represents a IntendSettle event raised by the ISingleSession contract.
type ISingleSessionIntendSettle struct {
	Seq *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterIntendSettle is a free log retrieval operation binding the contract event 0xce68db27527c6058059e8cbd1c6de0528ef1c417fe1c21c545919c7da3466d2a.
//
// Solidity: event IntendSettle(uint256 seq)
func (_ISingleSession *ISingleSessionFilterer) FilterIntendSettle(opts *bind.FilterOpts) (*ISingleSessionIntendSettleIterator, error) {

	logs, sub, err := _ISingleSession.contract.FilterLogs(opts, "IntendSettle")
	if err != nil {
		return nil, err
	}
	return &ISingleSessionIntendSettleIterator{contract: _ISingleSession.contract, event: "IntendSettle", logs: logs, sub: sub}, nil
}

// WatchIntendSettle is a free log subscription operation binding the contract event 0xce68db27527c6058059e8cbd1c6de0528ef1c417fe1c21c545919c7da3466d2a.
//
// Solidity: event IntendSettle(uint256 seq)
func (_ISingleSession *ISingleSessionFilterer) WatchIntendSettle(opts *bind.WatchOpts, sink chan<- *ISingleSessionIntendSettle) (event.Subscription, error) {

	logs, sub, err := _ISingleSession.contract.WatchLogs(opts, "IntendSettle")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ISingleSessionIntendSettle)
				if err := _ISingleSession.contract.UnpackLog(event, "IntendSettle", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
