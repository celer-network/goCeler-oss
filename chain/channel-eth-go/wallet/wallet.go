// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package wallet

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

// CelerWalletABI is the input ABI used to generate the binding from.
const CelerWalletABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"walletNum\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"account\",\"type\":\"address\"}],\"name\":\"isPauser\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renouncePauser\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"account\",\"type\":\"address\"}],\"name\":\"addPauser\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"account\",\"type\":\"address\"}],\"name\":\"PauserAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"account\",\"type\":\"address\"}],\"name\":\"PauserRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"walletId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"owners\",\"type\":\"address[]\"},{\"indexed\":true,\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"CreateWallet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"walletId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"tokenAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DepositToWallet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"walletId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"tokenAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawFromWallet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"fromWalletId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"toWalletId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"tokenAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TransferToWallet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"walletId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"oldOperator\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"ChangeOperator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"walletId\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"newOperator\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"proposer\",\"type\":\"address\"}],\"name\":\"ProposeNewOperator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"tokenAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DrainToken\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owners\",\"type\":\"address[]\"},{\"name\":\"_operator\",\"type\":\"address\"},{\"name\":\"_nonce\",\"type\":\"bytes32\"}],\"name\":\"create\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"}],\"name\":\"depositETH\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"},{\"name\":\"_tokenAddress\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"depositERC20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"},{\"name\":\"_tokenAddress\",\"type\":\"address\"},{\"name\":\"_receiver\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_fromWalletId\",\"type\":\"bytes32\"},{\"name\":\"_toWalletId\",\"type\":\"bytes32\"},{\"name\":\"_tokenAddress\",\"type\":\"address\"},{\"name\":\"_receiver\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transferToWallet\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"},{\"name\":\"_newOperator\",\"type\":\"address\"}],\"name\":\"transferOperatorship\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"},{\"name\":\"_newOperator\",\"type\":\"address\"}],\"name\":\"proposeNewOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_tokenAddress\",\"type\":\"address\"},{\"name\":\"_receiver\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"drainToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"}],\"name\":\"getWalletOwners\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"}],\"name\":\"getOperator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"},{\"name\":\"_tokenAddress\",\"type\":\"address\"}],\"name\":\"getBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"}],\"name\":\"getProposedNewOperator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_walletId\",\"type\":\"bytes32\"},{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"getProposalVote\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// CelerWalletBin is the compiled bytecode used for deploying new contracts.
const CelerWalletBin = `0x6080604052620000183364010000000062000028810204565b6001805460ff1916905562000109565b62000043600082640100000000620015796200007a82021704565b604051600160a060020a038216907f6719d08c1888103bea251a4ed56406bd0c3e69723c8a1686e017e7bbe159b6f890600090a250565b600160a060020a0381166200008e57600080fd5b620000a38282640100000000620000d3810204565b15620000ae57600080fd5b600160a060020a0316600090815260209190915260409020805460ff19166001179055565b6000600160a060020a038216620000e957600080fd5b50600160a060020a03166000908152602091909152604090205460ff1690565b61176f80620001196000396000f3fe60806040526004361061013c576000357c01000000000000000000000000000000000000000000000000000000009004806380ba952e116100bd578063a96a5f9411610081578063a96a5f94146104f9578063bfa2c1d214610523578063c108bb4014610566578063cafd4600146105a5578063d68d9d4e146105de5761013c565b806380ba952e146103e057806382dc1ec41461042f5780638456cb59146104625780638e0cc17614610477578063a0c89a8c146104c05761013c565b80633f4ba83a116101045780633f4ba83a1461032157806346fbf68e14610336578063530e931c1461037d5780635c975abb146103b65780636ef8d66d146103cb5761013c565b80630d63a1fd1461014157806314da2906146102115780631687cc6014610257578063323c4480146102d157806336cc9e8d1461030c575b600080fd5b34801561014d57600080fd5b506101ff6004803603606081101561016457600080fd5b81019060208101813564010000000081111561017f57600080fd5b82018360208201111561019157600080fd5b803590602001918460208302840111640100000000831117156101b357600080fd5b91908080602002602001604051908101604052809392919081815260200183836020028082843760009201919091525092955050600160a060020a0383351693505050602001356105fb565b60408051918252519081900360200190f35b34801561021d57600080fd5b5061023b6004803603602081101561023457600080fd5b50356107e5565b60408051600160a060020a039092168252519081900360200190f35b34801561026357600080fd5b506102816004803603602081101561027a57600080fd5b5035610807565b60408051602080825283518183015283519192839290830191858101910280838360005b838110156102bd5781810151838201526020016102a5565b505050509050019250505060405180910390f35b3480156102dd57600080fd5b5061030a600480360360408110156102f457600080fd5b5080359060200135600160a060020a0316610873565b005b34801561031857600080fd5b506101ff6109ed565b34801561032d57600080fd5b5061030a6109f3565b34801561034257600080fd5b506103696004803603602081101561035957600080fd5b5035600160a060020a0316610a53565b604080519115158252519081900360200190f35b34801561038957600080fd5b506101ff600480360360408110156103a057600080fd5b5080359060200135600160a060020a0316610a6b565b3480156103c257600080fd5b50610369610a97565b3480156103d757600080fd5b5061030a610aa1565b3480156103ec57600080fd5b5061030a600480360360a081101561040357600080fd5b50803590602081013590600160a060020a03604082013581169160608101359091169060800135610aac565b34801561043b57600080fd5b5061030a6004803603602081101561045257600080fd5b5035600160a060020a0316610c46565b34801561046e57600080fd5b5061030a610c64565b34801561048357600080fd5b5061030a6004803603608081101561049a57600080fd5b50803590600160a060020a03602082013581169160408101359091169060600135610cc7565b3480156104cc57600080fd5b5061030a600480360360408110156104e357600080fd5b5080359060200135600160a060020a0316610e01565b34801561050557600080fd5b5061023b6004803603602081101561051c57600080fd5b5035610e93565b34801561052f57600080fd5b5061030a6004803603606081101561054657600080fd5b50600160a060020a03813581169160208101359091169060400135610eb1565b34801561057257600080fd5b5061030a6004803603606081101561058957600080fd5b50803590600160a060020a036020820135169060400135610f28565b3480156105b157600080fd5b50610369600480360360408110156105c857600080fd5b5080359060200135600160a060020a0316610fa0565b61030a600480360360208110156105f457600080fd5b503561101e565b60015460009060ff161561060e57600080fd5b600160a060020a03831661066c576040805160e560020a62461bcd02815260206004820152601a60248201527f4e6577206f70657261746f722069732061646472657373283029000000000000604482015290519081900360640190fd5b604080516c010000000000000000000000003081026020808401919091523391909102603483015260488083018690528351808403909101815260689092018352815191810191909120600081815260039092529190206001810154600160a060020a031615610726576040805160e560020a62461bcd02815260206004820152601260248201527f4f636375706965642077616c6c65742069640000000000000000000000000000604482015290519081900360640190fd5b85516107389082906020890190611676565b506001818101805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03881690811790915560028054909201909155604051875188919081906020808501910280838360005b838110156107a0578181015183820152602001610788565b505060405192909401829003822095508894507fe778e91533ef049a5fc99752bc4efb2b50ca4c967dfc0d4bb4782fb128070c3493506000925050a450949350505050565b60008181526003602081905260409091200154600160a060020a03165b919050565b60008181526003602090815260409182902080548351818402810184019094528084526060939283018282801561086757602002820191906000526020600020905b8154600160a060020a03168152600190910190602001808311610849575b50505050509050919050565b813361087f8282611078565b6108bd5760405160e560020a62461bcd02815260040180806020018281038252602181526020018061171a6021913960400191505060405180910390fd5b600160a060020a03831661091b576040805160e560020a62461bcd02815260206004820152601a60248201527f4e6577206f70657261746f722069732061646472657373283029000000000000604482015290519081900360640190fd5b600084815260036020819052604090912090810154600160a060020a038581169116146109765761094b816110dd565b60038101805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0386161790555b336000818152600483016020526040808220805460ff1916600117905551600160a060020a0387169188917f71f9e7796b33cb192d1670169ee7f4af7c5364f8f01bab4b95466787593745c39190a46109ce81611140565b156109e6576109dd85856111a9565b6109e6816110dd565b5050505050565b60025481565b6109fc33610a53565b610a0557600080fd5b60015460ff16610a1457600080fd5b6001805460ff191690556040805133815290517f5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa9181900360200190a1565b6000610a65818363ffffffff61127b16565b92915050565b6000828152600360209081526040808320600160a060020a038516845260020190915290205492915050565b60015460ff165b90565b610aaa336112b0565b565b60015460ff1615610abc57600080fd5b6000858152600360205260409020600101548590600160a060020a03163314610b2f576040805160e560020a62461bcd02815260206004820152601a60248201527f6d73672e73656e646572206973206e6f74206f70657261746f72000000000000604482015290519081900360640190fd5b8583610b3b8282611078565b610b795760405160e560020a62461bcd02815260040180806020018281038252602181526020018061171a6021913960400191505060405180910390fd5b8685610b858282611078565b610bc35760405160e560020a62461bcd02815260040180806020018281038252602181526020018061171a6021913960400191505060405180910390fd5b610bd08a898860016112f8565b610bdd89898860006112f8565b87600160a060020a0316898b7f1b56f805e5edb1e61b0d3f46feffdcbab5e591aa0e70e978ada9fc22093601c88a8a6040518083600160a060020a0316600160a060020a031681526020018281526020019250505060405180910390a450505050505050505050565b610c4f33610a53565b610c5857600080fd5b610c61816113a5565b50565b610c6d33610a53565b610c7657600080fd5b60015460ff1615610c8657600080fd5b6001805460ff1916811790556040805133815290517f62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a2589181900360200190a1565b60015460ff1615610cd757600080fd5b6000848152600360205260409020600101548490600160a060020a03163314610d4a576040805160e560020a62461bcd02815260206004820152601a60248201527f6d73672e73656e646572206973206e6f74206f70657261746f72000000000000604482015290519081900360640190fd5b8483610d568282611078565b610d945760405160e560020a62461bcd02815260040180806020018281038252602181526020018061171a6021913960400191505060405180910390fd5b610da187878660016112f8565b84600160a060020a031686600160a060020a0316887fd897e862036b62a0f770979fbd2227f3210565bba2eb4d9acd1dc8ccc00c928b876040518082815260200191505060405180910390a4610df88686866113ed565b50505050505050565b60015460ff1615610e1157600080fd5b6000828152600360205260409020600101548290600160a060020a03163314610e84576040805160e560020a62461bcd02815260206004820152601a60248201527f6d73672e73656e646572206973206e6f74206f70657261746f72000000000000604482015290519081900360640190fd5b610e8e83836111a9565b505050565b600090815260036020526040902060010154600160a060020a031690565b60015460ff16610ec057600080fd5b610ec933610a53565b610ed257600080fd5b81600160a060020a031683600160a060020a03167f896ecb17b26927fb33933fc5f413873193bced3c59fe736c42968a9778bf6b58836040518082815260200191505060405180910390a3610e8e8383836113ed565b60015460ff1615610f3857600080fd5b610f4583838360006112f8565b604080518281529051600160a060020a0384169185917fbc8e388b96ba8b9f627cb6d72d3513182f763c33c6107ecd31191de1f71abc1a9181900360200190a3610e8e600160a060020a03831633308463ffffffff61145416565b60008282610fae8282611078565b610fec5760405160e560020a62461bcd02815260040180806020018281038252602181526020018061171a6021913960400191505060405180910390fd5b5050506000918252600360209081526040808420600160a060020a039390931684526004909201905290205460ff1690565b60015460ff161561102e57600080fd5b3461103c82600083816112f8565b60408051828152905160009184917fbc8e388b96ba8b9f627cb6d72d3513182f763c33c6107ecd31191de1f71abc1a9181900360200190a35050565b6000828152600360205260408120815b81548110156110d2578160000181815481106110a057fe5b600091825260209091200154600160a060020a03858116911614156110ca57600192505050610a65565b600101611088565b506000949350505050565b60005b815481101561113c57600082600401600084600001848154811061110057fe5b600091825260208083209190910154600160a060020a031683528201929092526040019020805460ff19169115159190911790556001016110e0565b5050565b6000805b82548110156111a05782600401600084600001838154811061116257fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff16611198576000915050610802565b600101611144565b50600192915050565b600160a060020a038116611207576040805160e560020a62461bcd02815260206004820152601a60248201527f4e6577206f70657261746f722069732061646472657373283029000000000000604482015290519081900360640190fd5b600082815260036020526040808220600181018054600160a060020a0386811673ffffffffffffffffffffffffffffffffffffffff1983168117909355935192949316929091839187917f118c3f8030bc3c8254e737a0bd0584403c33646afbcbee8321c3bd5b26543cda9190a450505050565b6000600160a060020a03821661129057600080fd5b50600160a060020a03166000908152602091909152604090205460ff1690565b6112c160008263ffffffff61150316565b604051600160a060020a038216907fcd265ebaf09df2871cc7bd4133404a235ba12eff2041bb89d9c714a2621c7c7e90600090a250565b60008481526003602052604081209082600181111561131357fe5b141561136457600160a060020a0384166000908152600282016020526040902054611344908463ffffffff61154b16565b600160a060020a03851660009081526002830160205260409020556109e6565b600182600181111561137257fe5b14156113a357600160a060020a0384166000908152600282016020526040902054611344908463ffffffff61156416565bfe5b6113b660008263ffffffff61157916565b604051600160a060020a038216907f6719d08c1888103bea251a4ed56406bd0c3e69723c8a1686e017e7bbe159b6f890600090a250565b600160a060020a03831661143a576040518290600160a060020a0382169083156108fc029084906000818181858888f19350505050158015611433573d6000803e3d6000fd5b5050610e8e565b610e8e600160a060020a038416838363ffffffff6115c516565b604080517f23b872dd000000000000000000000000000000000000000000000000000000008152600160a060020a0385811660048301528481166024830152604482018490529151918616916323b872dd916064808201926020929091908290030181600087803b1580156114c857600080fd5b505af11580156114dc573d6000803e3d6000fd5b505050506040513d60208110156114f257600080fd5b50516114fd57600080fd5b50505050565b600160a060020a03811661151657600080fd5b611520828261127b565b61152957600080fd5b600160a060020a0316600090815260209190915260409020805460ff19169055565b60008282018381101561155d57600080fd5b9392505050565b60008282111561157357600080fd5b50900390565b600160a060020a03811661158c57600080fd5b611596828261127b565b156115a057600080fd5b600160a060020a0316600090815260209190915260409020805460ff19166001179055565b82600160a060020a031663a9059cbb83836040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b15801561164157600080fd5b505af1158015611655573d6000803e3d6000fd5b505050506040513d602081101561166b57600080fd5b5051610e8e57600080fd5b8280548282559060005260206000209081019282156116d8579160200282015b828111156116d8578251825473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03909116178255602090920191600190910190611696565b506116e49291506116e8565b5090565b610a9e91905b808211156116e457805473ffffffffffffffffffffffffffffffffffffffff191681556001016116ee56fe476976656e2061646472657373206973206e6f742077616c6c6574206f776e6572a265627a7a72305820a7523d8741c499742d47f79cb59b91012c3a3ce44436946f82d115663c4da30d64736f6c634300050a0032`

