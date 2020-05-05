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
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// SimpleSingleSessionAppABI is the input ABI used to generate the binding from.
const SimpleSingleSessionAppABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_stateProof\",\"type\":\"bytes\"}],\"name\":\"intendSettle\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_action\",\"type\":\"bytes\"}],\"name\":\"applyAction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getSeqNum\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getSettleFinalizedTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getActionDeadline\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeOnActionTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_players\",\"type\":\"address[]\"},{\"name\":\"_nonce\",\"type\":\"uint256\"},{\"name\":\"_timeout\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"}],\"name\":\"IntendSettle\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"getOutcome\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SimpleSingleSessionAppBin is the compiled bytecode used for deploying new contracts.
const SimpleSingleSessionAppBin = `0x60806040523480156200001157600080fd5b5060405162001d9138038062001d91833981018060405260608110156200003757600080fd5b8101908080516401000000008111156200005057600080fd5b828101905060208101848111156200006757600080fd5b81518560208202830111640100000000821117156200008557600080fd5b505092919060200180519060200190929190805190602001909291905050508282828282828160008001819055508260006001019080519060200190620000ce92919062000119565b50806000600301819055506000806002018190555060008060050160006101000a81548160ff021916908360038111156200010557fe5b0217905550505050505050505050620001ee565b82805482825590600052602060002090810192821562000195579160200282015b82811115620001945782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550916020019190600101906200013a565b5b509050620001a49190620001a8565b5090565b620001eb91905b80821115620001e757600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff021916905550600101620001af565b5090565b90565b611b9380620001fe6000396000f3fe6080604052600436106100a4576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063130d33fe146100a95780631f2b71e51461012f57806344c9af28146101b55780634e69d560146102695780636d15c457146102a2578063b71ca01f146102cd578063bbc35280146102f8578063bcdbda9414610323578063ea4ba8eb146103c1578063fa5e7ff51461045f575b600080fd5b3480156100b557600080fd5b5061012d600480360360208110156100cc57600080fd5b81019080803590602001906401000000008111156100e957600080fd5b8201836020820111156100fb57600080fd5b8035906020019184600183028401116401000000008311171561011d57600080fd5b9091929391929390505050610476565b005b34801561013b57600080fd5b506101b36004803603602081101561015257600080fd5b810190808035906020019064010000000081111561016f57600080fd5b82018360208201111561018157600080fd5b803590602001918460018302840111640100000000831117156101a357600080fd5b9091929391929390505050610812565b005b3480156101c157600080fd5b506101ee600480360360208110156101d857600080fd5b8101908080359060200190929190505050610a3d565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561022e578082015181840152602081019050610213565b50505050905090810190601f16801561025b5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561027557600080fd5b5061027e610a9b565b6040518082600381111561028e57fe5b60ff16815260200191505060405180910390f35b3480156102ae57600080fd5b506102b7610ab4565b6040518082815260200191505060405180910390f35b3480156102d957600080fd5b506102e2610ac0565b6040518082815260200191505060405180910390f35b34801561030457600080fd5b5061030d610b08565b6040518082815260200191505060405180910390f35b34801561032f57600080fd5b506103a76004803603602081101561034657600080fd5b810190808035906020019064010000000081111561036357600080fd5b82018360208201111561037557600080fd5b8035906020019184600183028401116401000000008311171561039757600080fd5b9091929391929390505050610b95565b604051808215151515815260200191505060405180910390f35b3480156103cd57600080fd5b50610445600480360360208110156103e457600080fd5b810190808035906020019064010000000081111561040157600080fd5b82018360208201111561041357600080fd5b8035906020019184600183028401116401000000008311171561043557600080fd5b9091929391929390505050610bdb565b604051808215151515815260200191505060405180910390f35b34801561046b57600080fd5b50610474610d14565b005b60038081111561048257fe5b600060050160009054906101000a900460ff1660038111156104a057fe5b14151515610516576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b61051e611b0a565b61056b83838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050610dbe565b905061057f81600001518260200151610f49565b15156105f3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260118152602001807f696e76616c6964207369676e617475726500000000000000000000000000000081525060200191505060405180910390fd5b6105fb611b24565b610608826000015161109d565b905060008001548160000151141515610689576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600f8152602001807f6e6f6e6365206e6f74206d61746368000000000000000000000000000000000081525060200191505060405180910390fd5b8060200151600060020154101515610709576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260178152602001807f696e76616c69642073657175656e6365206e756d62657200000000000000000081525060200191505060405180910390fd5b80602001516000600201819055506001600060050160006101000a81548160ff0219169083600381111561073957fe5b0217905550600060030154430160006004018190555061075c816040015161118b565b15156107d0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f737461746520757064617465206661696c65640000000000000000000000000081525060200191505060405180910390fd5b7fce68db27527c6058059e8cbd1c6de0528ef1c417fe1c21c545919c7da3466d2a6000600201546040518082815260200191505060405180910390a150505050565b60038081111561081e57fe5b600060050160009054906101000a900460ff16600381111561083c57fe5b141515156108b2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b600160038111156108bf57fe5b600060050160009054906101000a900460ff1660038111156108dd57fe5b1480156108ee575060006004015443115b1561091b576002600060050160006101000a81548160ff0219169083600381111561091557fe5b02179055505b6002600381111561092857fe5b600060050160009054906101000a900460ff16600381111561094657fe5b1415156109bb576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f617070206e6f7420696e20616374696f6e206d6f64650000000000000000000081525060200191505060405180910390fd5b6000600201600081548092919060010191905055506000600301544301600060040181905550610a2e82828080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050611306565b1515610a3957600080fd5b5050565b60608060206040519080825280601f01601f191660200182016040528015610a745781602001600182028038833980820191505090505b5090506000600660009054906101000a900460ff1690508060208301528192505050919050565b60008060050160009054906101000a900460ff16905090565b60008060020154905090565b600060016003811115610acf57fe5b600060050160009054906101000a900460ff166003811115610aed57fe5b1415610b00576000600401549050610b05565b600090505b90565b600060026003811115610b1757fe5b600060050160009054906101000a900460ff166003811115610b3557fe5b1415610b48576000600401549050610b92565b60016003811115610b5557fe5b600060050160009054906101000a900460ff166003811115610b7357fe5b1415610b8d57600060030154600060040154019050610b92565b600090505b90565b60008083839050141515610ba857600080fd5b600380811115610bb457fe5b600060050160009054906101000a900460ff166003811115610bd257fe5b14905092915050565b6000600183839050141515610c58576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f696e76616c6964207175657279206c656e67746800000000000000000000000081525060200191505060405180910390fd5b828260008181101515610c6757fe5b905001357f010000000000000000000000000000000000000000000000000000000000000090047f0100000000000000000000000000000000000000000000000000000000000000027effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19167f0100000000000000000000000000000000000000000000000000000000000000900460ff16600660009054906101000a900460ff1660ff1614905092915050565b60026003811115610d2157fe5b600060050160009054906101000a900460ff166003811115610d3f57fe5b1415610d5d5760006004015443111515610d5857600080fd5b610db3565b60016003811115610d6a57fe5b600060050160009054906101000a900460ff166003811115610d8857fe5b1415610dad576000600301546000600401540143111515610da857600080fd5b610db2565b610dbc565b5b610dbb611481565b5b565b610dc6611b0a565b610dce611b4d565b610dd7836114aa565b90506060610def6002836114db90919063ffffffff16565b9050806002815181101515610e0057fe5b90602001906020020151604051908082528060200260200182016040528015610e3d57816020015b6060815260200190600190039081610e285790505b5083602001819052506000816002815181101515610e5757fe5b90602001906020020181815250506000805b610e7284611584565b15610f4057610e8084611599565b8092508193505050600015610e9457610f3b565b6001821415610eb357610ea6846115cf565b8560000181905250610f3a565b6002821415610f2557610ec5846115cf565b8560200151846002815181101515610ed957fe5b90602001906020020151815181101515610eef57fe5b90602001906020020181905250826002815181101515610f0b57fe5b906020019060200201805180919060010181525050610f39565b610f38818561168a90919063ffffffff16565b5b5b5b610e69565b50505050919050565b600080600101805490508251141515610fca576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601c8152602001807f696e76616c6964206e756d626572206f66207369676e6174757265730000000081525060200191505060405180910390fd5b6060610fd88484600061171c565b905060008090505b600060010180549050811015611090578181815181101515610ffe57fe5b9060200190602002015173ffffffffffffffffffffffffffffffffffffffff1660006001018281548110151561103057fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614151561108357600092505050611097565b8080600101915050610fe0565b5060019150505b92915050565b6110a5611b24565b6110ad611b4d565b6110b6836114aa565b90506000805b6110c583611584565b15611183576110d383611599565b80925081935050506000156110e75761117e565b6001821415611107576110f983611947565b84600001818152505061117d565b60028214156111275761111983611947565b84602001818152505061117c565b600382141561114657611139836115cf565b846040018190525061117b565b60048214156111665761115883611947565b84606001818152505061117a565b611179818461168a90919063ffffffff16565b5b5b5b5b5b6110bc565b505050919050565b600060018251141515611206576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f696e76616c6964207374617465206c656e67746800000000000000000000000081525060200191505060405180910390fd5b81600081518110151561121557fe5b9060200101517f010000000000000000000000000000000000000000000000000000000000000090047f0100000000000000000000000000000000000000000000000000000000000000027f01000000000000000000000000000000000000000000000000000000000000009004600660006101000a81548160ff021916908360ff1602179055506001600660009054906101000a900460ff1660ff1614806112d057506002600660009054906101000a900460ff1660ff16145b156112fd576003600060050160006101000a81548160ff021916908360038111156112f757fe5b02179055505b60019050919050565b600060018251141515611381576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f696e76616c696420616374696f6e206c656e677468000000000000000000000081525060200191505060405180910390fd5b81600081518110151561139057fe5b9060200101517f010000000000000000000000000000000000000000000000000000000000000090047f0100000000000000000000000000000000000000000000000000000000000000027f01000000000000000000000000000000000000000000000000000000000000009004600660006101000a81548160ff021916908360ff1602179055506001600660009054906101000a900460ff1660ff16148061144b57506002600660009054906101000a900460ff1660ff16145b15611478576003600060050160006101000a81548160ff0219169083600381111561147257fe5b02179055505b60019050919050565b6003600060050160006101000a81548160ff021916908360038111156114a357fe5b0217905550565b6114b2611b4d565b600182511115156114c257600080fd5b8181602001819052506000816000018181525050919050565b6060600083600001519050600183016040519080825280602002602001820160405280156115185781602001602082028038833980820191505090505b5091506000805b61152886611584565b156115715761153686611599565b80925081935050506001848381518110151561154e57fe5b906020019060200201818151019150818152505061156c868261168a565b61151f565b8286600001818152505050505092915050565b60008160200151518260000151109050919050565b60008060006115a784611947565b90506008818115156115b557fe5b0492506007811660058111156115c757fe5b915050915091565b606060006115dc83611947565b9050600081846000015101905083602001515181111515156115fd57600080fd5b816040519080825280601f01601f1916602001820160405280156116305781602001600182028038833980820191505090505b50925060608460200151905060008086600001519050602086019150806020840101905060008090505b8581101561167557808201518184015260208101905061165a565b50838760000181815250505050505050919050565b6000600581111561169757fe5b8160058111156116a357fe5b14156116b8576116b282611947565b50611718565b600260058111156116c557fe5b8160058111156116d157fe5b14156117125760006116e283611947565b90508083600001818151019150818152505082602001515183600001511115151561170c57600080fd5b50611717565b600080fd5b5b5050565b606080835160405190808252806020026020018201604052801561174f5781602001602082028038833980820191505090505b50905060006117d0866040516020018082805190602001908083835b602083101515611790578051825260208201915060208101905060208303925061176b565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051602081830303815290604052805190602001206119d0565b9050600080905060008090505b8651811015611939576118078388838151811015156117f857fe5b90602001906020020151611a28565b848281518110151561181557fe5b9060200190602002019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff1681525050851561192c578173ffffffffffffffffffffffffffffffffffffffff16848281518110151561187c57fe5b9060200190602002015173ffffffffffffffffffffffffffffffffffffffff16111515611911576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f7369676e657273206e6f7420696e20617363656e64696e67206f72646572000081525060200191505060405180910390fd5b838181518110151561191f57fe5b9060200190602002015191505b80806001019150506117dd565b508293505050509392505050565b60008060608360200151905083600001519250826020820101519150600080935060008090505b600a8110156119c55783811a915060078102607f83169060020a028517945060006080831614156119b85760018101866000018181510191508181525050849450505050506119cb565b808060010191505061196e565b50600080fd5b919050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b600060418251141515611a3e5760009050611b04565b60008060006020850151925060408501519150606085015160001a9050601b8160ff161015611a6e57601b810190505b601b8160ff1614158015611a865750601c8160ff1614155b15611a975760009350505050611b04565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015611af4573d6000803e3d6000fd5b5050506020604051035193505050505b92915050565b604080519081016040528060608152602001606081525090565b608060405190810160405280600081526020016000815260200160608152602001600081525090565b60408051908101604052806000815260200160608152509056fea165627a7a72305820461a83b4d578ada2d2b7c8a25dacedb9bed0d2eca2cc7e7d3582c999dc396d480029`

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
// Solidity: function getActionDeadline() constant returns(uint256)
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
// Solidity: function getActionDeadline() constant returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetActionDeadline() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetActionDeadline(&_SimpleSingleSessionApp.CallOpts)
}

