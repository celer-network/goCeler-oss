// Copyright 2018-2019 Celer Network

package cooperativewithdraw

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler-oss/chain"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/cnode/jobs"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/common/structs"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/ledgerview"
	"github.com/celer-network/goCeler-oss/metrics"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/transactor"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"
)

type utils interface {
	GetCurrentBlockNumber() *big.Int
}

type Processor struct {
	selfAddress       string
	signer            common.Crypto
	transactorPool    *transactor.Pool
	connectionManager *rpc.ConnectionManager
	monitorService    intfs.MonitorService
	dal               *storage.DAL
	ledger            *chain.BoundContract
	streamWriter      common.StreamWriter
	callbacks         map[string]Callback
	callbacksLock     sync.Mutex
	initLock          sync.Mutex
	runningJobs       map[string]bool
	runningJobsLock   sync.Mutex
	utils             utils
	eventMonitorName  string
	enableJobs        bool
}

func StartProcessor(
	selfAddress string,
	signer common.Crypto,
	transactorPool *transactor.Pool,
	connectionManager *rpc.ConnectionManager,
	monitorService intfs.MonitorService,
	dal *storage.DAL,
	streamWriter common.StreamWriter,
	ledger *chain.BoundContract,
	utils utils,
	keepMonitor bool,
	enableJobs bool) (*Processor, error) {
	p := &Processor{
		selfAddress:       selfAddress,
		signer:            signer,
		transactorPool:    transactorPool,
		connectionManager: connectionManager,
		monitorService:    monitorService,
		dal:               dal,
		streamWriter:      streamWriter,
		ledger:            ledger,
		callbacks:         make(map[string]Callback),
		runningJobs:       make(map[string]bool),
		utils:             utils,
		eventMonitorName: fmt.Sprintf(
			"%s-%s", ctype.Addr2Hex(ledger.GetAddr()), event.CooperativeWithdraw),
		enableJobs: enableJobs,
	}
	if keepMonitor {
		p.monitorEvent()
	} else {
		// Restore event monitoring
		has, err := p.dal.HasEventMonitorBit(p.eventMonitorName)
		if err != nil {
			return nil, err
		}
		if has {
			p.monitorSingleEvent(false)
		}
	}

	return p, nil
}

func (p *Processor) generateWithdrawHash(cid ctype.CidType, seqNum uint64) string {
	seqNumBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(seqNumBytes, seqNum)
	hashBytes :=
		sha3.Sum256(append(p.ledger.GetAddr().Bytes(), append(cid[:], seqNumBytes...)...))
	return ctype.Bytes2Hex(hashBytes[:])
}

func (p *Processor) maybeHandleEvent(eLog *types.Log) bool {
	e := &ledger.CelerLedgerCooperativeWithdraw{}
	err := p.ledger.ParseEvent(event.CooperativeWithdraw, *eLog, e)
	if err != nil {
		log.Error(err)
		return false
	}
	cid := ctype.CidType(e.ChannelId)
	exists, err := p.dal.HasPeer(cid)
	if err != nil {
		log.Error(err)
		return false
	} else if !exists {
		return false
	}
	self := ctype.Hex2Addr(p.selfAddress)
	peerStr, err := p.dal.GetPeer(cid)
	if err != nil {
		log.Error(err)
		return false
	}
	peer := ctype.Hex2Addr(peerStr)
	receiver := e.Receiver
	if receiver != self && receiver != peer {
		return false
	}
	txHash := eLog.TxHash.String()
	log.Debugln("CooperativeWithdraw event txHash", txHash)
	log.Infoln("Caught new CooperativeWithdraw from channel ID", cid.Hex())
	metrics.IncCoopWithdrawEventCnt()
	p.updateOnChainBalance(
		cid,
		ctype.Addr2Hex(self),
		ctype.Addr2Hex(peer),
		e,
		txHash)
	return true
}

func (p *Processor) monitorEvent() {
	_, err := p.monitorService.Monitor(
		event.CooperativeWithdraw,
		p.ledger,
		p.utils.GetCurrentBlockNumber(),
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

func (p *Processor) monitorSingleEvent(reset bool) {
	startBlock := p.utils.GetCurrentBlockNumber()
	duration := new(big.Int)
	duration.SetUint64(config.CooperativeWithdrawTimeout)
	endBlock := new(big.Int).Add(startBlock, duration)
	_, err := p.monitorService.Monitor(
		event.CooperativeWithdraw,
		p.ledger,
		startBlock,
		endBlock,
		false, /* quickCatch */
		reset,
		func(id monitor.CallbackID, eLog types.Log) {
			if p.maybeHandleEvent(&eLog) {
				p.monitorService.RemoveEvent(id)
				p.dal.DeleteEventMonitorBit(p.eventMonitorName)
			}
		})
	if err != nil {
		log.Error(err)
	}
}

func (p *Processor) updateOnChainBalance(
	cid ctype.CidType,
	self string,
	peer string,
	e *ledger.CelerLedgerCooperativeWithdraw,
	txHash string) {
	if len(e.Deposits) != 2 || len(e.Withdrawals) != 2 {
		log.Error("on chain balances length not match")
		return
	}
	var myIndex int
	if strings.Compare(self, peer) < 0 {
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
	if err := p.dal.PutOnChainBalance(cid, onChainBalance); err != nil {
		log.Error(err)
	}

	if !p.enableJobs {
		return
	}

	withdrawHash := p.generateWithdrawHash(cid, e.SeqNum.Uint64())
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
	job.State = jobs.CooperativeWithdrawSucceeded
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
	blkNum := p.utils.GetCurrentBlockNumber().Uint64()
	balance, err := ledgerview.GetBalanceTx(tx, cid, p.selfAddress, blkNum)
	if err != nil {
		log.Error(err)
		return err
	}

	// verify no pending withdraw
	onChainBalance, err := tx.GetOnChainBalance(cid)
	if err != nil {
		log.Error(err)
		return err
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
	if receiver == ctype.Hex2Addr(p.selfAddress) {
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
	err = tx.PutOnChainBalance(cid, onChainBalance)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
