// Copyright 2019-2020 Celer Network

package route

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	rt "github.com/celer-network/goCeler/chain/channel-eth-go/routerregistry"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/route/ospreport"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
)

type BcastSendCallback func(info *rpc.RoutingRequest, ospAddrs []string)

// Controller configs to handle onchain router-related event
type Controller struct {
	nodeConfig        common.GlobalNodeConfig
	transactor        *transactor.Transactor
	monitorService    intfs.MonitorService
	dal               *storage.DAL
	signer            common.Signer
	bcastSendCallback BcastSendCallback
	rtBuilder         *routingTableBuilder
	explorerReport    *ospreport.OspInfo
	explorerUrl       string // explorer url

	// Dynamic routing updates from OSPs are gathered here then
	// used for recomputing the routing table.
	routingBatch     map[ctype.Addr]*rpc.RoutingUpdate
	routingBatchLock sync.Mutex
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
	routeTTL              = 15
)

const expireIntervalBlock = uint64(46500) // estimation of block numbers during one week, fluctuation tolerant
const refreshThreshold = uint64(10000)    // check if Osp needs to refresh when starting

// NewController creates a new process for router controller
func NewController(
	nodeConfig common.GlobalNodeConfig,
	transactor *transactor.Transactor,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	signer common.Signer,
	bcastSendCallback BcastSendCallback,
	routingData []byte,
	rpcHost string,
	explorerUrl string) (*Controller, error) {
	c := &Controller{
		nodeConfig:        nodeConfig,
		transactor:        transactor,
		monitorService:    monitorService,
		dal:               dal,
		signer:            signer,
		bcastSendCallback: bcastSendCallback,
		explorerUrl:       explorerUrl,
	}
	c.rtBuilder = newRoutingTableBuilder(nodeConfig.GetOnChainAddr(), dal)
	if c.rtBuilder == nil {
		return c, fmt.Errorf("fail to initialize routing table builder")
	}
	c.explorerReport = &ospreport.OspInfo{
		EthAddr:    nodeConfig.GetOnChainAddr().Hex(), // format required by explorer
		RpcHost:    rpcHost,
		OpenAccept: true,
	}
	err := c.startRoutingRecoverProcess(monitorService.GetCurrentBlockNumber(), routingData, nodeConfig)
	return c, err
}

// Start starts router process to instantiate OSP as a router.
func (c *Controller) Start() {
	// check if OSP is registered on-chain as a router
	blknum, err := c.queryRouterRegistry()
	if err != nil {
		log.Errorf("query router registry failed: %s", err)
		return
	}
	if blknum != 0 {
		log.Infoln("router registered / refreshed at block", blknum)
		// check if OSP needs to send refresh transaction
		currentBlk := c.monitorService.GetCurrentBlockNumber().Uint64()
		if currentBlk-blknum > refreshThreshold {
			c.refreshRouterRegistry()
		}
		// start onchain events monitor
		c.monitorRouterUpdatedEvent()
		// start routine job
		go c.runRoutersRoutineJob()
	} else {
		log.Warn("NOT able to join the OSP network because this node is not registered on-chain as a router")
	}
}

// monitors the RouterUpdated event onchain
// backtrack from one interval before the current block
func (c *Controller) monitorRouterUpdatedEvent() {
	startBlk := c.calculateStartBlockNumber()
	_, err := c.monitorService.Monitor(
		event.RouterUpdated,
		c.nodeConfig.GetRouterRegistryContract(),
		startBlk,
		nil,   // endBlock
		false, // quickCatch
		true,  // reset
		func(id monitor.CallbackID, eLog types.Log) {
			e := &rt.RouterRegistryRouterUpdated{} // event RouterUpdated
			if err := c.nodeConfig.GetRouterRegistryContract().ParseEvent(event.RouterUpdated, eLog, e); err != nil {
				log.Error(err)
				return
			}

			// Only used in log
			routerAddr := ctype.Addr2Hex(e.RouterAddress)
			txHash := fmt.Sprintf("%x", eLog.TxHash)
			log.Infoln("Seeing RouterUpdated event, router addr:", routerAddr, "tx hash:", txHash, "callback id:", id, "blkNum:", eLog.BlockNumber)

			c.processRouterUpdatedEvent(e, eLog.BlockNumber)
		},
	)

	if err != nil {
		log.Error(err)
	}
}

// processes the RouterUpdated event according to various router opeartion
func (c *Controller) processRouterUpdatedEvent(e *rt.RouterRegistryRouterUpdated, blkNum uint64) {
	switch e.Op {
	case routerAdded:
		c.addRouter(e.RouterAddress, blkNum)
	case routerRemoved:
		c.removeRouter(e.RouterAddress)
	case routerRefreshed:
		c.refreshRouter(e.RouterAddress, blkNum)
	default:
		log.Warnf("Unknown router operation from router registry contract: %v", e.Op)
	}
}

