// Copyright 2018-2019 Celer Network

package cnode

import (
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler-oss/chain"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/cnode/jobs"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/intfs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/metrics"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/transactor"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

type DepositCallback interface {
	OnDeposit(jobID string, txHash string)
	OnError(jobID string, err string)
}

type depositUtils interface {
	GetCurrentBlockNumber() *big.Int
}

type depositProcessor struct {
	nodeConfig      common.GlobalNodeConfig
	transactor      *transactor.Transactor
	dal             *storage.DAL
	ledger          *chain.BoundContract
	monitorService  intfs.MonitorService
	callbacks       map[string]DepositCallback
	callbacksLock   sync.Mutex
	runningJobs     map[string]bool
	runningJobsLock sync.Mutex
	utils           depositUtils
	enableJobs      bool
}

func startDepositProcessor(
	nodeConfig common.GlobalNodeConfig,
	transactor *transactor.Transactor,
	dal *storage.DAL,
	ledger *chain.BoundContract,
	monitorService intfs.MonitorService,
	utils depositUtils,
	enableJobs bool) (*depositProcessor, error) {
	p := &depositProcessor{
		nodeConfig:     nodeConfig,
		transactor:     transactor,
		dal:            dal,
		ledger:         ledger,
		monitorService: monitorService,
		callbacks:      make(map[string]DepositCallback),
		runningJobs:    make(map[string]bool),
		utils:          utils,
		enableJobs:     enableJobs,
	}
	err := p.resumeJobs()
	if err != nil {
		return nil, err
	}
	go p.monitorEvent()
	return p, nil
}

// Depending on whether the deposit is for ETH or ERC20, and whether the account has enough
// allowance for the ERC20 token, this function sends either a deposit() or an approve()
// transaction. Upon successfully sending the transaction, a deposit job is initialized and
// persisted.
func (p *depositProcessor) prepareJob(amt string, cid ctype.CidType) (*jobs.DepositJob, error) {
	amtInt := new(big.Int)
	amtInt.SetString(amt, 16)
	log.Infoln("Depositing", amt, "wei into channel", cid)
	open, err := p.dal.HasPeer(cid)
	if err != nil {
		return nil, err
	}
	if !open {
		return nil, errors.New("CHANNEL_NOT_OPEN")
	}
	if err != nil {
		return nil, err
	}
	tokenAddr, err := p.dal.GetTokenContractAddr(cid)
	if err != nil {
		return nil, err
	}

	// Deposit ETH
	if tokenAddr == "" || tokenAddr == common.EthContractAddr {
		depositTxHash, depositErr := p.sendDepositTx(cid, amtInt, big.NewInt(0))
		if depositErr != nil {
			return nil, depositErr
		}
		job, jobErr := p.initJob(amt, cid, "", depositTxHash)
		if jobErr != nil {
			return nil, jobErr
		}
		return job, nil
	}
	// Deposit ERC20
	tokenAddress := ctype.Hex2Addr(tokenAddr)
	log.Debugln("Token address:", ctype.Addr2Hex(tokenAddress))

	// Check allowance to avoid unnecessary approve() tx
	erc20, erc20Err := chain.NewERC20Caller(tokenAddress, p.transactor.ContractCaller())
	if erc20Err != nil {
		return nil, erc20Err
	}
	owner := p.transactor.Address()
	spender := p.nodeConfig.GetLedgerContract().GetAddr()
	allowance, allowanceErr := erc20.Allowance(&bind.CallOpts{}, owner, spender)
	if allowanceErr != nil {
		return nil, allowanceErr
	}
	if allowance.Cmp(amtInt) >= 0 {
		depositTxHash, depositErr := p.sendDepositTx(cid, big.NewInt(0), amtInt)
		if depositErr != nil {
			return nil, depositErr
		}
		job, jobErr := p.initJob(amt, cid, "", depositTxHash)
		if jobErr != nil {
			return nil, jobErr
		}
		return job, nil
	}
	approveTxHash, approveErr := p.sendApproveTx(tokenAddress, spender, amtInt)
	if approveErr != nil {
		return nil, approveErr
	}
	job, jobErr := p.initJob(amt, cid, approveTxHash, "")
	if jobErr != nil {
		return nil, jobErr
	}
	return job, nil
}

func (p *depositProcessor) sendApproveTx(
	tokenAddress ethcommon.Address,
	spender ethcommon.Address,
	amtInt *big.Int) (string, error) {
	tx, err := p.transactor.Transact(
		nil,
		big.NewInt(0),
		func(
			transactor bind.ContractTransactor,
			opts *bind.TransactOpts) (*types.Transaction, error) {
			erc20, erc20Err := chain.NewERC20Transactor(tokenAddress, transactor)
			if erc20Err != nil {
				return nil, erc20Err
			}
			return erc20.Approve(opts, spender, amtInt)
		})
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

func (p *depositProcessor) sendDepositTx(
	cid ctype.CidType, txValue *big.Int, amt *big.Int) (string, error) {
	tx, err := p.transactor.Transact(
		nil,
		txValue,
		func(
			transactor bind.ContractTransactor,
			opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, contractErr :=
				ledger.NewCelerLedgerTransactor(p.ledger.GetAddr(), transactor)
			if contractErr != nil {
				return nil, contractErr
			}
			return contract.Deposit(opts, cid, p.transactor.Address(), amt)
		})
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// Update the state deposit according to an on-chain Deposit event. If the txHash maps to a known
// deposit job, mark it as "Succeeded" and tries to fire the callback.
func (p *depositProcessor) handleEvent(event *ledger.CelerLedgerDeposit, txHash string) {
	cid := ctype.CidType(event.ChannelId)
	updateOnChainBalanceTx := func(tx *storage.DALTx, args ...interface{}) error {
		balance, err := tx.GetOnChainBalance(cid)
		if err != nil {
			return err
		}
		if event.PeerAddrs[0] == p.transactor.Address() {
			balance.MyDeposit = event.Deposits[0]
			balance.PeerDeposit = event.Deposits[1]
		} else if event.PeerAddrs[1] == p.transactor.Address() {
			balance.MyDeposit = event.Deposits[1]
			balance.PeerDeposit = event.Deposits[0]
		} else {
			return common.ErrInvalidAccountAddress
		}
		return tx.PutOnChainBalance(cid, balance)
	}

	if err := p.dal.Transactional(updateOnChainBalanceTx); err != nil {
		log.Error(err)
		return
	}
	if !p.enableJobs {
		return
	}
	hasJob, err := p.dal.HasDepositTxHashToJobID(txHash)
	if err != nil {
		log.Error(err)
		return
	} else if !hasJob {
		return
	}
	jobID, err := p.dal.GetDepositTxHashToJobID(txHash)
	if err != nil {
		log.Error(err)
		return
	}
	job, err := p.dal.GetDepositJob(jobID)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot retrieve deposit job %s: %s", jobID, err)
		log.Error(errMsg)
		p.maybeFireErrCallbackWithJobID(jobID, errMsg)
		return
	}
	job.State = jobs.DepositSucceeded
	putErr := p.dal.PutDepositJob(jobID, job)
	if putErr != nil {
		log.Error(putErr)
		// Even if we just lost persistence, still dispatch the job and try to invoke the callback.
	}
	p.dispatchJob(job)
}

// Continuously monitor the on-chain "Deposit" event.
func (p *depositProcessor) monitorEvent() {
	p.monitorService.Monitor(
		event.Deposit,
		p.ledger,
		p.utils.GetCurrentBlockNumber(),
		nil,
		false, /* quickCatch */
		false,
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerDeposit{}
			err := p.ledger.ParseEvent(event.Deposit, eLog, e)
			if err != nil {
				log.Error(err)
				return
			}
			self := p.transactor.Address()
			if e.PeerAddrs[0] != self && e.PeerAddrs[1] != self {
				return
			}
			txHash := eLog.TxHash.String()
			log.Debugln("Deposit event txHash", txHash)
			log.Infoln("Caught new deposit made to channel ID ", ctype.CidType(e.ChannelId).Hex())
			p.handleEvent(e, txHash)
			metrics.IncCNodeDepositEventCnt()
		})
}

// Wait for the ERC20 approve() tx and send the deposit() tx. If successful, transition the job into
// the "WaitDeposit" state and re-dispatch it.
func (p *depositProcessor) waitApproveAndSendDepositTx(job *jobs.DepositJob) {
	approveTxHash := job.ApproveTxHash
	receipt, waitErr := p.transactor.WaitMined(approveTxHash)
	if waitErr != nil {
		p.abortJob(job, waitErr)
		return
	}
	log.Debugf(
		"Approve tx %s mined, status: %d, gas used: %d",
		approveTxHash,
		receipt.Status,
		receipt.GasUsed)
	amtInt := new(big.Int)
	amtInt.SetString(job.Amount, 16)
	depositTxHash, depositErr := p.sendDepositTx(job.ChannelID, big.NewInt(0), amtInt)
	if depositErr != nil {
		p.abortJob(job, depositErr)
		return
	}
	job.State = jobs.DepositWaitDeposit
	job.DepositTxHash = depositTxHash
	_, persistErr := p.persistJobAndDepositTxHash(job)
	if persistErr != nil {
		p.abortJob(job, persistErr)
		return
	}
	p.dispatchJob(job)
}

// Wait for the deposit() tx.
func (p *depositProcessor) waitDeposit(job *jobs.DepositJob) {
	txHash := job.DepositTxHash
	receipt, err := p.transactor.WaitMined(txHash)
	if err != nil {
		p.abortJob(job, err)
		return
	}
	log.Debugf(
		"Deposit tx %s mined, status: %d, gas used: %d",
		txHash,
		receipt.Status,
		receipt.GasUsed)
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Debugf("Deposit tx %s succeeded", txHash)
	} else {
		p.abortJob(job, fmt.Errorf("Deposit tx %s failed", txHash))
	}
}

