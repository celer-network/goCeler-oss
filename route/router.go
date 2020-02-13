// Copyright 2018-2019 Celer Network

package route

import (
	"fmt"
	"math/big"
	"time"

	rt "github.com/celer-network/goCeler-oss/chain/channel-eth-go/routerregistry"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/celer-network/goCeler-oss/transactor"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ec "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// RoutingInfoKeeper keeps routers info and provides api to opearate
type RoutingInfoKeeper interface {
	GetOspInfo() map[ec.Address]uint64
	MarkOsp(osp ec.Address, blknum uint64)
	UnmarkOsp(osp ec.Address)
	IsOspExisting(osp ec.Address) bool
}

// RouterProcessor configs to handle onchain router-related event
type RouterProcessor struct {
	nodeConfig          common.GlobalNodeConfig
	transactor          *transactor.Transactor
	monitorService      intfs.MonitorService
	keeper              RoutingInfoKeeper
	isRegisteredOnchain bool // If Osp has been registered in contract, no need to add Mutex when operating since it would only be changed by onchain events
	blockDelay          uint64
	adminWebHostAndPort string
}

// Enum corrsponding to the onchain router operation
const (
	routerAdded uint8 = iota
	routerRemoved
	routerRefreshed
)

// Time duration for tickers
const (
	refreshRouterInterval = 5 * 24 * time.Hour
	scanRouterInterval    = 1 * time.Hour
)

const expireIntervalBlock = uint64(46500) // estimation of block numbers during one week, fluctuation tolerant
const refreshThreshold = uint64(10000)    // check if Osp needs to refresh when starting

// NewRouterProcessor creates a new process for routing
func NewRouterProcessor(
	nodeConfig common.GlobalNodeConfig,
	transactor *transactor.Transactor,
	monitorService intfs.MonitorService,
	keeper RoutingInfoKeeper,
	blockDelay uint64,
	adminWebHostAndPort string,
) *RouterProcessor {
	p := &RouterProcessor{
		nodeConfig:          nodeConfig,
		transactor:          transactor,
		monitorService:      monitorService,
		keeper:              keeper,
		isRegisteredOnchain: false,
		blockDelay:          blockDelay,
		adminWebHostAndPort: adminWebHostAndPort,
	}

	return p
}

// Start starts router process to instantiate Osp as a router.
func (p *RouterProcessor) Start() {
	// check if Osp is registered in contract
	blk, err := p.SendRouterInfoCall()
	if err != nil {
		log.Errorf("RouterInfo call failed: %s", err)
	} else if blkNum := blk.Uint64(); blkNum != 0 {
		p.isRegisteredOnchain = true
		// check if Osp needs to send refresh transaction
		currentBlk := p.monitorService.GetCurrentBlockNumber().Uint64()
		if currentBlk-blkNum > refreshThreshold {
			p.SendRefreshRouterTransaction()
		}
	}

	// start onchain events monitor
	p.monitorRouterUpdatedEvent()

	// start routine job
	go p.runRoutersRoutineJob()
}

// monitors the RouterAdded event onchain
// backtrack from one interval before the current block
func (p *RouterProcessor) monitorRouterUpdatedEvent() {
	startBlk := p.calculateStartBlockNumber()
	_, err := p.monitorService.Monitor(
		event.RouterUpdated,
		p.nodeConfig.GetRouterRegistryContract(),
		startBlk,
		nil,   // endBlock
		false, // quickCatch
		true,  // reset
		func(id monitor.CallbackID, eLog types.Log) {
			// ignore events if Osp does not register as a router
			if !p.isRegisteredOnchain {
				return
			}
			e := &rt.RouterRegistryRouterUpdated{} // event RouterAdded
			if err := p.nodeConfig.GetRouterRegistryContract().ParseEvent(event.RouterUpdated, eLog, e); err != nil {
				log.Error(err)
				return
			}

			// Only used in log
			routerAddr := ctype.Addr2Hex(e.RouterAddress)
			txHash := fmt.Sprintf("%x", eLog.TxHash)
			log.Infoln("Seeing RouterUpdated event, router addr:", routerAddr, "tx hash:", txHash, "callback id:", id, "blkNum:", eLog.BlockNumber)

			p.processRouterUpdatedEvent(e, eLog.BlockNumber)
			if p.adminWebHostAndPort != "" {
				postErr := utils.RequestBuildRoutingTable(p.adminWebHostAndPort)
				if postErr != nil {
					log.Error(postErr)
				}
			}
		},
	)

	if err != nil {
		log.Error(err)
	}
}

// processes the RouterUpdated event according to various router opeartion
func (p *RouterProcessor) processRouterUpdatedEvent(e *rt.RouterRegistryRouterUpdated, blkNum uint64) {
	switch e.Op {
	case routerAdded:
		p.addRouter(e.RouterAddress, blkNum)
	case routerRemoved:
		p.removeRouter(e.RouterAddress)
	case routerRefreshed:
		p.refreshRouter(e.RouterAddress, blkNum)
	default:
		log.Infof("Unknown router operation from router registry contract: %v", e.Op)
	}
}

// adds router node and record the block number
func (p *RouterProcessor) addRouter(routerAddr ec.Address, blkNum uint64) {
	p.keeper.MarkOsp(routerAddr, blkNum)
}

// removes router node and delete it from the map
func (p *RouterProcessor) removeRouter(routerAddr ec.Address) {
	if !p.keeper.IsOspExisting(routerAddr) {
		return
	}

	p.keeper.UnmarkOsp(routerAddr)
}

// refreshes a router node and update block number in the map
func (p *RouterProcessor) refreshRouter(routerAddr ec.Address, blkNum uint64) {
	p.keeper.MarkOsp(routerAddr, blkNum)
}

// calculates the start block number for event monitor service.
// No matter whether Osp starts from scratch or starts from existing database,
// Osp only backtracks one interval back from the current block number.
// Interval is the same as the expire interval in rtconfig
func (p *RouterProcessor) calculateStartBlockNumber() *big.Int {
	currentBlk := p.monitorService.GetCurrentBlockNumber()
	blkDelay := new(big.Int).SetUint64(p.blockDelay)

	// take into delay block consideration
	currentBlk.Sub(currentBlk, blkDelay)

	interval := big.NewInt(0).SetUint64(expireIntervalBlock)
	if interval.Cmp(currentBlk) == 1 {
		return big.NewInt(0)
	}
	return currentBlk.Sub(currentBlk, interval) // start block number for onchain monitor service
}

// SendRouterInfoCall calls routerInfo in router registry contract to check if Osp has been registered.
// Return value is the block number corresponding to Osp address
func (p *RouterProcessor) SendRouterInfoCall() (*big.Int, error) {
	log.Infoln("Sending RouterInfo call", ctype.Addr2Hex(p.nodeConfig.GetRouterRegistryContract().GetAddr()), p.transactor.ContractCaller() != nil)
	caller, err := rt.NewRouterRegistryCaller(p.nodeConfig.GetRouterRegistryContract().GetAddr(), p.transactor.ContractCaller())
	if err != nil {
		log.Errorf("Fail to create router registry caller: %s", err)
		return nil, err
	}

	selfAddr := p.transactor.Address()
	return caller.RouterInfo(&bind.CallOpts{}, selfAddr)
}

// SendRegisterRouterTransaction sends transaction to register as a router
func (p *RouterProcessor) SendRegisterRouterTransaction(txValue *big.Int) {
	selfAddr := p.transactor.Address()

	log.Infoln("Sending RegisterRouter tx")
	// If router node already exists, there is no need to send transaction
	if !p.keeper.IsOspExisting(selfAddr) {
		return
	}

	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				txHash := receipt.TxHash
				if receipt.Status == types.ReceiptStatusSuccessful {
					log.Debugf("RegisterRouter transaction 0x%x succeeded", txHash)
				} else {
					log.Errorf("RegisterRouter transaction 0x%x failed", txHash)
				}
			},
		},
		txValue,
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err := rt.NewRouterRegistryTransactor(p.nodeConfig.GetRouterRegistryContract().GetAddr(), transactor)
			if err != nil {
				log.Errorf("Fail to create router registry transactor when registering: %s", err)
				return nil, err
			}

			return contract.RegisterRouter(opts)
		},
	)
	if err != nil {
		log.Errorf("Fail to register as a router: %s", err)
	}
}

