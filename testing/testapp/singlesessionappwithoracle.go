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

// SimpleSingleSessionAppWithOracleABI is the input ABI used to generate the binding from.
const SimpleSingleSessionAppWithOracleABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"}],\"name\":\"settleBySigTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"},{\"name\":\"_cosignedStateProof\",\"type\":\"bytes\"}],\"name\":\"settleByInvalidTurn\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"isFinalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"}],\"name\":\"settleByMoveTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_oracleProof\",\"type\":\"bytes\"},{\"name\":\"_cosignedStateProof\",\"type\":\"bytes\"}],\"name\":\"settleByInvalidState\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_players\",\"type\":\"address[]\"},{\"name\":\"_nonce\",\"type\":\"uint256\"},{\"name\":\"_sigTimeout\",\"type\":\"uint256\"},{\"name\":\"_moveTimeout\",\"type\":\"uint256\"},{\"name\":\"_oracle\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"SigTimeoutDispute\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"MoveTimeoutDispute\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"InvalidTurnDispute\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"InvalidStateDispute\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_query\",\"type\":\"bytes\"}],\"name\":\"getOutcome\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"uint256\"}],\"name\":\"getState\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SimpleSingleSessionAppWithOracleBin is the compiled bytecode used for deploying new contracts.
var SimpleSingleSessionAppWithOracleBin = "0x60806040523480156200001157600080fd5b506040516200255f3803806200255f833981018060405260a08110156200003757600080fd5b8101908080516401000000008111156200005057600080fd5b828101905060208101848111156200006757600080fd5b81518560208202830111640100000000821117156200008557600080fd5b505092919060200180519060200190929190805190602001909291908051906020019092919080519060200190929190505050848484848484848484848360008001819055508460006001019080519060200190620000e692919062000178565b50826000600201819055508160006003018190555060008060040160006101000a81548160ff021916908360018111156200011d57fe5b021790555080600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505050505050505050505050505050506200024d565b828054828255906000526020600020908101928215620001f4579160200282015b82811115620001f35782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055509160200191906001019062000199565b5b50905062000203919062000207565b5090565b6200024a91905b808211156200024657600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff0219169055506001016200020e565b5090565b90565b612302806200025d6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063bcdbda941161005b578063bcdbda94146102a7578063ea4ba8eb14610338578063f26285b2146103c9578063fb3fe8061461044257610088565b80632141dbda1461008d57806344c9af28146101065780634e69d560146101ad578063a428cd3b146101d9575b600080fd5b610104600480360360208110156100a357600080fd5b81019080803590602001906401000000008111156100c057600080fd5b8201836020820111156100d257600080fd5b803590602001918460018302840111640100000000831117156100f457600080fd5b9091929391929390505050610510565b005b6101326004803603602081101561011c57600080fd5b810190808035906020019092919050505061072b565b6040518080602001828103825283818151815260200191508051906020019080838360005b83811015610172578082015181840152602081019050610157565b50505050905090810190601f16801561019f5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6101b5610789565b604051808260018111156101c557fe5b60ff16815260200191505060405180910390f35b6102a5600480360360408110156101ef57600080fd5b810190808035906020019064010000000081111561020c57600080fd5b82018360208201111561021e57600080fd5b8035906020019184600183028401116401000000008311171561024057600080fd5b90919293919293908035906020019064010000000081111561026157600080fd5b82018360208201111561027357600080fd5b8035906020019184600183028401116401000000008311171561029557600080fd5b90919293919293905050506107a2565b005b61031e600480360360208110156102bd57600080fd5b81019080803590602001906401000000008111156102da57600080fd5b8201836020820111156102ec57600080fd5b8035906020019184600183028401116401000000008311171561030e57600080fd5b9091929391929390505050610a1c565b604051808215151515815260200191505060405180910390f35b6103af6004803603602081101561034e57600080fd5b810190808035906020019064010000000081111561036b57600080fd5b82018360208201111561037d57600080fd5b8035906020019184600183028401116401000000008311171561039f57600080fd5b9091929391929390505050610a60565b604051808215151515815260200191505060405180910390f35b610440600480360360208110156103df57600080fd5b81019080803590602001906401000000008111156103fc57600080fd5b82018360208201111561040e57600080fd5b8035906020019184600183028401116401000000008311171561043057600080fd5b9091929391929390505050610b14565b005b61050e6004803603604081101561045857600080fd5b810190808035906020019064010000000081111561047557600080fd5b82018360208201111561048757600080fd5b803590602001918460018302840111640100000000831117156104a957600080fd5b9091929391929390803590602001906401000000008111156104ca57600080fd5b8201836020820111156104dc57600080fd5b803590602001918460018302840111640100000000831117156104fe57600080fd5b9091929391929390505050610cd6565b005b60018081111561051c57fe5b600060040160009054906101000a900460ff16600181111561053a57fe5b14156105ae576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b6105b661215e565b61060383838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050610f10565b905080606001516000600201548260400151011061066c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602781526020018061226d6027913960400191505060405180910390fd5b6106746121a3565b606061068583600001516001611099565b915091506000600101805490508151106106ea576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602b815260200180612242602b913960400191505060405180910390fd5b6106f88260400151826110ea565b7f4ede5136b5765b273092424d192085d14577fe5c0b0512401b5d3706b46c8e2060405160405180910390a15050505050565b60608060206040519080825280601f01601f1916602001820160405280156107625781602001600182028038833980820191505090505b5090506000600560149054906101000a900460ff1690508060208301528192505050919050565b60008060040160009054906101000a900460ff16905090565b6001808111156107ae57fe5b600060040160009054906101000a900460ff1660018111156107cc57fe5b1415610840576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b6108486121a3565b606061089984848080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050506001611099565b915091506108a68161114a565b60006108b583604001516112c1565b90506108bf61215e565b61090c88888080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050610f10565b9050806020015173ffffffffffffffffffffffffffffffffffffffff1660006001018360ff168154811061093c57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156109d4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260218152602001806122946021913960400191505060405180910390fd5b6109e6846040015182602001516112e5565b7f12583b36c4ceb396f9b64ae9b3055f92e067e9e27f25624468e49465c7d6d25860405160405180910390a15050505050505050565b6000808383905014610a2d57600080fd5b600180811115610a3957fe5b600060040160009054906101000a900460ff166001811115610a5757fe5b14905092915050565b600060018383905014610adb576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260148152602001807f696e76616c6964207175657279206c656e67746800000000000000000000000081525060200191505060405180910390fd5b82826000818110610ae857fe5b9050013560f81c60f81b60f81c60ff16600560149054906101000a900460ff1660ff1614905092915050565b600180811115610b2057fe5b600060040160009054906101000a900460ff166001811115610b3e57fe5b1415610bb2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b610bba61215e565b610c0783838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050610f10565b9050806060015160006003015482604001510110610c70576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602881526020018061221a6028913960400191505060405180910390fd5b610c786121a3565b6060610c8983600001516001611099565b91509150610c968161114a565b610ca38260400151611345565b7f5ca2c3e34ecb38f5b6e98856236647247d4acb6a797617631c7815dc16ac0c1160405160405180910390a15050505050565b600180811115610ce257fe5b600060040160009054906101000a900460ff166001811115610d0057fe5b1415610d74576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f6170702073746174652069732066696e616c697a65640000000000000000000081525060200191505060405180910390fd5b610d7c6121a3565b6060610dcd84848080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050506001611099565b91509150610dda8161114a565b610de261215e565b610e2f87878080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050610f10565b9050610e396121a3565b610e4882600001516000611099565b509050610e5584826113a4565b15610ec8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601f8152602001807f6f7261636c65206170707374617465206d75737420626520696e76616c69640081525060200191505060405180910390fd5b610eda846040015183602001516113e8565b7f58857d7e5d46d366a88b9b19c3c4bfb7573323b8e0602e3966d582bac5aa5fca60405160405180910390a15050505050505050565b610f1861215e565b610f206121cb565b610f2983611448565b90506000610fab82600001516040516020018082805190602001908083835b60208310610f6b5780518252602082019150602081019050602083039250610f48565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051602081830303815290604052805190602001206114f3565b90506000610fbd82846020015161154b565b9050600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614611082576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f7369676e6572206d757374206265206f7261636c65000000000000000000000081525060200191505060405180910390fd5b61108f836000015161162b565b9350505050919050565b6110a16121a3565b60606110ab6121e5565b6110b48561186d565b90506110c381600001516119e4565b925083156110e2576110df836040015182602001516000611ad2565b91505b509250929050565b6001600060040160006101000a81548160ff0219169083600181111561110c57fe5b02179055508160018151811061111e57fe5b602001015160f81c60f81b60f81c600560146101000a81548160ff021916908360ff1602179055505050565b6000600101805490508151146111c8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260198152602001807f696e76616c6964206e756d626572206f66207369676e6572730000000000000081525060200191505060405180910390fd5b60008090505b6000600101805490508110156112bd578181815181106111ea57fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff166000600101828154811061121857fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156112b0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806122b56022913960400191505060405180910390fd5b80806001019150506111ce565b5050565b6000816000815181106112d057fe5b602001015160f81c60f81b60f81c9050919050565b6001600060040160006101000a81548160ff0219169083600181111561130757fe5b02179055508160018151811061131957fe5b602001015160f81c60f81b60f81c600560146101000a81548160ff021916908360ff1602179055505050565b6001600060040160006101000a81548160ff0219169083600181111561136757fe5b02179055508060018151811061137957fe5b602001015160f81c60f81b60f81c600560146101000a81548160ff021916908360ff16021790555050565b600081602001518360200151106113be57600090506113e2565b6113d083604001518360400151611ce9565b6113dd57600090506113e2565b600190505b92915050565b6001600060040160006101000a81548160ff0219169083600181111561140a57fe5b02179055508160018151811061141c57fe5b602001015160f81c60f81b60f81c600560146101000a81548160ff021916908360ff1602179055505050565b6114506121cb565b6114586121ff565b61146183611e35565b90506000805b61147083611e64565b156114eb5761147e83611e79565b8092508193505050600015611492576114e6565b60018214156114b1576114a483611ead565b84600001819052506114e5565b60028214156114d0576114c383611ead565b84602001819052506114e4565b6114e38184611f6690919063ffffffff16565b5b5b5b611467565b505050919050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b6000604182511461155f5760009050611625565b60008060006020850151925060408501519150606085015160001a9050601b8160ff16101561158f57601b810190505b601b8160ff16141580156115a75750601c8160ff1614155b156115b85760009350505050611625565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015611615573d6000803e3d6000fd5b5050506020604051035193505050505b92915050565b61163361215e565b61163b6121ff565b61164483611e35565b9050606061165c600583611ff690919063ffffffff16565b90508060058151811061166b57fe5b60200260200101516040519080825280602002602001820160405280156116a15781602001602082028038833980820191505090505b5083608001819052506000816005815181106116b957fe5b6020026020010181815250506000805b6116d284611e64565b15611864576116e084611e79565b80925081935050506000156116f45761185f565b60018214156117135761170684611ead565b856000018190525061185e565b60028214156117695761172d61172885611ead565b61209b565b856020019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505061185d565b60038214156117895761177b846120ad565b85604001818152505061185c565b60048214156117a95761179b846120ad565b85606001818152505061185b565b6005821415611846576117c36117be85611ead565b61209b565b8560800151846005815181106117d557fe5b6020026020010151815181106117e757fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508260058151811061182e57fe5b6020026020010180518091906001018152505061185a565b6118598185611f6690919063ffffffff16565b5b5b5b5b5b5b6116c9565b50505050919050565b6118756121e5565b61187d6121ff565b61188683611e35565b9050606061189e600283611ff690919063ffffffff16565b9050806002815181106118ad57fe5b60200260200101516040519080825280602002602001820160405280156118e857816020015b60608152602001906001900390816118d35790505b50836020018190525060008160028151811061190057fe5b6020026020010181815250506000805b61191984611e64565b156119db5761192784611e79565b809250819350505060001561193b576119d6565b600182141561195a5761194d84611ead565b85600001819052506119d5565b60028214156119c05761196c84611ead565b85602001518460028151811061197e57fe5b60200260200101518151811061199057fe5b6020026020010181905250826002815181106119a857fe5b602002602001018051809190600101815250506119d4565b6119d38185611f6690919063ffffffff16565b5b5b5b611910565b50505050919050565b6119ec6121a3565b6119f46121ff565b6119fd83611e35565b90506000805b611a0c83611e64565b15611aca57611a1a83611e79565b8092508193505050600015611a2e57611ac5565b6001821415611a4e57611a40836120ad565b846000018181525050611ac4565b6002821415611a6e57611a60836120ad565b846020018181525050611ac3565b6003821415611a8d57611a8083611ead565b8460400181905250611ac2565b6004821415611aad57611a9f836120ad565b846060018181525050611ac1565b611ac08184611f6690919063ffffffff16565b5b5b5b5b5b611a03565b505050919050565b6060808351604051908082528060200260200182016040528015611b055781602001602082028038833980820191505090505b5090506000611b84866040516020018082805190602001908083835b60208310611b445780518252602082019150602081019050602083039250611b21565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051602081830303815290604052805190602001206114f3565b9050600080905060008090505b8651811015611cdb57611bb783888381518110611baa57fe5b602002602001015161154b565b848281518110611bc357fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508515611cce578173ffffffffffffffffffffffffffffffffffffffff16848281518110611c2657fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff1611611cb7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f7369676e657273206e6f7420696e20617363656e64696e67206f72646572000081525060200191505060405180910390fd5b838181518110611cc357fe5b602002602001015191505b8080600101915050611b91565b508293505050509392505050565b60006002835114611d62576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f696e76616c696420636f7369676e6564207374617465206c656e67746800000081525060200191505060405180910390fd5b6002825114611d745760009050611e2f565b6000611d7f846112c1565b90506000611d8c846112c1565b90508060ff166000600101805490506001840160ff1681611da957fe5b0614611dba57600092505050611e2f565b600085600181518110611dc957fe5b602001015160f81c60f81b60f81c9050600085600181518110611de857fe5b602001015160f81c60f81b60f81c905060018260ff161480611e0d575060028260ff16145b15611e26578060ff168260ff1614945050505050611e2f565b60019450505050505b92915050565b611e3d6121ff565b6001825111611e4b57600080fd5b8181602001819052506000816000018181525050919050565b60008160200151518260000151109050919050565b6000806000611e87846120ad565b905060088181611e9357fe5b049250600781166005811115611ea557fe5b915050915091565b60606000611eba836120ad565b90506000818460000151019050836020015151811115611ed957600080fd5b816040519080825280601f01601f191660200182016040528015611f0c5781602001600182028038833980820191505090505b50925060608460200151905060008086600001519050602086019150806020840101905060008090505b85811015611f51578082015181840152602081019050611f36565b50838760000181815250505050505050919050565b60006005811115611f7357fe5b816005811115611f7f57fe5b1415611f9457611f8e826120ad565b50611ff2565b60026005811115611fa157fe5b816005811115611fad57fe5b1415611fec576000611fbe836120ad565b90508083600001818151019150818152505082602001515183600001511115611fe657600080fd5b50611ff1565b600080fd5b5b5050565b6060600083600001519050600183016040519080825280602002602001820160405280156120335781602001602082028038833980820191505090505b5091506000805b61204386611e64565b156120885761205186611e79565b8092508193505050600184838151811061206757fe5b6020026020010181815101915081815250506120838682611f66565b61203a565b8286600001818152505050505092915050565b60006120a682612133565b9050919050565b60008060608360200151905083600001519250826020820101519150600080935060008090505b600a8110156121285783811a915060078102607f8316901b85179450600060808316141561211b57600181018660000181815101915081815250508494505050505061212e565b80806001019150506120d4565b50600080fd5b919050565b6000601482511461214357600080fd5b6c010000000000000000000000006020830151049050919050565b6040518060a0016040528060608152602001600073ffffffffffffffffffffffffffffffffffffffff1681526020016000815260200160008152602001606081525090565b6040518060800160405280600081526020016000815260200160608152602001600081525090565b604051806040016040528060608152602001606081525090565b604051806040016040528060608152602001606081525090565b60405180604001604052806000815260200160608152509056fe6f7261636c65207374617465206d757374206265206166746572206d6f766520646561646c696e657369676e657273206c656e677468206d75737420626520736d616c6c6572207468616e20706c61796572736f7261636c65207374617465206d7573742062652061667465722073696720646561646c696e656f7261636c652073746174652075736572206d75737420626520696e76616c696473746174652070726f6f66207369676e6572206d75737420626520636f7272656374a165627a7a7230582088b2d65da9605b4a847897fd4e15a3191ca2d3d8a925d14d0e030115a0b590d30029"

// DeploySimpleSingleSessionAppWithOracle deploys a new Ethereum contract, binding an instance of SimpleSingleSessionAppWithOracle to it.
func DeploySimpleSingleSessionAppWithOracle(auth *bind.TransactOpts, backend bind.ContractBackend, _players []common.Address, _nonce *big.Int, _sigTimeout *big.Int, _moveTimeout *big.Int, _oracle common.Address) (common.Address, *types.Transaction, *SimpleSingleSessionAppWithOracle, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSingleSessionAppWithOracleABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleSingleSessionAppWithOracleBin), backend, _players, _nonce, _sigTimeout, _moveTimeout, _oracle)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleSingleSessionAppWithOracle{SimpleSingleSessionAppWithOracleCaller: SimpleSingleSessionAppWithOracleCaller{contract: contract}, SimpleSingleSessionAppWithOracleTransactor: SimpleSingleSessionAppWithOracleTransactor{contract: contract}, SimpleSingleSessionAppWithOracleFilterer: SimpleSingleSessionAppWithOracleFilterer{contract: contract}}, nil
}