// DeployCelerWallet deploys a new Ethereum contract, binding an instance of CelerWallet to it.
func DeployCelerWallet(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *CelerWallet, error) {
	parsed, err := abi.JSON(strings.NewReader(CelerWalletABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(CelerWalletBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CelerWallet{CelerWalletCaller: CelerWalletCaller{contract: contract}, CelerWalletTransactor: CelerWalletTransactor{contract: contract}, CelerWalletFilterer: CelerWalletFilterer{contract: contract}}, nil
}

// CelerWallet is an auto generated Go binding around an Ethereum contract.
type CelerWallet struct {
	CelerWalletCaller     // Read-only binding to the contract
	CelerWalletTransactor // Write-only binding to the contract
	CelerWalletFilterer   // Log filterer for contract events
}

// CelerWalletCaller is an auto generated read-only Go binding around an Ethereum contract.
type CelerWalletCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CelerWalletTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CelerWalletTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CelerWalletFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CelerWalletFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CelerWalletSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CelerWalletSession struct {
	Contract     *CelerWallet      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CelerWalletCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CelerWalletCallerSession struct {
	Contract *CelerWalletCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// CelerWalletTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CelerWalletTransactorSession struct {
	Contract     *CelerWalletTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// CelerWalletRaw is an auto generated low-level Go binding around an Ethereum contract.
type CelerWalletRaw struct {
	Contract *CelerWallet // Generic contract binding to access the raw methods on
}

// CelerWalletCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CelerWalletCallerRaw struct {
	Contract *CelerWalletCaller // Generic read-only contract binding to access the raw methods on
}

// CelerWalletTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CelerWalletTransactorRaw struct {
	Contract *CelerWalletTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCelerWallet creates a new instance of CelerWallet, bound to a specific deployed contract.
func NewCelerWallet(address common.Address, backend bind.ContractBackend) (*CelerWallet, error) {
	contract, err := bindCelerWallet(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CelerWallet{CelerWalletCaller: CelerWalletCaller{contract: contract}, CelerWalletTransactor: CelerWalletTransactor{contract: contract}, CelerWalletFilterer: CelerWalletFilterer{contract: contract}}, nil
}

// NewCelerWalletCaller creates a new read-only instance of CelerWallet, bound to a specific deployed contract.
func NewCelerWalletCaller(address common.Address, caller bind.ContractCaller) (*CelerWalletCaller, error) {
	contract, err := bindCelerWallet(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CelerWalletCaller{contract: contract}, nil
}

// NewCelerWalletTransactor creates a new write-only instance of CelerWallet, bound to a specific deployed contract.
func NewCelerWalletTransactor(address common.Address, transactor bind.ContractTransactor) (*CelerWalletTransactor, error) {
	contract, err := bindCelerWallet(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CelerWalletTransactor{contract: contract}, nil
}

// NewCelerWalletFilterer creates a new log filterer instance of CelerWallet, bound to a specific deployed contract.
func NewCelerWalletFilterer(address common.Address, filterer bind.ContractFilterer) (*CelerWalletFilterer, error) {
	contract, err := bindCelerWallet(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CelerWalletFilterer{contract: contract}, nil
}

// bindCelerWallet binds a generic wrapper to an already deployed contract.
func bindCelerWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(CelerWalletABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CelerWallet *CelerWalletRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _CelerWallet.Contract.CelerWalletCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CelerWallet *CelerWalletRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CelerWallet.Contract.CelerWalletTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CelerWallet *CelerWalletRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CelerWallet.Contract.CelerWalletTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CelerWallet *CelerWalletCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _CelerWallet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CelerWallet *CelerWalletTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CelerWallet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CelerWallet *CelerWalletTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CelerWallet.Contract.contract.Transact(opts, method, params...)
}

// GetBalance is a free data retrieval call binding the contract method 0x530e931c.
//
// Solidity: function getBalance(bytes32 _walletId, address _tokenAddress) constant returns(uint256)
func (_CelerWallet *CelerWalletCaller) GetBalance(opts *bind.CallOpts, _walletId [32]byte, _tokenAddress common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "getBalance", _walletId, _tokenAddress)
	return *ret0, err
}

// GetBalance is a free data retrieval call binding the contract method 0x530e931c.
//
// Solidity: function getBalance(bytes32 _walletId, address _tokenAddress) constant returns(uint256)
func (_CelerWallet *CelerWalletSession) GetBalance(_walletId [32]byte, _tokenAddress common.Address) (*big.Int, error) {
	return _CelerWallet.Contract.GetBalance(&_CelerWallet.CallOpts, _walletId, _tokenAddress)
}

// GetBalance is a free data retrieval call binding the contract method 0x530e931c.
//
// Solidity: function getBalance(bytes32 _walletId, address _tokenAddress) constant returns(uint256)
func (_CelerWallet *CelerWalletCallerSession) GetBalance(_walletId [32]byte, _tokenAddress common.Address) (*big.Int, error) {
	return _CelerWallet.Contract.GetBalance(&_CelerWallet.CallOpts, _walletId, _tokenAddress)
}

// GetOperator is a free data retrieval call binding the contract method 0xa96a5f94.
//
// Solidity: function getOperator(bytes32 _walletId) constant returns(address)
func (_CelerWallet *CelerWalletCaller) GetOperator(opts *bind.CallOpts, _walletId [32]byte) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "getOperator", _walletId)
	return *ret0, err
}

// GetOperator is a free data retrieval call binding the contract method 0xa96a5f94.
//
// Solidity: function getOperator(bytes32 _walletId) constant returns(address)
func (_CelerWallet *CelerWalletSession) GetOperator(_walletId [32]byte) (common.Address, error) {
	return _CelerWallet.Contract.GetOperator(&_CelerWallet.CallOpts, _walletId)
}

// GetOperator is a free data retrieval call binding the contract method 0xa96a5f94.
//
// Solidity: function getOperator(bytes32 _walletId) constant returns(address)
func (_CelerWallet *CelerWalletCallerSession) GetOperator(_walletId [32]byte) (common.Address, error) {
	return _CelerWallet.Contract.GetOperator(&_CelerWallet.CallOpts, _walletId)
}

// GetProposalVote is a free data retrieval call binding the contract method 0xcafd4600.
//
// Solidity: function getProposalVote(bytes32 _walletId, address _owner) constant returns(bool)
func (_CelerWallet *CelerWalletCaller) GetProposalVote(opts *bind.CallOpts, _walletId [32]byte, _owner common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "getProposalVote", _walletId, _owner)
	return *ret0, err
}

// GetProposalVote is a free data retrieval call binding the contract method 0xcafd4600.
//
// Solidity: function getProposalVote(bytes32 _walletId, address _owner) constant returns(bool)
func (_CelerWallet *CelerWalletSession) GetProposalVote(_walletId [32]byte, _owner common.Address) (bool, error) {
	return _CelerWallet.Contract.GetProposalVote(&_CelerWallet.CallOpts, _walletId, _owner)
}

// GetProposalVote is a free data retrieval call binding the contract method 0xcafd4600.
//
// Solidity: function getProposalVote(bytes32 _walletId, address _owner) constant returns(bool)
func (_CelerWallet *CelerWalletCallerSession) GetProposalVote(_walletId [32]byte, _owner common.Address) (bool, error) {
	return _CelerWallet.Contract.GetProposalVote(&_CelerWallet.CallOpts, _walletId, _owner)
}

// GetProposedNewOperator is a free data retrieval call binding the contract method 0x14da2906.
//
// Solidity: function getProposedNewOperator(bytes32 _walletId) constant returns(address)
func (_CelerWallet *CelerWalletCaller) GetProposedNewOperator(opts *bind.CallOpts, _walletId [32]byte) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "getProposedNewOperator", _walletId)
	return *ret0, err
}

// GetProposedNewOperator is a free data retrieval call binding the contract method 0x14da2906.
//
// Solidity: function getProposedNewOperator(bytes32 _walletId) constant returns(address)
func (_CelerWallet *CelerWalletSession) GetProposedNewOperator(_walletId [32]byte) (common.Address, error) {
	return _CelerWallet.Contract.GetProposedNewOperator(&_CelerWallet.CallOpts, _walletId)
}

// GetProposedNewOperator is a free data retrieval call binding the contract method 0x14da2906.
//
// Solidity: function getProposedNewOperator(bytes32 _walletId) constant returns(address)
func (_CelerWallet *CelerWalletCallerSession) GetProposedNewOperator(_walletId [32]byte) (common.Address, error) {
	return _CelerWallet.Contract.GetProposedNewOperator(&_CelerWallet.CallOpts, _walletId)
}

// GetWalletOwners is a free data retrieval call binding the contract method 0x1687cc60.
//
// Solidity: function getWalletOwners(bytes32 _walletId) constant returns(address[])
func (_CelerWallet *CelerWalletCaller) GetWalletOwners(opts *bind.CallOpts, _walletId [32]byte) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "getWalletOwners", _walletId)
	return *ret0, err
}

