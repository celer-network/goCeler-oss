// Copyright 2018-2019 Celer Network

package dispute

import (
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/transactor"
)

type routingTableBuilder interface {
	RemoveEdge(ctype.CidType) error
}

// Processor struct implements the actual disputing logic
type Processor struct {
	nodeConfig     common.GlobalNodeConfig
	transactor     *transactor.Transactor
	transactorPool *transactor.Pool
	rtBuilder      routingTableBuilder
	monitorService intfs.MonitorService
	dal            *storage.DAL
	isOSP          bool
}

// NewProcessor creates a new Disputer struct
func NewProcessor(
	nodeConfig common.GlobalNodeConfig,
	transactor *transactor.Transactor,
	transactorPool *transactor.Pool,
	rtBuilder routingTableBuilder,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	isOSP bool,
) *Processor {
	p := &Processor{
		nodeConfig:     nodeConfig,
		transactor:     transactor,
		transactorPool: transactorPool,
		rtBuilder:      rtBuilder,
		monitorService: monitorService,
		dal:            dal,
		isOSP:          isOSP,
	}
	if isOSP {
		p.monitorPaymentChannelSettleEvent()
		p.monitorNoncooperativeWithdrawEvent()
	}
	return p
}