// SendDeregisterRouterTransaction sends transaction to deregister as a router
func (p *RouterProcessor) SendDeregisterRouterTransaction() {
	selfAddr := p.transactor.Address()

	log.Infoln("Sending DeregisterRouter tx")
	// If an Osp does not want to be a router anymore, it should remove itself
	// no matter if it could send the deregister transaction.
	p.removeRouter(selfAddr)

	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				txHash := receipt.TxHash
				if receipt.Status == types.ReceiptStatusSuccessful {
					log.Debugf("DeregisterRouter transaction 0x%x succeeded", txHash)
				} else {
					log.Errorf("DeregisterRouter transaction 0x%x failed", txHash)
				}
			},
		},
		big.NewInt(0), // value
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err := rt.NewRouterRegistryTransactor(p.nodeConfig.GetRouterRegistryContract().GetAddr(), transactor)
			if err != nil {
				log.Errorf("Fail to create router registry transactor when deregistering: %s", err)
				return nil, err
			}

			return contract.DeregisterRouter(opts)
		},
	)
	if err != nil {
		log.Errorf("Fail to deregister router's registry: %s", err)
	}
}

// SendRefreshRouterTransaction sends transaction to refresh the block number of Osp address.
// CAUTION: need to pay attention if it fails to refresh
func (p *RouterProcessor) SendRefreshRouterTransaction() {
	log.Infoln("Sending RefreshRouter tx")

	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				txHash := receipt.TxHash
				if receipt.Status == types.ReceiptStatusSuccessful {
					log.Debugf("RefreshRouter transaction 0x%x succeeded", txHash)
				} else {
					log.Errorf("RefreshRouter transaction 0x%x failed", txHash)
				}
			},
		},
		big.NewInt(0), // value
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err := rt.NewRouterRegistryTransactor(p.nodeConfig.GetRouterRegistryContract().GetAddr(), transactor)
			if err != nil {
				log.Errorf("Fail to crate router registry transactor when refreshing: %s", err)
			}

			return contract.RefreshRouter(opts)
		},
	)
	if err != nil {
		log.Errorf("Fail to refresh the router: %s", err)
	}
}

// starts some routine jobs
// CAUTION: This should be run in goroutine
func (p *RouterProcessor) runRoutersRoutineJob() {
	scanTicker := time.NewTicker(scanRouterInterval)
	refreshTicker := time.NewTicker(refreshRouterInterval)
	defer func() {
		scanTicker.Stop()
		refreshTicker.Stop()
	}()

	for {
		select {
		case <-scanTicker.C:
			p.removeExpiredRouters()
		case <-refreshTicker.C:
			// Do not refresh if Osp is not yet a router
			if !p.isRegisteredOnchain {
				continue
			}
			p.SendRefreshRouterTransaction()
		}
	}
}

// Traverses the map and remove the expired routers.
func (p *RouterProcessor) removeExpiredRouters() {
	if !p.isRegisteredOnchain {
		return
	}
	currentBlk := p.monitorService.GetCurrentBlockNumber().Uint64()

	ospInfo := p.keeper.GetOspInfo()
	for addr := range ospInfo {
		blk := ospInfo[addr]

		if isRouterExpired(blk, currentBlk) {
			p.keeper.UnmarkOsp(addr)
		}
	}
}

func isRouterExpired(routerBlk, currentBlk uint64) bool {
	return routerBlk+expireIntervalBlock < currentBlk
}