// GetWalletOwners is a free data retrieval call binding the contract method 0x1687cc60.
//
// Solidity: function getWalletOwners(bytes32 _walletId) constant returns(address[])
func (_CelerWallet *CelerWalletSession) GetWalletOwners(_walletId [32]byte) ([]common.Address, error) {
	return _CelerWallet.Contract.GetWalletOwners(&_CelerWallet.CallOpts, _walletId)
}

// GetWalletOwners is a free data retrieval call binding the contract method 0x1687cc60.
//
// Solidity: function getWalletOwners(bytes32 _walletId) constant returns(address[])
func (_CelerWallet *CelerWalletCallerSession) GetWalletOwners(_walletId [32]byte) ([]common.Address, error) {
	return _CelerWallet.Contract.GetWalletOwners(&_CelerWallet.CallOpts, _walletId)
}

// IsPauser is a free data retrieval call binding the contract method 0x46fbf68e.
//
// Solidity: function isPauser(address account) constant returns(bool)
func (_CelerWallet *CelerWalletCaller) IsPauser(opts *bind.CallOpts, account common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "isPauser", account)
	return *ret0, err
}

// IsPauser is a free data retrieval call binding the contract method 0x46fbf68e.
//
// Solidity: function isPauser(address account) constant returns(bool)
func (_CelerWallet *CelerWalletSession) IsPauser(account common.Address) (bool, error) {
	return _CelerWallet.Contract.IsPauser(&_CelerWallet.CallOpts, account)
}

