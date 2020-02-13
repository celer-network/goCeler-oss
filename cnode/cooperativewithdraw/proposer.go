// Copyright 2018-2019 Celer Network

package cooperativewithdraw

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler-oss/chain"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/cnode/jobs"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
)

type Callback interface {
	OnWithdraw(withdrawHash string, txHash string)
	OnError(withdrawHash string, err string)
}

func (p *Processor) ProcessResponse(response *rpc.CooperativeWithdrawResponse) error {
	withdrawInfo := response.WithdrawInfo
	withdrawHash := p.generateWithdrawHash(ctype.Bytes2Cid(withdrawInfo.ChannelId), withdrawInfo.SeqNum)
	p.initLock.Lock()
	hasJob, err := p.dal.HasCooperativeWithdrawJob(withdrawHash)
	p.initLock.Unlock()
	if err != nil {
		log.Error(err)
	}
	errNoRequest := errors.New("Received CooperativeWithdrawResponse without request")
	if err != nil || !hasJob {
		return errNoRequest
	}
	job, err := p.dal.GetCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		log.Error(err)
	}
	if err != nil || job.State != jobs.CooperativeWithdrawWaitResponse {
		return errNoRequest
	}

	// Send the tx asynchronously
	go func() {
		err = p.sendCooperativeWithdrawTx(job, response)
		if err != nil {
			p.abortJob(job, err)
		}
	}()
	return nil
}

func (p *Processor) prepareJob(
	cid ctype.CidType,
	tokenType entity.TokenType,
	tokenAddress string,
	amount string) (*jobs.CooperativeWithdrawJob, error) {

	contract, err :=
		ledger.NewCelerLedgerCaller(p.ledger.GetAddr(), p.transactorPool.ContractCaller())
	if err != nil {
		return nil, err
	}
	seqNum, err := contract.GetCooperativeWithdrawSeqNum(&bind.CallOpts{}, cid)
	if err != nil {
		return nil, err
	}
	newSeqNum := big.NewInt(0)
	newSeqNum.Add(seqNum, big.NewInt(1))
	newSeqNumUint64 := newSeqNum.Uint64()

	withdrawHash := p.generateWithdrawHash(cid, newSeqNumUint64)
	hasJob, err := p.dal.HasCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		return nil, err
	} else if hasJob {
		return nil, errors.New("Previous withdraw request still pending")
	}

	log.Infof("Withdrawing %s from cid %s", amount, ctype.Cid2Hex(cid))
	_, success := big.NewInt(0).SetString(amount, 16)
	if !success {
		return nil, fmt.Errorf("Invalid withdraw amount %s", amount)
	}

	withdraw := &entity.AccountAmtPair{
		Account: ctype.Hex2Bytes(p.selfAddress),
		Amt:     ctype.Hex2Bytes(amount),
	}
	withdrawInfo := &entity.CooperativeWithdrawInfo{
		ChannelId: cid[:],
		SeqNum:    newSeqNumUint64,
		Withdraw:  withdraw,
		WithdrawDeadline: p.utils.GetCurrentBlockNumber().Uint64() +
			config.CooperativeWithdrawTimeout,
	}
	serializedInfo, err := proto.Marshal(withdrawInfo)
	if err != nil {
		return nil, err
	}
	requesterSig, err := p.signer.Sign(serializedInfo)
	if err != nil {
		return nil, err
	}
	request := &rpc.CooperativeWithdrawRequest{
		WithdrawInfo: withdrawInfo,
		RequesterSig: requesterSig,
	}
	peer, err := p.dal.GetPeer(cid)
	if err != nil {
		return nil, err
	}

	err = p.dal.Transactional(p.checkWithdrawBalanceTx, cid, withdrawInfo)
	if err != nil {
		return nil, err
	}

	msg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_WithdrawRequest{
			WithdrawRequest: request,
		},
	}
	log.Infoln("Sending withdraw request of seq", newSeqNum.String(), "to", peer)
	p.initLock.Lock()
	err = p.streamWriter.WriteCelerMsg(peer, msg)
	if err != nil {
		return nil, err
	}
	job, err := p.initJob(withdrawHash)
	p.initLock.Unlock()
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (p *Processor) sendCooperativeWithdrawTx(
	job *jobs.CooperativeWithdrawJob,
	response *rpc.CooperativeWithdrawResponse) error {
	requesterSig := response.RequesterSig
	approverSig := response.ApproverSig
	withdrawInfo := response.WithdrawInfo
	serializedInfo, err := proto.Marshal(withdrawInfo)
	if err != nil {
		return err
	}
	cid := ctype.Bytes2Cid(response.WithdrawInfo.ChannelId)
	peer, err := p.dal.GetPeer(cid)
	if err != nil {
		return err
	}
	if !p.signer.SigIsValid(peer, serializedInfo, response.ApproverSig) {
		errMsg := "Invalid CooperativeWithdrawResponse signature"
		return errors.New(errMsg)
	}
	var sigs [][]byte
	if strings.Compare(p.selfAddress, peer) < 0 {
		sigs = [][]byte{requesterSig, approverSig}
	} else {
		sigs = [][]byte{approverSig, requesterSig}
	}
	onChainRequest := &chain.CooperativeWithdrawRequest{
		WithdrawInfo: serializedInfo,
		Sigs:         sigs,
	}
	serializedRequest, err := proto.Marshal(onChainRequest)
	if err != nil {
		return err
	}

	p.monitorSingleEvent(true)
	err = p.dal.PutEventMonitorBit(p.eventMonitorName)
	if err != nil {
		return err
	}

	tx, err := p.transactorPool.Submit(
		nil,
		big.NewInt(0),
		func(
			transactor bind.ContractTransactor,
			opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, contractErr :=
				ledger.NewCelerLedgerTransactor(p.ledger.GetAddr(), transactor)
			if contractErr != nil {
				return nil, contractErr
			}
			return contract.CooperativeWithdraw(opts, serializedRequest)
		})
	if err != nil {
		return err
	}
	job.State = jobs.CooperativeWithdrawWaitTx
	job.TxHash = tx.Hash().Hex()
	err = p.dal.PutCooperativeWithdrawJob(job.WithdrawHash, job)
	if err != nil {
		return err
	}
	p.dispatchJob(job)
	return nil
}

