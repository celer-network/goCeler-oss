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

// IMultiSessionABI is the input ABI used to generate the binding from.
const IMultiSessionABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"session\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"}],\"name\":\"IntendSettle\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_stateProof\",\"type\":\"bytes\"}],\"name\":\"intendSettle\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getSettleFinalizedTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"},{\"name\":\"_action\",\"type\":\"bytes\"}],\"name\":\"applyAction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"finalizeOnActionTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getActionDeadline\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getSeqNum\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"},{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nonce\",\"type\":\"uint256\"},{\"name\":\"_signers\",\"type\":\"address[]\"}],\"name\":\"getSessionID\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// IMultiSession is an auto generated Go binding around an Ethereum contract.
type IMultiSession struct {
	IMultiSessionCaller     // Read-only binding to the contract
	IMultiSessionTransactor // Write-only binding to the contract
	IMultiSessionFilterer   // Log filterer for contract events
}

// IMultiSessionCaller is an auto generated read-only Go binding around an Ethereum contract.
type IMultiSessionCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMultiSessionTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IMultiSessionTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMultiSessionFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IMultiSessionFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMultiSessionSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IMultiSessionSession struct {
	Contract     *IMultiSession    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IMultiSessionCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IMultiSessionCallerSession struct {
	Contract *IMultiSessionCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// IMultiSessionTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IMultiSessionTransactorSession struct {
	Contract     *IMultiSessionTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// IMultiSessionRaw is an auto generated low-level Go binding around an Ethereum contract.
type IMultiSessionRaw struct {
	Contract *IMultiSession // Generic contract binding to access the raw methods on
}

// IMultiSessionCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IMultiSessionCallerRaw struct {
	Contract *IMultiSessionCaller // Generic read-only contract binding to access the raw methods on
}

// IMultiSessionTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IMultiSessionTransactorRaw struct {
	Contract *IMultiSessionTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIMultiSession creates a new instance of IMultiSession, bound to a specific deployed contract.
func NewIMultiSession(address common.Address, backend bind.ContractBackend) (*IMultiSession, error) {
	contract, err := bindIMultiSession(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IMultiSession{IMultiSessionCaller: IMultiSessionCaller{contract: contract}, IMultiSessionTransactor: IMultiSessionTransactor{contract: contract}, IMultiSessionFilterer: IMultiSessionFilterer{contract: contract}}, nil
}

// NewIMultiSessionCaller creates a new read-only instance of IMultiSession, bound to a specific deployed contract.
func NewIMultiSessionCaller(address common.Address, caller bind.ContractCaller) (*IMultiSessionCaller, error) {
	contract, err := bindIMultiSession(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionCaller{contract: contract}, nil
}

// NewIMultiSessionTransactor creates a new write-only instance of IMultiSession, bound to a specific deployed contract.
func NewIMultiSessionTransactor(address common.Address, transactor bind.ContractTransactor) (*IMultiSessionTransactor, error) {
	contract, err := bindIMultiSession(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionTransactor{contract: contract}, nil
}

// NewIMultiSessionFilterer creates a new log filterer instance of IMultiSession, bound to a specific deployed contract.
func NewIMultiSessionFilterer(address common.Address, filterer bind.ContractFilterer) (*IMultiSessionFilterer, error) {
	contract, err := bindIMultiSession(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionFilterer{contract: contract}, nil
}

// bindIMultiSession binds a generic wrapper to an already deployed contract.
func bindIMultiSession(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IMultiSessionABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMultiSession *IMultiSessionRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IMultiSession.Contract.IMultiSessionCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMultiSession *IMultiSessionRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMultiSession.Contract.IMultiSessionTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMultiSession *IMultiSessionRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMultiSession.Contract.IMultiSessionTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMultiSession *IMultiSessionCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IMultiSession.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMultiSession *IMultiSessionTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMultiSession.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMultiSession *IMultiSessionTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMultiSession.Contract.contract.Transact(opts, method, params...)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xcab92446.
//
// Solidity: function getActionDeadline(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionCaller) GetActionDeadline(opts *bind.CallOpts, _session [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IMultiSession.contract.Call(opts, out, "getActionDeadline", _session)
	return *ret0, err
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xcab92446.
//
// Solidity: function getActionDeadline(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionSession) GetActionDeadline(_session [32]byte) (*big.Int, error) {
	return _IMultiSession.Contract.GetActionDeadline(&_IMultiSession.CallOpts, _session)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xcab92446.
//
// Solidity: function getActionDeadline(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionCallerSession) GetActionDeadline(_session [32]byte) (*big.Int, error) {
	return _IMultiSession.Contract.GetActionDeadline(&_IMultiSession.CallOpts, _session)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x3b6de66f.
//
// Solidity: function getSeqNum(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionCaller) GetSeqNum(opts *bind.CallOpts, _session [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IMultiSession.contract.Call(opts, out, "getSeqNum", _session)
	return *ret0, err
}

// GetSeqNum is a free data retrieval call binding the contract method 0x3b6de66f.
//
// Solidity: function getSeqNum(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionSession) GetSeqNum(_session [32]byte) (*big.Int, error) {
	return _IMultiSession.Contract.GetSeqNum(&_IMultiSession.CallOpts, _session)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x3b6de66f.
//
// Solidity: function getSeqNum(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionCallerSession) GetSeqNum(_session [32]byte) (*big.Int, error) {
	return _IMultiSession.Contract.GetSeqNum(&_IMultiSession.CallOpts, _session)
}

// GetSessionID is a free data retrieval call binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) pure returns(bytes32)
func (_IMultiSession *IMultiSessionCaller) GetSessionID(opts *bind.CallOpts, _nonce *big.Int, _signers []common.Address) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _IMultiSession.contract.Call(opts, out, "getSessionID", _nonce, _signers)
	return *ret0, err
}

// GetSessionID is a free data retrieval call binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) pure returns(bytes32)
func (_IMultiSession *IMultiSessionSession) GetSessionID(_nonce *big.Int, _signers []common.Address) ([32]byte, error) {
	return _IMultiSession.Contract.GetSessionID(&_IMultiSession.CallOpts, _nonce, _signers)
}

// GetSessionID is a free data retrieval call binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) pure returns(bytes32)
func (_IMultiSession *IMultiSessionCallerSession) GetSessionID(_nonce *big.Int, _signers []common.Address) ([32]byte, error) {
	return _IMultiSession.Contract.GetSessionID(&_IMultiSession.CallOpts, _nonce, _signers)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0x09b65d86.
//
// Solidity: function getSettleFinalizedTime(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionCaller) GetSettleFinalizedTime(opts *bind.CallOpts, _session [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IMultiSession.contract.Call(opts, out, "getSettleFinalizedTime", _session)
	return *ret0, err
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0x09b65d86.
//
// Solidity: function getSettleFinalizedTime(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionSession) GetSettleFinalizedTime(_session [32]byte) (*big.Int, error) {
	return _IMultiSession.Contract.GetSettleFinalizedTime(&_IMultiSession.CallOpts, _session)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0x09b65d86.
//
// Solidity: function getSettleFinalizedTime(bytes32 _session) view returns(uint256)
func (_IMultiSession *IMultiSessionCallerSession) GetSettleFinalizedTime(_session [32]byte) (*big.Int, error) {
	return _IMultiSession.Contract.GetSettleFinalizedTime(&_IMultiSession.CallOpts, _session)
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) view returns(bytes)
func (_IMultiSession *IMultiSessionCaller) GetState(opts *bind.CallOpts, _session [32]byte, _key *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _IMultiSession.contract.Call(opts, out, "getState", _session, _key)
	return *ret0, err
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) view returns(bytes)
func (_IMultiSession *IMultiSessionSession) GetState(_session [32]byte, _key *big.Int) ([]byte, error) {
	return _IMultiSession.Contract.GetState(&_IMultiSession.CallOpts, _session, _key)
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) view returns(bytes)
func (_IMultiSession *IMultiSessionCallerSession) GetState(_session [32]byte, _key *big.Int) ([]byte, error) {
	return _IMultiSession.Contract.GetState(&_IMultiSession.CallOpts, _session, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) view returns(uint8)
func (_IMultiSession *IMultiSessionCaller) GetStatus(opts *bind.CallOpts, _session [32]byte) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _IMultiSession.contract.Call(opts, out, "getStatus", _session)
	return *ret0, err
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) view returns(uint8)
func (_IMultiSession *IMultiSessionSession) GetStatus(_session [32]byte) (uint8, error) {
	return _IMultiSession.Contract.GetStatus(&_IMultiSession.CallOpts, _session)
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) view returns(uint8)
func (_IMultiSession *IMultiSessionCallerSession) GetStatus(_session [32]byte) (uint8, error) {
	return _IMultiSession.Contract.GetStatus(&_IMultiSession.CallOpts, _session)
}

// ApplyAction is a paid mutator transaction binding the contract method 0xf3c77192.
//
// Solidity: function applyAction(bytes32 _session, bytes _action) returns()
func (_IMultiSession *IMultiSessionTransactor) ApplyAction(opts *bind.TransactOpts, _session [32]byte, _action []byte) (*types.Transaction, error) {
	return _IMultiSession.contract.Transact(opts, "applyAction", _session, _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0xf3c77192.
//
// Solidity: function applyAction(bytes32 _session, bytes _action) returns()
func (_IMultiSession *IMultiSessionSession) ApplyAction(_session [32]byte, _action []byte) (*types.Transaction, error) {
	return _IMultiSession.Contract.ApplyAction(&_IMultiSession.TransactOpts, _session, _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0xf3c77192.
//
// Solidity: function applyAction(bytes32 _session, bytes _action) returns()
func (_IMultiSession *IMultiSessionTransactorSession) ApplyAction(_session [32]byte, _action []byte) (*types.Transaction, error) {
	return _IMultiSession.Contract.ApplyAction(&_IMultiSession.TransactOpts, _session, _action)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xb89fa28b.
//
// Solidity: function finalizeOnActionTimeout(bytes32 _session) returns()
func (_IMultiSession *IMultiSessionTransactor) FinalizeOnActionTimeout(opts *bind.TransactOpts, _session [32]byte) (*types.Transaction, error) {
	return _IMultiSession.contract.Transact(opts, "finalizeOnActionTimeout", _session)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xb89fa28b.
//
// Solidity: function finalizeOnActionTimeout(bytes32 _session) returns()
func (_IMultiSession *IMultiSessionSession) FinalizeOnActionTimeout(_session [32]byte) (*types.Transaction, error) {
	return _IMultiSession.Contract.FinalizeOnActionTimeout(&_IMultiSession.TransactOpts, _session)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xb89fa28b.
//
// Solidity: function finalizeOnActionTimeout(bytes32 _session) returns()
func (_IMultiSession *IMultiSessionTransactorSession) FinalizeOnActionTimeout(_session [32]byte) (*types.Transaction, error) {
	return _IMultiSession.Contract.FinalizeOnActionTimeout(&_IMultiSession.TransactOpts, _session)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_IMultiSession *IMultiSessionTransactor) IntendSettle(opts *bind.TransactOpts, _stateProof []byte) (*types.Transaction, error) {
	return _IMultiSession.contract.Transact(opts, "intendSettle", _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_IMultiSession *IMultiSessionSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _IMultiSession.Contract.IntendSettle(&_IMultiSession.TransactOpts, _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_IMultiSession *IMultiSessionTransactorSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _IMultiSession.Contract.IntendSettle(&_IMultiSession.TransactOpts, _stateProof)
}

// IMultiSessionIntendSettleIterator is returned from FilterIntendSettle and is used to iterate over the raw logs and unpacked data for IntendSettle events raised by the IMultiSession contract.
type IMultiSessionIntendSettleIterator struct {
	Event *IMultiSessionIntendSettle // Event containing the contract specifics and raw log

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
func (it *IMultiSessionIntendSettleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IMultiSessionIntendSettle)
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
		it.Event = new(IMultiSessionIntendSettle)
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
func (it *IMultiSessionIntendSettleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IMultiSessionIntendSettleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IMultiSessionIntendSettle represents a IntendSettle event raised by the IMultiSession contract.
type IMultiSessionIntendSettle struct {
	Session [32]byte
	Seq     *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterIntendSettle is a free log retrieval operation binding the contract event 0x82c4eeba939ff9358877334330e22a5cdb0472113cd14f90625ea634b60d2e5b.
//
// Solidity: event IntendSettle(bytes32 indexed session, uint256 seq)
func (_IMultiSession *IMultiSessionFilterer) FilterIntendSettle(opts *bind.FilterOpts, session [][32]byte) (*IMultiSessionIntendSettleIterator, error) {

	var sessionRule []interface{}
	for _, sessionItem := range session {
		sessionRule = append(sessionRule, sessionItem)
	}

	logs, sub, err := _IMultiSession.contract.FilterLogs(opts, "IntendSettle", sessionRule)
	if err != nil {
		return nil, err
	}
	return &IMultiSessionIntendSettleIterator{contract: _IMultiSession.contract, event: "IntendSettle", logs: logs, sub: sub}, nil
}

// WatchIntendSettle is a free log subscription operation binding the contract event 0x82c4eeba939ff9358877334330e22a5cdb0472113cd14f90625ea634b60d2e5b.
//
// Solidity: event IntendSettle(bytes32 indexed session, uint256 seq)
func (_IMultiSession *IMultiSessionFilterer) WatchIntendSettle(opts *bind.WatchOpts, sink chan<- *IMultiSessionIntendSettle, session [][32]byte) (event.Subscription, error) {

	var sessionRule []interface{}
	for _, sessionItem := range session {
		sessionRule = append(sessionRule, sessionItem)
	}

	logs, sub, err := _IMultiSession.contract.WatchLogs(opts, "IntendSettle", sessionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IMultiSessionIntendSettle)
				if err := _IMultiSession.contract.UnpackLog(event, "IntendSettle", log); err != nil {
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

// ParseIntendSettle is a log parse operation binding the contract event 0x82c4eeba939ff9358877334330e22a5cdb0472113cd14f90625ea634b60d2e5b.
//
// Solidity: event IntendSettle(bytes32 indexed session, uint256 seq)
func (_IMultiSession *IMultiSessionFilterer) ParseIntendSettle(log types.Log) (*IMultiSessionIntendSettle, error) {
	event := new(IMultiSessionIntendSettle)
	if err := _IMultiSession.contract.UnpackLog(event, "IntendSettle", log); err != nil {
		return nil, err
	}
	return event, nil
}
