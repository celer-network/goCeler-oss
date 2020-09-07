// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package testapp

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

// SimpleSingleSessionAppABI is the input ABI used to generate the binding from.
const SimpleSingleSessionAppABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_stateProof\",\"type\":\"bytes\"}],\"name\":\"intendSettle\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_action\",\"type\":\"bytes\"}],\"name\":\"applyAction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getSeqNum\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getSettleFinalizedTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getActionDeadline\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeOnActionTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_players\",\"type\":\"address[]\"},{\"name\":\"_nonce\",\"type\":\"uint256\"},{\"name\":\"_timeout\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"}],\"name\":\"IntendSettle\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"getOutcome\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SimpleSingleSessionAppBin is the compiled bytecode used for deploying new contracts.
var SimpleSingleSessionAppBin = "0x60806040523480156200001157600080fd5b5060405162001b6238038062001b62833981018060405260608110156200003757600080fd5b8101908080516401000000008111156200005057600080fd5b828101905060208101848111156200006757600080fd5b81518560208202830111640100000000821117156200008557600080fd5b505092919060200180519060200190929190805190602001909291905050508282828282828160008001819055508260006001019080519060200190620000ce92919062000119565b50806000600301819055506000806002018190555060008060050160006101000a81548160ff021916908360038111156200010557fe5b0217905550505050505050505050620001ee565b82805482825590600052602060002090810192821562000195579160200282015b82811115620001945782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550916020019190600101906200013a565b5b509050620001a49190620001a8565b5090565b620001eb91905b80821115620001e757600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff021916905550600101620001af565b5090565b90565b61196480620001fe6000396000f3fe608060405234801561001057600080fd5b506004361061009e5760003560e01c8063b71ca01f11610066578063b71ca01f14610286578063bbc35280146102a4578063bcdbda94146102c2578063ea4ba8eb14610353578063fa5e7ff5146103e45761009e565b8063130d33fe146100a35780631f2b71e51461011c57806344c9af28146101955780634e69d5601461023c5780636d15c45714610268575b600080fd5b61011a600480360360208110156100b957600080fd5b81019080803590602001906401000000008111156100d657600080fd5b8201836020820111156100e857600080fd5b8035906020019184600183028401116401000000008311171561010a57600080fd5b90919293919293905050506103ee565b005b6101936004803603602081101561013257600080fd5b810190808035906020019064010000000081111561014f57600080fd5b82018360208201111561016157600080fd5b8035906020019184600183028401116401000000008311171561018357600080fd5b9091929391929390505050610780565b005b6101c1600480360360208110156101ab57600080fd5b81019080803590602001909291905050506109a5565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156102015780820151818401526020810190506101e6565b50505050905090810190601f16801561022e5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b610244610a03565b6040518082600381111561025457fe5b60ff16815260200191505060405180910390f35b610270610a1c565b6040518082815260200191505060405180910390f35b61028e610a28565b6040518082815260200191505060405180910390f35b6102ac610a70565b6040518082815260200191505060405180910390f35b610339600480360360208110156102d857600080fd5b81019080803590602001906401000000008111156102f557600080fd5b82018360208201111561030757600080fd5b8035906020019184600183028401116401000000008311171561032957600080fd5b9091929391929390505050610afd565b604051808215151515815260200191505060405180910390f35b6103ca6004803603602081101561036957600080fd5b810190808035906020019064010000000081111561038657600080fd5b82018360208201111561039857600080fd5b803590602001918460018302840111640100000000831117156103ba57600080fd5b9091929391929390505050610b41565b604051808215151515815260200191505060405180910390f35b6103ec610bf5565b005b6003808111156103fa57fe5b600060050160009054906101000a900460ff16600381111561041857fe5b141561048c576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b6104946118dc565b6104e183838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050610c9b565b90506104f581600001518260200151610e12565b610567576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260118152602001807f696e76616c6964207369676e617475726500000000000000000000000000000081525060200191505060405180910390fd5b61056f6118f6565b61057c8260000151610f5c565b905060008001548160000151146105fb576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600f8152602001807f6e6f6e6365206e6f74206d61746368000000000000000000000000000000000081525060200191505060405180910390fd5b806020015160006002015410610679576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260178152602001807f696e76616c69642073657175656e6365206e756d62657200000000000000000081525060200191505060405180910390fd5b80602001516000600201819055506001600060050160006101000a81548160ff021916908360038111156106a957fe5b021790555060006003015443016000600401819055506106cc816040015161104a565b61073e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f737461746520757064617465206661696c65640000000000000000000000000081525060200191505060405180910390fd5b7fce68db27527c6058059e8cbd1c6de0528ef1c417fe1c21c545919c7da3466d2a6000600201546040518082815260200191505060405180910390a150505050565b60038081111561078c57fe5b600060050160009054906101000a900460ff1660038111156107aa57fe5b141561081e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b6001600381111561082b57fe5b600060050160009054906101000a900460ff16600381111561084957fe5b14801561085a575060006004015443115b15610887576002600060050160006101000a81548160ff0219169083600381111561088157fe5b02179055505b6002600381111561089457fe5b600060050160009054906101000a900460ff1660038111156108b257fe5b14610925576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f617070206e6f7420696e20616374696f6e206d6f64650000000000000000000081525060200191505060405180910390fd5b600060020160008154809291906001019190505550600060030154430160006004018190555061099882828080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050611161565b6109a157600080fd5b5050565b60608060206040519080825280601f01601f1916602001820160405280156109dc5781602001600182028038833980820191505090505b5090506000600660009054906101000a900460ff1690508060208301528192505050919050565b60008060050160009054906101000a900460ff16905090565b60008060020154905090565b600060016003811115610a3757fe5b600060050160009054906101000a900460ff166003811115610a5557fe5b1415610a68576000600401549050610a6d565b600090505b90565b600060026003811115610a7f57fe5b600060050160009054906101000a900460ff166003811115610a9d57fe5b1415610ab0576000600401549050610afa565b60016003811115610abd57fe5b600060050160009054906101000a900460ff166003811115610adb57fe5b1415610af557600060030154600060040154019050610afa565b600090505b90565b6000808383905014610b0e57600080fd5b600380811115610b1a57fe5b600060050160009054906101000a900460ff166003811115610b3857fe5b14905092915050565b600060018383905014610bbc576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f696e76616c6964207175657279206c656e67746800000000000000000000000081525060200191505060405180910390fd5b82826000818110610bc957fe5b9050013560f81c60f81b60f81c60ff16600660009054906101000a900460ff1660ff1614905092915050565b60026003811115610c0257fe5b600060050160009054906101000a900460ff166003811115610c2057fe5b1415610c3c576000600401544311610c3757600080fd5b610c90565b60016003811115610c4957fe5b600060050160009054906101000a900460ff166003811115610c6757fe5b1415610c8a57600060030154600060040154014311610c8557600080fd5b610c8f565b610c99565b5b610c98611278565b5b565b610ca36118dc565b610cab61191e565b610cb4836112a1565b90506060610ccc6002836112d090919063ffffffff16565b905080600281518110610cdb57fe5b6020026020010151604051908082528060200260200182016040528015610d1657816020015b6060815260200190600190039081610d015790505b508360200181905250600081600281518110610d2e57fe5b6020026020010181815250506000805b610d4784611375565b15610e0957610d558461138a565b8092508193505050600015610d6957610e04565b6001821415610d8857610d7b846113be565b8560000181905250610e03565b6002821415610dee57610d9a846113be565b856020015184600281518110610dac57fe5b602002602001015181518110610dbe57fe5b602002602001018190525082600281518110610dd657fe5b60200260200101805180919060010181525050610e02565b610e01818561147790919063ffffffff16565b5b5b5b610d3e565b50505050919050565b60008060010180549050825114610e91576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601c8152602001807f696e76616c6964206e756d626572206f66207369676e6174757265730000000081525060200191505060405180910390fd5b6060610e9f84846000611507565b905060008090505b600060010180549050811015610f4f57818181518110610ec357fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff1660006001018281548110610ef157fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614610f4257600092505050610f56565b8080600101915050610ea7565b5060019150505b92915050565b610f646118f6565b610f6c61191e565b610f75836112a1565b90506000805b610f8483611375565b1561104257610f928361138a565b8092508193505050600015610fa65761103d565b6001821415610fc657610fb88361171e565b84600001818152505061103c565b6002821415610fe657610fd88361171e565b84602001818152505061103b565b600382141561100557610ff8836113be565b846040018190525061103a565b6004821415611025576110178361171e565b846060018181525050611039565b611038818461147790919063ffffffff16565b5b5b5b5b5b610f7b565b505050919050565b600060018251146110c3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f696e76616c6964207374617465206c656e67746800000000000000000000000081525060200191505060405180910390fd5b816000815181106110d057fe5b602001015160f81c60f81b60f81c600660006101000a81548160ff021916908360ff1602179055506001600660009054906101000a900460ff1660ff16148061112b57506002600660009054906101000a900460ff1660ff16145b15611158576003600060050160006101000a81548160ff0219169083600381111561115257fe5b02179055505b60019050919050565b600060018251146111da576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f696e76616c696420616374696f6e206c656e677468000000000000000000000081525060200191505060405180910390fd5b816000815181106111e757fe5b602001015160f81c60f81b60f81c600660006101000a81548160ff021916908360ff1602179055506001600660009054906101000a900460ff1660ff16148061124257506002600660009054906101000a900460ff1660ff16145b1561126f576003600060050160006101000a81548160ff0219169083600381111561126957fe5b02179055505b60019050919050565b6003600060050160006101000a81548160ff0219169083600381111561129a57fe5b0217905550565b6112a961191e565b60018251116112b757600080fd5b8181602001819052506000816000018181525050919050565b60606000836000015190506001830160405190808252806020026020018201604052801561130d5781602001602082028038833980820191505090505b5091506000805b61131d86611375565b156113625761132b8661138a565b8092508193505050600184838151811061134157fe5b60200260200101818151019150818152505061135d8682611477565b611314565b8286600001818152505050505092915050565b60008160200151518260000151109050919050565b60008060006113988461171e565b9050600881816113a457fe5b0492506007811660058111156113b657fe5b915050915091565b606060006113cb8361171e565b905060008184600001510190508360200151518111156113ea57600080fd5b816040519080825280601f01601f19166020018201604052801561141d5781602001600182028038833980820191505090505b50925060608460200151905060008086600001519050602086019150806020840101905060008090505b85811015611462578082015181840152602081019050611447565b50838760000181815250505050505050919050565b6000600581111561148457fe5b81600581111561149057fe5b14156114a55761149f8261171e565b50611503565b600260058111156114b257fe5b8160058111156114be57fe5b14156114fd5760006114cf8361171e565b905080836000018181510191508181525050826020015151836000015111156114f757600080fd5b50611502565b600080fd5b5b5050565b606080835160405190808252806020026020018201604052801561153a5781602001602082028038833980820191505090505b50905060006115b9866040516020018082805190602001908083835b602083106115795780518252602082019150602081019050602083039250611556565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051602081830303815290604052805190602001206117a4565b9050600080905060008090505b8651811015611710576115ec838883815181106115df57fe5b60200260200101516117fc565b8482815181106115f857fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508515611703578173ffffffffffffffffffffffffffffffffffffffff1684828151811061165b57fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff16116116ec576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f7369676e657273206e6f7420696e20617363656e64696e67206f72646572000081525060200191505060405180910390fd5b8381815181106116f857fe5b602002602001015191505b80806001019150506115c6565b508293505050509392505050565b60008060608360200151905083600001519250826020820101519150600080935060008090505b600a8110156117995783811a915060078102607f8316901b85179450600060808316141561178c57600181018660000181815101915081815250508494505050505061179f565b8080600101915050611745565b50600080fd5b919050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b6000604182511461181057600090506118d6565b60008060006020850151925060408501519150606085015160001a9050601b8160ff16101561184057601b810190505b601b8160ff16141580156118585750601c8160ff1614155b1561186957600093505050506118d6565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa1580156118c6573d6000803e3d6000fd5b5050506020604051035193505050505b92915050565b604051806040016040528060608152602001606081525090565b6040518060800160405280600081526020016000815260200160608152602001600081525090565b60405180604001604052806000815260200160608152509056fea165627a7a723058202cff3ccecd57e354a68effd0ab29670d09f8325852c811d185e3f6dfcef547310029"

