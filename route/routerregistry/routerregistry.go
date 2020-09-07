// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package routerregistry

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

// RouterRegistryABI is the input ABI used to generate the binding from.
const RouterRegistryABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"routerInfo\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"op\",\"type\":\"uint8\"},{\"indexed\":true,\"name\":\"routerAddress\",\"type\":\"address\"}],\"name\":\"RouterUpdated\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[],\"name\":\"registerRouter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"deregisterRouter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"refreshRouter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// RouterRegistryBin is the compiled bytecode used for deploying new contracts.
var RouterRegistryBin = "0x608060405234801561001057600080fd5b506102f3806100206000396000f3fe608060405234801561001057600080fd5b5060043610610068577c0100000000000000000000000000000000000000000000000000000000600035046324f277d2811461006d5780632ff0282b14610077578063788094561461007f578063d1cf70d1146100c4575b600080fd5b6100756100cc565b005b610075610186565b6100b26004803603602081101561009557600080fd5b503573ffffffffffffffffffffffffffffffffffffffff1661021a565b60408051918252519081900360200190f35b61007561022c565b336000908152602081905260409020541561014857604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f526f75746572206164647265737320616c726561647920657869737473000000604482015290519081900360640190fd5b3360008181526020819052604081204390555b6040517fed739f5df64012854c2039ba144af8e3af26211fc7f10a959c6a592ae58c449190600090a3565b3360009081526020819052604090205461020157604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f526f75746572206164647265737320646f6573206e6f74206578697374000000604482015290519081900360640190fd5b336000818152602081905260409020439055600261015b565b60006020819052908152604090205481565b336000908152602081905260409020546102a757604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f526f75746572206164647265737320646f6573206e6f74206578697374000000604482015290519081900360640190fd5b33600081815260208190526040812055600161015b56fea265627a7a72305820a6cb3406d5c8b09132b542c813a59bbd1dc3ae874cfeed6b4d800c190e6ffff164736f6c634300050a0032"

// DeployRouterRegistry deploys a new Ethereum contract, binding an instance of RouterRegistry to it.
func DeployRouterRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *RouterRegistry, error) {
	parsed, err := abi.JSON(strings.NewReader(RouterRegistryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(RouterRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &RouterRegistry{RouterRegistryCaller: RouterRegistryCaller{contract: contract}, RouterRegistryTransactor: RouterRegistryTransactor{contract: contract}, RouterRegistryFilterer: RouterRegistryFilterer{contract: contract}}, nil
}

// RouterRegistry is an auto generated Go binding around an Ethereum contract.
type RouterRegistry struct {
	RouterRegistryCaller     // Read-only binding to the contract
	RouterRegistryTransactor // Write-only binding to the contract
	RouterRegistryFilterer   // Log filterer for contract events
}

// RouterRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type RouterRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RouterRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RouterRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RouterRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RouterRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RouterRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RouterRegistrySession struct {
	Contract     *RouterRegistry   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RouterRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RouterRegistryCallerSession struct {
	Contract *RouterRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// RouterRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RouterRegistryTransactorSession struct {
	Contract     *RouterRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// RouterRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type RouterRegistryRaw struct {
	Contract *RouterRegistry // Generic contract binding to access the raw methods on
}

// RouterRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RouterRegistryCallerRaw struct {
	Contract *RouterRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// RouterRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RouterRegistryTransactorRaw struct {
	Contract *RouterRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRouterRegistry creates a new instance of RouterRegistry, bound to a specific deployed contract.
func NewRouterRegistry(address common.Address, backend bind.ContractBackend) (*RouterRegistry, error) {
	contract, err := bindRouterRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RouterRegistry{RouterRegistryCaller: RouterRegistryCaller{contract: contract}, RouterRegistryTransactor: RouterRegistryTransactor{contract: contract}, RouterRegistryFilterer: RouterRegistryFilterer{contract: contract}}, nil
}

// NewRouterRegistryCaller creates a new read-only instance of RouterRegistry, bound to a specific deployed contract.
func NewRouterRegistryCaller(address common.Address, caller bind.ContractCaller) (*RouterRegistryCaller, error) {
	contract, err := bindRouterRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RouterRegistryCaller{contract: contract}, nil
}

// NewRouterRegistryTransactor creates a new write-only instance of RouterRegistry, bound to a specific deployed contract.
func NewRouterRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*RouterRegistryTransactor, error) {
	contract, err := bindRouterRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RouterRegistryTransactor{contract: contract}, nil
}

// NewRouterRegistryFilterer creates a new log filterer instance of RouterRegistry, bound to a specific deployed contract.
func NewRouterRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*RouterRegistryFilterer, error) {
	contract, err := bindRouterRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RouterRegistryFilterer{contract: contract}, nil
}

// bindRouterRegistry binds a generic wrapper to an already deployed contract.
func bindRouterRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RouterRegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RouterRegistry *RouterRegistryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _RouterRegistry.Contract.RouterRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RouterRegistry *RouterRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RouterRegistry.Contract.RouterRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RouterRegistry *RouterRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RouterRegistry.Contract.RouterRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RouterRegistry *RouterRegistryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _RouterRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RouterRegistry *RouterRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RouterRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RouterRegistry *RouterRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RouterRegistry.Contract.contract.Transact(opts, method, params...)
}

