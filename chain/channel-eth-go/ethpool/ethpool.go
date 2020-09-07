// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ethpool

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

// EthPoolABI is the input ABI used to generate the binding from.
const EthPoolABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_receiver\",\"type\":\"address\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_walletAddr\",\"type\":\"address\"},{\"name\":\"_walletId\",\"type\":\"bytes32\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferToCelerWallet\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// EthPoolBin is the compiled bytecode used for deploying new contracts.
var EthPoolBin = "0x608060405234801561001057600080fd5b50610b3f806100206000396000f3fe6080604052600436106100c4576000357c01000000000000000000000000000000000000000000000000000000009004806370a082311161008157806370a08231146102735780637e1cd431146102b857806395d89b4114610301578063a457c2d714610316578063dd62ed3e1461034f578063f340fa011461038a576100c4565b806306fdde03146100c9578063095ea7b31461015357806323b872dd146101a05780632e1a7d4d146101e3578063313ce5671461020f578063395093511461023a575b600080fd5b3480156100d557600080fd5b506100de6103b0565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610118578181015183820152602001610100565b50505050905090810190601f1680156101455780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561015f57600080fd5b5061018c6004803603604081101561017657600080fd5b50600160a060020a0381351690602001356103e9565b604080519115158252519081900360200190f35b3480156101ac57600080fd5b5061018c600480360360608110156101c357600080fd5b50600160a060020a0381358116916020810135909116906040013561049e565b3480156101ef57600080fd5b5061020d6004803603602081101561020657600080fd5b5035610531565b005b34801561021b57600080fd5b5061022461053f565b6040805160ff9092168252519081900360200190f35b34801561024657600080fd5b5061018c6004803603604081101561025d57600080fd5b50600160a060020a038135169060200135610544565b34801561027f57600080fd5b506102a66004803603602081101561029657600080fd5b5035600160a060020a031661062b565b60408051918252519081900360200190f35b3480156102c457600080fd5b5061018c600480360360808110156102db57600080fd5b50600160a060020a03813581169160208101359091169060408101359060600135610646565b34801561030d57600080fd5b506100de6107d0565b34801561032257600080fd5b5061018c6004803603604081101561033957600080fd5b50600160a060020a038135169060200135610809565b34801561035b57600080fd5b506102a66004803603604081101561037257600080fd5b50600160a060020a038135811691602001351661089d565b61020d600480360360208110156103a057600080fd5b5035600160a060020a03166108c8565b6040518060400160405280600981526020017f457468496e506f6f6c000000000000000000000000000000000000000000000081525081565b6000600160a060020a038316610449576040805160e560020a62461bcd02815260206004820152601460248201527f5370656e64657220616464726573732069732030000000000000000000000000604482015290519081900360640190fd5b336000818152600160209081526040808320600160a060020a0388168085529083529281902086905580518681529051929392600080516020610aeb833981519152929181900390910190a350600192915050565b600160a060020a03831660009081526001602090815260408083203384529091528120546104d2908363ffffffff6109a216565b600160a060020a038516600081815260016020908152604080832033808552908352928190208590558051948552519193600080516020610aeb833981519152929081900390910190a36105278484846109b7565b5060019392505050565b61053c3333836109b7565b50565b601281565b6000600160a060020a0383166105a4576040805160e560020a62461bcd02815260206004820152601460248201527f5370656e64657220616464726573732069732030000000000000000000000000604482015290519081900360640190fd5b336000908152600160209081526040808320600160a060020a03871684529091529020546105d8908363ffffffff610ad116565b336000818152600160209081526040808320600160a060020a038916808552908352928190208590558051948552519193600080516020610aeb833981519152929081900390910190a350600192915050565b600160a060020a031660009081526020819052604090205490565b600160a060020a038416600090815260016020908152604080832033845290915281205461067a908363ffffffff6109a216565b600160a060020a038616600081815260016020908152604080832033808552908352928190208590558051948552519193600080516020610aeb833981519152929081900390910190a3600160a060020a0385166000908152602081905260409020546106ed908363ffffffff6109a216565b600160a060020a038087166000818152602081815260409182902094909455805186815290519288169391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a3600084905080600160a060020a031663d68d9d4e84866040518363ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808281526020019150506000604051808303818588803b1580156107ab57600080fd5b505af11580156107bf573d6000803e3d6000fd5b5060019a9950505050505050505050565b6040518060400160405280600581526020017f457468495000000000000000000000000000000000000000000000000000000081525081565b6000600160a060020a038316610869576040805160e560020a62461bcd02815260206004820152601460248201527f5370656e64657220616464726573732069732030000000000000000000000000604482015290519081900360640190fd5b336000908152600160209081526040808320600160a060020a03871684529091529020546105d8908363ffffffff6109a216565b600160a060020a03918216600090815260016020908152604080832093909416825291909152205490565b600160a060020a038116610926576040805160e560020a62461bcd02815260206004820152601560248201527f5265636569766572206164647265737320697320300000000000000000000000604482015290519081900360640190fd5b600160a060020a03811660009081526020819052604090205461094f903463ffffffff610ad116565b600160a060020a03821660008181526020818152604091829020939093558051348152905191927fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c92918290030190a250565b6000828211156109b157600080fd5b50900390565b600160a060020a038216610a15576040805160e560020a62461bcd02815260206004820152600f60248201527f546f206164647265737320697320300000000000000000000000000000000000604482015290519081900360640190fd5b600160a060020a038316600090815260208190526040902054610a3e908263ffffffff6109a216565b600160a060020a038085166000818152602081815260409182902094909455805185815290519286169391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a3604051600160a060020a0383169082156108fc029083906000818181858888f19350505050158015610acb573d6000803e3d6000fd5b50505050565b600082820183811015610ae357600080fd5b939250505056fe8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925a265627a7a72305820e51338a3c6ac672fc2d366d593d9a46aebef69c804363974cd743a137e9b05e064736f6c634300050a0032"