// DeploySimpleSingleSessionApp deploys a new Ethereum contract, binding an instance of SimpleSingleSessionApp to it.
func DeploySimpleSingleSessionApp(auth *bind.TransactOpts, backend bind.ContractBackend, _players []common.Address, _nonce *big.Int, _timeout *big.Int) (common.Address, *types.Transaction, *SimpleSingleSessionApp, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSingleSessionAppABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleSingleSessionAppBin), backend, _players, _nonce, _timeout)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleSingleSessionApp{SimpleSingleSessionAppCaller: SimpleSingleSessionAppCaller{contract: contract}, SimpleSingleSessionAppTransactor: SimpleSingleSessionAppTransactor{contract: contract}, SimpleSingleSessionAppFilterer: SimpleSingleSessionAppFilterer{contract: contract}}, nil
}

// SimpleSingleSessionApp is an auto generated Go binding around an Ethereum contract.
type SimpleSingleSessionApp struct {
	SimpleSingleSessionAppCaller     // Read-only binding to the contract
	SimpleSingleSessionAppTransactor // Write-only binding to the contract
	SimpleSingleSessionAppFilterer   // Log filterer for contract events
}

// SimpleSingleSessionAppCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSingleSessionAppTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSingleSessionAppFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleSingleSessionAppFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSingleSessionAppSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleSingleSessionAppSession struct {
	Contract     *SimpleSingleSessionApp // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// SimpleSingleSessionAppCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleSingleSessionAppCallerSession struct {
	Contract *SimpleSingleSessionAppCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// SimpleSingleSessionAppTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleSingleSessionAppTransactorSession struct {
	Contract     *SimpleSingleSessionAppTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// SimpleSingleSessionAppRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleSingleSessionAppRaw struct {
	Contract *SimpleSingleSessionApp // Generic contract binding to access the raw methods on
}

// SimpleSingleSessionAppCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppCallerRaw struct {
	Contract *SimpleSingleSessionAppCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleSingleSessionAppTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppTransactorRaw struct {
	Contract *SimpleSingleSessionAppTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleSingleSessionApp creates a new instance of SimpleSingleSessionApp, bound to a specific deployed contract.
func NewSimpleSingleSessionApp(address common.Address, backend bind.ContractBackend) (*SimpleSingleSessionApp, error) {
	contract, err := bindSimpleSingleSessionApp(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionApp{SimpleSingleSessionAppCaller: SimpleSingleSessionAppCaller{contract: contract}, SimpleSingleSessionAppTransactor: SimpleSingleSessionAppTransactor{contract: contract}, SimpleSingleSessionAppFilterer: SimpleSingleSessionAppFilterer{contract: contract}}, nil
}

// NewSimpleSingleSessionAppCaller creates a new read-only instance of SimpleSingleSessionApp, bound to a specific deployed contract.
func NewSimpleSingleSessionAppCaller(address common.Address, caller bind.ContractCaller) (*SimpleSingleSessionAppCaller, error) {
	contract, err := bindSimpleSingleSessionApp(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppCaller{contract: contract}, nil
}

// NewSimpleSingleSessionAppTransactor creates a new write-only instance of SimpleSingleSessionApp, bound to a specific deployed contract.
func NewSimpleSingleSessionAppTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleSingleSessionAppTransactor, error) {
	contract, err := bindSimpleSingleSessionApp(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppTransactor{contract: contract}, nil
}

// NewSimpleSingleSessionAppFilterer creates a new log filterer instance of SimpleSingleSessionApp, bound to a specific deployed contract.
func NewSimpleSingleSessionAppFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleSingleSessionAppFilterer, error) {
	contract, err := bindSimpleSingleSessionApp(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppFilterer{contract: contract}, nil
}

// bindSimpleSingleSessionApp binds a generic wrapper to an already deployed contract.
func bindSimpleSingleSessionApp(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSingleSessionAppABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSingleSessionApp *SimpleSingleSessionAppRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSingleSessionApp.Contract.SimpleSingleSessionAppCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSingleSessionApp *SimpleSingleSessionAppRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.SimpleSingleSessionAppTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSingleSessionApp *SimpleSingleSessionAppRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.SimpleSingleSessionAppTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSingleSessionApp.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.contract.Transact(opts, method, params...)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xbbc35280.
//
// Solidity: function getActionDeadline() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCaller) GetActionDeadline(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSingleSessionApp.contract.Call(opts, out, "getActionDeadline")
	return *ret0, err
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xbbc35280.
//
// Solidity: function getActionDeadline() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetActionDeadline() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetActionDeadline(&_SimpleSingleSessionApp.CallOpts)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xbbc35280.
//
// Solidity: function getActionDeadline() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetActionDeadline() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetActionDeadline(&_SimpleSingleSessionApp.CallOpts)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCaller) GetOutcome(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleSingleSessionApp.contract.Call(opts, out, "getOutcome", _query)
	return *ret0, err
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleSingleSessionApp.Contract.GetOutcome(&_SimpleSingleSessionApp.CallOpts, _query)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleSingleSessionApp.Contract.GetOutcome(&_SimpleSingleSessionApp.CallOpts, _query)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCaller) GetSeqNum(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSingleSessionApp.contract.Call(opts, out, "getSeqNum")
	return *ret0, err
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetSeqNum() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSeqNum(&_SimpleSingleSessionApp.CallOpts)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetSeqNum() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSeqNum(&_SimpleSingleSessionApp.CallOpts)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCaller) GetSettleFinalizedTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSingleSessionApp.contract.Call(opts, out, "getSettleFinalizedTime")
	return *ret0, err
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetSettleFinalizedTime() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSettleFinalizedTime(&_SimpleSingleSessionApp.CallOpts)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() view returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetSettleFinalizedTime() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSettleFinalizedTime(&_SimpleSingleSessionApp.CallOpts)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) view returns(bytes)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCaller) GetState(opts *bind.CallOpts, _key *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _SimpleSingleSessionApp.contract.Call(opts, out, "getState", _key)
	return *ret0, err
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) view returns(bytes)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetState(_key *big.Int) ([]byte, error) {
	return _SimpleSingleSessionApp.Contract.GetState(&_SimpleSingleSessionApp.CallOpts, _key)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) view returns(bytes)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetState(_key *big.Int) ([]byte, error) {
	return _SimpleSingleSessionApp.Contract.GetState(&_SimpleSingleSessionApp.CallOpts, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint8)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCaller) GetStatus(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _SimpleSingleSessionApp.contract.Call(opts, out, "getStatus")
	return *ret0, err
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint8)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetStatus() (uint8, error) {
	return _SimpleSingleSessionApp.Contract.GetStatus(&_SimpleSingleSessionApp.CallOpts)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint8)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetStatus() (uint8, error) {
	return _SimpleSingleSessionApp.Contract.GetStatus(&_SimpleSingleSessionApp.CallOpts)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCaller) IsFinalized(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleSingleSessionApp.contract.Call(opts, out, "isFinalized", _query)
	return *ret0, err
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) IsFinalized(_query []byte) (bool, error) {
	return _SimpleSingleSessionApp.Contract.IsFinalized(&_SimpleSingleSessionApp.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) IsFinalized(_query []byte) (bool, error) {
	return _SimpleSingleSessionApp.Contract.IsFinalized(&_SimpleSingleSessionApp.CallOpts, _query)
}

// ApplyAction is a paid mutator transaction binding the contract method 0x1f2b71e5.
//
// Solidity: function applyAction(bytes _action) returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactor) ApplyAction(opts *bind.TransactOpts, _action []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.contract.Transact(opts, "applyAction", _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0x1f2b71e5.
//
// Solidity: function applyAction(bytes _action) returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) ApplyAction(_action []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.ApplyAction(&_SimpleSingleSessionApp.TransactOpts, _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0x1f2b71e5.
//
// Solidity: function applyAction(bytes _action) returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactorSession) ApplyAction(_action []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.ApplyAction(&_SimpleSingleSessionApp.TransactOpts, _action)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xfa5e7ff5.
//
// Solidity: function finalizeOnActionTimeout() returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactor) FinalizeOnActionTimeout(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.contract.Transact(opts, "finalizeOnActionTimeout")
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xfa5e7ff5.
//
// Solidity: function finalizeOnActionTimeout() returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) FinalizeOnActionTimeout() (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.FinalizeOnActionTimeout(&_SimpleSingleSessionApp.TransactOpts)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xfa5e7ff5.
//
// Solidity: function finalizeOnActionTimeout() returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactorSession) FinalizeOnActionTimeout() (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.FinalizeOnActionTimeout(&_SimpleSingleSessionApp.TransactOpts)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactor) IntendSettle(opts *bind.TransactOpts, _stateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.contract.Transact(opts, "intendSettle", _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.IntendSettle(&_SimpleSingleSessionApp.TransactOpts, _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_SimpleSingleSessionApp *SimpleSingleSessionAppTransactorSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionApp.Contract.IntendSettle(&_SimpleSingleSessionApp.TransactOpts, _stateProof)
}

// SimpleSingleSessionAppIntendSettleIterator is returned from FilterIntendSettle and is used to iterate over the raw logs and unpacked data for IntendSettle events raised by the SimpleSingleSessionApp contract.
type SimpleSingleSessionAppIntendSettleIterator struct {
	Event *SimpleSingleSessionAppIntendSettle // Event containing the contract specifics and raw log

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
func (it *SimpleSingleSessionAppIntendSettleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSingleSessionAppIntendSettle)
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
		it.Event = new(SimpleSingleSessionAppIntendSettle)
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
func (it *SimpleSingleSessionAppIntendSettleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSingleSessionAppIntendSettleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSingleSessionAppIntendSettle represents a IntendSettle event raised by the SimpleSingleSessionApp contract.
type SimpleSingleSessionAppIntendSettle struct {
	Seq *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterIntendSettle is a free log retrieval operation binding the contract event 0xce68db27527c6058059e8cbd1c6de0528ef1c417fe1c21c545919c7da3466d2a.
//
// Solidity: event IntendSettle(uint256 seq)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppFilterer) FilterIntendSettle(opts *bind.FilterOpts) (*SimpleSingleSessionAppIntendSettleIterator, error) {

	logs, sub, err := _SimpleSingleSessionApp.contract.FilterLogs(opts, "IntendSettle")
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppIntendSettleIterator{contract: _SimpleSingleSessionApp.contract, event: "IntendSettle", logs: logs, sub: sub}, nil
}

// WatchIntendSettle is a free log subscription operation binding the contract event 0xce68db27527c6058059e8cbd1c6de0528ef1c417fe1c21c545919c7da3466d2a.
//
// Solidity: event IntendSettle(uint256 seq)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppFilterer) WatchIntendSettle(opts *bind.WatchOpts, sink chan<- *SimpleSingleSessionAppIntendSettle) (event.Subscription, error) {

	logs, sub, err := _SimpleSingleSessionApp.contract.WatchLogs(opts, "IntendSettle")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSingleSessionAppIntendSettle)
				if err := _SimpleSingleSessionApp.contract.UnpackLog(event, "IntendSettle", log); err != nil {
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

// ParseIntendSettle is a log parse operation binding the contract event 0xce68db27527c6058059e8cbd1c6de0528ef1c417fe1c21c545919c7da3466d2a.
//
// Solidity: event IntendSettle(uint256 seq)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppFilterer) ParseIntendSettle(log types.Log) (*SimpleSingleSessionAppIntendSettle, error) {
	event := new(SimpleSingleSessionAppIntendSettle)
	if err := _SimpleSingleSessionApp.contract.UnpackLog(event, "IntendSettle", log); err != nil {
		return nil, err
	}
	return event, nil
}