// IsPauser is a free data retrieval call binding the contract method 0x46fbf68e.
//
// Solidity: function isPauser(address account) constant returns(bool)
func (_CelerWallet *CelerWalletCallerSession) IsPauser(account common.Address) (bool, error) {
	return _CelerWallet.Contract.IsPauser(&_CelerWallet.CallOpts, account)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() constant returns(bool)
func (_CelerWallet *CelerWalletCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "paused")
	return *ret0, err
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() constant returns(bool)
func (_CelerWallet *CelerWalletSession) Paused() (bool, error) {
	return _CelerWallet.Contract.Paused(&_CelerWallet.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() constant returns(bool)
func (_CelerWallet *CelerWalletCallerSession) Paused() (bool, error) {
	return _CelerWallet.Contract.Paused(&_CelerWallet.CallOpts)
}

// WalletNum is a free data retrieval call binding the contract method 0x36cc9e8d.
//
// Solidity: function walletNum() constant returns(uint256)
func (_CelerWallet *CelerWalletCaller) WalletNum(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _CelerWallet.contract.Call(opts, out, "walletNum")
	return *ret0, err
}

// WalletNum is a free data retrieval call binding the contract method 0x36cc9e8d.
//
// Solidity: function walletNum() constant returns(uint256)
func (_CelerWallet *CelerWalletSession) WalletNum() (*big.Int, error) {
	return _CelerWallet.Contract.WalletNum(&_CelerWallet.CallOpts)
}

// WalletNum is a free data retrieval call binding the contract method 0x36cc9e8d.
//
// Solidity: function walletNum() constant returns(uint256)
func (_CelerWallet *CelerWalletCallerSession) WalletNum() (*big.Int, error) {
	return _CelerWallet.Contract.WalletNum(&_CelerWallet.CallOpts)
}

// AddPauser is a paid mutator transaction binding the contract method 0x82dc1ec4.
//
// Solidity: function addPauser(address account) returns()
func (_CelerWallet *CelerWalletTransactor) AddPauser(opts *bind.TransactOpts, account common.Address) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "addPauser", account)
}

// AddPauser is a paid mutator transaction binding the contract method 0x82dc1ec4.
//
// Solidity: function addPauser(address account) returns()
func (_CelerWallet *CelerWalletSession) AddPauser(account common.Address) (*types.Transaction, error) {
	return _CelerWallet.Contract.AddPauser(&_CelerWallet.TransactOpts, account)
}

// AddPauser is a paid mutator transaction binding the contract method 0x82dc1ec4.
//
// Solidity: function addPauser(address account) returns()
func (_CelerWallet *CelerWalletTransactorSession) AddPauser(account common.Address) (*types.Transaction, error) {
	return _CelerWallet.Contract.AddPauser(&_CelerWallet.TransactOpts, account)
}

// Create is a paid mutator transaction binding the contract method 0x0d63a1fd.
//
// Solidity: function create(address[] _owners, address _operator, bytes32 _nonce) returns(bytes32)
func (_CelerWallet *CelerWalletTransactor) Create(opts *bind.TransactOpts, _owners []common.Address, _operator common.Address, _nonce [32]byte) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "create", _owners, _operator, _nonce)
}

// Create is a paid mutator transaction binding the contract method 0x0d63a1fd.
//
// Solidity: function create(address[] _owners, address _operator, bytes32 _nonce) returns(bytes32)
func (_CelerWallet *CelerWalletSession) Create(_owners []common.Address, _operator common.Address, _nonce [32]byte) (*types.Transaction, error) {
	return _CelerWallet.Contract.Create(&_CelerWallet.TransactOpts, _owners, _operator, _nonce)
}

