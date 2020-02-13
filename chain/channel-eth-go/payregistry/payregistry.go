// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package payregistry

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

// PayRegistryABI is the input ABI used to generate the binding from.
const PayRegistryABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"payInfoMap\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"resolveDeadline\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"payId\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"resolveDeadline\",\"type\":\"uint256\"}],\"name\":\"PayInfoUpdate\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_payHash\",\"type\":\"bytes32\"},{\"name\":\"_setter\",\"type\":\"address\"}],\"name\":\"calculatePayId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_payHash\",\"type\":\"bytes32\"},{\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"setPayAmount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_payHash\",\"type\":\"bytes32\"},{\"name\":\"_deadline\",\"type\":\"uint256\"}],\"name\":\"setPayDeadline\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_payHash\",\"type\":\"bytes32\"},{\"name\":\"_amt\",\"type\":\"uint256\"},{\"name\":\"_deadline\",\"type\":\"uint256\"}],\"name\":\"setPayInfo\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_payHashes\",\"type\":\"bytes32[]\"},{\"name\":\"_amts\",\"type\":\"uint256[]\"}],\"name\":\"setPayAmounts\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_payHashes\",\"type\":\"bytes32[]\"},{\"name\":\"_deadlines\",\"type\":\"uint256[]\"}],\"name\":\"setPayDeadlines\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_payHashes\",\"type\":\"bytes32[]\"},{\"name\":\"_amts\",\"type\":\"uint256[]\"},{\"name\":\"_deadlines\",\"type\":\"uint256[]\"}],\"name\":\"setPayInfos\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_payIds\",\"type\":\"bytes32[]\"},{\"name\":\"_lastPayResolveDeadline\",\"type\":\"uint256\"}],\"name\":\"getPayAmounts\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_payId\",\"type\":\"bytes32\"}],\"name\":\"getPayInfo\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// PayRegistryBin is the compiled bytecode used for deploying new contracts.
const PayRegistryBin = `0x608060405234801561001057600080fd5b50610b96806100206000396000f3fe608060405234801561001057600080fd5b50600436106100bb576000357c0100000000000000000000000000000000000000000000000000000000900480638f13b2f5116100835780638f13b2f51461045057806396efe5731461046d578063cdfa146b146104b8578063e1e35490146104db578063f8fb012f14610504576100bb565b80630daddd34146100c0578063204a95ee1461018457806327b0e05814610246578063414f7e0e1461027c5780637cac39cf14610390575b600080fd5b610182600480360360408110156100d657600080fd5b8101906020810181356401000000008111156100f157600080fd5b82018360208201111561010357600080fd5b8035906020019184602083028401116401000000008311171561012557600080fd5b91939092909160208101903564010000000081111561014357600080fd5b82018360208201111561015557600080fd5b8035906020019184602083028401116401000000008311171561017757600080fd5b509092509050610527565b005b6101826004803603604081101561019a57600080fd5b8101906020810181356401000000008111156101b557600080fd5b8201836020820111156101c757600080fd5b803590602001918460208302840111640100000000831117156101e957600080fd5b91939092909160208101903564010000000081111561020757600080fd5b82018360208201111561021957600080fd5b8035906020019184602083028401116401000000008311171561023b57600080fd5b509092509050610623565b6102636004803603602081101561025c57600080fd5b5035610707565b6040805192835260208301919091528051918290030190f35b6101826004803603606081101561029257600080fd5b8101906020810181356401000000008111156102ad57600080fd5b8201836020820111156102bf57600080fd5b803590602001918460208302840111640100000000831117156102e157600080fd5b9193909290916020810190356401000000008111156102ff57600080fd5b82018360208201111561031157600080fd5b8035906020019184602083028401116401000000008311171561033357600080fd5b91939092909160208101903564010000000081111561035157600080fd5b82018360208201111561036357600080fd5b8035906020019184602083028401116401000000008311171561038557600080fd5b509092509050610721565b610400600480360360408110156103a657600080fd5b8101906020810181356401000000008111156103c157600080fd5b8201836020820111156103d357600080fd5b803590602001918460208302840111640100000000831117156103f557600080fd5b919350915035610843565b60408051602080825283518183015283519192839290830191858101910280838360005b8381101561043c578181015183820152602001610424565b505050509050019250505060405180910390f35b6102636004803603602081101561046657600080fd5b50356109db565b6104a66004803603604081101561048357600080fd5b508035906020013573ffffffffffffffffffffffffffffffffffffffff166109f4565b60408051918252519081900360200190f35b610182600480360360408110156104ce57600080fd5b5080359060200135610a46565b610182600480360360608110156104f157600080fd5b5080359060208101359060400135610a9a565b6101826004803603604081101561051a57600080fd5b5080359060200135610af0565b82811461057e576040805160e560020a62461bcd02815260206004820152601460248201527f4c656e6774687320646f206e6f74206d61746368000000000000000000000000604482015290519081900360640190fd5b600033815b8581101561061a576105a787878381811061059a57fe5b90506020020135836109f4565b60008181526020819052604090209093508585838181106105c457fe5b602002919091013560018301555080548490600080516020610b42833981519152908888868181106105f257fe5b604080519485526020918202939093013590840152508051918290030190a250600101610583565b50505050505050565b82811461067a576040805160e560020a62461bcd02815260206004820152601460248201527f4c656e6774687320646f206e6f74206d61746368000000000000000000000000604482015290519081900360640190fd5b600033815b8581101561061a5761069687878381811061059a57fe5b60008181526020819052604090209093508585838181106106b357fe5b602002919091013582555083600080516020610b428339815191528787858181106106da57fe5b6001860154604080516020938402959095013585529184015280519283900301919050a25060010161067f565b600090815260208190526040902080546001909101549091565b848314801561072f57508481145b610783576040805160e560020a62461bcd02815260206004820152601460248201527f4c656e6774687320646f206e6f74206d61746368000000000000000000000000604482015290519081900360640190fd5b600033815b878110156108385761079f89898381811061059a57fe5b60008181526020819052604090209093508787838181106107bc57fe5b60200291909101358255508585838181106107d357fe5b602002919091013560018301555083600080516020610b428339815191528989858181106107fd57fe5b9050602002013588888681811061081057fe5b604080519485526020918202939093013590840152508051918290030190a250600101610788565b505050505050505050565b60608084849050604051908082528060200260200182016040528015610873578160200160208202803883390190505b50905060005b848110156109d25760008087878481811061089057fe5b905060200201358152602001908152602001600020600101546000141561090d57834311610908576040805160e560020a62461bcd02815260206004820152601860248201527f5061796d656e74206973206e6f742066696e616c697a65640000000000000000604482015290519081900360640190fd5b61098b565b60008087878481811061091c57fe5b90506020020135815260200190815260200160002060010154431161098b576040805160e560020a62461bcd02815260206004820152601860248201527f5061796d656e74206973206e6f742066696e616c697a65640000000000000000604482015290519081900360640190fd5b60008087878481811061099a57fe5b905060200201358152602001908152602001600020600001548282815181106109bf57fe5b6020908102919091010152600101610879565b50949350505050565b6000602081905290815260409020805460019091015482565b6040805160208082019490945273ffffffffffffffffffffffffffffffffffffffff929092166c0100000000000000000000000002828201528051808303603401815260549092019052805191012090565b6000610a5283336109f4565b6000818152602081815260409182902060018101869055805483519081529182018690528251939450928492600080516020610b42833981519152928290030190a250505050565b6000610aa684336109f4565b600081815260208181526040918290208681556001810186905582518781529182018690528251939450928492600080516020610b42833981519152928290030190a25050505050565b6000610afc83336109f4565b6000818152602081815260409182902085815560018101548351878152928301528251939450928492600080516020610b42833981519152928290030190a25050505056fe9e9acc6d43d5d7bd6fa143ef0ee1d224cfe2bb010b7e3acf44878d6314ebc607a265627a7a723058207fb2d6e01db7df745dcc11670cc9d62a37205b7774c19a8d372e5d099e892ee164736f6c634300050a0032`