func (p *depositProcessor) registerCallback(jobID string, cb DepositCallback) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	p.callbacks[jobID] = cb
}

func (p *depositProcessor) maybeFireDepositCallback(job *jobs.DepositJob) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	jobID := job.JobID
	callback := p.callbacks[jobID]
	if callback != nil {
		go callback.OnDeposit(job.JobID, job.DepositTxHash)
		p.callbacks[jobID] = nil
	}
}

func (p *depositProcessor) maybeFireErrCallback(job *jobs.DepositJob, err string) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	jobID := job.JobID
	callback := p.callbacks[jobID]
	if callback != nil {
		go callback.OnError(jobID, err)
		p.callbacks[jobID] = nil
	}
}

func (p *depositProcessor) maybeFireErrCallbackWithJobID(jobID string, err string) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	callback := p.callbacks[jobID]
	if callback != nil {
		go callback.OnError(jobID, err)
		p.callbacks[jobID] = nil
	}
}

func (p *depositProcessor) markRunningJobAndStartDispatcher(job *jobs.DepositJob) {
	jobID := job.JobID
	p.runningJobsLock.Lock()
	defer p.runningJobsLock.Unlock()
	running := p.runningJobs[jobID]
	if running {
		return
	}
	p.runningJobs[jobID] = true
	go p.dispatchJob(job)
}