// SimpleSingleSessionAppWithOracle is an auto generated Go binding around an Ethereum contract.
type SimpleSingleSessionAppWithOracle struct {
	SimpleSingleSessionAppWithOracleCaller     // Read-only binding to the contract
	SimpleSingleSessionAppWithOracleTransactor // Write-only binding to the contract
	SimpleSingleSessionAppWithOracleFilterer   // Log filterer for contract events
}

// SimpleSingleSessionAppWithOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppWithOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSingleSessionAppWithOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppWithOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSingleSessionAppWithOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleSingleSessionAppWithOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSingleSessionAppWithOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleSingleSessionAppWithOracleSession struct {
	Contract     *SimpleSingleSessionAppWithOracle // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                     // Call options to use throughout this session
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// SimpleSingleSessionAppWithOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleSingleSessionAppWithOracleCallerSession struct {
	Contract *SimpleSingleSessionAppWithOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                           // Call options to use throughout this session
}

// SimpleSingleSessionAppWithOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleSingleSessionAppWithOracleTransactorSession struct {
	Contract     *SimpleSingleSessionAppWithOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                           // Transaction auth options to use throughout this session
}

// SimpleSingleSessionAppWithOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleSingleSessionAppWithOracleRaw struct {
	Contract *SimpleSingleSessionAppWithOracle // Generic contract binding to access the raw methods on
}

// SimpleSingleSessionAppWithOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppWithOracleCallerRaw struct {
	Contract *SimpleSingleSessionAppWithOracleCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleSingleSessionAppWithOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleSingleSessionAppWithOracleTransactorRaw struct {
	Contract *SimpleSingleSessionAppWithOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleSingleSessionAppWithOracle creates a new instance of SimpleSingleSessionAppWithOracle, bound to a specific deployed contract.
func NewSimpleSingleSessionAppWithOracle(address common.Address, backend bind.ContractBackend) (*SimpleSingleSessionAppWithOracle, error) {
	contract, err := bindSimpleSingleSessionAppWithOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracle{SimpleSingleSessionAppWithOracleCaller: SimpleSingleSessionAppWithOracleCaller{contract: contract}, SimpleSingleSessionAppWithOracleTransactor: SimpleSingleSessionAppWithOracleTransactor{contract: contract}, SimpleSingleSessionAppWithOracleFilterer: SimpleSingleSessionAppWithOracleFilterer{contract: contract}}, nil
}

// NewSimpleSingleSessionAppWithOracleCaller creates a new read-only instance of SimpleSingleSessionAppWithOracle, bound to a specific deployed contract.
func NewSimpleSingleSessionAppWithOracleCaller(address common.Address, caller bind.ContractCaller) (*SimpleSingleSessionAppWithOracleCaller, error) {
	contract, err := bindSimpleSingleSessionAppWithOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracleCaller{contract: contract}, nil
}

// NewSimpleSingleSessionAppWithOracleTransactor creates a new write-only instance of SimpleSingleSessionAppWithOracle, bound to a specific deployed contract.
func NewSimpleSingleSessionAppWithOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleSingleSessionAppWithOracleTransactor, error) {
	contract, err := bindSimpleSingleSessionAppWithOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracleTransactor{contract: contract}, nil
}

