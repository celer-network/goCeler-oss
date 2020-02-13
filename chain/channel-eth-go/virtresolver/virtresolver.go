// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package virtresolver

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

// VirtContractResolverABI is the input ABI used to generate the binding from.
const VirtContractResolverABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"virtAddr\",\"type\":\"bytes32\"}],\"name\":\"Deploy\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_code\",\"type\":\"bytes\"},{\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"deploy\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_virtAddr\",\"type\":\"bytes32\"}],\"name\":\"resolve\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// VirtContractResolverBin is the compiled bytecode used for deploying new contracts.
const VirtContractResolverBin = `0x608060405234801561001057600080fd5b5061040e806100206000396000f3fe608060405234801561001057600080fd5b5060043610610052577c010000000000000000000000000000000000000000000000000000000060003504635c23bdf581146100575780639c4ae2d01461009d575b600080fd5b6100746004803603602081101561006d57600080fd5b5035610121565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b61010d600480360360408110156100b357600080fd5b8101906020810181356401000000008111156100ce57600080fd5b8201836020820111156100e057600080fd5b8035906020019184600183028401116401000000008311171561010257600080fd5b9193509150356101da565b604080519115158252519081900360200190f35b60008181526020819052604081205473ffffffffffffffffffffffffffffffffffffffff166101b157604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601b60248201527f4e6f6e6578697374656e74207669727475616c20616464726573730000000000604482015290519081900360640190fd5b5060009081526020819052604090205473ffffffffffffffffffffffffffffffffffffffff1690565b600080848484604051602001808484808284379190910192835250506040805180830381526020808401808452825192820192909220601f8b0182900490910284018301835289825295506060945092508891889182910183828082843760009201829052508681526020819052604090205493945050505073ffffffffffffffffffffffffffffffffffffffff16156102d557604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f43757272656e74207265616c2061646472657373206973206e6f742030000000604482015290519081900360640190fd5b60008151602083016000f0905073ffffffffffffffffffffffffffffffffffffffff811661036457604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f43726561746520636f6e7472616374206661696c65642e000000000000000000604482015290519081900360640190fd5b600083815260208190526040808220805473ffffffffffffffffffffffffffffffffffffffff191673ffffffffffffffffffffffffffffffffffffffff85161790555184917f149208daa30a9306858cc9c171c3510e0e50ab5d59ed2027a37a728430dd02e491a2506001969550505050505056fea265627a7a7230582086095bccbd167fc366930c118a12ab7e1eaee166a47c07e081e0ddf4eb420d6064736f6c634300050a0032`

