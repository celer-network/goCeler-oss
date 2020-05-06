// Copyright 2018-2020 Celer Network

package cooperativewithdraw

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"
)

type Processor struct {
	nodeConfig        common.GlobalNodeConfig
	selfAddress       ctype.Addr
	signer            common.Signer
	transactorPool    *transactor.Pool
	connectionManager *rpc.ConnectionManager
	monitorService    intfs.MonitorService
	dal               *storage.DAL
	streamWriter      common.StreamWriter
	callbacks         map[string]Callback
	callbacksLock     sync.Mutex
	initLock          sync.Mutex
	runningJobs       map[string]bool
	runningJobsLock   sync.Mutex
	enableJobs        bool
}

func StartProcessor(
	nodeConfig common.GlobalNodeConfig,
	selfAddress ctype.Addr,
	signer common.Signer,
	transactorPool *transactor.Pool,
	connectionManager *rpc.ConnectionManager,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	streamWriter common.StreamWriter,
	keepMonitor bool,
	enableJobs bool) (*Processor, error) {
	p := &Processor{
		nodeConfig:        nodeConfig,
		selfAddress:       selfAddress,
		signer:            signer,
		transactorPool:    transactorPool,
		connectionManager: connectionManager,
		monitorService:    monitorService,
		dal:               dal,
		streamWriter:      streamWriter,
		callbacks:         make(map[string]Callback),
		runningJobs:       make(map[string]bool),
		enableJobs:        enableJobs,
	}
	if keepMonitor {
		p.monitorOnAllLedgers()
	} else {
		// Restore event monitoring
		ledgerAddrs, err := p.dal.GetMonitorAddrsByEventAndRestart(event.CooperativeWithdraw, true /*restart*/)
		if err != nil {
			return nil, err
		}

		for _, ledgerAddr := range ledgerAddrs {
			contract := p.nodeConfig.GetLedgerContractOn(ledgerAddr)
			if contract != nil {
				p.monitorSingleEvent(contract, false /*reset*/)
			}
		}
	}

	return p, nil
}

func (p *Processor) generateWithdrawHash(cid ctype.CidType, ledgerAddr ctype.Addr, seqNum uint64) string {
	seqNumBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(seqNumBytes, seqNum)
	hashBytes := sha3.Sum256(append(ledgerAddr.Bytes(), append(cid[:], seqNumBytes...)...))
	return ctype.Bytes2Hex(hashBytes[:])
}

func (p *Processor) maybeHandleEvent(eLog *types.Log) bool {
	e := &ledger.CelerLedgerCooperativeWithdraw{}
	err := p.nodeConfig.GetLedgerContract().ParseEvent(event.CooperativeWithdraw, *eLog, e)
	if err != nil {
		log.Error(err)
		return false
	}
	cid := ctype.CidType(e.ChannelId)
	peer, found, err := p.dal.GetChanPeer(cid)
	if err != nil {
		log.Error(err, cid.Hex())
		return false
	}
	if !found {
		return false
	}
	receiver := e.Receiver
	self := p.selfAddress
	if receiver != self && receiver != peer {
		return false
	}
	txHash := eLog.TxHash.String()
	log.Debugln("CooperativeWithdraw event txHash", txHash)
	log.Infoln("Caught new CooperativeWithdraw from channel ID", cid.Hex())
	metrics.IncCoopWithdrawEventCnt()
	p.updateOnChainBalance(
		cid,
		self,
		peer,
		e,
		txHash)
	return true
}

func (p *Processor) monitorOnAllLedgers() {
	ledgers := p.nodeConfig.GetAllLedgerContracts()

	for _, contract := range ledgers {
		if contract != nil {
			p.monitorEvent(contract)
		}
	}
}

func (p *Processor) monitorEvent(ledgerContract chain.Contract) {
	_, err := p.monitorService.Monitor(
		event.CooperativeWithdraw,
		ledgerContract,
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /* endBlock */
		false, /* quickCatch */
		false, /* reset */
		func(id monitor.CallbackID, eLog types.Log) {
			p.maybeHandleEvent(&eLog)
		})
	if err != nil {
		log.Error(err)
	}
}

