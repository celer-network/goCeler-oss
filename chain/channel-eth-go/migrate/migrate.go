// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package migrate

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

// LedgerMigrateABI is the input ABI used to generate the binding from.
const LedgerMigrateABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"newLedgerAddr\",\"type\":\"address\"}],\"name\":\"MigrateChannelTo\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"oldLedgerAddr\",\"type\":\"address\"}],\"name\":\"MigrateChannelFrom\",\"type\":\"event\"}]"

// LedgerMigrateBin is the compiled bytecode used for deploying new contracts.
const LedgerMigrateBin = `0x6112f8610026600b82828239805160001a60731461001957fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600436106100435760e060020a60003504633c50ec72811461004857806382b4338a146100de575b600080fd5b81801561005457600080fd5b506100cc6004803603604081101561006b57600080fd5b8135919081019060408101602082013564010000000081111561008d57600080fd5b82018360208201111561009f57600080fd5b803590602001918460018302840111640100000000831117156100c157600080fd5b509092509050610172565b60408051918252519081900360200190f35b8180156100ea57600080fd5b506101706004803603606081101561010157600080fd5b813591600160a060020a036020820135169181019060608101604082013564010000000081111561013157600080fd5b82018360208201111561014357600080fd5b8035906020019184600183028401116401000000008311171561016557600080fd5b509092509050610505565b005b600061017c611206565b6101bb84848080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061078b92505050565b90506101c5611220565b81516101d0906108e2565b80516000818152600689016020526040908190209083015192935090916001600383015460ff16600481111561020257fe5b148061022057506002600383015460ff16600481111561021e57fe5b145b61022957600080fd5b600085600001516040518082805190602001908083835b6020831061025f5780518252601f199092019160209182019101610240565b51815160209384036101000a6000190180199092169116179052604051919093018190039020918a015191945061029d9350869250849190506109cc565b6102f1576040805160e560020a62461bcd02815260206004820152601460248201527f436865636b20636f2d73696773206661696c6564000000000000000000000000604482015290519081900360640190fd5b6020850151600160a060020a03163014610355576040805160e560020a62461bcd02815260206004820152601f60248201527f46726f6d206c65646765722061646472657373206973206e6f74207468697300604482015290519081900360640190fd5b600160a060020a038216331461039f5760405160e560020a62461bcd0281526004018080602001828103825260238152602001806112806023913960400191505060405180910390fd5b84606001514311156103fb576040805160e560020a62461bcd02815260206004820152601960248201527f506173736564206d6967726174696f6e20646561646c696e6500000000000000604482015290519081900360640190fd5b61040d8a84600463ffffffff610a6816565b60038301805474ffffffffffffffffffffffffffffffffffffffff001916610100600160a060020a0385169081029190911790915560405185907fdefb8a94bbfc44ef5297b035407a7dd1314f369e39c3301f5b90f8810fb9fe4f90600090a360038a0154604080517fa0c89a8c00000000000000000000000000000000000000000000000000000000815260048101879052600160a060020a0385811660248301529151919092169163a0c89a8c91604480830192600092919082900301818387803b1580156104dd57600080fd5b505af11580156104f1573d6000803e3d6000fd5b5095985050505050505050505b9392505050565b6040517fe0a515b7000000000000000000000000000000000000000000000000000000008152602060048201908152602482018390528491600091600160a060020a0384169163e0a515b79187918791908190604401848480828437600081840152601f19601f8201169050808301925050509350505050602060405180830381600087803b15801561059757600080fd5b505af11580156105ab573d6000803e3d6000fd5b505050506040513d60208110156105c157600080fd5b505160008181526006880160205260408120919250600382015460ff1660048111156105e957fe5b146106285760405160e560020a62461bcd0281526004018080602001828103825260218152602001806112a36021913960400191505060405180910390fd5b6003870154604080517fa96a5f940000000000000000000000000000000000000000000000000000000081526004810185905290513092600160a060020a03169163a96a5f94916024808301926020929190829003018186803b15801561068e57600080fd5b505afa1580156106a2573d6000803e3d6000fd5b505050506040513d60208110156106b857600080fd5b5051600160a060020a031614610718576040805160e560020a62461bcd02815260206004820152601c60248201527f4f70657261746f7273686970206e6f74207472616e7366657272656400000000604482015290519081900360640190fd5b61072a8782600163ffffffff610a6816565b61073b81848463ffffffff610b8916565b61074c81848463ffffffff610c8916565b604051600160a060020a0387169083907f141a72a1d915a7c4205104b6e564cc991aa827c5f2c672a5d6a1da8bef99d6eb90600090a350505050505050565b610793611206565b61079b611247565b6107a483610e3c565b905060606107b982600263ffffffff610e5316565b9050806002815181106107c857fe5b602002602001015160405190808252806020026020018201604052801561080357816020015b60608152602001906001900390816107ee5790505b50836020018190525060008160028151811061081b57fe5b6020026020010181815250506000805b61083484610ee3565b156108d95761084284610ef2565b909250905081600114156108605761085984610f1f565b85526108d4565b81600214156108c45761087284610f1f565b85602001518460028151811061088457fe5b60200260200101518151811061089657fe5b6020026020010181905250826002815181106108ae57fe5b60209081029190910101805160010190526108d4565b6108d4848263ffffffff610fac16565b61082b565b50505050919050565b6108ea611220565b6108f2611247565b6108fb83610e3c565b90506000805b61090a83610ee3565b156109c45761091883610ef2565b9092509050816001141561093e5761093761093284610f1f565b61100d565b84526109bf565b816002141561096b5761095861095384610f1f565b611025565b600160a060020a031660208501526109bf565b81600314156109935761098061095384610f1f565b600160a060020a031660408501526109bf565b81600414156109af576109a583611036565b60608501526109bf565b6109bf838263ffffffff610fac16565b610901565b505050919050565b600081516002146109df575060006104fe565b60006109ea84611094565b90506000805b6002811015610a5b57610a1f858281518110610a0857fe5b6020026020010151846110e590919063ffffffff16565b9150866004018160028110610a3057fe5b6008020154600160a060020a03838116911614610a5357600093505050506104fe565b6001016109f0565b5060019695505050505050565b806004811115610a7457fe5b600383015460ff166004811115610a8757fe5b1415610a9257610b84565b6000600383015460ff166004811115610aa757fe5b14610b12576003820154610ae890600190859060009060ff166004811115610acb57fe5b8152602001908152602001600020546111b790919063ffffffff16565b6003830154849060009060ff166004811115610b0057fe5b81526020810191909152604001600020555b610b436001846000846004811115610b2657fe5b8152602001908152602001600020546111cc90919063ffffffff16565b836000836004811115610b5257fe5b815260208101919091526040016000205560038201805482919060ff19166001836004811115610b7e57fe5b02179055505b505050565b600082600160a060020a0316632f0ac304836040518263ffffffff1660e060020a0281526004018082815260200191505060806040518083038186803b158015610bd257600080fd5b505afa158015610be6573d6000803e3d6000fd5b505050506040513d6080811015610bfc57600080fd5b50805160208201516040830151606090930151601488015560028088018054600160a060020a039095166101000274ffffffffffffffffffffffffffffffffffffffff001990951694909417909355600187019190915591508190811115610c6057fe5b60028086018054909160ff19909116906001908490811115610c7e57fe5b021790555050505050565b610c91611261565b610c99611261565b610ca1611261565b610ca9611261565b610cb1611261565b610cb9611261565b87600160a060020a03166388f41465886040518263ffffffff1660e060020a028152600401808281526020019150506101806040518083038186803b158015610d0157600080fd5b505afa158015610d15573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250610180811015610d3b57600080fd5b50955050604085019350506080840191505060c083016101008401610140850160005b6002811015610e305760008a6004018260028110610d7857fe5b600802019050878260028110610d8a57fe5b6020020151815473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03909116178155868260028110610dc257fe5b60200201516001820155858260028110610dd857fe5b60200201518160020181905550848260028110610df157fe5b60200201516003820155838260028110610e0757fe5b60200201516004820155828260028110610e1d57fe5b6020020151600790910155600101610d5e565b50505050505050505050565b610e44611247565b60208101919091526000815290565b815160408051600184018082526020808202830101909252606092918015610e85578160200160208202803883390190505b5091506000805b610e9586610ee3565b15610eda57610ea386610ef2565b80925081935050506001848381518110610eb957fe5b602002602001018181510191508181525050610ed58682610fac565b610e8c565b50509092525090565b6020810151518151105b919050565b6000806000610f0084611036565b9050600881049250806007166005811115610f1757fe5b915050915091565b60606000610f2c83611036565b8351602085015151919250820190811115610f4657600080fd5b816040519080825280601f01601f191660200182016040528015610f71576020820181803883390190505b50602080860151865192955091818601919083010160005b85811015610fa1578181015183820152602001610f89565b505050935250919050565b6000816005811115610fba57fe5b1415610fcf57610fc982611036565b50611009565b6002816005811115610fdd57fe5b1415610043576000610fee83611036565b835181018085526020850151519192501115610fc957600080fd5b5050565b6000815160201461101d57600080fd5b506020015190565b6000611030826111de565b92915050565b602080820151825181019091015160009182805b600a81101561108e5783811a91508060070282607f169060020a0285179450816080166000141561108657855101600101855250610eed915050565b60010161104a565b50600080fd5b604080517f19457468657265756d205369676e6564204d6573736167653a0a333200000000602080830191909152603c8083019490945282518083039094018452605c909101909152815191012090565b60008060008084516041146111005760009350505050611030565b50505060208201516040830151606084015160001a601b81101561112257601b015b8060ff16601b1415801561113a57508060ff16601c14155b1561114b5760009350505050611030565b6040805160008152602080820180845289905260ff8416828401526060820186905260808201859052915160019260a0808401939192601f1981019281900390910190855afa1580156111a2573d6000803e3d6000fd5b5050604051601f190151979650505050505050565b6000828211156111c657600080fd5b50900390565b6000828201838110156104fe57600080fd5b600081516014146111ee57600080fd5b50602001516c01000000000000000000000000900490565b604051806040016040528060608152602001606081525090565b60408051608081018252600080825260208201819052918101829052606081019190915290565b604051806040016040528060008152602001606081525090565b6040518060400160405280600290602082028038833950919291505056fe546f206c65646765722061646472657373206973206e6f74206d73672e73656e646572496d6d69677261746564206368616e6e656c20616c726561647920657869737473a265627a7a72305820a023589bbfaa8f8c149ab9c063f9e92add5e0946c733e8f1fc6a89d36a81094564736f6c634300050a0032`