// DeployPayRegistry deploys a new Ethereum contract, binding an instance of PayRegistry to it.
func DeployPayRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *PayRegistry, error) {
	parsed, err := abi.JSON(strings.NewReader(PayRegistryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PayRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PayRegistry{PayRegistryCaller: PayRegistryCaller{contract: contract}, PayRegistryTransactor: PayRegistryTransactor{contract: contract}, PayRegistryFilterer: PayRegistryFilterer{contract: contract}}, nil
}

// PayRegistry is an auto generated Go binding around an Ethereum contract.
type PayRegistry struct {
	PayRegistryCaller     // Read-only binding to the contract
	PayRegistryTransactor // Write-only binding to the contract
	PayRegistryFilterer   // Log filterer for contract events
}

// PayRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type PayRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PayRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PayRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PayRegistrySession struct {
	Contract     *PayRegistry      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PayRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PayRegistryCallerSession struct {
	Contract *PayRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// PayRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PayRegistryTransactorSession struct {
	Contract     *PayRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// PayRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type PayRegistryRaw struct {
	Contract *PayRegistry // Generic contract binding to access the raw methods on
}

// PayRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PayRegistryCallerRaw struct {
	Contract *PayRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// PayRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PayRegistryTransactorRaw struct {
	Contract *PayRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPayRegistry creates a new instance of PayRegistry, bound to a specific deployed contract.
func NewPayRegistry(address common.Address, backend bind.ContractBackend) (*PayRegistry, error) {
	contract, err := bindPayRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PayRegistry{PayRegistryCaller: PayRegistryCaller{contract: contract}, PayRegistryTransactor: PayRegistryTransactor{contract: contract}, PayRegistryFilterer: PayRegistryFilterer{contract: contract}}, nil
}

// NewPayRegistryCaller creates a new read-only instance of PayRegistry, bound to a specific deployed contract.
func NewPayRegistryCaller(address common.Address, caller bind.ContractCaller) (*PayRegistryCaller, error) {
	contract, err := bindPayRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PayRegistryCaller{contract: contract}, nil
}

// NewPayRegistryTransactor creates a new write-only instance of PayRegistry, bound to a specific deployed contract.
func NewPayRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*PayRegistryTransactor, error) {
	contract, err := bindPayRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PayRegistryTransactor{contract: contract}, nil
}

// NewPayRegistryFilterer creates a new log filterer instance of PayRegistry, bound to a specific deployed contract.
func NewPayRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*PayRegistryFilterer, error) {
	contract, err := bindPayRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PayRegistryFilterer{contract: contract}, nil
}

