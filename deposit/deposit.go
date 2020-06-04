// Copyright 2018-2020 Celer Network

package deposit

import (
	"fmt"
	"sync"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/core/types"
)

type DepositCallback interface {
	OnDeposit(jobID string, txHash string)
	OnError(jobID string, err string)
}

type Processor struct {
	nodeConfig     common.GlobalNodeConfig
	transactor     *eth.Transactor
	dal            *storage.DAL
	monitorService intfs.MonitorService
	isOSP          bool // server mode (true) or client mode (false)

	// fields for client mode
	callbacks       map[string]DepositCallback
	callbacksLock   sync.Mutex
	runningJobs     map[string]bool
	runningJobsLock sync.Mutex

	// fields for server mode
	// map(ledgerAddr -> list of channelDeposit)
	chanDeposits map[ctype.Addr][]*channelDeposit
	// channel deposits that were submitting but not recorded before last shutdown
	// map(tx batch time -> map(cid:topeer -> channelDeposit))
	unrecordedDeposits     map[string]map[string]*channelDeposit
	unrecordedDepositsLock sync.Mutex
	lastAlertTime          time.Time
}

func StartProcessor(
	nodeConfig common.GlobalNodeConfig,
	transactor *eth.Transactor,
	dal *storage.DAL,
	monitorService intfs.MonitorService,
	isOSP bool,
	isEventListener bool,
	quit chan bool) (*Processor, error) {
	p := &Processor{
		nodeConfig:     nodeConfig,
		transactor:     transactor,
		dal:            dal,
		monitorService: monitorService,
		isOSP:          isOSP,
		callbacks:      make(map[string]DepositCallback),
		runningJobs:    make(map[string]bool),
		chanDeposits:   make(map[ctype.Addr][]*channelDeposit),
	}
	if isOSP && isEventListener {
		err := p.resumeServerJobs()
		if err != nil {
			return nil, err
		}
		p.lastAlertTime = now()
		go p.serverDepositJobPolling(quit)
	} else {
		err := p.resumeClientJobs()
		if err != nil {
			return nil, err
		}
	}
	go p.monitorOnAllLedgers()
	return p, nil
}

func (p *Processor) GetDepositState(jobID string) (int, string, error) {
	state, msg, found, err := p.dal.GetDepositState(jobID)
	if err != nil {
		metrics.IncDepositErrCnt()
		return structs.DepositState_NULL, "", fmt.Errorf("GetDepositState err: %w", err)
	}
	if !found {
		return structs.DepositState_NULL, "", common.ErrDepositNotFound
	}
	return state, msg, nil
}

func (p *Processor) monitorOnAllLedgers() {
	ledgers := p.nodeConfig.GetAllLedgerContracts()

	for _, contract := range ledgers {
		if contract != nil {
			p.monitorEvent(contract)
		}
	}
}

// Continuously monitor the on-chain "Deposit" event.
func (p *Processor) monitorEvent(ledgerContract chain.Contract) {
	monitorCfg := &monitor.Config{
		EventName:  event.Deposit,
		Contract:   ledgerContract,
		StartBlock: p.monitorService.GetCurrentBlockNumber(),
	}
	p.monitorService.Monitor(monitorCfg,
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerDeposit{}
			err := ledgerContract.ParseEvent(event.Deposit, eLog, e)
			if err != nil {
				log.Error(err)
				return
			}
			self := p.nodeConfig.GetOnChainAddr()
			if e.PeerAddrs[0] != self && e.PeerAddrs[1] != self {
				return
			}
			txHash := eLog.TxHash
			log.Infoln("Caught new deposit made to channel", ctype.CidType(e.ChannelId).Hex(), "tx", txHash.Hex())
			p.handleEvent(e, txHash)
		})
}

// Update balance and deposit jobs according to an on-chain Deposit event.
func (p *Processor) handleEvent(event *ledger.CelerLedgerDeposit, txHash ctype.Hash) {
	metrics.IncDepositEventCnt()
	cid := ctype.CidType(event.ChannelId)
	updateOnChainBalanceTx := func(tx *storage.DALTx, args ...interface{}) error {
		balance, found, err := tx.GetOnChainBalance(cid)
		if err != nil {
			return err
		}
		if !found {
			return common.ErrChannelNotFound
		}
		if event.PeerAddrs[0] == p.nodeConfig.GetOnChainAddr() {
			balance.MyDeposit = event.Deposits[0]
			balance.PeerDeposit = event.Deposits[1]
		} else if event.PeerAddrs[1] == p.nodeConfig.GetOnChainAddr() {
			balance.MyDeposit = event.Deposits[1]
			balance.PeerDeposit = event.Deposits[0]
		} else {
			return common.ErrInvalidAccountAddress
		}
		err = tx.UpdateOnChainBalance(cid, balance)
		if err != nil {
			return err
		}
		return tx.UpdateDepositStatesByTxHashAndCid(txHash.Hex(), cid, structs.DepositState_SUCCEEDED)
	}
	if err := p.dal.Transactional(updateOnChainBalanceTx); err != nil {
		log.Error(err)
		metrics.IncDepositErrCnt()
		return
	}
	if p.isOSP {
		p.handleBatchJobEvent(txHash)
	} else {
		p.handleSingleJobEvent(txHash.Hex())
	}
}

func PrintDepositJob(d *structs.DepositJob) string {
	str := fmt.Sprintf("uuid %s cid %s amount %s deadline %s state %s",
		d.UUID, ctype.Cid2Hex(d.Cid), d.Amount, d.Deadline, depositStateName(d.State))
	if d.ToPeer {
		str += " toPeer"
	}
	if d.Refill {
		str += " refill"
	}
	if d.TxHash != "" {
		str += " txhash " + d.TxHash
	}
	if d.ErrMsg != "" {
		str += " errmsg " + d.ErrMsg
	}
	return str
}

func depositStateName(state int) string {
	switch state {
	case structs.DepositState_NULL:
		return "NULL"
	case structs.DepositState_QUEUED:
		return "QUEUED"
	case structs.DepositState_APPROVING_ERC20:
		return "APPROVING_ERC20"
	case structs.DepositState_TX_SUBMITTING:
		return "TX_SUBMITTING"
	case structs.DepositState_TX_SUBMITTED:
		return "TX_SUBMITTED"
	case structs.DepositState_SUCCEEDED:
		return "SUCCEEDED"
	case structs.DepositState_FAILED:
		return "FAILED"
	default:
		return "ERROR"
	}
}

func now() time.Time {
	return time.Now().UTC()
}