// DeployLedgerMigrate deploys a new Ethereum contract, binding an instance of LedgerMigrate to it.
func DeployLedgerMigrate(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *LedgerMigrate, error) {
	parsed, err := abi.JSON(strings.NewReader(LedgerMigrateABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(LedgerMigrateBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &LedgerMigrate{LedgerMigrateCaller: LedgerMigrateCaller{contract: contract}, LedgerMigrateTransactor: LedgerMigrateTransactor{contract: contract}, LedgerMigrateFilterer: LedgerMigrateFilterer{contract: contract}}, nil
}

// LedgerMigrate is an auto generated Go binding around an Ethereum contract.
type LedgerMigrate struct {
	LedgerMigrateCaller     // Read-only binding to the contract
	LedgerMigrateTransactor // Write-only binding to the contract
	LedgerMigrateFilterer   // Log filterer for contract events
}

// LedgerMigrateCaller is an auto generated read-only Go binding around an Ethereum contract.
type LedgerMigrateCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerMigrateTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LedgerMigrateTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerMigrateFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LedgerMigrateFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerMigrateSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LedgerMigrateSession struct {
	Contract     *LedgerMigrate    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LedgerMigrateCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LedgerMigrateCallerSession struct {
	Contract *LedgerMigrateCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// LedgerMigrateTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LedgerMigrateTransactorSession struct {
	Contract     *LedgerMigrateTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// LedgerMigrateRaw is an auto generated low-level Go binding around an Ethereum contract.
type LedgerMigrateRaw struct {
	Contract *LedgerMigrate // Generic contract binding to access the raw methods on
}

// LedgerMigrateCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LedgerMigrateCallerRaw struct {
	Contract *LedgerMigrateCaller // Generic read-only contract binding to access the raw methods on
}

// LedgerMigrateTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LedgerMigrateTransactorRaw struct {
	Contract *LedgerMigrateTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLedgerMigrate creates a new instance of LedgerMigrate, bound to a specific deployed contract.
func NewLedgerMigrate(address common.Address, backend bind.ContractBackend) (*LedgerMigrate, error) {
	contract, err := bindLedgerMigrate(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LedgerMigrate{LedgerMigrateCaller: LedgerMigrateCaller{contract: contract}, LedgerMigrateTransactor: LedgerMigrateTransactor{contract: contract}, LedgerMigrateFilterer: LedgerMigrateFilterer{contract: contract}}, nil
}

// NewLedgerMigrateCaller creates a new read-only instance of LedgerMigrate, bound to a specific deployed contract.
func NewLedgerMigrateCaller(address common.Address, caller bind.ContractCaller) (*LedgerMigrateCaller, error) {
	contract, err := bindLedgerMigrate(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LedgerMigrateCaller{contract: contract}, nil
}

// NewLedgerMigrateTransactor creates a new write-only instance of LedgerMigrate, bound to a specific deployed contract.
func NewLedgerMigrateTransactor(address common.Address, transactor bind.ContractTransactor) (*LedgerMigrateTransactor, error) {
	contract, err := bindLedgerMigrate(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LedgerMigrateTransactor{contract: contract}, nil
}

// NewLedgerMigrateFilterer creates a new log filterer instance of LedgerMigrate, bound to a specific deployed contract.
func NewLedgerMigrateFilterer(address common.Address, filterer bind.ContractFilterer) (*LedgerMigrateFilterer, error) {
	contract, err := bindLedgerMigrate(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LedgerMigrateFilterer{contract: contract}, nil
}

// bindLedgerMigrate binds a generic wrapper to an already deployed contract.
func bindLedgerMigrate(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(LedgerMigrateABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LedgerMigrate *LedgerMigrateRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LedgerMigrate.Contract.LedgerMigrateCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LedgerMigrate *LedgerMigrateRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LedgerMigrate.Contract.LedgerMigrateTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LedgerMigrate *LedgerMigrateRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LedgerMigrate.Contract.LedgerMigrateTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LedgerMigrate *LedgerMigrateCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LedgerMigrate.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LedgerMigrate *LedgerMigrateTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LedgerMigrate.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LedgerMigrate *LedgerMigrateTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LedgerMigrate.Contract.contract.Transact(opts, method, params...)
}

// LedgerMigrateMigrateChannelFromIterator is returned from FilterMigrateChannelFrom and is used to iterate over the raw logs and unpacked data for MigrateChannelFrom events raised by the LedgerMigrate contract.
type LedgerMigrateMigrateChannelFromIterator struct {
	Event *LedgerMigrateMigrateChannelFrom // Event containing the contract specifics and raw log

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
func (it *LedgerMigrateMigrateChannelFromIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LedgerMigrateMigrateChannelFrom)
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
		it.Event = new(LedgerMigrateMigrateChannelFrom)
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
func (it *LedgerMigrateMigrateChannelFromIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LedgerMigrateMigrateChannelFromIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LedgerMigrateMigrateChannelFrom represents a MigrateChannelFrom event raised by the LedgerMigrate contract.
type LedgerMigrateMigrateChannelFrom struct {
	ChannelId     [32]byte
	OldLedgerAddr common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMigrateChannelFrom is a free log retrieval operation binding the contract event 0x141a72a1d915a7c4205104b6e564cc991aa827c5f2c672a5d6a1da8bef99d6eb.
//
// Solidity: event MigrateChannelFrom(bytes32 indexed channelId, address indexed oldLedgerAddr)
func (_LedgerMigrate *LedgerMigrateFilterer) FilterMigrateChannelFrom(opts *bind.FilterOpts, channelId [][32]byte, oldLedgerAddr []common.Address) (*LedgerMigrateMigrateChannelFromIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var oldLedgerAddrRule []interface{}
	for _, oldLedgerAddrItem := range oldLedgerAddr {
		oldLedgerAddrRule = append(oldLedgerAddrRule, oldLedgerAddrItem)
	}

	logs, sub, err := _LedgerMigrate.contract.FilterLogs(opts, "MigrateChannelFrom", channelIdRule, oldLedgerAddrRule)
	if err != nil {
		return nil, err
	}
	return &LedgerMigrateMigrateChannelFromIterator{contract: _LedgerMigrate.contract, event: "MigrateChannelFrom", logs: logs, sub: sub}, nil
}

// WatchMigrateChannelFrom is a free log subscription operation binding the contract event 0x141a72a1d915a7c4205104b6e564cc991aa827c5f2c672a5d6a1da8bef99d6eb.
//
// Solidity: event MigrateChannelFrom(bytes32 indexed channelId, address indexed oldLedgerAddr)
func (_LedgerMigrate *LedgerMigrateFilterer) WatchMigrateChannelFrom(opts *bind.WatchOpts, sink chan<- *LedgerMigrateMigrateChannelFrom, channelId [][32]byte, oldLedgerAddr []common.Address) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var oldLedgerAddrRule []interface{}
	for _, oldLedgerAddrItem := range oldLedgerAddr {
		oldLedgerAddrRule = append(oldLedgerAddrRule, oldLedgerAddrItem)
	}

	logs, sub, err := _LedgerMigrate.contract.WatchLogs(opts, "MigrateChannelFrom", channelIdRule, oldLedgerAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LedgerMigrateMigrateChannelFrom)
				if err := _LedgerMigrate.contract.UnpackLog(event, "MigrateChannelFrom", log); err != nil {
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

// LedgerMigrateMigrateChannelToIterator is returned from FilterMigrateChannelTo and is used to iterate over the raw logs and unpacked data for MigrateChannelTo events raised by the LedgerMigrate contract.
type LedgerMigrateMigrateChannelToIterator struct {
	Event *LedgerMigrateMigrateChannelTo // Event containing the contract specifics and raw log

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
func (it *LedgerMigrateMigrateChannelToIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LedgerMigrateMigrateChannelTo)
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
		it.Event = new(LedgerMigrateMigrateChannelTo)
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
func (it *LedgerMigrateMigrateChannelToIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LedgerMigrateMigrateChannelToIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LedgerMigrateMigrateChannelTo represents a MigrateChannelTo event raised by the LedgerMigrate contract.
type LedgerMigrateMigrateChannelTo struct {
	ChannelId     [32]byte
	NewLedgerAddr common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMigrateChannelTo is a free log retrieval operation binding the contract event 0xdefb8a94bbfc44ef5297b035407a7dd1314f369e39c3301f5b90f8810fb9fe4f.
//
// Solidity: event MigrateChannelTo(bytes32 indexed channelId, address indexed newLedgerAddr)
func (_LedgerMigrate *LedgerMigrateFilterer) FilterMigrateChannelTo(opts *bind.FilterOpts, channelId [][32]byte, newLedgerAddr []common.Address) (*LedgerMigrateMigrateChannelToIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var newLedgerAddrRule []interface{}
	for _, newLedgerAddrItem := range newLedgerAddr {
		newLedgerAddrRule = append(newLedgerAddrRule, newLedgerAddrItem)
	}

	logs, sub, err := _LedgerMigrate.contract.FilterLogs(opts, "MigrateChannelTo", channelIdRule, newLedgerAddrRule)
	if err != nil {
		return nil, err
	}
	return &LedgerMigrateMigrateChannelToIterator{contract: _LedgerMigrate.contract, event: "MigrateChannelTo", logs: logs, sub: sub}, nil
}

// WatchMigrateChannelTo is a free log subscription operation binding the contract event 0xdefb8a94bbfc44ef5297b035407a7dd1314f369e39c3301f5b90f8810fb9fe4f.
//
// Solidity: event MigrateChannelTo(bytes32 indexed channelId, address indexed newLedgerAddr)
func (_LedgerMigrate *LedgerMigrateFilterer) WatchMigrateChannelTo(opts *bind.WatchOpts, sink chan<- *LedgerMigrateMigrateChannelTo, channelId [][32]byte, newLedgerAddr []common.Address) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var newLedgerAddrRule []interface{}
	for _, newLedgerAddrItem := range newLedgerAddr {
		newLedgerAddrRule = append(newLedgerAddrRule, newLedgerAddrItem)
	}

	logs, sub, err := _LedgerMigrate.contract.WatchLogs(opts, "MigrateChannelTo", channelIdRule, newLedgerAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LedgerMigrateMigrateChannelTo)
				if err := _LedgerMigrate.contract.UnpackLog(event, "MigrateChannelTo", log); err != nil {
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