// NewSimpleSingleSessionAppWithOracleFilterer creates a new log filterer instance of SimpleSingleSessionAppWithOracle, bound to a specific deployed contract.
func NewSimpleSingleSessionAppWithOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleSingleSessionAppWithOracleFilterer, error) {
	contract, err := bindSimpleSingleSessionAppWithOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracleFilterer{contract: contract}, nil
}

// bindSimpleSingleSessionAppWithOracle binds a generic wrapper to an already deployed contract.
func bindSimpleSingleSessionAppWithOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSingleSessionAppWithOracleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSingleSessionAppWithOracle.Contract.SimpleSingleSessionAppWithOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SimpleSingleSessionAppWithOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SimpleSingleSessionAppWithOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSingleSessionAppWithOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.contract.Transact(opts, method, params...)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCaller) GetOutcome(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleSingleSessionAppWithOracle.contract.Call(opts, out, "getOutcome", _query)
	return *ret0, err
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.GetOutcome(&_SimpleSingleSessionAppWithOracle.CallOpts, _query)
}

// GetOutcome is a free data retrieval call binding the contract method 0xea4ba8eb.
//
// Solidity: function getOutcome(bytes _query) view returns(bool)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCallerSession) GetOutcome(_query []byte) (bool, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.GetOutcome(&_SimpleSingleSessionAppWithOracle.CallOpts, _query)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) view returns(bytes)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCaller) GetState(opts *bind.CallOpts, _key *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _SimpleSingleSessionAppWithOracle.contract.Call(opts, out, "getState", _key)
	return *ret0, err
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) view returns(bytes)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) GetState(_key *big.Int) ([]byte, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.GetState(&_SimpleSingleSessionAppWithOracle.CallOpts, _key)
}

