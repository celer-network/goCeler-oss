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

// SimpleMultiSessionAppABI is the input ABI used to generate the binding from.
const SimpleMultiSessionAppABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getSettleFinalizedTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_stateProof\",\"type\":\"bytes\"}],\"name\":\"intendSettle\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getSeqNum\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nonce\",\"type\":\"uint256\"},{\"name\":\"_signers\",\"type\":\"address[]\"}],\"name\":\"getSessionID\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"finalizeOnActionTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"}],\"name\":\"getActionDeadline\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"},{\"name\":\"_action\",\"type\":\"bytes\"}],\"name\":\"applyAction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_playerNum\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"session\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"}],\"name\":\"IntendSettle\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"getOutcome\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_session\",\"type\":\"bytes32\"},{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SimpleMultiSessionAppBin is the compiled bytecode used for deploying new contracts.
var SimpleMultiSessionAppBin = "0x608060405234801561001057600080fd5b50604051602080611e9f8339810180604052602081101561003057600080fd5b81019080805190602001909291905050508080806000800181905550505050611e418061005e6000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c80635de28ae0116100715780635de28ae0146102f3578063b89fa28b14610343578063bcdbda9414610371578063cab9244614610402578063ea4ba8eb14610444578063f3c77192146104d5576100a9565b806309b65d86146100ae578063130d33fe146100f057806329dd2f8e146101695780633b6de66f1461021a5780634d8bedec1461025c575b600080fd5b6100da600480360360208110156100c457600080fd5b8101908080359060200190929190505050610558565b6040518082815260200191505060405180910390f35b6101676004803603602081101561010657600080fd5b810190808035906020019064010000000081111561012357600080fd5b82018360208201111561013557600080fd5b8035906020019184600183028401116401000000008311171561015757600080fd5b90919293919293905050506105c4565b005b61019f6004803603604081101561017f57600080fd5b8101908080359060200190929190803590602001909291905050506109cb565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156101df5780820151818401526020810190506101c4565b50505050905090810190601f16801561020c5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6102466004803603602081101561023057600080fd5b8101908080359060200190929190505050610a3e565b6040518082815260200191505060405180910390f35b6102dd6004803603604081101561027257600080fd5b81019080803590602001909291908035906020019064010000000081111561029957600080fd5b8201836020820111156102ab57600080fd5b803590602001918460208302840111640100000000831117156102cd57600080fd5b9091929391929390505050610a5e565b6040518082815260200191505060405180910390f35b61031f6004803603602081101561030957600080fd5b8101908080359060200190929190505050610ac7565b6040518082600381111561032f57fe5b60ff16815260200191505060405180910390f35b61036f6004803603602081101561035957600080fd5b8101908080359060200190929190505050610af4565b005b6103e86004803603602081101561038757600080fd5b81019080803590602001906401000000008111156103a457600080fd5b8201836020820111156103b657600080fd5b803590602001918460018302840111640100000000831117156103d857600080fd5b9091929391929390505050610ba9565b604051808215151515815260200191505060405180910390f35b61042e6004803603602081101561041857600080fd5b8101908080359060200190929190505050610c40565b6040518082815260200191505060405180910390f35b6104bb6004803603602081101561045a57600080fd5b810190808035906020019064010000000081111561047757600080fd5b82018360208201111561048957600080fd5b803590602001918460018302840111640100000000831117156104ab57600080fd5b9091929391929390505050610d24565b604051808215151515815260200191505060405180910390f35b610556600480360360408110156104eb57600080fd5b81019080803590602001909291908035906020019064010000000081111561051257600080fd5b82018360208201111561052457600080fd5b8035906020019184600183028401116401000000008311171561054657600080fd5b9091929391929390505050610d98565b005b60006001600381111561056757fe5b6001600084815260200190815260200160002060040160009054906101000a900460ff16600381111561059657fe5b14156105ba57600160008381526020019081526020016000206003015490506105bf565b600090505b919050565b6105cc611ccf565b61061983838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050610fcf565b90506060610631826000015183602001516001611146565b905060008001548151146106ad576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260198152602001807f696e76616c6964206e756d626572206f6620706c61796572730000000000000081525060200191505060405180910390fd5b6106b5611ce9565b6106c2836000015161135d565b905060008160000151836040516020018083815260200180602001828103825283818151815260200191508051906020019060200280838360005b838110156107185780820151818401526020810190506106fd565b50505050905001935050505060405160208183030381529060405280519060200120905060006001600083815260200190815260200160002090506000600381111561076057fe5b8160040160009054906101000a900460ff16600381111561077d57fe5b141561079d578381600001908051906020019061079b929190611d11565b505b6003808111156107a957fe5b8160040160009054906101000a900460ff1660038111156107c657fe5b141561083a576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260128152602001807f73746174652069732066696e616c697a6564000000000000000000000000000081525060200191505060405180910390fd5b82602001518160010154106108b7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260178152602001807f696e76616c69642073657175656e6365206e756d62657200000000000000000081525060200191505060405180910390fd5b8260200151816001018190555060018160040160006101000a81548160ff021916908360038111156108e557fe5b02179055508260600151816002018190555082606001514301816003018190555061091482846040015161144b565b610986576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f737461746520757064617465206661696c65640000000000000000000000000081525060200191505060405180910390fd5b817f82c4eeba939ff9358877334330e22a5cdb0472113cd14f90625ea634b60d2e5b82600101546040518082815260200191505060405180910390a250505050505050565b60608060206040519080825280601f01601f191660200182016040528015610a025781602001600182028038833980820191505090505b50905060006002600086815260200190815260200160002060000160009054906101000a900460ff169050806020830152819250505092915050565b600060016000838152602001908152602001600020600101549050919050565b600083838360405160200180848152602001806020018281038252848482818152602001925060200280828437600081840152601f19601f8201169050808301925050509450505050506040516020818303038152906040528051906020012090509392505050565b60006001600083815260200190815260200160002060040160009054906101000a900460ff169050919050565b600060016000838152602001908152602001600020905060008160040160009054906101000a900460ff16905060008260030154905060026003811115610b3757fe5b826003811115610b4357fe5b1415610b5a57804311610b5557600080fd5b610b99565b60016003811115610b6757fe5b826003811115610b7357fe5b1415610b9057826002015481014311610b8b57600080fd5b610b98565b505050610ba6565b5b610ba284611592565b5050505b50565b600080610bf984848080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050506115cd565b9050600380811115610c0757fe5b6001600083815260200190815260200160002060040160009054906101000a900460ff166003811115610c3657fe5b1491505092915050565b600060026003811115610c4f57fe5b6001600084815260200190815260200160002060040160009054906101000a900460ff166003811115610c7e57fe5b1415610ca25760016000838152602001908152602001600020600301549050610d1f565b60016003811115610caf57fe5b6001600084815260200190815260200160002060040160009054906101000a900460ff166003811115610cde57fe5b1415610d1a5760016000838152602001908152602001600020600201546001600084815260200190815260200160002060030154019050610d1f565b600090505b919050565b6000610d2e611d9b565b610d7b84848080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050506115e9565b9050610d8f8160000151826020015161169d565b91505092915050565b6000600160008581526020019081526020016000209050600380811115610dbb57fe5b8160040160009054906101000a900460ff166003811115610dd857fe5b1415610e4c576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b60016003811115610e5957fe5b8160040160009054906101000a900460ff166003811115610e7657fe5b148015610e865750438160030154105b15610eb25760028160040160006101000a81548160ff02191690836003811115610eac57fe5b02179055505b60026003811115610ebf57fe5b8160040160009054906101000a900460ff166003811115610edc57fe5b14610f4f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f617070206e6f7420696e20616374696f6e206d6f64650000000000000000000081525060200191505060405180910390fd5b8060010160008154809291906001019190505550806002015443018160030181905550610fc08484848080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050611764565b610fc957600080fd5b50505050565b610fd7611ccf565b610fdf611db8565b610fe8836118ab565b905060606110006002836118da90919063ffffffff16565b90508060028151811061100f57fe5b602002602001015160405190808252806020026020018201604052801561104a57816020015b60608152602001906001900390816110355790505b50836020018190525060008160028151811061106257fe5b6020026020010181815250506000805b61107b8461197f565b1561113d5761108984611994565b809250819350505060001561109d57611138565b60018214156110bc576110af846119c8565b8560000181905250611137565b6002821415611122576110ce846119c8565b8560200151846002815181106110e057fe5b6020026020010151815181106110f257fe5b60200260200101819052508260028151811061110a57fe5b60200260200101805180919060010181525050611136565b6111358185611a8190919063ffffffff16565b5b5b5b611072565b50505050919050565b60608083516040519080825280602002602001820160405280156111795781602001602082028038833980820191505090505b50905060006111f8866040516020018082805190602001908083835b602083106111b85780518252602082019150602081019050602083039250611195565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405160208183030381529060405280519060200120611b11565b9050600080905060008090505b865181101561134f5761122b8388838151811061121e57fe5b6020026020010151611b69565b84828151811061123757fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508515611342578173ffffffffffffffffffffffffffffffffffffffff1684828151811061129a57fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff161161132b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f7369676e657273206e6f7420696e20617363656e64696e67206f72646572000081525060200191505060405180910390fd5b83818151811061133757fe5b602002602001015191505b8080600101915050611205565b508293505050509392505050565b611365611ce9565b61136d611db8565b611376836118ab565b90506000805b6113858361197f565b156114435761139383611994565b80925081935050506000156113a75761143e565b60018214156113c7576113b983611c49565b84600001818152505061143d565b60028214156113e7576113d983611c49565b84602001818152505061143c565b6003821415611406576113f9836119c8565b846040018190525061143b565b60048214156114265761141883611c49565b84606001818152505061143a565b6114398184611a8190919063ffffffff16565b5b5b5b5b5b61137c565b505050919050565b600060018251146114c4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f696e76616c6964207374617465206c656e67746800000000000000000000000081525060200191505060405180910390fd5b6000600260008581526020019081526020016000209050826000815181106114e857fe5b602001015160f81c60f81b60f81c8160000160006101000a81548160ff021916908360ff16021790555060018160000160009054906101000a900460ff1660ff161480611549575060028160000160009054906101000a900460ff1660ff16145b156115875760036001600086815260200190815260200160002060040160006101000a81548160ff0219169083600381111561158157fe5b02179055505b600191505092915050565b60036001600083815260200190815260200160002060040160006101000a81548160ff021916908360038111156115c557fe5b021790555050565b600060208251146115dd57600080fd5b60208201519050919050565b6115f1611d9b565b6115f9611db8565b611602836118ab565b90506000805b6116118361197f565b156116955761161f83611994565b809250819350505060001561163357611690565b600182141561165b5761164d611648846119c8565b6115cd565b84600001818152505061168f565b600282141561167a5761166d836119c8565b846020018190525061168e565b61168d8184611a8190919063ffffffff16565b5b5b5b611608565b505050919050565b60006001825114611716576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f696e76616c6964207175657279206c656e67746800000000000000000000000081525060200191505060405180910390fd5b8160008151811061172357fe5b602001015160f81c60f81b60f81c60ff166002600085815260200190815260200160002060000160009054906101000a900460ff1660ff1614905092915050565b600060018251146117dd576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f696e76616c696420616374696f6e206c656e677468000000000000000000000081525060200191505060405180910390fd5b60006002600085815260200190815260200160002090508260008151811061180157fe5b602001015160f81c60f81b60f81c8160000160006101000a81548160ff021916908360ff16021790555060018160000160009054906101000a900460ff1660ff161480611862575060028160000160009054906101000a900460ff1660ff16145b156118a05760036001600086815260200190815260200160002060040160006101000a81548160ff0219169083600381111561189a57fe5b02179055505b600191505092915050565b6118b3611db8565b60018251116118c157600080fd5b8181602001819052506000816000018181525050919050565b6060600083600001519050600183016040519080825280602002602001820160405280156119175781602001602082028038833980820191505090505b5091506000805b6119278661197f565b1561196c5761193586611994565b8092508193505050600184838151811061194b57fe5b6020026020010181815101915081815250506119678682611a81565b61191e565b8286600001818152505050505092915050565b60008160200151518260000151109050919050565b60008060006119a284611c49565b9050600881816119ae57fe5b0492506007811660058111156119c057fe5b915050915091565b606060006119d583611c49565b905060008184600001510190508360200151518111156119f457600080fd5b816040519080825280601f01601f191660200182016040528015611a275781602001600182028038833980820191505090505b50925060608460200151905060008086600001519050602086019150806020840101905060008090505b85811015611a6c578082015181840152602081019050611a51565b50838760000181815250505050505050919050565b60006005811115611a8e57fe5b816005811115611a9a57fe5b1415611aaf57611aa982611c49565b50611b0d565b60026005811115611abc57fe5b816005811115611ac857fe5b1415611b07576000611ad983611c49565b90508083600001818151019150818152505082602001515183600001511115611b0157600080fd5b50611b0c565b600080fd5b5b5050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b60006041825114611b7d5760009050611c43565b60008060006020850151925060408501519150606085015160001a9050601b8160ff161015611bad57601b810190505b601b8160ff1614158015611bc55750601c8160ff1614155b15611bd65760009350505050611c43565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015611c33573d6000803e3d6000fd5b5050506020604051035193505050505b92915050565b60008060608360200151905083600001519250826020820101519150600080935060008090505b600a811015611cc45783811a915060078102607f8316901b851794506000608083161415611cb7576001810186600001818151019150818152505084945050505050611cca565b8080600101915050611c70565b50600080fd5b919050565b604051806040016040528060608152602001606081525090565b6040518060800160405280600081526020016000815260200160608152602001600081525090565b828054828255906000526020600020908101928215611d8a579160200282015b82811115611d895782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555091602001919060010190611d31565b5b509050611d979190611dd2565b5090565b604051806040016040528060008019168152602001606081525090565b604051806040016040528060008152602001606081525090565b611e1291905b80821115611e0e57600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff021916905550600101611dd8565b5090565b9056fea165627a7a723058205e5fba456536ddf9f1ce69a4488227c0f3b08c317467eb5d659a10f907e41eb40029"

// DeploySimpleMultiSessionApp deploys a new Ethereum contract, binding an instance of SimpleMultiSessionApp to it.
func DeploySimpleMultiSessionApp(auth *bind.TransactOpts, backend bind.ContractBackend, _playerNum *big.Int) (common.Address, *types.Transaction, *SimpleMultiSessionApp, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleMultiSessionAppABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleMultiSessionAppBin), backend, _playerNum)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleMultiSessionApp{SimpleMultiSessionAppCaller: SimpleMultiSessionAppCaller{contract: contract}, SimpleMultiSessionAppTransactor: SimpleMultiSessionAppTransactor{contract: contract}, SimpleMultiSessionAppFilterer: SimpleMultiSessionAppFilterer{contract: contract}}, nil
}