// bindPayRegistry binds a generic wrapper to an already deployed contract.
func bindPayRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PayRegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayRegistry *PayRegistryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PayRegistry.Contract.PayRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayRegistry *PayRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayRegistry.Contract.PayRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayRegistry *PayRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayRegistry.Contract.PayRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayRegistry *PayRegistryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PayRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayRegistry *PayRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayRegistry *PayRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayRegistry.Contract.contract.Transact(opts, method, params...)
}

// CalculatePayId is a free data retrieval call binding the contract method 0x96efe573.
//
// Solidity: function calculatePayId(bytes32 _payHash, address _setter) constant returns(bytes32)
func (_PayRegistry *PayRegistryCaller) CalculatePayId(opts *bind.CallOpts, _payHash [32]byte, _setter common.Address) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _PayRegistry.contract.Call(opts, out, "calculatePayId", _payHash, _setter)
	return *ret0, err
}

// CalculatePayId is a free data retrieval call binding the contract method 0x96efe573.
//
// Solidity: function calculatePayId(bytes32 _payHash, address _setter) constant returns(bytes32)
func (_PayRegistry *PayRegistrySession) CalculatePayId(_payHash [32]byte, _setter common.Address) ([32]byte, error) {
	return _PayRegistry.Contract.CalculatePayId(&_PayRegistry.CallOpts, _payHash, _setter)
}