func (p *depositProcessor) unmarkRunningJob(jobID string) {
	p.runningJobsLock.Lock()
	defer p.runningJobsLock.Unlock()
	delete(p.runningJobs, jobID)
}

// Dispatch a job depending on its state.
func (p *depositProcessor) dispatchJob(job *jobs.DepositJob) {
	switch job.State {
	case jobs.DepositWaitApprove:
		p.waitApproveAndSendDepositTx(job)
	case jobs.DepositWaitDeposit:
		p.waitDeposit(job)
	case jobs.DepositSucceeded:
		p.maybeFireDepositCallback(job)
		p.unmarkRunningJob(job.JobID)
	case jobs.DepositFailed:
		p.maybeFireErrCallback(job, job.Error)
		p.unmarkRunningJob(job.JobID)
	}
}

func (p *depositProcessor) resumeJobs() error {
	jobIDs, err := p.dal.GetAllDepositJobKeys()
	if err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		p.resumeJob(jobID)
	}
	return nil
}

func (p *depositProcessor) resumeJob(jobID string) {
	job, err := p.dal.GetDepositJob(jobID)
	if err != nil {
		p.maybeFireErrCallbackWithJobID(
			jobID, fmt.Sprintf("Cannot retrieve deposit job %s: %s", jobID, err))
		return
	}
	p.markRunningJobAndStartDispatcher(job)
}