// adds router node and record the block number
func (c *Controller) addRouter(routerAddr ctype.Addr, blkNum uint64) {
	c.rtBuilder.markOsp(routerAddr, blkNum)
}

// removes router node and delete it from the map
func (c *Controller) removeRouter(routerAddr ctype.Addr) {
	if !c.rtBuilder.hasOsp(routerAddr) {
		return
	}
	c.rtBuilder.unmarkOsp(routerAddr)
}

// refreshes a router node and update block number in the map
func (c *Controller) refreshRouter(routerAddr ctype.Addr, blkNum uint64) {
	c.rtBuilder.markOsp(routerAddr, blkNum)
}

// calculates the start block number for event monitor service.
// No matter whether Osp starts from scratch or starts from existing database,
// Osp only backtracks one interval back from the current block number.
// Interval is the same as the expire interval in rtconfig
func (c *Controller) calculateStartBlockNumber() *big.Int {
	currentBlk := c.monitorService.GetCurrentBlockNumber()
	interval := big.NewInt(0).SetUint64(expireIntervalBlock)
	if interval.Cmp(currentBlk) == 1 {
		return big.NewInt(0)
	}
	return currentBlk.Sub(currentBlk, interval) // start block number for onchain monitor service
}

// call routerInfo in router registry contract to check if Osp has been registered.
// Return value is the block number corresponding to Osp address
func (c *Controller) queryRouterRegistry() (uint64, error) {
	routerRegistryAddr := c.nodeConfig.GetRouterRegistryContract().GetAddr()
	caller, err := rt.NewRouterRegistryCaller(routerRegistryAddr, c.transactor.ContractCaller())
	if err != nil {
		return 0, err
	}
	blknum, err := caller.RouterInfo(&bind.CallOpts{}, c.transactor.Address())
	if err != nil {
		return 0, err
	}
	return blknum.Uint64(), nil
}

// send on-chain transaction to refresh the block number of Osp address.
// CAUTION: need to pay attention if it fails to refresh
func (c *Controller) refreshRouterRegistry() {
	log.Infoln("sending RefreshRouter tx")
	routerRegistryAddr := c.nodeConfig.GetRouterRegistryContract().GetAddr()
	_, err := c.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				if receipt.Status == types.ReceiptStatusSuccessful {
					log.Infof("RefreshRouter transaction %x succeeded", receipt.TxHash)
				} else {
					log.Errorf("RefreshRouter transaction %x failed", receipt.TxHash)
				}
			},
		},
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := rt.NewRouterRegistryTransactor(routerRegistryAddr, transactor)
			if err2 != nil {
				log.Errorln("NewRouterRegistryTransactor err:", err2)
				return nil, err2
			}
			return contract.RefreshRouter(opts)
		})
	if err != nil {
		log.Errorf("Fail to refresh the router: %s", err)
	}
}

// starts some routine jobs
// CAUTION: This should be run in goroutine
func (c *Controller) runRoutersRoutineJob() {
	scanTicker := time.NewTicker(scanRouterInterval)
	refreshTicker := time.NewTicker(refreshRouterInterval)
	bcastTicker := time.NewTicker(config.RouterBcastInterval)
	buildTicker := time.NewTicker(config.RouterBuildInterval)
	reportTicker := time.NewTicker(config.OspReportInverval)
	defer func() {
		scanTicker.Stop()
		refreshTicker.Stop()
		bcastTicker.Stop()
		buildTicker.Stop()
		reportTicker.Stop()
	}()

	for {
		select {
		case <-scanTicker.C:
			c.removeExpiredRouters()
		case <-refreshTicker.C:
			c.refreshRouterRegistry()
		case <-bcastTicker.C:
			c.bcastRouterInfo()
		case <-buildTicker.C:
			c.buildRoutingTable()
		case <-reportTicker.C:
			c.reportOspInfoToExplorer()
		}
	}
}

// Traverses the map and remove the expired routers.
func (c *Controller) removeExpiredRouters() {
	currentBlk := c.monitorService.GetCurrentBlockNumber().Uint64()
	ospInfo := c.rtBuilder.getAllOsps()
	for addr := range ospInfo {
		blk := ospInfo[addr].RegistryBlock

		if isRouterExpired(blk, currentBlk) {
			c.rtBuilder.unmarkOsp(addr)
		}
	}
}

func isRouterExpired(routerBlk, currentBlk uint64) bool {
	return routerBlk+expireIntervalBlock < currentBlk
}

