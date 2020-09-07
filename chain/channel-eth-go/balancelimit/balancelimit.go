// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package balancelimit

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

// LedgerBalanceLimitABI is the input ABI used to generate the binding from.
const LedgerBalanceLimitABI = "[]"

// LedgerBalanceLimitBin is the compiled bytecode used for deploying new contracts.
var LedgerBalanceLimitBin = "0x610390610026600b82828239805160001a60731461001957fe5b30600052607381538281f3fe730000000000000000000000000000000000000000301460806040526004361061007d577c010000000000000000000000000000000000000000000000000000000060003504635930e0e181146100825780636ad1dc2d146100ae5780636ae97472146100d8578063bdca79a714610109578063c88c626514610154575b600080fd5b81801561008e57600080fd5b506100ac600480360360208110156100a557600080fd5b503561022a565b005b8180156100ba57600080fd5b506100ac600480360360208110156100d157600080fd5b5035610237565b6100f5600480360360208110156100ee57600080fd5b5035610247565b604080519115158252519081900360200190f35b6101426004803603604081101561011f57600080fd5b508035906020013573ffffffffffffffffffffffffffffffffffffffff16610251565b60408051918252519081900360200190f35b81801561016057600080fd5b506100ac6004803603606081101561017757600080fd5b8135919081019060408101602082013564010000000081111561019957600080fd5b8201836020820111156101ab57600080fd5b803590602001918460208302840111640100000000831117156101cd57600080fd5b9193909290916020810190356401000000008111156101eb57600080fd5b8201836020820111156101fd57600080fd5b8035906020019184602083028401116401000000008311171561021f57600080fd5b50909250905061027d565b600501805460ff19169055565b600501805460ff19166001179055565b6005015460ff1690565b73ffffffffffffffffffffffffffffffffffffffff166000908152600491909101602052604090205490565b8281146102eb57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601460248201527f4c656e6774687320646f206e6f74206d61746368000000000000000000000000604482015290519081900360640190fd5b60005b838110156103535782828281811061030257fe5b9050602002013586600401600087878581811061031b57fe5b6020908102929092013573ffffffffffffffffffffffffffffffffffffffff16835250810191909152604001600020556001016102ee565b50505050505056fea265627a7a7230582069dfaa8fe74f4bb6a0086c0f09df2f47928eaffad54d66d4db2bf3b08bbd49cd64736f6c634300050a0032"

// DeployLedgerBalanceLimit deploys a new Ethereum contract, binding an instance of LedgerBalanceLimit to it.
func DeployLedgerBalanceLimit(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *LedgerBalanceLimit, error) {
	parsed, err := abi.JSON(strings.NewReader(LedgerBalanceLimitABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(LedgerBalanceLimitBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &LedgerBalanceLimit{LedgerBalanceLimitCaller: LedgerBalanceLimitCaller{contract: contract}, LedgerBalanceLimitTransactor: LedgerBalanceLimitTransactor{contract: contract}, LedgerBalanceLimitFilterer: LedgerBalanceLimitFilterer{contract: contract}}, nil
}

// LedgerBalanceLimit is an auto generated Go binding around an Ethereum contract.
type LedgerBalanceLimit struct {
	LedgerBalanceLimitCaller     // Read-only binding to the contract
	LedgerBalanceLimitTransactor // Write-only binding to the contract
	LedgerBalanceLimitFilterer   // Log filterer for contract events
}

// LedgerBalanceLimitCaller is an auto generated read-only Go binding around an Ethereum contract.
type LedgerBalanceLimitCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerBalanceLimitTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LedgerBalanceLimitTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerBalanceLimitFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LedgerBalanceLimitFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerBalanceLimitSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LedgerBalanceLimitSession struct {
	Contract     *LedgerBalanceLimit // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// LedgerBalanceLimitCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LedgerBalanceLimitCallerSession struct {
	Contract *LedgerBalanceLimitCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// LedgerBalanceLimitTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LedgerBalanceLimitTransactorSession struct {
	Contract     *LedgerBalanceLimitTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// LedgerBalanceLimitRaw is an auto generated low-level Go binding around an Ethereum contract.
type LedgerBalanceLimitRaw struct {
	Contract *LedgerBalanceLimit // Generic contract binding to access the raw methods on
}

// LedgerBalanceLimitCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LedgerBalanceLimitCallerRaw struct {
	Contract *LedgerBalanceLimitCaller // Generic read-only contract binding to access the raw methods on
}

// LedgerBalanceLimitTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LedgerBalanceLimitTransactorRaw struct {
	Contract *LedgerBalanceLimitTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLedgerBalanceLimit creates a new instance of LedgerBalanceLimit, bound to a specific deployed contract.
func NewLedgerBalanceLimit(address common.Address, backend bind.ContractBackend) (*LedgerBalanceLimit, error) {
	contract, err := bindLedgerBalanceLimit(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LedgerBalanceLimit{LedgerBalanceLimitCaller: LedgerBalanceLimitCaller{contract: contract}, LedgerBalanceLimitTransactor: LedgerBalanceLimitTransactor{contract: contract}, LedgerBalanceLimitFilterer: LedgerBalanceLimitFilterer{contract: contract}}, nil
}

// NewLedgerBalanceLimitCaller creates a new read-only instance of LedgerBalanceLimit, bound to a specific deployed contract.
func NewLedgerBalanceLimitCaller(address common.Address, caller bind.ContractCaller) (*LedgerBalanceLimitCaller, error) {
	contract, err := bindLedgerBalanceLimit(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LedgerBalanceLimitCaller{contract: contract}, nil
}

// NewLedgerBalanceLimitTransactor creates a new write-only instance of LedgerBalanceLimit, bound to a specific deployed contract.
func NewLedgerBalanceLimitTransactor(address common.Address, transactor bind.ContractTransactor) (*LedgerBalanceLimitTransactor, error) {
	contract, err := bindLedgerBalanceLimit(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LedgerBalanceLimitTransactor{contract: contract}, nil
}

// NewLedgerBalanceLimitFilterer creates a new log filterer instance of LedgerBalanceLimit, bound to a specific deployed contract.
func NewLedgerBalanceLimitFilterer(address common.Address, filterer bind.ContractFilterer) (*LedgerBalanceLimitFilterer, error) {
	contract, err := bindLedgerBalanceLimit(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LedgerBalanceLimitFilterer{contract: contract}, nil
}

// bindLedgerBalanceLimit binds a generic wrapper to an already deployed contract.
func bindLedgerBalanceLimit(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(LedgerBalanceLimitABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LedgerBalanceLimit *LedgerBalanceLimitRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LedgerBalanceLimit.Contract.LedgerBalanceLimitCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LedgerBalanceLimit *LedgerBalanceLimitRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LedgerBalanceLimit.Contract.LedgerBalanceLimitTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LedgerBalanceLimit *LedgerBalanceLimitRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LedgerBalanceLimit.Contract.LedgerBalanceLimitTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LedgerBalanceLimit *LedgerBalanceLimitCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LedgerBalanceLimit.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LedgerBalanceLimit *LedgerBalanceLimitTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LedgerBalanceLimit.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LedgerBalanceLimit *LedgerBalanceLimitTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LedgerBalanceLimit.Contract.contract.Transact(opts, method, params...)
}