// RouterInfo is a free data retrieval call binding the contract method 0x78809456.
//
// Solidity: function routerInfo(address ) view returns(uint256)
func (_RouterRegistry *RouterRegistryCaller) RouterInfo(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _RouterRegistry.contract.Call(opts, out, "routerInfo", arg0)
	return *ret0, err
}

// RouterInfo is a free data retrieval call binding the contract method 0x78809456.
//
// Solidity: function routerInfo(address ) view returns(uint256)
func (_RouterRegistry *RouterRegistrySession) RouterInfo(arg0 common.Address) (*big.Int, error) {
	return _RouterRegistry.Contract.RouterInfo(&_RouterRegistry.CallOpts, arg0)
}

// RouterInfo is a free data retrieval call binding the contract method 0x78809456.
//
// Solidity: function routerInfo(address ) view returns(uint256)
func (_RouterRegistry *RouterRegistryCallerSession) RouterInfo(arg0 common.Address) (*big.Int, error) {
	return _RouterRegistry.Contract.RouterInfo(&_RouterRegistry.CallOpts, arg0)
}

// DeregisterRouter is a paid mutator transaction binding the contract method 0xd1cf70d1.
//
// Solidity: function deregisterRouter() returns()
func (_RouterRegistry *RouterRegistryTransactor) DeregisterRouter(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RouterRegistry.contract.Transact(opts, "deregisterRouter")
}

// DeregisterRouter is a paid mutator transaction binding the contract method 0xd1cf70d1.
//
// Solidity: function deregisterRouter() returns()
func (_RouterRegistry *RouterRegistrySession) DeregisterRouter() (*types.Transaction, error) {
	return _RouterRegistry.Contract.DeregisterRouter(&_RouterRegistry.TransactOpts)
}

// DeregisterRouter is a paid mutator transaction binding the contract method 0xd1cf70d1.
//
// Solidity: function deregisterRouter() returns()
func (_RouterRegistry *RouterRegistryTransactorSession) DeregisterRouter() (*types.Transaction, error) {
	return _RouterRegistry.Contract.DeregisterRouter(&_RouterRegistry.TransactOpts)
}

// RefreshRouter is a paid mutator transaction binding the contract method 0x2ff0282b.
//
// Solidity: function refreshRouter() returns()
func (_RouterRegistry *RouterRegistryTransactor) RefreshRouter(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RouterRegistry.contract.Transact(opts, "refreshRouter")
}

// RefreshRouter is a paid mutator transaction binding the contract method 0x2ff0282b.
//
// Solidity: function refreshRouter() returns()
func (_RouterRegistry *RouterRegistrySession) RefreshRouter() (*types.Transaction, error) {
	return _RouterRegistry.Contract.RefreshRouter(&_RouterRegistry.TransactOpts)
}

// RefreshRouter is a paid mutator transaction binding the contract method 0x2ff0282b.
//
// Solidity: function refreshRouter() returns()
func (_RouterRegistry *RouterRegistryTransactorSession) RefreshRouter() (*types.Transaction, error) {
	return _RouterRegistry.Contract.RefreshRouter(&_RouterRegistry.TransactOpts)
}

// RegisterRouter is a paid mutator transaction binding the contract method 0x24f277d2.
//
// Solidity: function registerRouter() returns()
func (_RouterRegistry *RouterRegistryTransactor) RegisterRouter(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RouterRegistry.contract.Transact(opts, "registerRouter")
}