// CalculatePayId is a free data retrieval call binding the contract method 0x96efe573.
//
// Solidity: function calculatePayId(bytes32 _payHash, address _setter) constant returns(bytes32)
func (_PayRegistry *PayRegistryCallerSession) CalculatePayId(_payHash [32]byte, _setter common.Address) ([32]byte, error) {
	return _PayRegistry.Contract.CalculatePayId(&_PayRegistry.CallOpts, _payHash, _setter)
}

// GetPayAmounts is a free data retrieval call binding the contract method 0x7cac39cf.
//
// Solidity: function getPayAmounts(bytes32[] _payIds, uint256 _lastPayResolveDeadline) constant returns(uint256[])
func (_PayRegistry *PayRegistryCaller) GetPayAmounts(opts *bind.CallOpts, _payIds [][32]byte, _lastPayResolveDeadline *big.Int) ([]*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
	)
	out := ret0
	err := _PayRegistry.contract.Call(opts, out, "getPayAmounts", _payIds, _lastPayResolveDeadline)
	return *ret0, err
}

// GetPayAmounts is a free data retrieval call binding the contract method 0x7cac39cf.
//
// Solidity: function getPayAmounts(bytes32[] _payIds, uint256 _lastPayResolveDeadline) constant returns(uint256[])
func (_PayRegistry *PayRegistrySession) GetPayAmounts(_payIds [][32]byte, _lastPayResolveDeadline *big.Int) ([]*big.Int, error) {
	return _PayRegistry.Contract.GetPayAmounts(&_PayRegistry.CallOpts, _payIds, _lastPayResolveDeadline)
}

// GetPayAmounts is a free data retrieval call binding the contract method 0x7cac39cf.
//
// Solidity: function getPayAmounts(bytes32[] _payIds, uint256 _lastPayResolveDeadline) constant returns(uint256[])
func (_PayRegistry *PayRegistryCallerSession) GetPayAmounts(_payIds [][32]byte, _lastPayResolveDeadline *big.Int) ([]*big.Int, error) {
	return _PayRegistry.Contract.GetPayAmounts(&_PayRegistry.CallOpts, _payIds, _lastPayResolveDeadline)
}

// GetPayInfo is a free data retrieval call binding the contract method 0x27b0e058.
//
// Solidity: function getPayInfo(bytes32 _payId) constant returns(uint256, uint256)
func (_PayRegistry *PayRegistryCaller) GetPayInfo(opts *bind.CallOpts, _payId [32]byte) (*big.Int, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _PayRegistry.contract.Call(opts, out, "getPayInfo", _payId)
	return *ret0, *ret1, err
}

// GetPayInfo is a free data retrieval call binding the contract method 0x27b0e058.
//
// Solidity: function getPayInfo(bytes32 _payId) constant returns(uint256, uint256)
func (_PayRegistry *PayRegistrySession) GetPayInfo(_payId [32]byte) (*big.Int, *big.Int, error) {
	return _PayRegistry.Contract.GetPayInfo(&_PayRegistry.CallOpts, _payId)
}

// GetPayInfo is a free data retrieval call binding the contract method 0x27b0e058.
//
// Solidity: function getPayInfo(bytes32 _payId) constant returns(uint256, uint256)
func (_PayRegistry *PayRegistryCallerSession) GetPayInfo(_payId [32]byte) (*big.Int, *big.Int, error) {
	return _PayRegistry.Contract.GetPayInfo(&_PayRegistry.CallOpts, _payId)
}

// PayInfoMap is a free data retrieval call binding the contract method 0x8f13b2f5.
//
// Solidity: function payInfoMap(bytes32 ) constant returns(uint256 amount, uint256 resolveDeadline)
func (_PayRegistry *PayRegistryCaller) PayInfoMap(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Amount          *big.Int
	ResolveDeadline *big.Int
}, error) {
	ret := new(struct {
		Amount          *big.Int
		ResolveDeadline *big.Int
	})
	out := ret
	err := _PayRegistry.contract.Call(opts, out, "payInfoMap", arg0)
	return *ret, err
}