// SimpleMultiSessionApp is an auto generated Go binding around an Ethereum contract.
type SimpleMultiSessionApp struct {
	SimpleMultiSessionAppCaller     // Read-only binding to the contract
	SimpleMultiSessionAppTransactor // Write-only binding to the contract
	SimpleMultiSessionAppFilterer   // Log filterer for contract events
}

// SimpleMultiSessionAppCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleMultiSessionAppCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleMultiSessionAppTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleMultiSessionAppTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleMultiSessionAppFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleMultiSessionAppFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleMultiSessionAppSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleMultiSessionAppSession struct {
	Contract     *SimpleMultiSessionApp // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// SimpleMultiSessionAppCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleMultiSessionAppCallerSession struct {
	Contract *SimpleMultiSessionAppCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// SimpleMultiSessionAppTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleMultiSessionAppTransactorSession struct {
	Contract     *SimpleMultiSessionAppTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// SimpleMultiSessionAppRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleMultiSessionAppRaw struct {
	Contract *SimpleMultiSessionApp // Generic contract binding to access the raw methods on
}

// SimpleMultiSessionAppCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleMultiSessionAppCallerRaw struct {
	Contract *SimpleMultiSessionAppCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleMultiSessionAppTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleMultiSessionAppTransactorRaw struct {
	Contract *SimpleMultiSessionAppTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleMultiSessionApp creates a new instance of SimpleMultiSessionApp, bound to a specific deployed contract.
func NewSimpleMultiSessionApp(address common.Address, backend bind.ContractBackend) (*SimpleMultiSessionApp, error) {
	contract, err := bindSimpleMultiSessionApp(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleMultiSessionApp{SimpleMultiSessionAppCaller: SimpleMultiSessionAppCaller{contract: contract}, SimpleMultiSessionAppTransactor: SimpleMultiSessionAppTransactor{contract: contract}, SimpleMultiSessionAppFilterer: SimpleMultiSessionAppFilterer{contract: contract}}, nil
}

// NewSimpleMultiSessionAppCaller creates a new read-only instance of SimpleMultiSessionApp, bound to a specific deployed contract.
func NewSimpleMultiSessionAppCaller(address common.Address, caller bind.ContractCaller) (*SimpleMultiSessionAppCaller, error) {
	contract, err := bindSimpleMultiSessionApp(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleMultiSessionAppCaller{contract: contract}, nil
}

// NewSimpleMultiSessionAppTransactor creates a new write-only instance of SimpleMultiSessionApp, bound to a specific deployed contract.
func NewSimpleMultiSessionAppTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleMultiSessionAppTransactor, error) {
	contract, err := bindSimpleMultiSessionApp(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleMultiSessionAppTransactor{contract: contract}, nil
}

// NewSimpleMultiSessionAppFilterer creates a new log filterer instance of SimpleMultiSessionApp, bound to a specific deployed contract.
func NewSimpleMultiSessionAppFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleMultiSessionAppFilterer, error) {
	contract, err := bindSimpleMultiSessionApp(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleMultiSessionAppFilterer{contract: contract}, nil
}

// bindSimpleMultiSessionApp binds a generic wrapper to an already deployed contract.
func bindSimpleMultiSessionApp(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleMultiSessionAppABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleMultiSessionApp *SimpleMultiSessionAppRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleMultiSessionApp.Contract.SimpleMultiSessionAppCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleMultiSessionApp *SimpleMultiSessionAppRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.SimpleMultiSessionAppTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleMultiSessionApp *SimpleMultiSessionAppRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.SimpleMultiSessionAppTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleMultiSessionApp.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.contract.Transact(opts, method, params...)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xcab92446.
//
// Solidity: function getActionDeadline(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) GetActionDeadline(opts *bind.CallOpts, _session [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "getActionDeadline", _session)
	return *ret0, err
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xcab92446.
//
// Solidity: function getActionDeadline(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) GetActionDeadline(_session [32]byte) (*big.Int, error) {
	return _SimpleMultiSessionApp.Contract.GetActionDeadline(&_SimpleMultiSessionApp.CallOpts, _session)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xcab92446.
//
// Solidity: function getActionDeadline(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) GetActionDeadline(_session [32]byte) (*big.Int, error) {
	return _SimpleMultiSessionApp.Contract.GetActionDeadline(&_SimpleMultiSessionApp.CallOpts, _session)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) GetOutcome(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "getOutcome", _query)
	return *ret0, err
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleMultiSessionApp.Contract.GetOutcome(&_SimpleMultiSessionApp.CallOpts, _query)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleMultiSessionApp.Contract.GetOutcome(&_SimpleMultiSessionApp.CallOpts, _query)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x3b6de66f.
//
// Solidity: function getSeqNum(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) GetSeqNum(opts *bind.CallOpts, _session [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "getSeqNum", _session)
	return *ret0, err
}

// GetSeqNum is a free data retrieval call binding the contract method 0x3b6de66f.
//
// Solidity: function getSeqNum(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) GetSeqNum(_session [32]byte) (*big.Int, error) {
	return _SimpleMultiSessionApp.Contract.GetSeqNum(&_SimpleMultiSessionApp.CallOpts, _session)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x3b6de66f.
//
// Solidity: function getSeqNum(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) GetSeqNum(_session [32]byte) (*big.Int, error) {
	return _SimpleMultiSessionApp.Contract.GetSeqNum(&_SimpleMultiSessionApp.CallOpts, _session)
}

// GetSessionID is a free data retrieval call binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) pure returns(bytes32)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) GetSessionID(opts *bind.CallOpts, _nonce *big.Int, _signers []common.Address) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "getSessionID", _nonce, _signers)
	return *ret0, err
}

