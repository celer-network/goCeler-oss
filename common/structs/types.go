// Copyright 2018 Celer Network
//
// Common data structures used across multiple Celer packages.
// It exists to avoid package dependency cycles.

package structs

import (
	"math/big"

	"github.com/celer-network/goCeler-oss/ctype"
)

// LogEventID tracks the position of a watch event in the event log.
type LogEventID struct {
	BlockNumber uint64 // Number of the block containing the event
	Index       uint   // Index of the event within the block
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
