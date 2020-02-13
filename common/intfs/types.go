// Copyright 2018 Celer Network
//
// Common interfaces used across multiple Celer packages.
// It exists to avoid package dependency cycles.

package intfs

import (
	"math/big"

	"github.com/celer-network/goCeler-oss/chain"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/ethereum/go-ethereum/core/types"
)

type MonitorService interface {
	GetCurrentBlockNumber() *big.Int
	RegisterDeadline(deadline monitor.Deadline) monitor.CallbackID
	Monitor(
		eventName string,
		contract chain.Contract,
		startBlock *big.Int,
		endBlock *big.Int,
		quickCatch bool,
		reset bool,
		callback func(monitor.CallbackID, types.Log)) (monitor.CallbackID, error)
	MonitorEvent(monitor.Event, bool) (monitor.CallbackID, error)
	RemoveDeadline(id monitor.CallbackID)
	RemoveEvent(id monitor.CallbackID)
}