// GetSessionID is a free data retrieval call binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) pure returns(bytes32)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) GetSessionID(_nonce *big.Int, _signers []common.Address) ([32]byte, error) {
	return _SimpleMultiSessionApp.Contract.GetSessionID(&_SimpleMultiSessionApp.CallOpts, _nonce, _signers)
}

// GetSessionID is a free data retrieval call binding the contract method 0x4d8bedec.
//
// Solidity: function getSessionID(uint256 _nonce, address[] _signers) pure returns(bytes32)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) GetSessionID(_nonce *big.Int, _signers []common.Address) ([32]byte, error) {
	return _SimpleMultiSessionApp.Contract.GetSessionID(&_SimpleMultiSessionApp.CallOpts, _nonce, _signers)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0x09b65d86.
//
// Solidity: function getSettleFinalizedTime(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) GetSettleFinalizedTime(opts *bind.CallOpts, _session [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "getSettleFinalizedTime", _session)
	return *ret0, err
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0x09b65d86.
//
// Solidity: function getSettleFinalizedTime(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) GetSettleFinalizedTime(_session [32]byte) (*big.Int, error) {
	return _SimpleMultiSessionApp.Contract.GetSettleFinalizedTime(&_SimpleMultiSessionApp.CallOpts, _session)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0x09b65d86.
