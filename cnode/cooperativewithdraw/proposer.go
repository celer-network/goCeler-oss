// Copyright 2018-2020 Celer Network

package cooperativewithdraw

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
)

type Callback interface {
	OnWithdraw(withdrawHash string, txHash string)
	OnError(withdrawHash string, err string)
}

func (p *Processor) ProcessResponse(frame *common.MsgFrame) error {
	response := frame.Message.GetWithdrawResponse()
	cid := ctype.Bytes2Cid(response.WithdrawInfo.ChannelId)
	frame.LogEntry.FromCid = ctype.Cid2Hex(cid)
	withdrawInfo := response.WithdrawInfo
	chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		return fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	withdrawHash := p.generateWithdrawHash(cid, chanLedger.GetAddr(), withdrawInfo.SeqNum)
	p.initLock.Lock()
	hasJob, err := p.dal.HasCooperativeWithdrawJob(withdrawHash)
	p.initLock.Unlock()
	if err != nil {
		log.Error(err)
	}
	errNoRequest := fmt.Errorf("Received CooperativeWithdrawResponse without request")
	if err != nil || !hasJob {
		return errNoRequest
	}
	job, err := p.dal.GetCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		log.Error(err)
	}
	if err != nil || job.State != structs.CooperativeWithdrawWaitResponse {
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

func (p *Processor) prepareJob(cid ctype.CidType, amount *big.Int) (*structs.CooperativeWithdrawJob, error) {

	chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		return nil, fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	contract, err :=
		ledger.NewCelerLedgerCaller(chanLedger.GetAddr(), p.transactorPool.ContractCaller())
	if err != nil {
		return nil, fmt.Errorf("NewCelerLedgerCaller err: %w", err)
	}
	seqNum, err := contract.GetCooperativeWithdrawSeqNum(&bind.CallOpts{}, cid)
	if err != nil {
		return nil, fmt.Errorf("GetCooperativeWithdrawSeqNum err: %w", err)
	}
	newSeqNum := big.NewInt(0)
	newSeqNum.Add(seqNum, big.NewInt(1))
	newSeqNumUint64 := newSeqNum.Uint64()

	withdrawHash := p.generateWithdrawHash(cid, chanLedger.GetAddr(), newSeqNumUint64)
	hasJob, err := p.dal.HasCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		return nil, fmt.Errorf("HasCooperativeWithdrawJob err: %w", err)
	} else if hasJob {
		return nil, fmt.Errorf("Previous withdraw request still pending")
	}

	withdraw := &entity.AccountAmtPair{
		Account: p.selfAddress.Bytes(),
		Amt:     amount.Bytes(),
	}
	withdrawInfo := &entity.CooperativeWithdrawInfo{
		ChannelId: cid[:],
		SeqNum:    newSeqNumUint64,
		Withdraw:  withdraw,
		WithdrawDeadline: p.monitorService.GetCurrentBlockNumber().Uint64() +
			config.CooperativeWithdrawTimeout,
	}
	serializedInfo, err := proto.Marshal(withdrawInfo)
	if err != nil {
		return nil, fmt.Errorf("proto Marshal err: %w", err)
	}
	requesterSig, err := p.signer.SignEthMessage(serializedInfo)
	if err != nil {
		return nil, fmt.Errorf("Sign err: %w", err)
	}
	request := &rpc.CooperativeWithdrawRequest{
		WithdrawInfo: withdrawInfo,
		RequesterSig: requesterSig,
	}
	peer, found, err := p.dal.GetChanPeer(cid)
	if err != nil {
		return nil, fmt.Errorf("GetChanPeer err: %w", err)
	}
	if !found {
		return nil, common.ErrChannelNotFound
	}

	err = p.dal.Transactional(p.checkWithdrawBalanceTx, cid, withdrawInfo)
	if err != nil {
		return nil, fmt.Errorf("checkWithdrawBalanceTx err: %w", err)
	}

	msg := &rpc.CelerMsg{
		Message: &rpc.CelerMsg_WithdrawRequest{
			WithdrawRequest: request,
		},
	}
	log.Infof("Sending cooperative withdraw request to %s. %s, withdraw hash: %s",
		ctype.Addr2Hex(peer), utils.PrintCooperativeWithdrawInfo(withdrawInfo), withdrawHash)
	p.initLock.Lock()
	defer p.initLock.Unlock()
	err = p.streamWriter.WriteCelerMsg(peer, msg)
	if err != nil {
		return nil, fmt.Errorf("WriteCelerMsg err: %w", err)
	}
	ledgerAddr, found, err := p.dal.GetChanLedger(cid)
	if err != nil {
		return nil, fmt.Errorf("Fail to get channel(%x) ledger: %w", cid, err)
	}
	if !found {
		return nil, fmt.Errorf("no channel found: %x", cid)
	}
	job, err := p.initJob(withdrawHash, ledgerAddr)
	if err != nil {
		return nil, fmt.Errorf("initJob err: %w", err)
	}
	return job, nil
}

