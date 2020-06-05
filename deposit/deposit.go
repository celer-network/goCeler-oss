// Copyright 2018-2020 Celer Network

package deposit

import (
	"fmt"
	"sync"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/eth"
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
	if isOSP {
		go p.monitorOnAllLedgers()
	}
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