// Get my dynamic routing information and broadcast it to peer OSPs.
// Also enqueue to the routing info batch to include it in the next
// routing recomputation.
func (c *Controller) bcastRouterInfo() {
	channels := c.gatherChannelInfo()
	myAddr := ctype.Addr2Hex(c.nodeConfig.GetOnChainAddr())
	update := &rpc.RoutingUpdate{
		Origin:   myAddr,
		Ts:       uint64(now().Unix()),
		Channels: channels,
	}

	updateBytes, err := proto.Marshal(update)
	if err != nil {
		log.Errorln("proto marshal signedUpdate err", err, update)
		return
	}
	sig, err := c.signer.SignEthMessage(updateBytes)

	signedUpdate := &rpc.SignedRoutingUpdate{
		Update: updateBytes,
		Sig:    sig,
		Ttl:    routeTTL,
	}

	info := &rpc.RoutingRequest{
		Updates: []*rpc.SignedRoutingUpdate{signedUpdate},
	}

	c.enqueueRouterInfo(update, signedUpdate.GetTtl())

	c.bcast(info, []string{myAddr}, "")
}

func (c *Controller) gatherChannelInfo() []*rpc.ChannelRoutingInfo {
	var channels []*rpc.ChannelRoutingInfo
	blkNum := c.monitorService.GetCurrentBlockNumber().Uint64()
	for _, neighbor := range c.rtBuilder.getAliveNeighbors() {
		for _, cid := range neighbor.TokenCids {
			bal, err := ledgerview.GetBalance(c.dal, cid, c.nodeConfig.GetOnChainAddr(), blkNum)
			if err != nil {
				log.Error(err)
				continue
			}
			channel := &rpc.ChannelRoutingInfo{
				Cid:     ctype.Cid2Hex(cid),
				Balance: bal.MyFree.String(),
			}
			channels = append(channels, channel)
		}
	}
	return channels
}

// Enqueue the dynamic routing information and return true if it should be
// propagated to peer OSPs in the broadcast.  The information is propagated
// if it's new to this OSP and still has time-to-live (hop counter).
//
// If this is the first information in the batch, trigger a delayed action
// to recompute the routing table, giving it some time for more routing info
// to be added to the batch.
func (c *Controller) enqueueRouterInfo(update *rpc.RoutingUpdate, ttl uint64) bool {
	if update == nil {
		return false
	}

	origin := ctype.Hex2Addr(update.GetOrigin())
	ts := update.GetTs()
	if ttl <= 0 {
		return false // this should not happen
	}

	c.routingBatchLock.Lock()
	defer c.routingBatchLock.Unlock()

	if c.routingBatch == nil {
		c.routingBatch = make(map[ctype.Addr]*rpc.RoutingUpdate)
	}

	oldUpdate, ok := c.routingBatch[origin]
	if ok && oldUpdate.GetTs() >= ts {
		return false // already have newer info from this origin
	}
	c.rtBuilder.keepOspAlive(origin, ts)

	c.routingBatch[origin] = update
	// Propagate the info if the incoming TTL was more than 1.
	return (ttl > 1)
}

func (c *Controller) buildRoutingTable() {
	for token := range c.rtBuilder.getAllTokens() {
		c.rtBuilder.buildTable(token)
	}

	// TODO: Recompute the routing table according to bcast info.
	/*
		// Grab the current batch of routing info and release the lock.
		c.routingBatchLock.Lock()
		batch := c.routingBatch
		c.routingBatch = nil
		c.routingBatchLock.Unlock()
		log.Debugf("computing routing table from %d OSP info", len(batch))
	*/
}

// New routing information arrived from another OSP. Enqueue it for
// a future route recomputation and, if needed, forward it to other
// peer OSPs in the broadcast.
func (c *Controller) RecvBcastRoutingInfo(info *rpc.RoutingRequest) error {
	// TODO: support batch updates
	if len(info.GetUpdates()) != 1 {
		return fmt.Errorf("invalid number of routing updates in one request, %d", len(info.GetUpdates()))
	}
	signedUpdate := info.Updates[0]
	var update rpc.RoutingUpdate
	err := proto.Unmarshal(signedUpdate.GetUpdate(), &update)
	if err != nil {
		return fmt.Errorf("unmarshal signed update err: %w", err)
	}
	if !utils.SigIsValid(ctype.Hex2Addr(update.GetOrigin()), signedUpdate.GetUpdate(), signedUpdate.GetSig()) {
		return fmt.Errorf("route update invalid sig for origin %s", update.GetOrigin())
	}

	log.Debugf("Receive router updates: origin:%s, sender:%s", update.GetOrigin(), info.GetSender())

	c.rtBuilder.keepNeighborAlive(ctype.Hex2Addr(info.GetSender()))
	if c.enqueueRouterInfo(&update, signedUpdate.GetTtl()) {
		info.Updates[0].Ttl--
		c.bcast(info, []string{update.GetOrigin()}, info.GetSender())
	}

	return nil
}