//
// Solidity: function getSettleFinalizedTime(bytes32 _session) view returns(uint256)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) GetSettleFinalizedTime(_session [32]byte) (*big.Int, error) {
	return _SimpleMultiSessionApp.Contract.GetSettleFinalizedTime(&_SimpleMultiSessionApp.CallOpts, _session)
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) view returns(bytes)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) GetState(opts *bind.CallOpts, _session [32]byte, _key *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "getState", _session, _key)
	return *ret0, err
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) view returns(bytes)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) GetState(_session [32]byte, _key *big.Int) ([]byte, error) {
	return _SimpleMultiSessionApp.Contract.GetState(&_SimpleMultiSessionApp.CallOpts, _session, _key)
}

// GetState is a free data retrieval call binding the contract method 0x29dd2f8e.
//
// Solidity: function getState(bytes32 _session, uint256 _key) view returns(bytes)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) GetState(_session [32]byte, _key *big.Int) ([]byte, error) {
	return _SimpleMultiSessionApp.Contract.GetState(&_SimpleMultiSessionApp.CallOpts, _session, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) view returns(uint8)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) GetStatus(opts *bind.CallOpts, _session [32]byte) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "getStatus", _session)
	return *ret0, err
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) view returns(uint8)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) GetStatus(_session [32]byte) (uint8, error) {
	return _SimpleMultiSessionApp.Contract.GetStatus(&_SimpleMultiSessionApp.CallOpts, _session)
}