// GetState is a free data retrieval call binding the contract method 0x44c9af28.
//
// Solidity: function getState(uint256 _key) view returns(bytes)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCallerSession) GetState(_key *big.Int) ([]byte, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.GetState(&_SimpleSingleSessionAppWithOracle.CallOpts, _key)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint8)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCaller) GetStatus(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _SimpleSingleSessionAppWithOracle.contract.Call(opts, out, "getStatus")
	return *ret0, err
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint8)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) GetStatus() (uint8, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.GetStatus(&_SimpleSingleSessionAppWithOracle.CallOpts)
}

// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
//
// Solidity: function getStatus() view returns(uint8)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCallerSession) GetStatus() (uint8, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.GetStatus(&_SimpleSingleSessionAppWithOracle.CallOpts)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCaller) IsFinalized(opts *bind.CallOpts, _query []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleSingleSessionAppWithOracle.contract.Call(opts, out, "isFinalized", _query)
	return *ret0, err
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) IsFinalized(_query []byte) (bool, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.IsFinalized(&_SimpleSingleSessionAppWithOracle.CallOpts, _query)
}

// IsFinalized is a free data retrieval call binding the contract method 0xbcdbda94.
//
// Solidity: function isFinalized(bytes _query) view returns(bool)
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleCallerSession) IsFinalized(_query []byte) (bool, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.IsFinalized(&_SimpleSingleSessionAppWithOracle.CallOpts, _query)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactor) SettleByInvalidState(opts *bind.TransactOpts, _oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.contract.Transact(opts, "settleByInvalidState", _oracleProof, _cosignedStateProof)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) SettleByInvalidState(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleByInvalidState(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidState is a paid mutator transaction binding the contract method 0xfb3fe806.
//
// Solidity: function settleByInvalidState(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactorSession) SettleByInvalidState(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleByInvalidState(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactor) SettleByInvalidTurn(opts *bind.TransactOpts, _oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.contract.Transact(opts, "settleByInvalidTurn", _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) SettleByInvalidTurn(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleByInvalidTurn(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByInvalidTurn is a paid mutator transaction binding the contract method 0xa428cd3b.
//
// Solidity: function settleByInvalidTurn(bytes _oracleProof, bytes _cosignedStateProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactorSession) SettleByInvalidTurn(_oracleProof []byte, _cosignedStateProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleByInvalidTurn(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof, _cosignedStateProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactor) SettleByMoveTimeout(opts *bind.TransactOpts, _oracleProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.contract.Transact(opts, "settleByMoveTimeout", _oracleProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) SettleByMoveTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleByMoveTimeout(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof)
}

// SettleByMoveTimeout is a paid mutator transaction binding the contract method 0xf26285b2.
//
// Solidity: function settleByMoveTimeout(bytes _oracleProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactorSession) SettleByMoveTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleByMoveTimeout(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactor) SettleBySigTimeout(opts *bind.TransactOpts, _oracleProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.contract.Transact(opts, "settleBySigTimeout", _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleSession) SettleBySigTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleBySigTimeout(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof)
}