// Send out the given routing information request to the peer OSPs
// excluding the direct sender of this message (if any).
func (c *Controller) bcast(info *rpc.RoutingRequest, origins []string, sender string) {
	// Get peer OSPs excluding me and the given direct sender.
	var ospAddrs []string
	// TODO: support batch updates
	origin := origins[0]
	myAddr := ctype.Addr2Hex(c.nodeConfig.GetOnChainAddr())
	neighborAddrs := c.rtBuilder.getNeighborAddrs()
	for _, ospAddr := range neighborAddrs {
		ospAddrStr := ctype.Addr2Hex(ospAddr)
		if ospAddrStr == myAddr || ospAddrStr == sender {
			continue
		}
		if origin == ospAddrStr {
			continue
		}
		ospAddrs = append(ospAddrs, ospAddrStr)
	}
	if len(ospAddrs) == 0 {
		return
	}
	info.Sender = myAddr
	log.Debugf("bcast router updates: origin %s, to %s", origin, ospAddrs)
	c.bcastSendCallback(info, ospAddrs)
}

func (c *Controller) reportOspInfoToExplorer() {
	if c.explorerUrl == "" {
		return
	}
	// set osp peers
	c.explorerReport.OspPeers = nil
	blkNum := c.monitorService.GetCurrentBlockNumber().Uint64()
	for addr, neighbor := range c.rtBuilder.getAliveNeighbors() {
		peerBalances := &ospreport.PeerBalances{
			Peer: addr.Hex(), // format required by explorer
		}
		for tk, cid := range neighbor.TokenCids {
			bal, err := ledgerview.GetBalance(c.dal, cid, c.nodeConfig.GetOnChainAddr(), blkNum)
			if err != nil {
				log.Error(err)
				continue
			}
			peerBalances.Balances = append(
				peerBalances.Balances,
				&ospreport.ChannelBalance{
					Cid:         cid.Hex(), // format required by explorer
					TokenAddr:   tk.Hex(),  // format required by explorer
					SelfBalance: bal.MyFree.String(),
					PeerBalance: bal.PeerFree.String(),
				})
		}
		c.explorerReport.OspPeers = append(c.explorerReport.OspPeers, peerBalances)
	}
	// set std openchan configs
	c.explorerReport.StdOpenchanConfigs = nil
	for _, cfg := range rtconfig.GetStandardConfigs().GetConfig() {
		if cfg != nil && cfg.Token != nil {
			cfgReport := &ospreport.StdOpenChanConfig{
				TokenAddr:  ctype.Hex2Addr(cfg.Token.Address).Hex(), // format required by explorer
				MinDeposit: cfg.MinDeposit,
				MaxDeposit: cfg.MaxDeposit,
			}
			c.explorerReport.StdOpenchanConfigs = append(c.explorerReport.StdOpenchanConfigs, cfgReport)
		}
	}
	// set pay count
	payCount, err := c.dal.CountPayments()
	if err != nil {
		log.Error("CountPayments err:", err)
	}
	c.explorerReport.Payments = int64(payCount)
	// set timestamp
	c.explorerReport.Timestamp = uint64(now().Unix())
	// marshal and sign
	reportBytes, err := proto.Marshal(c.explorerReport)
	if err != nil {
		log.Errorln("proto marshal OSP report err:", err, c.explorerReport)
		return
	}
	sig, err := c.signer.SignEthMessage(reportBytes)
	if err != nil {
		log.Error(err)
		return
	}
	// send report
	report := map[string]string{
		"ospInfo": ctype.Bytes2Hex(reportBytes),
		"sig":     ctype.Bytes2Hex(sig),
	}
	_, err = utils.HttpPost(c.explorerUrl, report)
	if err != nil {
		log.Warnln("explorer report error:", err)
	}
}

func (c *Controller) BuildTable(tokenAddr ctype.Addr) (map[ctype.Addr]ctype.CidType, error) {
	return c.rtBuilder.buildTable(tokenAddr)
}

func (c *Controller) AddEdge(p1 ctype.Addr, p2 ctype.Addr, cid ctype.CidType, tokenAddr ctype.Addr) error {
	return c.rtBuilder.addEdge(p1, p2, cid, tokenAddr)
}

func (c *Controller) RemoveEdge(cid ctype.CidType) error {
	return c.rtBuilder.removeEdge(cid)
}

func (c *Controller) GetAllNeighbors() map[ctype.Addr]*NeighborInfo {
	return c.rtBuilder.getAllNeighbors()
}

func now() time.Time {
	return time.Now().UTC()
}