func (p *Processor) monitorSingleEvent(ledgerContract chain.Contract, reset bool) {
	startBlock := p.monitorService.GetCurrentBlockNumber()
	duration := new(big.Int)
	duration.SetUint64(config.CooperativeWithdrawTimeout)
	endBlock := new(big.Int).Add(startBlock, duration)
	_, err := p.monitorService.Monitor(
		event.CooperativeWithdraw,
		ledgerContract,
		startBlock,
		endBlock,
		false, /* quickCatch */
		reset,
		func(id monitor.CallbackID, eLog types.Log) {
			if p.maybeHandleEvent(&eLog) {
				p.monitorService.RemoveEvent(id)
				p.dal.UpsertMonitorRestart(monitor.NewEventStr(ledgerContract.GetAddr(), event.CooperativeWithdraw), false)
			}
		})
	if err != nil {
		log.Error(err)
	}
}

func (p *Processor) updateOnChainBalance(
	cid ctype.CidType,
	self ctype.Addr,
	peer ctype.Addr,
	e *ledger.CelerLedgerCooperativeWithdraw,
	txHash string) {
	if len(e.Deposits) != 2 || len(e.Withdrawals) != 2 {
		log.Error("on chain balances length not match")
		return
	}
	var myIndex int
	if bytes.Compare(self.Bytes(), peer.Bytes()) < 0 {
		myIndex = 0
	} else {
		myIndex = 1
	}
	onChainBalance := &structs.OnChainBalance{
		MyDeposit:      e.Deposits[myIndex],
		MyWithdrawal:   e.Withdrawals[myIndex],
		PeerDeposit:    e.Deposits[1-myIndex],
		PeerWithdrawal: e.Withdrawals[1-myIndex],
		// overwrite pendingWithrawal with empty struct on withdraw event
	}
	if err := p.dal.UpdateOnChainBalance(cid, onChainBalance); err != nil {
		log.Error(err)
	}

	if !p.enableJobs {
		return
	}

	chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		log.Errorf("Fail to get ledger for channel: %x", cid)
		return
	}
	withdrawHash := p.generateWithdrawHash(cid, chanLedger.GetAddr(), e.SeqNum.Uint64())
	has, err := p.dal.HasCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		log.Error(err)
		return
	} else if !has {
		return
	}
	job, err := p.dal.GetCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot retrieve cooperative withdraw job %s: %s", withdrawHash, err)
		log.Error(errMsg)
		p.maybeFireErrCallbackWithWithdrawHash(withdrawHash, errMsg)
		return
	}
	job.State = structs.CooperativeWithdrawSucceeded
	err = p.dal.PutCooperativeWithdrawJob(withdrawHash, job)
	if err != nil {
		log.Error(err)
		// Even if we just lost persistence, still dispatch the job and try to invoke the callback.
	}
	p.dispatchJob(job)
}

func (p *Processor) checkWithdrawBalanceTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	withdrawInfo := args[1].(*entity.CooperativeWithdrawInfo)
	blkNum := p.monitorService.GetCurrentBlockNumber().Uint64()
	balance, err := ledgerview.GetBalanceTx(tx, cid, p.selfAddress, blkNum)
	if err != nil {
		log.Error(err)
		return err
	}

	// verify no pending withdraw
	onChainBalance, found, err := tx.GetOnChainBalance(cid)
	if err != nil {
		log.Error(err)
		return err
	}
	if !found {
		return common.ErrChannelNotFound
	}
	if blkNum <= onChainBalance.PendingWithdrawal.Deadline+config.WithdrawTimeoutSafeMargin {
		log.Errorln("previous withdraw still pending", onChainBalance.PendingWithdrawal)
		return errors.New("previous withdraw still pending")
	}

	// verify withdraw balance
	withdrawAmt := new(big.Int).SetBytes(withdrawInfo.Withdraw.Amt)
	receiver := ctype.Bytes2Addr(withdrawInfo.GetWithdraw().GetAccount())
	onChainBalance.PendingWithdrawal.Amount = withdrawAmt
	onChainBalance.PendingWithdrawal.Deadline = blkNum + config.CooperativeWithdrawTimeout
	onChainBalance.PendingWithdrawal.Receiver = receiver
	var freeBalance *big.Int
	if receiver == p.selfAddress {
		freeBalance = balance.MyFree
	} else {
		freeBalance = balance.PeerFree
	}
	if freeBalance.Cmp(withdrawAmt) < 0 {
		err2 := fmt.Errorf(
			"Insufficient balance, required %s got %s", withdrawAmt.String(), freeBalance.String())
		log.Error(err2)
		return err2
	}

	// update storage
	err = tx.UpdateOnChainBalance(cid, onChainBalance)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