// Create is a paid mutator transaction binding the contract method 0x0d63a1fd.
//
// Solidity: function create(address[] _owners, address _operator, bytes32 _nonce) returns(bytes32)
func (_CelerWallet *CelerWalletTransactorSession) Create(_owners []common.Address, _operator common.Address, _nonce [32]byte) (*types.Transaction, error) {
	return _CelerWallet.Contract.Create(&_CelerWallet.TransactOpts, _owners, _operator, _nonce)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0xc108bb40.
//
// Solidity: function depositERC20(bytes32 _walletId, address _tokenAddress, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactor) DepositERC20(opts *bind.TransactOpts, _walletId [32]byte, _tokenAddress common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "depositERC20", _walletId, _tokenAddress, _amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0xc108bb40.
//
// Solidity: function depositERC20(bytes32 _walletId, address _tokenAddress, uint256 _amount) returns()
func (_CelerWallet *CelerWalletSession) DepositERC20(_walletId [32]byte, _tokenAddress common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.DepositERC20(&_CelerWallet.TransactOpts, _walletId, _tokenAddress, _amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0xc108bb40.
//
// Solidity: function depositERC20(bytes32 _walletId, address _tokenAddress, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactorSession) DepositERC20(_walletId [32]byte, _tokenAddress common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.DepositERC20(&_CelerWallet.TransactOpts, _walletId, _tokenAddress, _amount)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 _walletId) returns()
func (_CelerWallet *CelerWalletTransactor) DepositETH(opts *bind.TransactOpts, _walletId [32]byte) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "depositETH", _walletId)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 _walletId) returns()
func (_CelerWallet *CelerWalletSession) DepositETH(_walletId [32]byte) (*types.Transaction, error) {
	return _CelerWallet.Contract.DepositETH(&_CelerWallet.TransactOpts, _walletId)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 _walletId) returns()
func (_CelerWallet *CelerWalletTransactorSession) DepositETH(_walletId [32]byte) (*types.Transaction, error) {
	return _CelerWallet.Contract.DepositETH(&_CelerWallet.TransactOpts, _walletId)
}

// DrainToken is a paid mutator transaction binding the contract method 0xbfa2c1d2.
//
// Solidity: function drainToken(address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactor) DrainToken(opts *bind.TransactOpts, _tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "drainToken", _tokenAddress, _receiver, _amount)
}

// DrainToken is a paid mutator transaction binding the contract method 0xbfa2c1d2.
//
// Solidity: function drainToken(address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletSession) DrainToken(_tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.DrainToken(&_CelerWallet.TransactOpts, _tokenAddress, _receiver, _amount)
}

// DrainToken is a paid mutator transaction binding the contract method 0xbfa2c1d2.
//
// Solidity: function drainToken(address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactorSession) DrainToken(_tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.DrainToken(&_CelerWallet.TransactOpts, _tokenAddress, _receiver, _amount)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_CelerWallet *CelerWalletTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_CelerWallet *CelerWalletSession) Pause() (*types.Transaction, error) {
	return _CelerWallet.Contract.Pause(&_CelerWallet.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_CelerWallet *CelerWalletTransactorSession) Pause() (*types.Transaction, error) {
	return _CelerWallet.Contract.Pause(&_CelerWallet.TransactOpts)
}

// ProposeNewOperator is a paid mutator transaction binding the contract method 0x323c4480.
//
// Solidity: function proposeNewOperator(bytes32 _walletId, address _newOperator) returns()
func (_CelerWallet *CelerWalletTransactor) ProposeNewOperator(opts *bind.TransactOpts, _walletId [32]byte, _newOperator common.Address) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "proposeNewOperator", _walletId, _newOperator)
}

// ProposeNewOperator is a paid mutator transaction binding the contract method 0x323c4480.
//
// Solidity: function proposeNewOperator(bytes32 _walletId, address _newOperator) returns()
func (_CelerWallet *CelerWalletSession) ProposeNewOperator(_walletId [32]byte, _newOperator common.Address) (*types.Transaction, error) {
	return _CelerWallet.Contract.ProposeNewOperator(&_CelerWallet.TransactOpts, _walletId, _newOperator)
}

// ProposeNewOperator is a paid mutator transaction binding the contract method 0x323c4480.
//
// Solidity: function proposeNewOperator(bytes32 _walletId, address _newOperator) returns()
func (_CelerWallet *CelerWalletTransactorSession) ProposeNewOperator(_walletId [32]byte, _newOperator common.Address) (*types.Transaction, error) {
	return _CelerWallet.Contract.ProposeNewOperator(&_CelerWallet.TransactOpts, _walletId, _newOperator)
}

// RenouncePauser is a paid mutator transaction binding the contract method 0x6ef8d66d.
//
// Solidity: function renouncePauser() returns()
func (_CelerWallet *CelerWalletTransactor) RenouncePauser(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "renouncePauser")
}

// RenouncePauser is a paid mutator transaction binding the contract method 0x6ef8d66d.
//
// Solidity: function renouncePauser() returns()
func (_CelerWallet *CelerWalletSession) RenouncePauser() (*types.Transaction, error) {
	return _CelerWallet.Contract.RenouncePauser(&_CelerWallet.TransactOpts)
}

// RenouncePauser is a paid mutator transaction binding the contract method 0x6ef8d66d.
//
// Solidity: function renouncePauser() returns()
func (_CelerWallet *CelerWalletTransactorSession) RenouncePauser() (*types.Transaction, error) {
	return _CelerWallet.Contract.RenouncePauser(&_CelerWallet.TransactOpts)
}

// TransferOperatorship is a paid mutator transaction binding the contract method 0xa0c89a8c.
//
// Solidity: function transferOperatorship(bytes32 _walletId, address _newOperator) returns()
func (_CelerWallet *CelerWalletTransactor) TransferOperatorship(opts *bind.TransactOpts, _walletId [32]byte, _newOperator common.Address) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "transferOperatorship", _walletId, _newOperator)
}

// TransferOperatorship is a paid mutator transaction binding the contract method 0xa0c89a8c.
//
// Solidity: function transferOperatorship(bytes32 _walletId, address _newOperator) returns()
func (_CelerWallet *CelerWalletSession) TransferOperatorship(_walletId [32]byte, _newOperator common.Address) (*types.Transaction, error) {
	return _CelerWallet.Contract.TransferOperatorship(&_CelerWallet.TransactOpts, _walletId, _newOperator)
}

// TransferOperatorship is a paid mutator transaction binding the contract method 0xa0c89a8c.
//
// Solidity: function transferOperatorship(bytes32 _walletId, address _newOperator) returns()
func (_CelerWallet *CelerWalletTransactorSession) TransferOperatorship(_walletId [32]byte, _newOperator common.Address) (*types.Transaction, error) {
	return _CelerWallet.Contract.TransferOperatorship(&_CelerWallet.TransactOpts, _walletId, _newOperator)
}

// TransferToWallet is a paid mutator transaction binding the contract method 0x80ba952e.
//
// Solidity: function transferToWallet(bytes32 _fromWalletId, bytes32 _toWalletId, address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactor) TransferToWallet(opts *bind.TransactOpts, _fromWalletId [32]byte, _toWalletId [32]byte, _tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "transferToWallet", _fromWalletId, _toWalletId, _tokenAddress, _receiver, _amount)
}

// TransferToWallet is a paid mutator transaction binding the contract method 0x80ba952e.
//
// Solidity: function transferToWallet(bytes32 _fromWalletId, bytes32 _toWalletId, address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletSession) TransferToWallet(_fromWalletId [32]byte, _toWalletId [32]byte, _tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.TransferToWallet(&_CelerWallet.TransactOpts, _fromWalletId, _toWalletId, _tokenAddress, _receiver, _amount)
}