// GetStatus is a free data retrieval call binding the contract method 0x5de28ae0.
//
// Solidity: function getStatus(bytes32 _session) view returns(uint8)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) GetStatus(_session [32]byte) (uint8, error) {
	return _SimpleMultiSessionApp.Contract.GetStatus(&_SimpleMultiSessionApp.CallOpts, _session)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCaller) IsFinalized(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleMultiSessionApp.contract.Call(opts, out, "isFinalized", _query)
	return *ret0, err
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) IsFinalized(_query []byte) (bool, error) {
	return _SimpleMultiSessionApp.Contract.IsFinalized(&_SimpleMultiSessionApp.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppCallerSession) IsFinalized(_query []byte) (bool, error) {
	return _SimpleMultiSessionApp.Contract.IsFinalized(&_SimpleMultiSessionApp.CallOpts, _query)
}

// ApplyAction is a paid mutator transaction binding the contract method 0xf3c77192.
//
// Solidity: function applyAction(bytes32 _session, bytes _action) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactor) ApplyAction(opts *bind.TransactOpts, _session [32]byte, _action []byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.contract.Transact(opts, "applyAction", _session, _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0xf3c77192.
//
// Solidity: function applyAction(bytes32 _session, bytes _action) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) ApplyAction(_session [32]byte, _action []byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.ApplyAction(&_SimpleMultiSessionApp.TransactOpts, _session, _action)
}

