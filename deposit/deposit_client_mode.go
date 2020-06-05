// Copyright 2018-2020 Celer Network

package deposit

import (
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

func (p *Processor) DepositWithCallback(amt *big.Int, cid ctype.CidType, cb DepositCallback) (string, error) {
	if p.isOSP {
		return "", fmt.Errorf("deposit client mode not supported")
	}
	job, err := p.prepareJob(amt, cid)
	if err != nil {
		return "", err
	}
	p.registerCallback(job.UUID, cb)
	p.markRunningJobAndStartDispatcher(job)
	return job.UUID, nil
}

func (p *Processor) MonitorJobWithCallback(jobID string, cb DepositCallback) {
	if p.isOSP {
		log.Error("deposit client mode not supported")
		return
	}
	p.registerCallback(jobID, cb)
	p.resumeJob(jobID)
}

// Remove the job from database.
func (p *Processor) RemoveJob(jobID string) error {
	return p.dal.DeleteDeposit(jobID)
}

// Depending on whether the deposit is for ETH or ERC20, and whether the account has enough
// allowance for the ERC20 token, this function sends either a deposit() or an approve()
// transaction. Upon successfully sending the transaction, a deposit job is initialized and
// persisted.
func (p *Processor) prepareJob(amount *big.Int, cid ctype.CidType) (*structs.DepositJob, error) {
	log.Infoln("Depositing", amount.String(), "wei into channel", cid.Hex())

	state, token, found, err := p.dal.GetChanStateToken(cid)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, common.ErrChannelNotFound
	}
	if state != structs.ChanState_OPENED {
		return nil, common.ErrInvalidChannelState
	}
	tokenAddr := utils.GetTokenAddr(token)
	// Deposit ETH
	if tokenAddr == ctype.EthTokenAddr {
		depositTxHash, depositErr := p.sendDepositTx(cid, amount, big.NewInt(0))
		if depositErr != nil {
			return nil, depositErr
		}
		job, jobErr := p.initJob(amount, cid, structs.DepositState_TX_SUBMITTED, depositTxHash)
		if jobErr != nil {
			return nil, jobErr
		}
		return job, nil
	}
	// Deposit ERC20
	log.Debugln("Token address:", ctype.Addr2Hex(tokenAddr))

	// Check allowance to avoid unnecessary approve() tx
	erc20, erc20Err := chain.NewERC20Caller(tokenAddr, p.transactor.ContractCaller())
	if erc20Err != nil {
		return nil, erc20Err
	}
	owner := p.transactor.Address()
	chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		return nil, fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	spender := chanLedger.GetAddr()
	allowance, allowanceErr := erc20.Allowance(&bind.CallOpts{}, owner, spender)
	if allowanceErr != nil {
		return nil, allowanceErr
	}
	if allowance.Cmp(amount) >= 0 {
		depositTxHash, depositErr := p.sendDepositTx(cid, big.NewInt(0), amount)
		if depositErr != nil {
			return nil, depositErr
		}
		job, jobErr := p.initJob(amount, cid, structs.DepositState_TX_SUBMITTED, depositTxHash)
		if jobErr != nil {
			return nil, jobErr
		}
		return job, nil
	}
	approveTxHash, approveErr := p.sendApproveTx(tokenAddr, spender, amount)
	if approveErr != nil {
		return nil, approveErr
	}
	job, jobErr := p.initJob(amount, cid, structs.DepositState_APPROVING_ERC20, approveTxHash)
	if jobErr != nil {
		return nil, jobErr
	}
	return job, nil
}