func (p *depositProcessor) initJob(
	amount string,
	cid ctype.CidType,
	approveTxHash string,
	depositTxHash string) (*jobs.DepositJob, error) {
	// Generate a random deposit job ID
	jobID := uuid.New().String()
	job := jobs.NewDepositJob(jobID, amount, cid, approveTxHash, depositTxHash)
	return p.persistJobAndDepositTxHash(job)
}

// Persist a job. If the deposit() tx hash exists, also persist a mapping from that to the jobID so
// that the asynchronous event monitor knows which callback to fire.
func (p *depositProcessor) persistJobAndDepositTxHash(
	job *jobs.DepositJob) (*jobs.DepositJob, error) {
	jobID := job.JobID
	depositTxHash := job.DepositTxHash
	init := func(tx *storage.DALTx, args ...interface{}) error {
		err := tx.PutDepositJob(jobID, job)
		if err != nil {
			return err
		}
		if depositTxHash != "" {
			return tx.PutDepositTxHashToJobID(depositTxHash, jobID)
		}
		return nil
	}

	err := p.dal.Transactional(init)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// Mark a job as "Failed" and try to call the error callback.
func (p *depositProcessor) abortJob(job *jobs.DepositJob, err error) {
	log.Error(fmt.Errorf("Deposit job %s failed: %s", job.JobID, err))
	job.State = jobs.DepositFailed
	job.Error = err.Error()
	jobID := job.JobID
	putErr := p.dal.PutDepositJob(jobID, job)
	if putErr != nil {
		log.Error(putErr)
	}
	p.dispatchJob(job)
}

func (p *depositProcessor) deposit(amt string, cid ctype.CidType, cb DepositCallback) (string, error) {
	job, err := p.prepareJob(amt, cid)
	if err != nil {
		return "", err
	}
	jobID := job.JobID
	p.registerCallback(jobID, cb)
	p.markRunningJobAndStartDispatcher(job)
	return jobID, nil
}

func (p *depositProcessor) monitorJob(jobID string, cb DepositCallback) {
	p.registerCallback(jobID, cb)
	p.resumeJob(jobID)
}

// Remove the job from persistence.
func (p *depositProcessor) removeJob(jobID string) error {
	job, err := p.dal.GetDepositJob(jobID)
	if err != nil {
		return err
	}
	depositTxHash := job.DepositTxHash
	delete := func(tx *storage.DALTx, args ...interface{}) error {
		has, err := tx.HasDepositTxHashToJobID(depositTxHash)
		if err != nil {
			return err
		}
		if has {
			depositTxHashToJobIDErr := tx.DeleteDepositTxHashToJobID(depositTxHash)
			if depositTxHashToJobIDErr != nil {
				return depositTxHashToJobIDErr
			}
		}
		return tx.DeleteDepositJob(job.JobID)
	}
	return p.dal.Transactional(delete)
}

func (c *CNode) Deposit(amt string, cid ctype.CidType, cb DepositCallback) (string, error) {
	return c.depositProcessor.deposit(amt, cid, cb)
}

func (c *CNode) MonitorDepositJob(jobID string, cb DepositCallback) {
	c.depositProcessor.monitorJob(jobID, cb)
}

func (c *CNode) RemoveDepositJob(jobID string) error {
	return c.depositProcessor.removeJob(jobID)
}