// SettleBySigTimeout is a paid mutator transaction binding the contract method 0x2141dbda.
//
// Solidity: function settleBySigTimeout(bytes _oracleProof) returns()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleTransactorSession) SettleBySigTimeout(_oracleProof []byte) (*types.Transaction, error) {
	return _SimpleSingleSessionAppWithOracle.Contract.SettleBySigTimeout(&_SimpleSingleSessionAppWithOracle.TransactOpts, _oracleProof)
}

// SimpleSingleSessionAppWithOracleInvalidStateDisputeIterator is returned from FilterInvalidStateDispute and is used to iterate over the raw logs and unpacked data for InvalidStateDispute events raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleInvalidStateDisputeIterator struct {
	Event *SimpleSingleSessionAppWithOracleInvalidStateDispute // Event containing the contract specifics and raw log

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
func (it *SimpleSingleSessionAppWithOracleInvalidStateDisputeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSingleSessionAppWithOracleInvalidStateDispute)
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
		it.Event = new(SimpleSingleSessionAppWithOracleInvalidStateDispute)
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
func (it *SimpleSingleSessionAppWithOracleInvalidStateDisputeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSingleSessionAppWithOracleInvalidStateDisputeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSingleSessionAppWithOracleInvalidStateDispute represents a InvalidStateDispute event raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleInvalidStateDispute struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterInvalidStateDispute is a free log retrieval operation binding the contract event 0x58857d7e5d46d366a88b9b19c3c4bfb7573323b8e0602e3966d582bac5aa5fca.
//
// Solidity: event InvalidStateDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) FilterInvalidStateDispute(opts *bind.FilterOpts) (*SimpleSingleSessionAppWithOracleInvalidStateDisputeIterator, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.FilterLogs(opts, "InvalidStateDispute")
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracleInvalidStateDisputeIterator{contract: _SimpleSingleSessionAppWithOracle.contract, event: "InvalidStateDispute", logs: logs, sub: sub}, nil
}