func (p *Processor) waitTx(job *jobs.CooperativeWithdrawJob) {
	txHash := job.TxHash
	receipt, err := p.transactorPool.WaitMined(txHash)
	if err != nil {
		p.abortJob(job, err)
		return
	}
	log.Debugf(
		"CooperativeWithdraw tx %s mined, status: %d, gas used: %d",
		txHash,
		receipt.Status,
		receipt.GasUsed)
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Debugf("CooperativeWithdraw tx %s succeeded", txHash)
	} else {
		p.abortJob(job, fmt.Errorf("CooperativeWithdraw tx %s failed", txHash))
	}
}

func (p *Processor) registerCallback(withdrawHash string, cb Callback) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	p.callbacks[withdrawHash] = cb
}

func (p *Processor) maybeFireWithdrawCallback(
	job *jobs.CooperativeWithdrawJob) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	withdrawHash := job.WithdrawHash
	callback := p.callbacks[withdrawHash]
	if callback != nil {
		go callback.OnWithdraw(withdrawHash, job.TxHash)
		p.callbacks[withdrawHash] = nil
	}
}

func (p *Processor) maybeFireErrCallback(job *jobs.CooperativeWithdrawJob) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	withdrawHash := job.WithdrawHash
	callback := p.callbacks[withdrawHash]
	if callback != nil {
		go callback.OnError(withdrawHash, job.Error)
		p.callbacks[withdrawHash] = nil
	}
}

func (p *Processor) maybeFireErrCallbackWithWithdrawHash(
	withdrawHash string, err string) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	callback := p.callbacks[withdrawHash]
	if callback != nil {
		go callback.OnError(withdrawHash, err)
		p.callbacks[withdrawHash] = nil
	}
}