// PayInfoMap is a free data retrieval call binding the contract method 0x8f13b2f5.
//
// Solidity: function payInfoMap(bytes32 ) constant returns(uint256 amount, uint256 resolveDeadline)
func (_PayRegistry *PayRegistrySession) PayInfoMap(arg0 [32]byte) (struct {
	Amount          *big.Int
	ResolveDeadline *big.Int
}, error) {
	return _PayRegistry.Contract.PayInfoMap(&_PayRegistry.CallOpts, arg0)
}

// PayInfoMap is a free data retrieval call binding the contract method 0x8f13b2f5.
//
// Solidity: function payInfoMap(bytes32 ) constant returns(uint256 amount, uint256 resolveDeadline)
func (_PayRegistry *PayRegistryCallerSession) PayInfoMap(arg0 [32]byte) (struct {
	Amount          *big.Int
	ResolveDeadline *big.Int
}, error) {
	return _PayRegistry.Contract.PayInfoMap(&_PayRegistry.CallOpts, arg0)
}

// SetPayAmount is a paid mutator transaction binding the contract method 0xf8fb012f.
//
// Solidity: function setPayAmount(bytes32 _payHash, uint256 _amt) returns()
func (_PayRegistry *PayRegistryTransactor) SetPayAmount(opts *bind.TransactOpts, _payHash [32]byte, _amt *big.Int) (*types.Transaction, error) {
	return _PayRegistry.contract.Transact(opts, "setPayAmount", _payHash, _amt)
}

// SetPayAmount is a paid mutator transaction binding the contract method 0xf8fb012f.
//
// Solidity: function setPayAmount(bytes32 _payHash, uint256 _amt) returns()
func (_PayRegistry *PayRegistrySession) SetPayAmount(_payHash [32]byte, _amt *big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayAmount(&_PayRegistry.TransactOpts, _payHash, _amt)
}

// SetPayAmount is a paid mutator transaction binding the contract method 0xf8fb012f.
//
// Solidity: function setPayAmount(bytes32 _payHash, uint256 _amt) returns()
func (_PayRegistry *PayRegistryTransactorSession) SetPayAmount(_payHash [32]byte, _amt *big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayAmount(&_PayRegistry.TransactOpts, _payHash, _amt)
}

// SetPayAmounts is a paid mutator transaction binding the contract method 0x204a95ee.
//
// Solidity: function setPayAmounts(bytes32[] _payHashes, uint256[] _amts) returns()
func (_PayRegistry *PayRegistryTransactor) SetPayAmounts(opts *bind.TransactOpts, _payHashes [][32]byte, _amts []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.contract.Transact(opts, "setPayAmounts", _payHashes, _amts)
}

// SetPayAmounts is a paid mutator transaction binding the contract method 0x204a95ee.
//
// Solidity: function setPayAmounts(bytes32[] _payHashes, uint256[] _amts) returns()
func (_PayRegistry *PayRegistrySession) SetPayAmounts(_payHashes [][32]byte, _amts []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayAmounts(&_PayRegistry.TransactOpts, _payHashes, _amts)
}

// SetPayAmounts is a paid mutator transaction binding the contract method 0x204a95ee.
//
// Solidity: function setPayAmounts(bytes32[] _payHashes, uint256[] _amts) returns()
func (_PayRegistry *PayRegistryTransactorSession) SetPayAmounts(_payHashes [][32]byte, _amts []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayAmounts(&_PayRegistry.TransactOpts, _payHashes, _amts)
}

// SetPayDeadline is a paid mutator transaction binding the contract method 0xcdfa146b.
//
// Solidity: function setPayDeadline(bytes32 _payHash, uint256 _deadline) returns()
func (_PayRegistry *PayRegistryTransactor) SetPayDeadline(opts *bind.TransactOpts, _payHash [32]byte, _deadline *big.Int) (*types.Transaction, error) {
	return _PayRegistry.contract.Transact(opts, "setPayDeadline", _payHash, _deadline)
}