// DeployVirtContractResolver deploys a new Ethereum contract, binding an instance of VirtContractResolver to it.
func DeployVirtContractResolver(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *VirtContractResolver, error) {
	parsed, err := abi.JSON(strings.NewReader(VirtContractResolverABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(VirtContractResolverBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &VirtContractResolver{VirtContractResolverCaller: VirtContractResolverCaller{contract: contract}, VirtContractResolverTransactor: VirtContractResolverTransactor{contract: contract}, VirtContractResolverFilterer: VirtContractResolverFilterer{contract: contract}}, nil
}

// VirtContractResolver is an auto generated Go binding around an Ethereum contract.
type VirtContractResolver struct {
	VirtContractResolverCaller     // Read-only binding to the contract
	VirtContractResolverTransactor // Write-only binding to the contract
	VirtContractResolverFilterer   // Log filterer for contract events
}

// VirtContractResolverCaller is an auto generated read-only Go binding around an Ethereum contract.
type VirtContractResolverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VirtContractResolverTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VirtContractResolverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VirtContractResolverFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VirtContractResolverFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VirtContractResolverSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VirtContractResolverSession struct {
	Contract     *VirtContractResolver // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// VirtContractResolverCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VirtContractResolverCallerSession struct {
	Contract *VirtContractResolverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// VirtContractResolverTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VirtContractResolverTransactorSession struct {
	Contract     *VirtContractResolverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// VirtContractResolverRaw is an auto generated low-level Go binding around an Ethereum contract.
type VirtContractResolverRaw struct {
	Contract *VirtContractResolver // Generic contract binding to access the raw methods on
}

// VirtContractResolverCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VirtContractResolverCallerRaw struct {
	Contract *VirtContractResolverCaller // Generic read-only contract binding to access the raw methods on
}

// VirtContractResolverTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VirtContractResolverTransactorRaw struct {
	Contract *VirtContractResolverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVirtContractResolver creates a new instance of VirtContractResolver, bound to a specific deployed contract.
func NewVirtContractResolver(address common.Address, backend bind.ContractBackend) (*VirtContractResolver, error) {
	contract, err := bindVirtContractResolver(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &VirtContractResolver{VirtContractResolverCaller: VirtContractResolverCaller{contract: contract}, VirtContractResolverTransactor: VirtContractResolverTransactor{contract: contract}, VirtContractResolverFilterer: VirtContractResolverFilterer{contract: contract}}, nil
}

// NewVirtContractResolverCaller creates a new read-only instance of VirtContractResolver, bound to a specific deployed contract.
func NewVirtContractResolverCaller(address common.Address, caller bind.ContractCaller) (*VirtContractResolverCaller, error) {
	contract, err := bindVirtContractResolver(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VirtContractResolverCaller{contract: contract}, nil
}

// NewVirtContractResolverTransactor creates a new write-only instance of VirtContractResolver, bound to a specific deployed contract.
func NewVirtContractResolverTransactor(address common.Address, transactor bind.ContractTransactor) (*VirtContractResolverTransactor, error) {
	contract, err := bindVirtContractResolver(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VirtContractResolverTransactor{contract: contract}, nil
}

// NewVirtContractResolverFilterer creates a new log filterer instance of VirtContractResolver, bound to a specific deployed contract.
func NewVirtContractResolverFilterer(address common.Address, filterer bind.ContractFilterer) (*VirtContractResolverFilterer, error) {
	contract, err := bindVirtContractResolver(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VirtContractResolverFilterer{contract: contract}, nil
}

// bindVirtContractResolver binds a generic wrapper to an already deployed contract.
func bindVirtContractResolver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(VirtContractResolverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VirtContractResolver *VirtContractResolverRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _VirtContractResolver.Contract.VirtContractResolverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VirtContractResolver *VirtContractResolverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VirtContractResolver.Contract.VirtContractResolverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VirtContractResolver *VirtContractResolverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VirtContractResolver.Contract.VirtContractResolverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VirtContractResolver *VirtContractResolverCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _VirtContractResolver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VirtContractResolver *VirtContractResolverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VirtContractResolver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VirtContractResolver *VirtContractResolverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VirtContractResolver.Contract.contract.Transact(opts, method, params...)
}

// Resolve is a free data retrieval call binding the contract method 0x5c23bdf5.
//
// Solidity: function resolve(bytes32 _virtAddr) constant returns(address)
func (_VirtContractResolver *VirtContractResolverCaller) Resolve(opts *bind.CallOpts, _virtAddr [32]byte) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _VirtContractResolver.contract.Call(opts, out, "resolve", _virtAddr)
	return *ret0, err
}

// Resolve is a free data retrieval call binding the contract method 0x5c23bdf5.
//
// Solidity: function resolve(bytes32 _virtAddr) constant returns(address)
func (_VirtContractResolver *VirtContractResolverSession) Resolve(_virtAddr [32]byte) (common.Address, error) {
	return _VirtContractResolver.Contract.Resolve(&_VirtContractResolver.CallOpts, _virtAddr)
}

// Resolve is a free data retrieval call binding the contract method 0x5c23bdf5.
//
// Solidity: function resolve(bytes32 _virtAddr) constant returns(address)
func (_VirtContractResolver *VirtContractResolverCallerSession) Resolve(_virtAddr [32]byte) (common.Address, error) {
	return _VirtContractResolver.Contract.Resolve(&_VirtContractResolver.CallOpts, _virtAddr)
}

// Deploy is a paid mutator transaction binding the contract method 0x9c4ae2d0.
//
// Solidity: function deploy(bytes _code, uint256 _nonce) returns(bool)
func (_VirtContractResolver *VirtContractResolverTransactor) Deploy(opts *bind.TransactOpts, _code []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _VirtContractResolver.contract.Transact(opts, "deploy", _code, _nonce)
}

// Deploy is a paid mutator transaction binding the contract method 0x9c4ae2d0.
//
// Solidity: function deploy(bytes _code, uint256 _nonce) returns(bool)
func (_VirtContractResolver *VirtContractResolverSession) Deploy(_code []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _VirtContractResolver.Contract.Deploy(&_VirtContractResolver.TransactOpts, _code, _nonce)
}

// Deploy is a paid mutator transaction binding the contract method 0x9c4ae2d0.
//
// Solidity: function deploy(bytes _code, uint256 _nonce) returns(bool)
func (_VirtContractResolver *VirtContractResolverTransactorSession) Deploy(_code []byte, _nonce *big.Int) (*types.Transaction, error) {
	return _VirtContractResolver.Contract.Deploy(&_VirtContractResolver.TransactOpts, _code, _nonce)
}

// VirtContractResolverDeployIterator is returned from FilterDeploy and is used to iterate over the raw logs and unpacked data for Deploy events raised by the VirtContractResolver contract.
type VirtContractResolverDeployIterator struct {
	Event *VirtContractResolverDeploy // Event containing the contract specifics and raw log

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
func (it *VirtContractResolverDeployIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VirtContractResolverDeploy)
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
		it.Event = new(VirtContractResolverDeploy)
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
func (it *VirtContractResolverDeployIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VirtContractResolverDeployIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VirtContractResolverDeploy represents a Deploy event raised by the VirtContractResolver contract.
type VirtContractResolverDeploy struct {
	VirtAddr [32]byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDeploy is a free log retrieval operation binding the contract event 0x149208daa30a9306858cc9c171c3510e0e50ab5d59ed2027a37a728430dd02e4.
//
// Solidity: event Deploy(bytes32 indexed virtAddr)
func (_VirtContractResolver *VirtContractResolverFilterer) FilterDeploy(opts *bind.FilterOpts, virtAddr [][32]byte) (*VirtContractResolverDeployIterator, error) {

	var virtAddrRule []interface{}
	for _, virtAddrItem := range virtAddr {
		virtAddrRule = append(virtAddrRule, virtAddrItem)
	}

	logs, sub, err := _VirtContractResolver.contract.FilterLogs(opts, "Deploy", virtAddrRule)
	if err != nil {
		return nil, err
	}
	return &VirtContractResolverDeployIterator{contract: _VirtContractResolver.contract, event: "Deploy", logs: logs, sub: sub}, nil
}

// WatchDeploy is a free log subscription operation binding the contract event 0x149208daa30a9306858cc9c171c3510e0e50ab5d59ed2027a37a728430dd02e4.
//
// Solidity: event Deploy(bytes32 indexed virtAddr)
func (_VirtContractResolver *VirtContractResolverFilterer) WatchDeploy(opts *bind.WatchOpts, sink chan<- *VirtContractResolverDeploy, virtAddr [][32]byte) (event.Subscription, error) {

	var virtAddrRule []interface{}
	for _, virtAddrItem := range virtAddr {
		virtAddrRule = append(virtAddrRule, virtAddrItem)
	}

	logs, sub, err := _VirtContractResolver.contract.WatchLogs(opts, "Deploy", virtAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VirtContractResolverDeploy)
				if err := _VirtContractResolver.contract.UnpackLog(event, "Deploy", log); err != nil {
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