// WatchInvalidStateDispute is a free log subscription operation binding the contract event 0x58857d7e5d46d366a88b9b19c3c4bfb7573323b8e0602e3966d582bac5aa5fca.
//
// Solidity: event InvalidStateDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) WatchInvalidStateDispute(opts *bind.WatchOpts, sink chan<- *SimpleSingleSessionAppWithOracleInvalidStateDispute) (event.Subscription, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.WatchLogs(opts, "InvalidStateDispute")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSingleSessionAppWithOracleInvalidStateDispute)
				if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "InvalidStateDispute", log); err != nil {
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

// ParseInvalidStateDispute is a log parse operation binding the contract event 0x58857d7e5d46d366a88b9b19c3c4bfb7573323b8e0602e3966d582bac5aa5fca.
//
// Solidity: event InvalidStateDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) ParseInvalidStateDispute(log types.Log) (*SimpleSingleSessionAppWithOracleInvalidStateDispute, error) {
	event := new(SimpleSingleSessionAppWithOracleInvalidStateDispute)
	if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "InvalidStateDispute", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SimpleSingleSessionAppWithOracleInvalidTurnDisputeIterator is returned from FilterInvalidTurnDispute and is used to iterate over the raw logs and unpacked data for InvalidTurnDispute events raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleInvalidTurnDisputeIterator struct {
	Event *SimpleSingleSessionAppWithOracleInvalidTurnDispute // Event containing the contract specifics and raw log

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
func (it *SimpleSingleSessionAppWithOracleInvalidTurnDisputeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSingleSessionAppWithOracleInvalidTurnDispute)
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
		it.Event = new(SimpleSingleSessionAppWithOracleInvalidTurnDispute)
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
func (it *SimpleSingleSessionAppWithOracleInvalidTurnDisputeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSingleSessionAppWithOracleInvalidTurnDisputeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSingleSessionAppWithOracleInvalidTurnDispute represents a InvalidTurnDispute event raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleInvalidTurnDispute struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterInvalidTurnDispute is a free log retrieval operation binding the contract event 0x12583b36c4ceb396f9b64ae9b3055f92e067e9e27f25624468e49465c7d6d258.
//
// Solidity: event InvalidTurnDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) FilterInvalidTurnDispute(opts *bind.FilterOpts) (*SimpleSingleSessionAppWithOracleInvalidTurnDisputeIterator, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.FilterLogs(opts, "InvalidTurnDispute")
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracleInvalidTurnDisputeIterator{contract: _SimpleSingleSessionAppWithOracle.contract, event: "InvalidTurnDispute", logs: logs, sub: sub}, nil
}

// WatchInvalidTurnDispute is a free log subscription operation binding the contract event 0x12583b36c4ceb396f9b64ae9b3055f92e067e9e27f25624468e49465c7d6d258.
//
// Solidity: event InvalidTurnDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) WatchInvalidTurnDispute(opts *bind.WatchOpts, sink chan<- *SimpleSingleSessionAppWithOracleInvalidTurnDispute) (event.Subscription, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.WatchLogs(opts, "InvalidTurnDispute")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSingleSessionAppWithOracleInvalidTurnDispute)
				if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "InvalidTurnDispute", log); err != nil {
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

// ParseInvalidTurnDispute is a log parse operation binding the contract event 0x12583b36c4ceb396f9b64ae9b3055f92e067e9e27f25624468e49465c7d6d258.
//
// Solidity: event InvalidTurnDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) ParseInvalidTurnDispute(log types.Log) (*SimpleSingleSessionAppWithOracleInvalidTurnDispute, error) {
	event := new(SimpleSingleSessionAppWithOracleInvalidTurnDispute)
	if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "InvalidTurnDispute", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SimpleSingleSessionAppWithOracleMoveTimeoutDisputeIterator is returned from FilterMoveTimeoutDispute and is used to iterate over the raw logs and unpacked data for MoveTimeoutDispute events raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleMoveTimeoutDisputeIterator struct {
	Event *SimpleSingleSessionAppWithOracleMoveTimeoutDispute // Event containing the contract specifics and raw log

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
func (it *SimpleSingleSessionAppWithOracleMoveTimeoutDisputeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSingleSessionAppWithOracleMoveTimeoutDispute)
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
		it.Event = new(SimpleSingleSessionAppWithOracleMoveTimeoutDispute)
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
func (it *SimpleSingleSessionAppWithOracleMoveTimeoutDisputeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSingleSessionAppWithOracleMoveTimeoutDisputeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSingleSessionAppWithOracleMoveTimeoutDispute represents a MoveTimeoutDispute event raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleMoveTimeoutDispute struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterMoveTimeoutDispute is a free log retrieval operation binding the contract event 0x5ca2c3e34ecb38f5b6e98856236647247d4acb6a797617631c7815dc16ac0c11.
//
// Solidity: event MoveTimeoutDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) FilterMoveTimeoutDispute(opts *bind.FilterOpts) (*SimpleSingleSessionAppWithOracleMoveTimeoutDisputeIterator, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.FilterLogs(opts, "MoveTimeoutDispute")
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracleMoveTimeoutDisputeIterator{contract: _SimpleSingleSessionAppWithOracle.contract, event: "MoveTimeoutDispute", logs: logs, sub: sub}, nil
}

// WatchMoveTimeoutDispute is a free log subscription operation binding the contract event 0x5ca2c3e34ecb38f5b6e98856236647247d4acb6a797617631c7815dc16ac0c11.
//
// Solidity: event MoveTimeoutDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) WatchMoveTimeoutDispute(opts *bind.WatchOpts, sink chan<- *SimpleSingleSessionAppWithOracleMoveTimeoutDispute) (event.Subscription, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.WatchLogs(opts, "MoveTimeoutDispute")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSingleSessionAppWithOracleMoveTimeoutDispute)
				if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "MoveTimeoutDispute", log); err != nil {
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

// ParseMoveTimeoutDispute is a log parse operation binding the contract event 0x5ca2c3e34ecb38f5b6e98856236647247d4acb6a797617631c7815dc16ac0c11.
//
// Solidity: event MoveTimeoutDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) ParseMoveTimeoutDispute(log types.Log) (*SimpleSingleSessionAppWithOracleMoveTimeoutDispute, error) {
	event := new(SimpleSingleSessionAppWithOracleMoveTimeoutDispute)
	if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "MoveTimeoutDispute", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SimpleSingleSessionAppWithOracleSigTimeoutDisputeIterator is returned from FilterSigTimeoutDispute and is used to iterate over the raw logs and unpacked data for SigTimeoutDispute events raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleSigTimeoutDisputeIterator struct {
	Event *SimpleSingleSessionAppWithOracleSigTimeoutDispute // Event containing the contract specifics and raw log

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
func (it *SimpleSingleSessionAppWithOracleSigTimeoutDisputeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSingleSessionAppWithOracleSigTimeoutDispute)
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
		it.Event = new(SimpleSingleSessionAppWithOracleSigTimeoutDispute)
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
func (it *SimpleSingleSessionAppWithOracleSigTimeoutDisputeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSingleSessionAppWithOracleSigTimeoutDisputeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSingleSessionAppWithOracleSigTimeoutDispute represents a SigTimeoutDispute event raised by the SimpleSingleSessionAppWithOracle contract.
type SimpleSingleSessionAppWithOracleSigTimeoutDispute struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterSigTimeoutDispute is a free log retrieval operation binding the contract event 0x4ede5136b5765b273092424d192085d14577fe5c0b0512401b5d3706b46c8e20.
//
// Solidity: event SigTimeoutDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) FilterSigTimeoutDispute(opts *bind.FilterOpts) (*SimpleSingleSessionAppWithOracleSigTimeoutDisputeIterator, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.FilterLogs(opts, "SigTimeoutDispute")
	if err != nil {
		return nil, err
	}
	return &SimpleSingleSessionAppWithOracleSigTimeoutDisputeIterator{contract: _SimpleSingleSessionAppWithOracle.contract, event: "SigTimeoutDispute", logs: logs, sub: sub}, nil
}

// WatchSigTimeoutDispute is a free log subscription operation binding the contract event 0x4ede5136b5765b273092424d192085d14577fe5c0b0512401b5d3706b46c8e20.
//
// Solidity: event SigTimeoutDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) WatchSigTimeoutDispute(opts *bind.WatchOpts, sink chan<- *SimpleSingleSessionAppWithOracleSigTimeoutDispute) (event.Subscription, error) {

	logs, sub, err := _SimpleSingleSessionAppWithOracle.contract.WatchLogs(opts, "SigTimeoutDispute")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSingleSessionAppWithOracleSigTimeoutDispute)
				if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "SigTimeoutDispute", log); err != nil {
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

// ParseSigTimeoutDispute is a log parse operation binding the contract event 0x4ede5136b5765b273092424d192085d14577fe5c0b0512401b5d3706b46c8e20.
//
// Solidity: event SigTimeoutDispute()
func (_SimpleSingleSessionAppWithOracle *SimpleSingleSessionAppWithOracleFilterer) ParseSigTimeoutDispute(log types.Log) (*SimpleSingleSessionAppWithOracleSigTimeoutDispute, error) {
	event := new(SimpleSingleSessionAppWithOracleSigTimeoutDispute)
	if err := _SimpleSingleSessionAppWithOracle.contract.UnpackLog(event, "SigTimeoutDispute", log); err != nil {
		return nil, err
	}
	return event, nil
}