// TransferToWallet is a paid mutator transaction binding the contract method 0x80ba952e.
//
// Solidity: function transferToWallet(bytes32 _fromWalletId, bytes32 _toWalletId, address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactorSession) TransferToWallet(_fromWalletId [32]byte, _toWalletId [32]byte, _tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.TransferToWallet(&_CelerWallet.TransactOpts, _fromWalletId, _toWalletId, _tokenAddress, _receiver, _amount)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_CelerWallet *CelerWalletTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_CelerWallet *CelerWalletSession) Unpause() (*types.Transaction, error) {
	return _CelerWallet.Contract.Unpause(&_CelerWallet.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_CelerWallet *CelerWalletTransactorSession) Unpause() (*types.Transaction, error) {
	return _CelerWallet.Contract.Unpause(&_CelerWallet.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x8e0cc176.
//
// Solidity: function withdraw(bytes32 _walletId, address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactor) Withdraw(opts *bind.TransactOpts, _walletId [32]byte, _tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.contract.Transact(opts, "withdraw", _walletId, _tokenAddress, _receiver, _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x8e0cc176.
//
// Solidity: function withdraw(bytes32 _walletId, address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletSession) Withdraw(_walletId [32]byte, _tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.Withdraw(&_CelerWallet.TransactOpts, _walletId, _tokenAddress, _receiver, _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x8e0cc176.
//
// Solidity: function withdraw(bytes32 _walletId, address _tokenAddress, address _receiver, uint256 _amount) returns()
func (_CelerWallet *CelerWalletTransactorSession) Withdraw(_walletId [32]byte, _tokenAddress common.Address, _receiver common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _CelerWallet.Contract.Withdraw(&_CelerWallet.TransactOpts, _walletId, _tokenAddress, _receiver, _amount)
}

// CelerWalletChangeOperatorIterator is returned from FilterChangeOperator and is used to iterate over the raw logs and unpacked data for ChangeOperator events raised by the CelerWallet contract.
type CelerWalletChangeOperatorIterator struct {
	Event *CelerWalletChangeOperator // Event containing the contract specifics and raw log

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
func (it *CelerWalletChangeOperatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletChangeOperator)
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
		it.Event = new(CelerWalletChangeOperator)
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
func (it *CelerWalletChangeOperatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletChangeOperatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletChangeOperator represents a ChangeOperator event raised by the CelerWallet contract.
type CelerWalletChangeOperator struct {
	WalletId    [32]byte
	OldOperator common.Address
	NewOperator common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChangeOperator is a free log retrieval operation binding the contract event 0x118c3f8030bc3c8254e737a0bd0584403c33646afbcbee8321c3bd5b26543cda.
//
// Solidity: event ChangeOperator(bytes32 indexed walletId, address indexed oldOperator, address indexed newOperator)
func (_CelerWallet *CelerWalletFilterer) FilterChangeOperator(opts *bind.FilterOpts, walletId [][32]byte, oldOperator []common.Address, newOperator []common.Address) (*CelerWalletChangeOperatorIterator, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var oldOperatorRule []interface{}
	for _, oldOperatorItem := range oldOperator {
		oldOperatorRule = append(oldOperatorRule, oldOperatorItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "ChangeOperator", walletIdRule, oldOperatorRule, newOperatorRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletChangeOperatorIterator{contract: _CelerWallet.contract, event: "ChangeOperator", logs: logs, sub: sub}, nil
}

// WatchChangeOperator is a free log subscription operation binding the contract event 0x118c3f8030bc3c8254e737a0bd0584403c33646afbcbee8321c3bd5b26543cda.
//
// Solidity: event ChangeOperator(bytes32 indexed walletId, address indexed oldOperator, address indexed newOperator)
func (_CelerWallet *CelerWalletFilterer) WatchChangeOperator(opts *bind.WatchOpts, sink chan<- *CelerWalletChangeOperator, walletId [][32]byte, oldOperator []common.Address, newOperator []common.Address) (event.Subscription, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var oldOperatorRule []interface{}
	for _, oldOperatorItem := range oldOperator {
		oldOperatorRule = append(oldOperatorRule, oldOperatorItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "ChangeOperator", walletIdRule, oldOperatorRule, newOperatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletChangeOperator)
				if err := _CelerWallet.contract.UnpackLog(event, "ChangeOperator", log); err != nil {
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

// CelerWalletCreateWalletIterator is returned from FilterCreateWallet and is used to iterate over the raw logs and unpacked data for CreateWallet events raised by the CelerWallet contract.
type CelerWalletCreateWalletIterator struct {
	Event *CelerWalletCreateWallet // Event containing the contract specifics and raw log

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
func (it *CelerWalletCreateWalletIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletCreateWallet)
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
		it.Event = new(CelerWalletCreateWallet)
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
func (it *CelerWalletCreateWalletIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletCreateWalletIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletCreateWallet represents a CreateWallet event raised by the CelerWallet contract.
type CelerWalletCreateWallet struct {
	WalletId [32]byte
	Owners   []common.Address
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterCreateWallet is a free log retrieval operation binding the contract event 0xe778e91533ef049a5fc99752bc4efb2b50ca4c967dfc0d4bb4782fb128070c34.
//
// Solidity: event CreateWallet(bytes32 indexed walletId, address[] indexed owners, address indexed operator)
func (_CelerWallet *CelerWalletFilterer) FilterCreateWallet(opts *bind.FilterOpts, walletId [][32]byte, owners [][]common.Address, operator []common.Address) (*CelerWalletCreateWalletIterator, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var ownersRule []interface{}
	for _, ownersItem := range owners {
		ownersRule = append(ownersRule, ownersItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "CreateWallet", walletIdRule, ownersRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletCreateWalletIterator{contract: _CelerWallet.contract, event: "CreateWallet", logs: logs, sub: sub}, nil
}

// WatchCreateWallet is a free log subscription operation binding the contract event 0xe778e91533ef049a5fc99752bc4efb2b50ca4c967dfc0d4bb4782fb128070c34.
//
// Solidity: event CreateWallet(bytes32 indexed walletId, address[] indexed owners, address indexed operator)
func (_CelerWallet *CelerWalletFilterer) WatchCreateWallet(opts *bind.WatchOpts, sink chan<- *CelerWalletCreateWallet, walletId [][32]byte, owners [][]common.Address, operator []common.Address) (event.Subscription, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var ownersRule []interface{}
	for _, ownersItem := range owners {
		ownersRule = append(ownersRule, ownersItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "CreateWallet", walletIdRule, ownersRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletCreateWallet)
				if err := _CelerWallet.contract.UnpackLog(event, "CreateWallet", log); err != nil {
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

// CelerWalletDepositToWalletIterator is returned from FilterDepositToWallet and is used to iterate over the raw logs and unpacked data for DepositToWallet events raised by the CelerWallet contract.
type CelerWalletDepositToWalletIterator struct {
	Event *CelerWalletDepositToWallet // Event containing the contract specifics and raw log

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
func (it *CelerWalletDepositToWalletIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletDepositToWallet)
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
		it.Event = new(CelerWalletDepositToWallet)
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
func (it *CelerWalletDepositToWalletIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletDepositToWalletIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletDepositToWallet represents a DepositToWallet event raised by the CelerWallet contract.
type CelerWalletDepositToWallet struct {
	WalletId     [32]byte
	TokenAddress common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDepositToWallet is a free log retrieval operation binding the contract event 0xbc8e388b96ba8b9f627cb6d72d3513182f763c33c6107ecd31191de1f71abc1a.
//
// Solidity: event DepositToWallet(bytes32 indexed walletId, address indexed tokenAddress, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) FilterDepositToWallet(opts *bind.FilterOpts, walletId [][32]byte, tokenAddress []common.Address) (*CelerWalletDepositToWalletIterator, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "DepositToWallet", walletIdRule, tokenAddressRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletDepositToWalletIterator{contract: _CelerWallet.contract, event: "DepositToWallet", logs: logs, sub: sub}, nil
}

// WatchDepositToWallet is a free log subscription operation binding the contract event 0xbc8e388b96ba8b9f627cb6d72d3513182f763c33c6107ecd31191de1f71abc1a.
//
// Solidity: event DepositToWallet(bytes32 indexed walletId, address indexed tokenAddress, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) WatchDepositToWallet(opts *bind.WatchOpts, sink chan<- *CelerWalletDepositToWallet, walletId [][32]byte, tokenAddress []common.Address) (event.Subscription, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "DepositToWallet", walletIdRule, tokenAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletDepositToWallet)
				if err := _CelerWallet.contract.UnpackLog(event, "DepositToWallet", log); err != nil {
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

// CelerWalletDrainTokenIterator is returned from FilterDrainToken and is used to iterate over the raw logs and unpacked data for DrainToken events raised by the CelerWallet contract.
type CelerWalletDrainTokenIterator struct {
	Event *CelerWalletDrainToken // Event containing the contract specifics and raw log

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
func (it *CelerWalletDrainTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletDrainToken)
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
		it.Event = new(CelerWalletDrainToken)
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
func (it *CelerWalletDrainTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletDrainTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletDrainToken represents a DrainToken event raised by the CelerWallet contract.
type CelerWalletDrainToken struct {
	TokenAddress common.Address
	Receiver     common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDrainToken is a free log retrieval operation binding the contract event 0x896ecb17b26927fb33933fc5f413873193bced3c59fe736c42968a9778bf6b58.
//
// Solidity: event DrainToken(address indexed tokenAddress, address indexed receiver, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) FilterDrainToken(opts *bind.FilterOpts, tokenAddress []common.Address, receiver []common.Address) (*CelerWalletDrainTokenIterator, error) {

	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "DrainToken", tokenAddressRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletDrainTokenIterator{contract: _CelerWallet.contract, event: "DrainToken", logs: logs, sub: sub}, nil
}

// WatchDrainToken is a free log subscription operation binding the contract event 0x896ecb17b26927fb33933fc5f413873193bced3c59fe736c42968a9778bf6b58.
//
// Solidity: event DrainToken(address indexed tokenAddress, address indexed receiver, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) WatchDrainToken(opts *bind.WatchOpts, sink chan<- *CelerWalletDrainToken, tokenAddress []common.Address, receiver []common.Address) (event.Subscription, error) {

	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "DrainToken", tokenAddressRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletDrainToken)
				if err := _CelerWallet.contract.UnpackLog(event, "DrainToken", log); err != nil {
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

// CelerWalletPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the CelerWallet contract.
type CelerWalletPausedIterator struct {
	Event *CelerWalletPaused // Event containing the contract specifics and raw log

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
func (it *CelerWalletPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletPaused)
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
		it.Event = new(CelerWalletPaused)
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
func (it *CelerWalletPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletPaused represents a Paused event raised by the CelerWallet contract.
type CelerWalletPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_CelerWallet *CelerWalletFilterer) FilterPaused(opts *bind.FilterOpts) (*CelerWalletPausedIterator, error) {

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &CelerWalletPausedIterator{contract: _CelerWallet.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_CelerWallet *CelerWalletFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *CelerWalletPaused) (event.Subscription, error) {

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletPaused)
				if err := _CelerWallet.contract.UnpackLog(event, "Paused", log); err != nil {
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

// CelerWalletPauserAddedIterator is returned from FilterPauserAdded and is used to iterate over the raw logs and unpacked data for PauserAdded events raised by the CelerWallet contract.
type CelerWalletPauserAddedIterator struct {
	Event *CelerWalletPauserAdded // Event containing the contract specifics and raw log

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
func (it *CelerWalletPauserAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletPauserAdded)
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
		it.Event = new(CelerWalletPauserAdded)
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
func (it *CelerWalletPauserAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletPauserAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletPauserAdded represents a PauserAdded event raised by the CelerWallet contract.
type CelerWalletPauserAdded struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPauserAdded is a free log retrieval operation binding the contract event 0x6719d08c1888103bea251a4ed56406bd0c3e69723c8a1686e017e7bbe159b6f8.
//
// Solidity: event PauserAdded(address indexed account)
func (_CelerWallet *CelerWalletFilterer) FilterPauserAdded(opts *bind.FilterOpts, account []common.Address) (*CelerWalletPauserAddedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "PauserAdded", accountRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletPauserAddedIterator{contract: _CelerWallet.contract, event: "PauserAdded", logs: logs, sub: sub}, nil
}

// WatchPauserAdded is a free log subscription operation binding the contract event 0x6719d08c1888103bea251a4ed56406bd0c3e69723c8a1686e017e7bbe159b6f8.
//
// Solidity: event PauserAdded(address indexed account)
func (_CelerWallet *CelerWalletFilterer) WatchPauserAdded(opts *bind.WatchOpts, sink chan<- *CelerWalletPauserAdded, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "PauserAdded", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletPauserAdded)
				if err := _CelerWallet.contract.UnpackLog(event, "PauserAdded", log); err != nil {
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

// CelerWalletPauserRemovedIterator is returned from FilterPauserRemoved and is used to iterate over the raw logs and unpacked data for PauserRemoved events raised by the CelerWallet contract.
type CelerWalletPauserRemovedIterator struct {
	Event *CelerWalletPauserRemoved // Event containing the contract specifics and raw log

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
func (it *CelerWalletPauserRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletPauserRemoved)
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
		it.Event = new(CelerWalletPauserRemoved)
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
func (it *CelerWalletPauserRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletPauserRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletPauserRemoved represents a PauserRemoved event raised by the CelerWallet contract.
type CelerWalletPauserRemoved struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPauserRemoved is a free log retrieval operation binding the contract event 0xcd265ebaf09df2871cc7bd4133404a235ba12eff2041bb89d9c714a2621c7c7e.
//
// Solidity: event PauserRemoved(address indexed account)
func (_CelerWallet *CelerWalletFilterer) FilterPauserRemoved(opts *bind.FilterOpts, account []common.Address) (*CelerWalletPauserRemovedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "PauserRemoved", accountRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletPauserRemovedIterator{contract: _CelerWallet.contract, event: "PauserRemoved", logs: logs, sub: sub}, nil
}

// WatchPauserRemoved is a free log subscription operation binding the contract event 0xcd265ebaf09df2871cc7bd4133404a235ba12eff2041bb89d9c714a2621c7c7e.
//
// Solidity: event PauserRemoved(address indexed account)
func (_CelerWallet *CelerWalletFilterer) WatchPauserRemoved(opts *bind.WatchOpts, sink chan<- *CelerWalletPauserRemoved, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "PauserRemoved", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletPauserRemoved)
				if err := _CelerWallet.contract.UnpackLog(event, "PauserRemoved", log); err != nil {
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

// CelerWalletProposeNewOperatorIterator is returned from FilterProposeNewOperator and is used to iterate over the raw logs and unpacked data for ProposeNewOperator events raised by the CelerWallet contract.
type CelerWalletProposeNewOperatorIterator struct {
	Event *CelerWalletProposeNewOperator // Event containing the contract specifics and raw log

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
func (it *CelerWalletProposeNewOperatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletProposeNewOperator)
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
		it.Event = new(CelerWalletProposeNewOperator)
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
func (it *CelerWalletProposeNewOperatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletProposeNewOperatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletProposeNewOperator represents a ProposeNewOperator event raised by the CelerWallet contract.
type CelerWalletProposeNewOperator struct {
	WalletId    [32]byte
	NewOperator common.Address
	Proposer    common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterProposeNewOperator is a free log retrieval operation binding the contract event 0x71f9e7796b33cb192d1670169ee7f4af7c5364f8f01bab4b95466787593745c3.
//
// Solidity: event ProposeNewOperator(bytes32 indexed walletId, address indexed newOperator, address indexed proposer)
func (_CelerWallet *CelerWalletFilterer) FilterProposeNewOperator(opts *bind.FilterOpts, walletId [][32]byte, newOperator []common.Address, proposer []common.Address) (*CelerWalletProposeNewOperatorIterator, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "ProposeNewOperator", walletIdRule, newOperatorRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletProposeNewOperatorIterator{contract: _CelerWallet.contract, event: "ProposeNewOperator", logs: logs, sub: sub}, nil
}

// WatchProposeNewOperator is a free log subscription operation binding the contract event 0x71f9e7796b33cb192d1670169ee7f4af7c5364f8f01bab4b95466787593745c3.
//
// Solidity: event ProposeNewOperator(bytes32 indexed walletId, address indexed newOperator, address indexed proposer)
func (_CelerWallet *CelerWalletFilterer) WatchProposeNewOperator(opts *bind.WatchOpts, sink chan<- *CelerWalletProposeNewOperator, walletId [][32]byte, newOperator []common.Address, proposer []common.Address) (event.Subscription, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "ProposeNewOperator", walletIdRule, newOperatorRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletProposeNewOperator)
				if err := _CelerWallet.contract.UnpackLog(event, "ProposeNewOperator", log); err != nil {
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

// CelerWalletTransferToWalletIterator is returned from FilterTransferToWallet and is used to iterate over the raw logs and unpacked data for TransferToWallet events raised by the CelerWallet contract.
type CelerWalletTransferToWalletIterator struct {
	Event *CelerWalletTransferToWallet // Event containing the contract specifics and raw log

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
func (it *CelerWalletTransferToWalletIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletTransferToWallet)
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
		it.Event = new(CelerWalletTransferToWallet)
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
func (it *CelerWalletTransferToWalletIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletTransferToWalletIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletTransferToWallet represents a TransferToWallet event raised by the CelerWallet contract.
type CelerWalletTransferToWallet struct {
	FromWalletId [32]byte
	ToWalletId   [32]byte
	TokenAddress common.Address
	Receiver     common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTransferToWallet is a free log retrieval operation binding the contract event 0x1b56f805e5edb1e61b0d3f46feffdcbab5e591aa0e70e978ada9fc22093601c8.
//
// Solidity: event TransferToWallet(bytes32 indexed fromWalletId, bytes32 indexed toWalletId, address indexed tokenAddress, address receiver, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) FilterTransferToWallet(opts *bind.FilterOpts, fromWalletId [][32]byte, toWalletId [][32]byte, tokenAddress []common.Address) (*CelerWalletTransferToWalletIterator, error) {

	var fromWalletIdRule []interface{}
	for _, fromWalletIdItem := range fromWalletId {
		fromWalletIdRule = append(fromWalletIdRule, fromWalletIdItem)
	}
	var toWalletIdRule []interface{}
	for _, toWalletIdItem := range toWalletId {
		toWalletIdRule = append(toWalletIdRule, toWalletIdItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "TransferToWallet", fromWalletIdRule, toWalletIdRule, tokenAddressRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletTransferToWalletIterator{contract: _CelerWallet.contract, event: "TransferToWallet", logs: logs, sub: sub}, nil
}

// WatchTransferToWallet is a free log subscription operation binding the contract event 0x1b56f805e5edb1e61b0d3f46feffdcbab5e591aa0e70e978ada9fc22093601c8.
//
// Solidity: event TransferToWallet(bytes32 indexed fromWalletId, bytes32 indexed toWalletId, address indexed tokenAddress, address receiver, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) WatchTransferToWallet(opts *bind.WatchOpts, sink chan<- *CelerWalletTransferToWallet, fromWalletId [][32]byte, toWalletId [][32]byte, tokenAddress []common.Address) (event.Subscription, error) {

	var fromWalletIdRule []interface{}
	for _, fromWalletIdItem := range fromWalletId {
		fromWalletIdRule = append(fromWalletIdRule, fromWalletIdItem)
	}
	var toWalletIdRule []interface{}
	for _, toWalletIdItem := range toWalletId {
		toWalletIdRule = append(toWalletIdRule, toWalletIdItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "TransferToWallet", fromWalletIdRule, toWalletIdRule, tokenAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletTransferToWallet)
				if err := _CelerWallet.contract.UnpackLog(event, "TransferToWallet", log); err != nil {
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

// CelerWalletUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the CelerWallet contract.
type CelerWalletUnpausedIterator struct {
	Event *CelerWalletUnpaused // Event containing the contract specifics and raw log

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
func (it *CelerWalletUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletUnpaused)
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
		it.Event = new(CelerWalletUnpaused)
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
func (it *CelerWalletUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletUnpaused represents a Unpaused event raised by the CelerWallet contract.
type CelerWalletUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_CelerWallet *CelerWalletFilterer) FilterUnpaused(opts *bind.FilterOpts) (*CelerWalletUnpausedIterator, error) {

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &CelerWalletUnpausedIterator{contract: _CelerWallet.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_CelerWallet *CelerWalletFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *CelerWalletUnpaused) (event.Subscription, error) {

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletUnpaused)
				if err := _CelerWallet.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// CelerWalletWithdrawFromWalletIterator is returned from FilterWithdrawFromWallet and is used to iterate over the raw logs and unpacked data for WithdrawFromWallet events raised by the CelerWallet contract.
type CelerWalletWithdrawFromWalletIterator struct {
	Event *CelerWalletWithdrawFromWallet // Event containing the contract specifics and raw log

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
func (it *CelerWalletWithdrawFromWalletIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CelerWalletWithdrawFromWallet)
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
		it.Event = new(CelerWalletWithdrawFromWallet)
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
func (it *CelerWalletWithdrawFromWalletIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CelerWalletWithdrawFromWalletIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CelerWalletWithdrawFromWallet represents a WithdrawFromWallet event raised by the CelerWallet contract.
type CelerWalletWithdrawFromWallet struct {
	WalletId     [32]byte
	TokenAddress common.Address
	Receiver     common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterWithdrawFromWallet is a free log retrieval operation binding the contract event 0xd897e862036b62a0f770979fbd2227f3210565bba2eb4d9acd1dc8ccc00c928b.
//
// Solidity: event WithdrawFromWallet(bytes32 indexed walletId, address indexed tokenAddress, address indexed receiver, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) FilterWithdrawFromWallet(opts *bind.FilterOpts, walletId [][32]byte, tokenAddress []common.Address, receiver []common.Address) (*CelerWalletWithdrawFromWalletIterator, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _CelerWallet.contract.FilterLogs(opts, "WithdrawFromWallet", walletIdRule, tokenAddressRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return &CelerWalletWithdrawFromWalletIterator{contract: _CelerWallet.contract, event: "WithdrawFromWallet", logs: logs, sub: sub}, nil
}

// WatchWithdrawFromWallet is a free log subscription operation binding the contract event 0xd897e862036b62a0f770979fbd2227f3210565bba2eb4d9acd1dc8ccc00c928b.
//
// Solidity: event WithdrawFromWallet(bytes32 indexed walletId, address indexed tokenAddress, address indexed receiver, uint256 amount)
func (_CelerWallet *CelerWalletFilterer) WatchWithdrawFromWallet(opts *bind.WatchOpts, sink chan<- *CelerWalletWithdrawFromWallet, walletId [][32]byte, tokenAddress []common.Address, receiver []common.Address) (event.Subscription, error) {

	var walletIdRule []interface{}
	for _, walletIdItem := range walletId {
		walletIdRule = append(walletIdRule, walletIdItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _CelerWallet.contract.WatchLogs(opts, "WithdrawFromWallet", walletIdRule, tokenAddressRule, receiverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CelerWalletWithdrawFromWallet)
				if err := _CelerWallet.contract.UnpackLog(event, "WithdrawFromWallet", log); err != nil {
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