// RegisterRouter is a paid mutator transaction binding the contract method 0x24f277d2.
//
// Solidity: function registerRouter() returns()
func (_RouterRegistry *RouterRegistrySession) RegisterRouter() (*types.Transaction, error) {
	return _RouterRegistry.Contract.RegisterRouter(&_RouterRegistry.TransactOpts)
}

// RegisterRouter is a paid mutator transaction binding the contract method 0x24f277d2.
//
// Solidity: function registerRouter() returns()
func (_RouterRegistry *RouterRegistryTransactorSession) RegisterRouter() (*types.Transaction, error) {
	return _RouterRegistry.Contract.RegisterRouter(&_RouterRegistry.TransactOpts)
}

// RouterRegistryRouterUpdatedIterator is returned from FilterRouterUpdated and is used to iterate over the raw logs and unpacked data for RouterUpdated events raised by the RouterRegistry contract.
type RouterRegistryRouterUpdatedIterator struct {
	Event *RouterRegistryRouterUpdated // Event containing the contract specifics and raw log

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
func (it *RouterRegistryRouterUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterRegistryRouterUpdated)
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
		it.Event = new(RouterRegistryRouterUpdated)
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
func (it *RouterRegistryRouterUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterRegistryRouterUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterRegistryRouterUpdated represents a RouterUpdated event raised by the RouterRegistry contract.
type RouterRegistryRouterUpdated struct {
	Op            uint8
	RouterAddress common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRouterUpdated is a free log retrieval operation binding the contract event 0xed739f5df64012854c2039ba144af8e3af26211fc7f10a959c6a592ae58c4491.
//
// Solidity: event RouterUpdated(uint8 indexed op, address indexed routerAddress)
func (_RouterRegistry *RouterRegistryFilterer) FilterRouterUpdated(opts *bind.FilterOpts, op []uint8, routerAddress []common.Address) (*RouterRegistryRouterUpdatedIterator, error) {

	var opRule []interface{}
	for _, opItem := range op {
		opRule = append(opRule, opItem)
	}
	var routerAddressRule []interface{}
	for _, routerAddressItem := range routerAddress {
		routerAddressRule = append(routerAddressRule, routerAddressItem)
	}

	logs, sub, err := _RouterRegistry.contract.FilterLogs(opts, "RouterUpdated", opRule, routerAddressRule)
	if err != nil {
		return nil, err
	}
	return &RouterRegistryRouterUpdatedIterator{contract: _RouterRegistry.contract, event: "RouterUpdated", logs: logs, sub: sub}, nil
}

// WatchRouterUpdated is a free log subscription operation binding the contract event 0xed739f5df64012854c2039ba144af8e3af26211fc7f10a959c6a592ae58c4491.
//
// Solidity: event RouterUpdated(uint8 indexed op, address indexed routerAddress)
func (_RouterRegistry *RouterRegistryFilterer) WatchRouterUpdated(opts *bind.WatchOpts, sink chan<- *RouterRegistryRouterUpdated, op []uint8, routerAddress []common.Address) (event.Subscription, error) {

	var opRule []interface{}
	for _, opItem := range op {
		opRule = append(opRule, opItem)
	}
	var routerAddressRule []interface{}
	for _, routerAddressItem := range routerAddress {
		routerAddressRule = append(routerAddressRule, routerAddressItem)
	}

	logs, sub, err := _RouterRegistry.contract.WatchLogs(opts, "RouterUpdated", opRule, routerAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterRegistryRouterUpdated)
				if err := _RouterRegistry.contract.UnpackLog(event, "RouterUpdated", log); err != nil {
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

// ParseRouterUpdated is a log parse operation binding the contract event 0xed739f5df64012854c2039ba144af8e3af26211fc7f10a959c6a592ae58c4491.
//
// Solidity: event RouterUpdated(uint8 indexed op, address indexed routerAddress)
func (_RouterRegistry *RouterRegistryFilterer) ParseRouterUpdated(log types.Log) (*RouterRegistryRouterUpdated, error) {
	event := new(RouterRegistryRouterUpdated)
	if err := _RouterRegistry.contract.UnpackLog(event, "RouterUpdated", log); err != nil {
		return nil, err
	}
	return event, nil
}
