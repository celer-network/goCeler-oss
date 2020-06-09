// Copyright 2018-2020 Celer Network
//
// Common interfaces used across multiple Celer packages.
// It exists to avoid package dependency cycles.

package intfs

import (
	"math/big"

	"github.com/celer-network/goutils/eth/monitor"
	"github.com/ethereum/go-ethereum/core/types"
)

type MonitorService interface {
	GetCurrentBlockNumber() *big.Int
	RegisterDeadline(deadline monitor.Deadline) monitor.CallbackID
	Monitor(cfg *monitor.Config, callback func(monitor.CallbackID, types.Log)) (monitor.CallbackID, error)
	RemoveDeadline(id monitor.CallbackID)
	RemoveEvent(id monitor.CallbackID)
	Close()
}