// ApplyAction is a paid mutator transaction binding the contract method 0xf3c77192.
//
// Solidity: function applyAction(bytes32 _session, bytes _action) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactorSession) ApplyAction(_session [32]byte, _action []byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.ApplyAction(&_SimpleMultiSessionApp.TransactOpts, _session, _action)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xb89fa28b.
//
// Solidity: function finalizeOnActionTimeout(bytes32 _session) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactor) FinalizeOnActionTimeout(opts *bind.TransactOpts, _session [32]byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.contract.Transact(opts, "finalizeOnActionTimeout", _session)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xb89fa28b.
//
// Solidity: function finalizeOnActionTimeout(bytes32 _session) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) FinalizeOnActionTimeout(_session [32]byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.FinalizeOnActionTimeout(&_SimpleMultiSessionApp.TransactOpts, _session)
}

// FinalizeOnActionTimeout is a paid mutator transaction binding the contract method 0xb89fa28b.
//
// Solidity: function finalizeOnActionTimeout(bytes32 _session) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactorSession) FinalizeOnActionTimeout(_session [32]byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.FinalizeOnActionTimeout(&_SimpleMultiSessionApp.TransactOpts, _session)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactor) IntendSettle(opts *bind.TransactOpts, _stateProof []byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.contract.Transact(opts, "intendSettle", _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.IntendSettle(&_SimpleMultiSessionApp.TransactOpts, _stateProof)
}