func (p *Processor) initJob(
	amount *big.Int,
	cid ctype.CidType,
	state int,
	txHash string) (*structs.DepositJob, error) {
	// Generate a random deposit job ID
	jobID := uuid.New().String()
	job := &structs.DepositJob{
		UUID:   jobID,
		Cid:    cid,
		Amount: amount,
		State:  state,
		TxHash: txHash,
	}
	err := p.dal.InsertDeposit(jobID, cid, false, amount, false, now(), state, txHash, "")
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (p *Processor) markRunningJobAndStartDispatcher(job *structs.DepositJob) {
	jobID := job.UUID
	p.runningJobsLock.Lock()
	defer p.runningJobsLock.Unlock()
	running := p.runningJobs[jobID]
	if running {
		return
	}
	p.runningJobs[jobID] = true
	go p.dispatchJob(job)
}

func (p *Processor) unmarkRunningJob(jobID string) {
	p.runningJobsLock.Lock()
	defer p.runningJobsLock.Unlock()
	delete(p.runningJobs, jobID)
}

// Dispatch a job depending on its state.
func (p *Processor) dispatchJob(job *structs.DepositJob) {
	switch job.State {
	case structs.DepositState_APPROVING_ERC20:
		p.waitApproveAndSendDepositTx(job)
	case structs.DepositState_TX_SUBMITTED:
		p.waitDeposit(job)
	case structs.DepositState_SUCCEEDED:
		p.maybeFireDepositCallback(job)
		p.unmarkRunningJob(job.UUID)
	case structs.DepositState_FAILED:
		p.maybeFireErrCallback(job.UUID, job.ErrMsg)
		p.unmarkRunningJob(job.UUID)
	}
}

func (p *Processor) sendApproveTx(
	tokenAddr ctype.Addr,
	spender ctype.Addr,
	amtInt *big.Int) (string, error) {
	tx, err := p.transactor.Transact(
		nil,
		&eth.TxConfig{},
		func(
			transactor bind.ContractTransactor,
			opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, contractErr := chain.NewERC20Transactor(tokenAddr, transactor)
			if contractErr != nil {
				return nil, contractErr
			}
			return contract.Approve(opts, spender, amtInt)
		})
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

func (p *Processor) sendDepositTx(
	cid ctype.CidType, txValue *big.Int, amt *big.Int) (string, error) {
	tx, err := p.transactor.Transact(
		nil,
		&eth.TxConfig{EthValue: txValue},
		func(
			transactor bind.ContractTransactor,
			opts *bind.TransactOpts) (*types.Transaction, error) {
			ledgerAddr, found, err := p.dal.GetChanLedger(cid)
			if err != nil {
				return nil, fmt.Errorf("Fail to get ledger address for channel(%x): %w", cid, err)
			}
			if !found {
				return nil, fmt.Errorf("no channel found: %x", cid)
			}
			contract, contractErr :=
				ledger.NewCelerLedgerTransactor(ledgerAddr, transactor)
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

// Wait for the ERC20 approve() tx and send the deposit() tx. If successful, transition the job into
// the "DepositState_TX_SUBMITTED" state and re-dispatch it.
func (p *Processor) waitApproveAndSendDepositTx(job *structs.DepositJob) {
	approveTxHash := job.TxHash
	log.Infof("waiting for deposit job %s erc20 approve tx %s to be mined", job.UUID, approveTxHash)
	receipt, err := p.transactor.WaitMined(approveTxHash)
	if err != nil {
		p.abortJob(job, err)
		return
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("deposit job %s erc20 approve tx %s succeeded", job.UUID, approveTxHash)
	} else {
		p.abortJob(job, fmt.Errorf("erc20 approve tx %s failed", approveTxHash))
		return
	}
	depositTxHash, err := p.sendDepositTx(job.Cid, big.NewInt(0), job.Amount)
	if err != nil {
		p.abortJob(job, err)
		return
	}
	job.State = structs.DepositState_TX_SUBMITTED
	job.TxHash = depositTxHash
	err = p.dal.UpdateDepositStateAndTxHash(job.UUID, structs.DepositState_TX_SUBMITTED, depositTxHash)
	if err != nil {
		p.abortJob(job, err)
		return
	}
	p.dispatchJob(job)
}

// Wait for the deposit() tx.
func (p *Processor) waitDeposit(job *structs.DepositJob) {
	depositTxHash := job.TxHash
	log.Infof("waiting for deposit job %s tx %s to be mined", job.UUID, depositTxHash)
	receipt, err := p.transactor.WaitMined(depositTxHash)
	if err != nil {
		p.abortJob(job, err)
		return
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("deposit job %s tx %s succeeded", job.UUID, depositTxHash)
		err = p.dal.UpdateDepositStatesByTxHashAndCid(depositTxHash, job.Cid, structs.DepositState_SUCCEEDED)
		if err != nil {
			log.Error(err)
			return
		}
		err = ledgerview.SyncOnChainBalance(p.dal, job.Cid, p.nodeConfig)
		if err != nil {
			log.Error(err)
		}
		job.State = structs.DepositState_SUCCEEDED
		p.dispatchJob(job)
	} else {
		p.abortJob(job, fmt.Errorf("deposit tx %s failed", depositTxHash))
	}
}

func (p *Processor) resumeClientJobs() error {
	jobs, err := p.dal.GetAllRunningDepositJobs()
	if err != nil {
		return err
	}
	for _, job := range jobs {
		p.markRunningJobAndStartDispatcher(job)
	}
	return nil
}

func (p *Processor) resumeJob(jobID string) {
	job, found, err := p.dal.GetDepositJob(jobID)
	if err != nil || !found {
		p.maybeFireErrCallback(jobID, fmt.Sprintf("Cannot retrieve deposit job %s: %v, exist %t", jobID, err, found))
		return
	}
	p.markRunningJobAndStartDispatcher(job)
}

// Mark a job as "Failed" and try to call the error callback.
func (p *Processor) abortJob(job *structs.DepositJob, err error) {
	log.Errorf("Deposit job %s failed: %s", job.UUID, err)
	job.State = structs.DepositState_FAILED
	job.ErrMsg = err.Error()
	putErr := p.dal.UpdateDepositErrMsg(job.UUID, job.ErrMsg)
	if putErr != nil {
		log.Error(putErr)
	}
	p.dispatchJob(job)
}

func (p *Processor) registerCallback(jobID string, cb DepositCallback) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	p.callbacks[jobID] = cb
}

func (p *Processor) maybeFireDepositCallback(job *structs.DepositJob) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	callback := p.callbacks[job.UUID]
	if callback != nil {
		go callback.OnDeposit(job.UUID, job.TxHash)
		p.callbacks[job.UUID] = nil
	}
}

func (p *Processor) maybeFireErrCallback(jobID, err string) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	callback := p.callbacks[jobID]
	if callback != nil {
		go callback.OnError(jobID, err)
		p.callbacks[jobID] = nil
	}
}