// SetPayDeadline is a paid mutator transaction binding the contract method 0xcdfa146b.
//
// Solidity: function setPayDeadline(bytes32 _payHash, uint256 _deadline) returns()
func (_PayRegistry *PayRegistrySession) SetPayDeadline(_payHash [32]byte, _deadline *big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayDeadline(&_PayRegistry.TransactOpts, _payHash, _deadline)
}

// SetPayDeadline is a paid mutator transaction binding the contract method 0xcdfa146b.
//
// Solidity: function setPayDeadline(bytes32 _payHash, uint256 _deadline) returns()
func (_PayRegistry *PayRegistryTransactorSession) SetPayDeadline(_payHash [32]byte, _deadline *big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayDeadline(&_PayRegistry.TransactOpts, _payHash, _deadline)
}

// SetPayDeadlines is a paid mutator transaction binding the contract method 0x0daddd34.
//
// Solidity: function setPayDeadlines(bytes32[] _payHashes, uint256[] _deadlines) returns()
func (_PayRegistry *PayRegistryTransactor) SetPayDeadlines(opts *bind.TransactOpts, _payHashes [][32]byte, _deadlines []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.contract.Transact(opts, "setPayDeadlines", _payHashes, _deadlines)
}

// SetPayDeadlines is a paid mutator transaction binding the contract method 0x0daddd34.
//
// Solidity: function setPayDeadlines(bytes32[] _payHashes, uint256[] _deadlines) returns()
func (_PayRegistry *PayRegistrySession) SetPayDeadlines(_payHashes [][32]byte, _deadlines []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayDeadlines(&_PayRegistry.TransactOpts, _payHashes, _deadlines)
}

// SetPayDeadlines is a paid mutator transaction binding the contract method 0x0daddd34.
//
// Solidity: function setPayDeadlines(bytes32[] _payHashes, uint256[] _deadlines) returns()
func (_PayRegistry *PayRegistryTransactorSession) SetPayDeadlines(_payHashes [][32]byte, _deadlines []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayDeadlines(&_PayRegistry.TransactOpts, _payHashes, _deadlines)
}

// SetPayInfo is a paid mutator transaction binding the contract method 0xe1e35490.
//
// Solidity: function setPayInfo(bytes32 _payHash, uint256 _amt, uint256 _deadline) returns()
func (_PayRegistry *PayRegistryTransactor) SetPayInfo(opts *bind.TransactOpts, _payHash [32]byte, _amt *big.Int, _deadline *big.Int) (*types.Transaction, error) {
	return _PayRegistry.contract.Transact(opts, "setPayInfo", _payHash, _amt, _deadline)
}

// SetPayInfo is a paid mutator transaction binding the contract method 0xe1e35490.
//
// Solidity: function setPayInfo(bytes32 _payHash, uint256 _amt, uint256 _deadline) returns()
func (_PayRegistry *PayRegistrySession) SetPayInfo(_payHash [32]byte, _amt *big.Int, _deadline *big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayInfo(&_PayRegistry.TransactOpts, _payHash, _amt, _deadline)
}

// SetPayInfo is a paid mutator transaction binding the contract method 0xe1e35490.
//
// Solidity: function setPayInfo(bytes32 _payHash, uint256 _amt, uint256 _deadline) returns()
func (_PayRegistry *PayRegistryTransactorSession) SetPayInfo(_payHash [32]byte, _amt *big.Int, _deadline *big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayInfo(&_PayRegistry.TransactOpts, _payHash, _amt, _deadline)
}

// SetPayInfos is a paid mutator transaction binding the contract method 0x414f7e0e.
//
// Solidity: function setPayInfos(bytes32[] _payHashes, uint256[] _amts, uint256[] _deadlines) returns()
func (_PayRegistry *PayRegistryTransactor) SetPayInfos(opts *bind.TransactOpts, _payHashes [][32]byte, _amts []*big.Int, _deadlines []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.contract.Transact(opts, "setPayInfos", _payHashes, _amts, _deadlines)
}

