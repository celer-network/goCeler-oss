// Copyright 2018-2020 Celer Network
//
// Common data structures used across multiple Celer packages.
// It exists to avoid package dependency cycles.

package structs

import (
	"math/big"
	"time"

	"github.com/celer-network/goCeler/ctype"
)

// LogEventID tracks the position of a watch event in the event log.
type LogEventID struct {
	BlockNumber uint64 // Number of the block containing the event
	Index       int64  // Index of the event within the block
}

// OnChainBalance tracks on chain balance values
type OnChainBalance struct {
	MyDeposit         *big.Int
	MyWithdrawal      *big.Int
	PeerDeposit       *big.Int
	PeerWithdrawal    *big.Int
	PendingWithdrawal *PendingWithdrawal
}

type PendingWithdrawal struct {
	Amount   *big.Int
	Receiver ctype.Addr
	Deadline uint64
}

const (
	DelegatedPayStatus_NULL    int = 0
	DelegatedPayStatus_RECVING int = 1
	DelegatedPayStatus_RECVD   int = 2
	DelegatedPayStatus_SENDING int = 3
	DelegatedPayStatus_DONE    int = 4

	PayState_NULL              int = 0
	PayState_ONESIG_PENDING    int = 1
	PayState_COSIGNED_PENDING  int = 2
	PayState_SECRET_REVEALED   int = 3
	PayState_ONESIG_PAID       int = 4
	PayState_COSIGNED_PAID     int = 5
	PayState_ONESIG_CANCELED   int = 6
	PayState_COSIGNED_CANCELED int = 7
	PayState_NACKED            int = 8
	PayState_INGRESS_REJECTED  int = 9

	DepositState_NULL            int = 0
	DepositState_QUEUED          int = 1
	DepositState_APPROVING_ERC20 int = 2
	DepositState_TX_SUBMITTING   int = 3
	DepositState_TX_SUBMITTED    int = 4
	DepositState_SUCCEEDED       int = 5
	DepositState_FAILED          int = 6

	ChanState_NULL          int = 0
	ChanState_TRUST_OPENED  int = 1
	ChanState_INSTANTIATING int = 2
	ChanState_OPENED        int = 3
	ChanState_SETTLING      int = 4
	ChanState_CLOSED        int = 5

	CrossNetPay_NULL    int = 0
	CrossNetPay_SRC     int = 1
	CrossNetPay_DST     int = 2
	CrossNetPay_INGRESS int = 3
	CrossNetPay_EGRESS  int = 4
)

type DepositJob struct {
	UUID     string
	Cid      ctype.CidType
	ToPeer   bool
	Amount   *big.Int
	Refill   bool
	Deadline time.Time
	State    int
	TxHash   string
	ErrMsg   string
}

type CooperativeWithdrawState int

const (
	CooperativeWithdrawWaitResponse CooperativeWithdrawState = iota
	CooperativeWithdrawWaitTx                                = 1
	CooperativeWithdrawSucceeded                             = 2
	CooperativeWithdrawFailed                                = 3
)

type CooperativeWithdrawJob struct {
	WithdrawHash string
	State        CooperativeWithdrawState
	TxHash       string
	LedgerAddr   ctype.Addr
	Error        string
}

// Edge describes all on-chain channels
type Edge struct {
	P1    ctype.Addr
	P2    ctype.Addr
	Cid   ctype.CidType
	Token ctype.Addr
}