// GetActionDeadline is a free data retrieval call binding the contract method 0xbbc35280.
//
// Solidity: function getActionDeadline() constant returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetActionDeadline() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetActionDeadline(&_SimpleSingleSessionApp.CallOpts)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) constant returns(bool)
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
// Solidity: function getOutcome(bytes _query) constant returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleSingleSessionApp.Contract.GetOutcome(&_SimpleSingleSessionApp.CallOpts, _query)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) constant returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleSingleSessionApp.Contract.GetOutcome(&_SimpleSingleSessionApp.CallOpts, _query)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() constant returns(uint256)
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
// Solidity: function getSeqNum() constant returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetSeqNum() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSeqNum(&_SimpleSingleSessionApp.CallOpts)
}

// GetSeqNum is a free data retrieval call binding the contract method 0x6d15c457.
//
// Solidity: function getSeqNum() constant returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetSeqNum() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSeqNum(&_SimpleSingleSessionApp.CallOpts)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() constant returns(uint256)
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
// Solidity: function getSettleFinalizedTime() constant returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetSettleFinalizedTime() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSettleFinalizedTime(&_SimpleSingleSessionApp.CallOpts)
}

// GetSettleFinalizedTime is a free data retrieval call binding the contract method 0xb71ca01f.
//
// Solidity: function getSettleFinalizedTime() constant returns(uint256)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetSettleFinalizedTime() (*big.Int, error) {
	return _SimpleSingleSessionApp.Contract.GetSettleFinalizedTime(&_SimpleSingleSessionApp.CallOpts)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
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
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetState(_key *big.Int) ([]byte, error) {
	return _SimpleSingleSessionApp.Contract.GetState(&_SimpleSingleSessionApp.CallOpts, _key)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) constant returns(bytes)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetState(_key *big.Int) ([]byte, error) {
	return _SimpleSingleSessionApp.Contract.GetState(&_SimpleSingleSessionApp.CallOpts, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
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
// Solidity: function getStatus() constant returns(uint8)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) GetStatus() (uint8, error) {
	return _SimpleSingleSessionApp.Contract.GetStatus(&_SimpleSingleSessionApp.CallOpts)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() constant returns(uint8)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppCallerSession) GetStatus() (uint8, error) {
	return _SimpleSingleSessionApp.Contract.GetStatus(&_SimpleSingleSessionApp.CallOpts)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
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
// Solidity: function isFinalized(bytes _query) constant returns(bool)
func (_SimpleSingleSessionApp *SimpleSingleSessionAppSession) IsFinalized(_query []byte) (bool, error) {
	return _SimpleSingleSessionApp.Contract.IsFinalized(&_SimpleSingleSessionApp.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) constant returns(bool)
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