func (p *Processor) markRunningJobAndStartDispatcher(job *jobs.CooperativeWithdrawJob) {
	withdrawHash := job.WithdrawHash
	p.runningJobsLock.Lock()
	defer p.runningJobsLock.Unlock()
	running := p.runningJobs[withdrawHash]
	if running {
		return
	}
	p.runningJobs[withdrawHash] = true
	go p.dispatchJob(job)
}

func (p *Processor) unmarkRunningJob(withdrawHash string) {
	p.runningJobsLock.Lock()
	defer p.runningJobsLock.Unlock()
	delete(p.runningJobs, withdrawHash)
}

func (p *Processor) dispatchJob(job *jobs.CooperativeWithdrawJob) {
	switch job.State {
	case jobs.CooperativeWithdrawWaitResponse:
		// TODO(mzhou): Monitor withdraw deadline
	case jobs.CooperativeWithdrawWaitTx:
		p.waitTx(job)
	case jobs.CooperativeWithdrawSucceeded:
		p.maybeFireWithdrawCallback(job)
		p.unmarkRunningJob(job.WithdrawHash)
	case jobs.CooperativeWithdrawFailed:
		p.maybeFireErrCallback(job)
		p.unmarkRunningJob(job.WithdrawHash)
	}
}

func (p *Processor) resumeJobs() error {
	withdrawHashes, err := p.dal.GetAllCooperativeWithdrawJobKeys()
	if err != nil {
		return err
	}
	for _, withdrawHash := range withdrawHashes {
		p.resumeJob(withdrawHash)
	}
	return nil
}

func (p *Processor) resumeJob(withdrawHash string) {
	job, err := p.dal.GetCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		p.maybeFireErrCallbackWithWithdrawHash(
			withdrawHash, fmt.Sprintf("Cannot retrieve deposit job %s", withdrawHash))
		return
	}
	p.markRunningJobAndStartDispatcher(job)
}

func (p *Processor) initJob(withdrawHash string) (*jobs.CooperativeWithdrawJob, error) {
	job := jobs.NewCooperativeWithdrawJob(withdrawHash)
	err := p.dal.PutCooperativeWithdrawJob(withdrawHash, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (p *Processor) abortJob(job *jobs.CooperativeWithdrawJob, err error) {
	log.Error(fmt.Errorf("CooperativeWithdraw job %s failed: %s", job.WithdrawHash, err))
	job.State = jobs.CooperativeWithdrawFailed
	job.Error = err.Error()
	withdrawHash := job.WithdrawHash
	delete := func(tx *storage.DALTx, args ...interface{}) error {
		err := tx.DeleteEventMonitorBit(p.eventMonitorName)
		if err != nil {
			return err
		}
		return tx.PutCooperativeWithdrawJob(withdrawHash, job)
	}
	txErr := p.dal.Transactional(delete)
	if txErr != nil {
		log.Error(txErr)
	}
	p.dispatchJob(job)
}

func (p *Processor) removeJob(job *jobs.CooperativeWithdrawJob) {
	err := p.dal.DeleteCooperativeWithdrawJob(job.WithdrawHash)
	if err != nil {
		log.Error(err)
	}
}

func (p *Processor) CooperativeWithdraw(
	cid ctype.CidType,
	tokenType entity.TokenType,
	tokenAddress string,
	amount string,
	cb Callback) (string, error) {
	job, err := p.prepareJob(cid, tokenType, tokenAddress, amount)
	if err != nil {
		return "", err
	}
	withdrawHash := job.WithdrawHash
	p.registerCallback(withdrawHash, cb)
	p.markRunningJobAndStartDispatcher(job)
	return withdrawHash, nil
}

func (p *Processor) MonitorCooperativeWithdrawJob(withdrawHash string, cb Callback) {
	p.registerCallback(withdrawHash, cb)
	p.resumeJob(withdrawHash)
}

func (p *Processor) RemoveCooperativeWithdrawJob(withdrawHash string) {
	err := p.dal.DeleteCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		log.Error(err)
	}
}