func (p *Processor) initJob(withdrawHash string, ledgerAddr ctype.Addr) (*structs.CooperativeWithdrawJob, error) {
	job := &structs.CooperativeWithdrawJob{
		WithdrawHash: withdrawHash,
		State:        structs.CooperativeWithdrawWaitResponse,
		LedgerAddr:   ledgerAddr,
	}
	err := p.dal.PutCooperativeWithdrawJob(withdrawHash, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (p *Processor) sendCooperativeWithdrawTx(
	job *structs.CooperativeWithdrawJob,
	response *rpc.CooperativeWithdrawResponse) error {
	requesterSig := response.RequesterSig
	approverSig := response.ApproverSig
	withdrawInfo := response.WithdrawInfo
	serializedInfo, err := proto.Marshal(withdrawInfo)
	if err != nil {
		return err
	}
	cid := ctype.Bytes2Cid(response.WithdrawInfo.ChannelId)
	peer, found, err := p.dal.GetChanPeer(cid)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrChannelNotFound
	}
	if !utils.SigIsValid(peer, serializedInfo, response.ApproverSig) {
		return fmt.Errorf("Invalid CooperativeWithdrawResponse signature")
	}
	var sigs [][]byte
	if bytes.Compare(p.selfAddress.Bytes(), peer.Bytes()) < 0 {
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

	ledgerContract := p.nodeConfig.GetLedgerContractOf(cid)
	if ledgerContract == nil {
		return fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	p.monitorSingleEvent(ledgerContract, true)
	err = p.dal.UpsertMonitorRestart(monitor.NewEventStr(ledgerContract.GetAddr(), event.CooperativeWithdraw), true)
	if err != nil {
		return err
	}

	tx, err := p.transactorPool.Submit(
		nil,
		&transactor.TxConfig{},
		func(
			transactor bind.ContractTransactor,
			opts *bind.TransactOpts) (*types.Transaction, error) {
			chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
			if chanLedger == nil {
				return nil, fmt.Errorf("Fail to get ledger for channel: %x", cid)
			}
			contract, contractErr :=
				ledger.NewCelerLedgerTransactor(chanLedger.GetAddr(), transactor)
			if contractErr != nil {
				return nil, contractErr
			}
			return contract.CooperativeWithdraw(opts, serializedRequest)
		})
	if err != nil {
		return err
	}
	job.State = structs.CooperativeWithdrawWaitTx
	job.TxHash = tx.Hash().Hex()
	err = p.dal.PutCooperativeWithdrawJob(job.WithdrawHash, job)
	if err != nil {
		return err
	}
	p.dispatchJob(job)
	return nil
}

func (p *Processor) waitTx(job *structs.CooperativeWithdrawJob) {
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
	job *structs.CooperativeWithdrawJob) {
	p.callbacksLock.Lock()
	defer p.callbacksLock.Unlock()
	withdrawHash := job.WithdrawHash
	callback := p.callbacks[withdrawHash]
	if callback != nil {
		go callback.OnWithdraw(withdrawHash, job.TxHash)
		p.callbacks[withdrawHash] = nil
	}
}

func (p *Processor) maybeFireErrCallback(job *structs.CooperativeWithdrawJob) {
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

func (p *Processor) markRunningJobAndStartDispatcher(job *structs.CooperativeWithdrawJob) {
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

func (p *Processor) dispatchJob(job *structs.CooperativeWithdrawJob) {
	switch job.State {
	case structs.CooperativeWithdrawWaitResponse:
		// TODO(mzhou): Monitor withdraw deadline
	case structs.CooperativeWithdrawWaitTx:
		p.waitTx(job)
	case structs.CooperativeWithdrawSucceeded:
		p.maybeFireWithdrawCallback(job)
		p.unmarkRunningJob(job.WithdrawHash)
	case structs.CooperativeWithdrawFailed:
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
			withdrawHash, fmt.Sprintf("Cannot retrieve cooperative withdraw job %s, err %s", withdrawHash, err))
		return
	}
	p.markRunningJobAndStartDispatcher(job)
}

func (p *Processor) abortJob(job *structs.CooperativeWithdrawJob, err error) {
	log.Error(fmt.Errorf("CooperativeWithdraw job %s failed: %w", job.WithdrawHash, err))
	job.State = structs.CooperativeWithdrawFailed
	job.Error = err.Error()
	withdrawHash := job.WithdrawHash
	delete := func(tx *storage.DALTx, args ...interface{}) error {
		err := tx.UpsertMonitorRestart(monitor.NewEventStr(job.LedgerAddr, event.CooperativeWithdraw), false)
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

func (p *Processor) CooperativeWithdraw(cid ctype.CidType, amount *big.Int, cb Callback) (string, error) {
	log.Infoln("cooperative withdraw", amount, "from cid", ctype.Cid2Hex(cid))
	job, err := p.prepareJob(cid, amount)
	if err != nil {
		log.Error(err)
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
	log.Infoln("remove cooperative withdraw job", withdrawHash)
	err := p.dal.DeleteCooperativeWithdrawJob(withdrawHash)
	if err != nil {
		log.Error(err)
	}
}