// IntendSettle is a paid mutator transaction binding the contract method 0x130d33fe.
//
// Solidity: function intendSettle(bytes _stateProof) returns()
func (_SimpleMultiSessionApp *SimpleMultiSessionAppTransactorSession) IntendSettle(_stateProof []byte) (*types.Transaction, error) {
	return _SimpleMultiSessionApp.Contract.IntendSettle(&_SimpleMultiSessionApp.TransactOpts, _stateProof)
}

// SimpleMultiSessionAppIntendSettleIterator is returned from FilterIntendSettle and is used to iterate over the raw logs and unpacked data for IntendSettle events raised by the SimpleMultiSessionApp contract.
type SimpleMultiSessionAppIntendSettleIterator struct {
	Event *SimpleMultiSessionAppIntendSettle // Event containing the contract specifics and raw log

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
func (it *SimpleMultiSessionAppIntendSettleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleMultiSessionAppIntendSettle)
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
		it.Event = new(SimpleMultiSessionAppIntendSettle)
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
func (it *SimpleMultiSessionAppIntendSettleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleMultiSessionAppIntendSettleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleMultiSessionAppIntendSettle represents a IntendSettle event raised by the SimpleMultiSessionApp contract.
type SimpleMultiSessionAppIntendSettle struct {
	Session [32]byte
	Seq     *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterIntendSettle is a free log retrieval operation binding the contract event 0x82c4eeba939ff9358877334330e22a5cdb0472113cd14f90625ea634b60d2e5b.
//
// Solidity: event IntendSettle(bytes32 indexed session, uint256 seq)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppFilterer) FilterIntendSettle(opts *bind.FilterOpts, session [][32]byte) (*SimpleMultiSessionAppIntendSettleIterator, error) {

	var sessionRule []interface{}
	for _, sessionItem := range session {
		sessionRule = append(sessionRule, sessionItem)
	}

	logs, sub, err := _SimpleMultiSessionApp.contract.FilterLogs(opts, "IntendSettle", sessionRule)
	if err != nil {
		return nil, err
	}
	return &SimpleMultiSessionAppIntendSettleIterator{contract: _SimpleMultiSessionApp.contract, event: "IntendSettle", logs: logs, sub: sub}, nil
}

// WatchIntendSettle is a free log subscription operation binding the contract event 0x82c4eeba939ff9358877334330e22a5cdb0472113cd14f90625ea634b60d2e5b.
//
// Solidity: event IntendSettle(bytes32 indexed session, uint256 seq)
func (_SimpleMultiSessionApp *SimpleMultiSessionAppFilterer) WatchIntendSettle(opts *bind.WatchOpts, sink chan<- *SimpleMultiSessionAppIntendSettle, session [][32]byte) (event.Subscription, error) {

	var sessionRule []interface{}
	for _, sessionItem := range session {
		sessionRule = append(sessionRule, sessionItem)
	}

	logs, sub, err := _SimpleMultiSessionApp.contract.WatchLogs(opts, "IntendSettle", sessionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleMultiSessionAppIntendSettle)
				if err := _SimpleMultiSessionApp.contract.UnpackLog(event, "IntendSettle", log); err != nil {
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
func (_SimpleMultiSessionApp *SimpleMultiSessionAppFilterer) ParseIntendSettle(log types.Log) (*SimpleMultiSessionAppIntendSettle, error) {
	event := new(SimpleMultiSessionAppIntendSettle)
	if err := _SimpleMultiSessionApp.contract.UnpackLog(event, "IntendSettle", log); err != nil {
		return nil, err
	}
	return event, nil
}
