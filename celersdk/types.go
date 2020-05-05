// Copyright 2018-2020 Celer Network

package celersdk

import (
	"github.com/celer-network/goCeler/celersdkintf"
)

type ClientCallback interface {
	HandleClientReady(c *Client)
	HandleClientInitErr(e *celersdkintf.E)
	HandleChannelOpened(token, cid string)
	HandleOpenChannelError(token, reason string)
	// Callback triggered when secret revealed.
	HandleRecvStart(pay *celersdkintf.Payment)
	// Callback triggered when pay settle request processed.
	HandleRecvDone(pay *celersdkintf.Payment)
	HandleSendComplete(pay *celersdkintf.Payment)
	HandleSendErr(pay *celersdkintf.Payment, e *celersdkintf.E)
}

type OnchainCallback interface {
	OnSubmitted(uid string)
	OnMined(tx string)
	OnErr(e *celersdkintf.E)
}

type DepositCallback interface {
	OnDeposit(jobID string, txHash string)
	OnError(jobID string, err string)
}

type CooperativeWithdrawCallback interface {
	OnWithdraw(withdrawHash string, txHash string)
	OnError(withdrawHash string, err string)
}

type Account struct {
	Keystore string
	Password string
}

type Deposit struct {
	Myamtwei   string
	Peeramtwei string
}

// CelerStatus defines a struct to store the join status and free balance of a celer endpoint
// For field JoinStatus, it has three values which are 0, 1 and 2. 0 means address queried does not
// join Celer Network. 1 means this address has a channel with Osp responsing this query(Local).
// 2 means this address has a channel with another Osp in Celer Network(Remote).
// For field FreeBalance, it uses a decimal string to represent the receiving capacity of address queried.
// When receiving this status, developers should first check JoinStatus, if it is not 1(Local), you should
// just ignore FreeBalance. Only when JoinStatus is 1 could the developer further use FreeBalance.
type CelerStatus struct {
	JoinStatus  int32
	FreeBalance string
}

// offchain balance struct
type Balance struct {
	Available    string
	Pending      string
	ReceivingCap string
}

type BooleanCondition struct {
	OnChainDeployed     bool
	OnChainAddress      string // onchain contract address if OnChainDeployed is true
	SessionID           string // offchain session hex string from NewAppSession
	ArgsForQueryOutcome []byte
	TimeoutBlockNum     int // timeout of one session. add current block num for pay deadline
}

type Token struct {
	Erctype string // ERC20, ERC721 etc.
	Addr    string // token contract addr
	Symbol  string // short name like gt, celr
}

type TokenType int32

const (
	tokenTypeInvalid TokenType = 0
	tokenTypeEth     TokenType = 1
	tokenTypeErc20   TokenType = 2
)

type TransferLogicType int32

const (
	transferLogicTypeBooleanAnd     TransferLogicType = 0
	transferLogicTypeBooleanOr      TransferLogicType = 1
	transferLogicTypeBooleanCircuit TransferLogicType = 2
	transferLogicTypeNumericAdd     TransferLogicType = 3
	transferLogicTypeNumericMax     TransferLogicType = 4
	transferLogicTypeNumericMin     TransferLogicType = 5
)

type TokenInfo struct {
	TokenType    TokenType
	TokenAddress string
}

type Condition struct {
	// Whether the condition is based on an on-chain deployed contract
	OnChainDeployed bool
	// On-chain contract address bytes if OnChainDeployed is true, or virtual contract address
	ContractAddress []byte
	// Args to isFinalized()
	IsFinalizedArgs []byte
	// Args to getOutcome()
	GetOutcomeArgs []byte
}

type OnChainPaymentInfo struct {
	Amount          string
	ResolveDeadline uint64
}

// UserInfo defines user info to send fiat-related request
type UserInfo struct {
	WalletAddr string
	EmailAddr  string
	Name       string
}
