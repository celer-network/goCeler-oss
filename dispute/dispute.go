// Copyright 2018-2020 Celer Network

package dispute

import (
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
)

// Processor struct implements the actual disputing logic
type Processor struct {
	nodeConfig      common.GlobalNodeConfig
	transactor      *transactor.Transactor
	transactorPool  *transactor.Pool
	routeController *route.Controller
	monitorService  intfs.MonitorService
	dal             *storage.DAL
	isOSP           bool
}

// NewProcessor creates a new Disputer struct
func NewProcessor(
	nodeConfig common.GlobalNodeConfig,
	transactor *transactor.Transactor,
	transactorPool *transactor.Pool,
	routeController *route.Controller,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	isOSP bool,
) *Processor {
	p := &Processor{
		nodeConfig:      nodeConfig,
		transactor:      transactor,
		transactorPool:  transactorPool,
		routeController: routeController,
		monitorService:  monitorService,
		dal:             dal,
		isOSP:           isOSP,
	}

	if isOSP {
		p.monitorOnAllLedgers()
	}
	return p
}

func (p *Processor) monitorOnAllLedgers() {
	ledgers := p.nodeConfig.GetAllLedgerContracts()

	for _, contract := range ledgers {
		if contract != nil {
			p.monitorPaymentChannelSettleEvent(contract)
			p.monitorNoncooperativeWithdrawEvent(contract)
		}
	}
}