// DeployEthPool deploys a new Ethereum contract, binding an instance of EthPool to it.
func DeployEthPool(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *EthPool, error) {
	parsed, err := abi.JSON(strings.NewReader(EthPoolABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(EthPoolBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &EthPool{EthPoolCaller: EthPoolCaller{contract: contract}, EthPoolTransactor: EthPoolTransactor{contract: contract}, EthPoolFilterer: EthPoolFilterer{contract: contract}}, nil
}

// EthPool is an auto generated Go binding around an Ethereum contract.
type EthPool struct {
	EthPoolCaller     // Read-only binding to the contract
	EthPoolTransactor // Write-only binding to the contract
	EthPoolFilterer   // Log filterer for contract events
}

// EthPoolCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthPoolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthPoolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthPoolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthPoolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthPoolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthPoolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthPoolSession struct {
	Contract     *EthPool          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EthPoolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthPoolCallerSession struct {
	Contract *EthPoolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// EthPoolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthPoolTransactorSession struct {
	Contract     *EthPoolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// EthPoolRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthPoolRaw struct {
	Contract *EthPool // Generic contract binding to access the raw methods on
}

// EthPoolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthPoolCallerRaw struct {
	Contract *EthPoolCaller // Generic read-only contract binding to access the raw methods on
}

// EthPoolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthPoolTransactorRaw struct {
	Contract *EthPoolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthPool creates a new instance of EthPool, bound to a specific deployed contract.
func NewEthPool(address common.Address, backend bind.ContractBackend) (*EthPool, error) {
	contract, err := bindEthPool(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthPool{EthPoolCaller: EthPoolCaller{contract: contract}, EthPoolTransactor: EthPoolTransactor{contract: contract}, EthPoolFilterer: EthPoolFilterer{contract: contract}}, nil
}

// NewEthPoolCaller creates a new read-only instance of EthPool, bound to a specific deployed contract.
func NewEthPoolCaller(address common.Address, caller bind.ContractCaller) (*EthPoolCaller, error) {
	contract, err := bindEthPool(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthPoolCaller{contract: contract}, nil
}

// NewEthPoolTransactor creates a new write-only instance of EthPool, bound to a specific deployed contract.
func NewEthPoolTransactor(address common.Address, transactor bind.ContractTransactor) (*EthPoolTransactor, error) {
	contract, err := bindEthPool(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthPoolTransactor{contract: contract}, nil
}

// NewEthPoolFilterer creates a new log filterer instance of EthPool, bound to a specific deployed contract.
func NewEthPoolFilterer(address common.Address, filterer bind.ContractFilterer) (*EthPoolFilterer, error) {
	contract, err := bindEthPool(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthPoolFilterer{contract: contract}, nil
}

// bindEthPool binds a generic wrapper to an already deployed contract.
func bindEthPool(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthPoolABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthPool *EthPoolRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthPool.Contract.EthPoolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthPool *EthPoolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthPool.Contract.EthPoolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthPool *EthPoolRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthPool.Contract.EthPoolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthPool *EthPoolCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthPool.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthPool *EthPoolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthPool.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthPool *EthPoolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthPool.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256)
func (_EthPool *EthPoolCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _EthPool.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256)
func (_EthPool *EthPoolSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _EthPool.Contract.Allowance(&_EthPool.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256)
func (_EthPool *EthPoolCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _EthPool.Contract.Allowance(&_EthPool.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_EthPool *EthPoolCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _EthPool.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_EthPool *EthPoolSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _EthPool.Contract.BalanceOf(&_EthPool.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_EthPool *EthPoolCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _EthPool.Contract.BalanceOf(&_EthPool.CallOpts, _owner)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_EthPool *EthPoolCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _EthPool.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_EthPool *EthPoolSession) Decimals() (uint8, error) {
	return _EthPool.Contract.Decimals(&_EthPool.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_EthPool *EthPoolCallerSession) Decimals() (uint8, error) {
	return _EthPool.Contract.Decimals(&_EthPool.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_EthPool *EthPoolCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _EthPool.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_EthPool *EthPoolSession) Name() (string, error) {
	return _EthPool.Contract.Name(&_EthPool.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_EthPool *EthPoolCallerSession) Name() (string, error) {
	return _EthPool.Contract.Name(&_EthPool.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_EthPool *EthPoolCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _EthPool.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_EthPool *EthPoolSession) Symbol() (string, error) {
	return _EthPool.Contract.Symbol(&_EthPool.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_EthPool *EthPoolCallerSession) Symbol() (string, error) {
	return _EthPool.Contract.Symbol(&_EthPool.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool)
func (_EthPool *EthPoolTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool)
func (_EthPool *EthPoolSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.Approve(&_EthPool.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool)
func (_EthPool *EthPoolTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.Approve(&_EthPool.TransactOpts, _spender, _value)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address _spender, uint256 _subtractedValue) returns(bool)
func (_EthPool *EthPoolTransactor) DecreaseAllowance(opts *bind.TransactOpts, _spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _EthPool.contract.Transact(opts, "decreaseAllowance", _spender, _subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address _spender, uint256 _subtractedValue) returns(bool)
func (_EthPool *EthPoolSession) DecreaseAllowance(_spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.DecreaseAllowance(&_EthPool.TransactOpts, _spender, _subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address _spender, uint256 _subtractedValue) returns(bool)
func (_EthPool *EthPoolTransactorSession) DecreaseAllowance(_spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.DecreaseAllowance(&_EthPool.TransactOpts, _spender, _subtractedValue)
}

// Deposit is a paid mutator transaction binding the contract method 0xf340fa01.
//
// Solidity: function deposit(address _receiver) payable returns()
func (_EthPool *EthPoolTransactor) Deposit(opts *bind.TransactOpts, _receiver common.Address) (*types.Transaction, error) {
	return _EthPool.contract.Transact(opts, "deposit", _receiver)
}

// Deposit is a paid mutator transaction binding the contract method 0xf340fa01.
//
// Solidity: function deposit(address _receiver) payable returns()
func (_EthPool *EthPoolSession) Deposit(_receiver common.Address) (*types.Transaction, error) {
	return _EthPool.Contract.Deposit(&_EthPool.TransactOpts, _receiver)
}

// Deposit is a paid mutator transaction binding the contract method 0xf340fa01.
//
// Solidity: function deposit(address _receiver) payable returns()
func (_EthPool *EthPoolTransactorSession) Deposit(_receiver common.Address) (*types.Transaction, error) {
	return _EthPool.Contract.Deposit(&_EthPool.TransactOpts, _receiver)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address _spender, uint256 _addedValue) returns(bool)
func (_EthPool *EthPoolTransactor) IncreaseAllowance(opts *bind.TransactOpts, _spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _EthPool.contract.Transact(opts, "increaseAllowance", _spender, _addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address _spender, uint256 _addedValue) returns(bool)
func (_EthPool *EthPoolSession) IncreaseAllowance(_spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.IncreaseAllowance(&_EthPool.TransactOpts, _spender, _addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address _spender, uint256 _addedValue) returns(bool)
func (_EthPool *EthPoolTransactorSession) IncreaseAllowance(_spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.IncreaseAllowance(&_EthPool.TransactOpts, _spender, _addedValue)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool)
func (_EthPool *EthPoolTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool)
func (_EthPool *EthPoolSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.TransferFrom(&_EthPool.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool)
func (_EthPool *EthPoolTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.TransferFrom(&_EthPool.TransactOpts, _from, _to, _value)
}

// TransferToCelerWallet is a paid mutator transaction binding the contract method 0x7e1cd431.
//
// Solidity: function transferToCelerWallet(address _from, address _walletAddr, bytes32 _walletId, uint256 _value) returns(bool)
func (_EthPool *EthPoolTransactor) TransferToCelerWallet(opts *bind.TransactOpts, _from common.Address, _walletAddr common.Address, _walletId [32]byte, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.contract.Transact(opts, "transferToCelerWallet", _from, _walletAddr, _walletId, _value)
}

// TransferToCelerWallet is a paid mutator transaction binding the contract method 0x7e1cd431.
//
// Solidity: function transferToCelerWallet(address _from, address _walletAddr, bytes32 _walletId, uint256 _value) returns(bool)
func (_EthPool *EthPoolSession) TransferToCelerWallet(_from common.Address, _walletAddr common.Address, _walletId [32]byte, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.TransferToCelerWallet(&_EthPool.TransactOpts, _from, _walletAddr, _walletId, _value)
}

// TransferToCelerWallet is a paid mutator transaction binding the contract method 0x7e1cd431.
//
// Solidity: function transferToCelerWallet(address _from, address _walletAddr, bytes32 _walletId, uint256 _value) returns(bool)
func (_EthPool *EthPoolTransactorSession) TransferToCelerWallet(_from common.Address, _walletAddr common.Address, _walletId [32]byte, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.TransferToCelerWallet(&_EthPool.TransactOpts, _from, _walletAddr, _walletId, _value)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _value) returns()
func (_EthPool *EthPoolTransactor) Withdraw(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _EthPool.contract.Transact(opts, "withdraw", _value)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _value) returns()
func (_EthPool *EthPoolSession) Withdraw(_value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.Withdraw(&_EthPool.TransactOpts, _value)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _value) returns()
func (_EthPool *EthPoolTransactorSession) Withdraw(_value *big.Int) (*types.Transaction, error) {
	return _EthPool.Contract.Withdraw(&_EthPool.TransactOpts, _value)
}

// EthPoolApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the EthPool contract.
type EthPoolApprovalIterator struct {
	Event *EthPoolApproval // Event containing the contract specifics and raw log

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
func (it *EthPoolApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthPoolApproval)
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
		it.Event = new(EthPoolApproval)
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
func (it *EthPoolApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthPoolApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthPoolApproval represents a Approval event raised by the EthPool contract.
type EthPoolApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_EthPool *EthPoolFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*EthPoolApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _EthPool.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &EthPoolApprovalIterator{contract: _EthPool.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_EthPool *EthPoolFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *EthPoolApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _EthPool.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthPoolApproval)
				if err := _EthPool.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_EthPool *EthPoolFilterer) ParseApproval(log types.Log) (*EthPoolApproval, error) {
	event := new(EthPoolApproval)
	if err := _EthPool.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	return event, nil
}

// EthPoolDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the EthPool contract.
type EthPoolDepositIterator struct {
	Event *EthPoolDeposit // Event containing the contract specifics and raw log

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
func (it *EthPoolDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthPoolDeposit)
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
		it.Event = new(EthPoolDeposit)
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
func (it *EthPoolDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthPoolDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthPoolDeposit represents a Deposit event raised by the EthPool contract.
type EthPoolDeposit struct {
	Receiver common.Address
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address indexed receiver, uint256 value)
func (_EthPool *EthPoolFilterer) FilterDeposit(opts *bind.FilterOpts, receiver []common.Address) (*EthPoolDepositIterator, error) {

	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _EthPool.contract.FilterLogs(opts, "Deposit", receiverRule)
	if err != nil {
		return nil, err
	}
	return &EthPoolDepositIterator{contract: _EthPool.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address indexed receiver, uint256 value)
func (_EthPool *EthPoolFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *EthPoolDeposit, receiver []common.Address) (event.Subscription, error) {

	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _EthPool.contract.WatchLogs(opts, "Deposit", receiverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthPoolDeposit)
				if err := _EthPool.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// ParseDeposit is a log parse operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address indexed receiver, uint256 value)
func (_EthPool *EthPoolFilterer) ParseDeposit(log types.Log) (*EthPoolDeposit, error) {
	event := new(EthPoolDeposit)
	if err := _EthPool.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	return event, nil
}

// EthPoolTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the EthPool contract.
type EthPoolTransferIterator struct {
	Event *EthPoolTransfer // Event containing the contract specifics and raw log

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
func (it *EthPoolTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthPoolTransfer)
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
		it.Event = new(EthPoolTransfer)
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
func (it *EthPoolTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthPoolTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthPoolTransfer represents a Transfer event raised by the EthPool contract.
type EthPoolTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_EthPool *EthPoolFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*EthPoolTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _EthPool.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &EthPoolTransferIterator{contract: _EthPool.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_EthPool *EthPoolFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *EthPoolTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _EthPool.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthPoolTransfer)
				if err := _EthPool.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_EthPool *EthPoolFilterer) ParseTransfer(log types.Log) (*EthPoolTransfer, error) {
	event := new(EthPoolTransfer)
	if err := _EthPool.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	return event, nil
}