// SetPayInfos is a paid mutator transaction binding the contract method 0x414f7e0e.
//
// Solidity: function setPayInfos(bytes32[] _payHashes, uint256[] _amts, uint256[] _deadlines) returns()
func (_PayRegistry *PayRegistrySession) SetPayInfos(_payHashes [][32]byte, _amts []*big.Int, _deadlines []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayInfos(&_PayRegistry.TransactOpts, _payHashes, _amts, _deadlines)
}

// SetPayInfos is a paid mutator transaction binding the contract method 0x414f7e0e.
//
// Solidity: function setPayInfos(bytes32[] _payHashes, uint256[] _amts, uint256[] _deadlines) returns()
func (_PayRegistry *PayRegistryTransactorSession) SetPayInfos(_payHashes [][32]byte, _amts []*big.Int, _deadlines []*big.Int) (*types.Transaction, error) {
	return _PayRegistry.Contract.SetPayInfos(&_PayRegistry.TransactOpts, _payHashes, _amts, _deadlines)
}

// PayRegistryPayInfoUpdateIterator is returned from FilterPayInfoUpdate and is used to iterate over the raw logs and unpacked data for PayInfoUpdate events raised by the PayRegistry contract.
type PayRegistryPayInfoUpdateIterator struct {
	Event *PayRegistryPayInfoUpdate // Event containing the contract specifics and raw log

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
func (it *PayRegistryPayInfoUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PayRegistryPayInfoUpdate)
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
		it.Event = new(PayRegistryPayInfoUpdate)
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
func (it *PayRegistryPayInfoUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PayRegistryPayInfoUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PayRegistryPayInfoUpdate represents a PayInfoUpdate event raised by the PayRegistry contract.
type PayRegistryPayInfoUpdate struct {
	PayId           [32]byte
	Amount          *big.Int
	ResolveDeadline *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterPayInfoUpdate is a free log retrieval operation binding the contract event 0x9e9acc6d43d5d7bd6fa143ef0ee1d224cfe2bb010b7e3acf44878d6314ebc607.
//
// Solidity: event PayInfoUpdate(bytes32 indexed payId, uint256 amount, uint256 resolveDeadline)
func (_PayRegistry *PayRegistryFilterer) FilterPayInfoUpdate(opts *bind.FilterOpts, payId [][32]byte) (*PayRegistryPayInfoUpdateIterator, error) {

	var payIdRule []interface{}
	for _, payIdItem := range payId {
		payIdRule = append(payIdRule, payIdItem)
	}

	logs, sub, err := _PayRegistry.contract.FilterLogs(opts, "PayInfoUpdate", payIdRule)
	if err != nil {
		return nil, err
	}
	return &PayRegistryPayInfoUpdateIterator{contract: _PayRegistry.contract, event: "PayInfoUpdate", logs: logs, sub: sub}, nil
}

// WatchPayInfoUpdate is a free log subscription operation binding the contract event 0x9e9acc6d43d5d7bd6fa143ef0ee1d224cfe2bb010b7e3acf44878d6314ebc607.
//
// Solidity: event PayInfoUpdate(bytes32 indexed payId, uint256 amount, uint256 resolveDeadline)
func (_PayRegistry *PayRegistryFilterer) WatchPayInfoUpdate(opts *bind.WatchOpts, sink chan<- *PayRegistryPayInfoUpdate, payId [][32]byte) (event.Subscription, error) {

	var payIdRule []interface{}
	for _, payIdItem := range payId {
		payIdRule = append(payIdRule, payIdItem)
	}

	logs, sub, err := _PayRegistry.contract.WatchLogs(opts, "PayInfoUpdate", payIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PayRegistryPayInfoUpdate)
				if err := _PayRegistry.contract.UnpackLog(event, "PayInfoUpdate", log); err != nil {
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
