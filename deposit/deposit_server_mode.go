// Copyright 2019-2020 Celer Network

package deposit

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/utils/lease"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

// TODO: add prometheus alerts for deposit error and low pool size/allowance

type channelDeposit struct {
	cid        ctype.CidType
	toPeer     bool
	receiver   ctype.Addr
	tokenAddr  ctype.Addr
	amount     *big.Int
	refillAmt  *big.Int
	jobs       []*structs.DepositJob
	deadline   time.Time
	ledgerAddr ctype.Addr
}

func chanDepositJobKey(cid ctype.CidType, toPeer bool) string {
	return fmt.Sprintf("%s:%t", ctype.Cid2Hex(cid), toPeer)
}

// RequestDeposit inserts a deposit job into the db, and return the deposit job ID
func (p *Processor) RequestDeposit(
	cid ctype.CidType, amount *big.Int, toPeer bool, maxWait time.Duration) (string, error) {
	if !p.isOSP {
		return "", fmt.Errorf("deposit server mode not supported")
	}
	jobID := uuid.New().String()
	err := p.dal.InsertDeposit(
		jobID, cid, toPeer, amount, false, now().Add(maxWait), structs.DepositState_QUEUED, "", "")
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Errorf("Insert deposit request error: cid %s, amount %s, topeer %t", ctype.Cid2Hex(cid), amount, toPeer)
	} else {
		metrics.IncDepositJobCnt()
		log.Infof("Inserted deposit request: cid %s, amount %s, topeer %t, maxWait %s, job ID %s",
			ctype.Cid2Hex(cid), amount, toPeer, maxWait, jobID)
	}
	return jobID, err
}

// RequestRefill inserts a refill deposit job into the db, and return the deposit job ID
func (p *Processor) RequestRefill(
	cid ctype.CidType, amount *big.Int, maxWait time.Duration) (string, error) {
	if !p.isOSP {
		return "", fmt.Errorf("deposit server mode not supported")
	}
	jobID := uuid.New().String()
	err := p.dal.Transactional(p.insertRefillTx, jobID, cid, amount, maxWait)
	if err != nil {
		if !errors.Is(err, common.ErrPendingRefill) {
			metrics.IncDepositErrCnt()
		}
		return "", err
	}
	metrics.IncDepositJobCnt()
	return jobID, nil
}

// RequestRefillTx inserts a refill deposit job into the db as part of a transaction, and return the deposit job ID
func (p *Processor) RequestRefillTx(
	tx *storage.DALTx, cid ctype.CidType, amount *big.Int, maxWait time.Duration) (string, error) {
	if !p.isOSP {
		return "", fmt.Errorf("deposit server mode not supported")
	}
	jobID := uuid.New().String()
	err := p.insertRefillTx(tx, jobID, cid, amount, maxWait)
	if err != nil {
		return "", err
	}
	return jobID, nil
}

func (p *Processor) insertRefillTx(tx *storage.DALTx, args ...interface{}) error {
	jobID := args[0].(string)
	cid := args[1].(ctype.CidType)
	amount := args[2].(*big.Int)
	maxWait := args[3].(time.Duration)
	found, err := tx.HasDepositRefillPending(cid)
	if err != nil {
		return err
	}
	if found {
		return common.ErrPendingRefill
	}
	return tx.InsertDeposit(jobID, cid, false, amount, true, now().Add(maxWait), structs.DepositState_QUEUED, "", "")
}

// serverDepositJobPolling periodically processes queued deposit jobs
func (p *Processor) serverDepositJobPolling(quit chan bool) {
	ticker := time.NewTicker(time.Duration(rtconfig.GetDepositPollingInterval()) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			p.processQueuedJobs()
		}
	}
}

func (p *Processor) processQueuedJobs() {
	if !lease.CheckOwner(p.dal, config.EventListenerLeaseName, p.nodeConfig.GetSvrName()) {
		return
	}

	var err error
	p.chanDeposits, err = p.getQueuedJobs()
	if err != nil {
		return
	}
	minDepositBatchSize := int(rtconfig.GetDepositMinBatchSize())
	maxDepositBatchSize := int(rtconfig.GetDepositMaxBatchSize())
	if minDepositBatchSize > maxDepositBatchSize {
		minDepositBatchSize, maxDepositBatchSize = maxDepositBatchSize, minDepositBatchSize
	}
	for ledgerAddr, chanDeposits := range p.chanDeposits {
		tokenSet := make(map[ctype.Addr]bool)
		for i := 0; i < len(chanDeposits); i += maxDepositBatchSize {
			var batch []*channelDeposit
			if i+maxDepositBatchSize < len(chanDeposits) {
				batch = chanDeposits[i : i+maxDepositBatchSize]
			} else {
				batch = chanDeposits[i:]
			}
			if len(batch) == 0 {
				continue
			}
			if batch[0].deadline.After(now()) && len(batch) < minDepositBatchSize {
				// Do not proceed if the earliest job deadline is after the current time, and not enough jobs for batch.
				// It only happens on the last batch in the array.
				continue
			}

			var uuids []string
			var batchSummary string
			for _, dep := range batch {
				tokenSet[dep.tokenAddr] = true
				if dep.toPeer {
					batchSummary += fmt.Sprintf("[cid %s to %s amt %s jobnum %d] ",
						ctype.Cid2Hex(dep.cid), ctype.Addr2Hex(dep.receiver), dep.amount, len(dep.jobs))
				} else {
					batchSummary += fmt.Sprintf("[cid %s to self amt %s jobnum %d] ",
						ctype.Cid2Hex(dep.cid), dep.amount, len(dep.jobs))
				}
				for _, job := range dep.jobs {
					uuids = append(uuids, job.UUID)
				}
			}
			batchSummary = fmt.Sprintf("batch size %d total jobnum %d ", len(batch), len(uuids)) + batchSummary
			// mark all jobs in the batch as SUBMITTING state
			// use batch time as temporary tx hash to differentiate batches
			err = p.dal.UpdateDepositsStateAndTxHash(
				uuids, structs.DepositState_TX_SUBMITTING, fmt.Sprintf("tx@%d", now().UnixNano()))
			if err != nil {
				metrics.IncDepositErrCnt()
				log.Errorln(err, uuids)
			}
			txHash, txErr := p.depositInBatch(batch, ledgerAddr)
			if txErr != nil {
				metrics.IncDepositErrCnt()
				log.Errorf("deposit tx err %s: %s", txErr, batchSummary)
				err = p.dal.UpdateDepositsErrMsg(uuids, txErr.Error())
			} else {
				metrics.IncDepositTxCnt()
				log.Infof("sent deposit tx %s: %s", txHash, batchSummary)
				err = p.dal.UpdateDepositsStateAndTxHash(uuids, structs.DepositState_TX_SUBMITTED, txHash)
				go p.waitDepositTxMined(txHash)
			}
			if err != nil {
				metrics.IncDepositErrCnt()
				log.Errorln(err)
			}
		}
		if len(chanDeposits) > 0 {
			p.checkRefillPool(ledgerAddr, tokenSet)
		}
	}
}

func (p *Processor) depositInBatch(chanDeposits []*channelDeposit, ledgerAddr ctype.Addr) (string, error) {
	var cids [][32]byte
	var receivers []ctype.Addr
	var amounts []*big.Int
	for _, chanDeposit := range chanDeposits {
		cids = append(cids, chanDeposit.cid)
		receivers = append(receivers, chanDeposit.receiver)
		amounts = append(amounts, chanDeposit.amount)
	}
	var tx *types.Transaction
	var err error
	if len(chanDeposits) == 0 {
		return "", fmt.Errorf("empty deposit batch list")
	} else {
		tx, err = p.transactor.Transact(
			nil,
			&transactor.TxConfig{},
			func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
				contract, contractErr := ledger.NewCelerLedgerTransactor(ledgerAddr, transactor)
				if contractErr != nil {
					return nil, contractErr
				}
				return contract.DepositInBatch(opts, cids, receivers, amounts)
			})
	}
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

func (p *Processor) waitDepositTxMined(txHash string) {
	receipt, err := p.transactor.WaitMined(txHash)
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Errorln(err, txHash)
		err = p.dal.UpdateDepositErrMsgByTxHash(txHash, err.Error())
		if err != nil {
			metrics.IncDepositErrCnt()
			log.Errorln(err, txHash)
		}
		return
	}
	log.Debugf("Deposit tx %s mined, status: %d, gas used: %d", txHash, receipt.Status, receipt.GasUsed)
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Debugf("Deposit tx %s succeeded", txHash)
	} else {
		err = fmt.Errorf("Deposit tx %s failed", txHash)
		metrics.IncDepositErrCnt()
		log.Errorln(err, txHash)
		err = p.dal.UpdateDepositErrMsgByTxHash(txHash, err.Error())
		if err != nil {
			metrics.IncDepositErrCnt()
			log.Errorln(err, txHash)
		}
	}
}

func (p *Processor) resumeServerJobs() error {
	// resume submitted jobs
	txHashes, err := p.dal.GetAllSubmittedDepositTxHashes()
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Error(err)
		return err
	}
	for _, txHash := range txHashes {
		go p.waitDepositTxMined(txHash)
	}

	// load jobs that were submitting but not recorded before last shutdown (rare cases)
	jobs, err := p.dal.GetAllDepositJobsByState(structs.DepositState_TX_SUBMITTING)
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Error(err)
		return err
	}
	if len(jobs) == 0 {
		return nil
	}
	var uuids []string
	for _, job := range jobs {
		uuids = append(uuids, job.UUID)
	}
	log.Warnln("Server was down when deposit jobs were being submitted but not recorded:", uuids)
	p.unrecordedDepositsLock.Lock()
	defer p.unrecordedDepositsLock.Unlock()

	p.unrecordedDeposits = make(map[string]map[string]*channelDeposit)
	txJobsmap := make(map[string][]*structs.DepositJob)
	for _, job := range jobs {
		txJobsmap[job.TxHash] = append(txJobsmap[job.TxHash], job)
	}
	for txtime, txjobs := range txJobsmap {
		p.unrecordedDeposits[txtime] = p.aggregateJobs(txjobs)
	}
	go p.clearUnrecordedDeposits()
	return nil
}

func (p *Processor) getQueuedJobs() (map[ctype.Addr][]*channelDeposit, error) {
	jobs, err := p.dal.GetAllDepositJobsByState(structs.DepositState_QUEUED)
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Error(err)
		return nil, err
	}
	chanDepositMap := p.aggregateJobs(jobs)
	chanDeposits := make(map[ctype.Addr][]*channelDeposit)
	for _, chanDeposit := range chanDepositMap {
		chanDeposits[chanDeposit.ledgerAddr] = append(chanDeposits[chanDeposit.ledgerAddr], chanDeposit)
	}
	for ledger := range chanDeposits {
		chanDeposits[ledger] = sortChannelDeposits(chanDeposits[ledger])
	}
	return chanDeposits, nil
}

func (p *Processor) aggregateJobs(jobs []*structs.DepositJob) map[string]*channelDeposit {
	chanDepositMap := make(map[string]*channelDeposit)
	for _, job := range jobs {
		key := chanDepositJobKey(job.Cid, job.ToPeer)
		if chanDepositMap[key] == nil {
			tokenAddr, ledgerAddr, receiver, err2 := p.getTokenAndLedgerAddr(job)
			if err2 != nil {
				metrics.IncDepositErrCnt()
				log.Errorln("deposit getTokenAndLedgerAddr err", err2, job.Cid.Hex())
				err2 = p.dal.UpdateDepositErrMsg(job.UUID, err2.Error())
				if err2 != nil {
					metrics.IncDepositErrCnt()
					log.Errorln(err2, job.UUID)
				}
				continue
			}
			chanDepositMap[key] = &channelDeposit{
				cid:        job.Cid,
				toPeer:     job.ToPeer,
				receiver:   receiver,
				tokenAddr:  tokenAddr,
				amount:     big.NewInt(0),
				refillAmt:  big.NewInt(0),
				deadline:   job.Deadline,
				ledgerAddr: ledgerAddr,
			}
		}
		chanDeposit := chanDepositMap[key]
		if job.Refill {
			if job.Amount.Cmp(chanDeposit.refillAmt) == 1 {
				chanDeposit.refillAmt = job.Amount
			}
		} else {
			chanDeposit.amount.Add(chanDeposit.amount, job.Amount)
		}
		chanDeposit.jobs = append(chanDeposit.jobs, job)
		if job.Deadline.Before(chanDeposit.deadline) {
			chanDeposit.deadline = job.Deadline
		}
	}
	for _, chanDeposit := range chanDepositMap {
		if chanDeposit.refillAmt.Cmp(chanDeposit.amount) == 1 {
			chanDeposit.amount = chanDeposit.refillAmt
		}
	}
	return chanDepositMap
}

func (p *Processor) getTokenAndLedgerAddr(job *structs.DepositJob) (ctype.Addr, ctype.Addr, ctype.Addr, error) {
	state, token, peer, ledger, found, err := p.dal.GetChanForDeposit(job.Cid)
	if err != nil {
		return ctype.ZeroAddr, ctype.ZeroAddr, ctype.ZeroAddr, err
	}
	if !found {
		return ctype.ZeroAddr, ctype.ZeroAddr, ctype.ZeroAddr, common.ErrChannelNotFound
	}
	if state != structs.ChanState_OPENED {
		return ctype.ZeroAddr, ctype.ZeroAddr, ctype.ZeroAddr, common.ErrInvalidChannelState
	}
	receiver := p.nodeConfig.GetOnChainAddr()
	if job.ToPeer {
		receiver = peer
	}
	return utils.GetTokenAddr(token), ledger, receiver, nil
}

func (p *Processor) checkRefillPool(ledgerAddr ctype.Addr, tokenSet map[ctype.Addr]bool) {
	for tokenAddr := range tokenSet {
		poolThreshold := rtconfig.GetRefillPoolThreshold(ctype.Addr2Hex(tokenAddr))
		if poolThreshold.Cmp(big.NewInt(0)) == 0 {
			// no refill for this token
			continue
		}
		if tokenAddr == ctype.EthTokenAddr {
			tokenAddr = p.nodeConfig.GetEthPoolAddr()
		}
		erc20, err := chain.NewERC20Caller(tokenAddr, p.nodeConfig.GetEthConn())
		if err != nil {
			metrics.IncDepositErrCnt()
			log.Error(err)
			continue
		}

		balance, err := erc20.BalanceOf(&bind.CallOpts{}, p.transactor.Address())
		if err != nil {
			metrics.IncDepositErrCnt()
			log.Error(err)
			continue
		}
		if balance.Cmp(poolThreshold) == -1 {
			// send at most one low pool balance alert per hour
			if now().Sub(p.lastAlertTime) > time.Hour {
				metrics.IncDepositPoolAlertCnt()
				p.lastAlertTime = now()
			}
			log.Warnf("Refiller's balance is low. Refiller: %x; token/ethpool: %x; balance: %s; threshold: %s",
				p.transactor.Address(), tokenAddr, balance, poolThreshold)
		}
		allowance, err := erc20.Allowance(&bind.CallOpts{}, p.transactor.Address(), ledgerAddr)
		if err != nil {
			metrics.IncDepositErrCnt()
			log.Error(err)
			continue
		}
		if allowance.Cmp(poolThreshold) == -1 {
			if now().Sub(p.lastAlertTime) > time.Hour {
				// send at most one low pool balance alert per hour
				metrics.IncDepositPoolAlertCnt()
				p.lastAlertTime = now()
			}
			log.Warnf("Refiller's allowance to ledger %x is low. Refiller: %x; token/ethpool: %x; balance: %s; threshold: %s",
				ledgerAddr, p.transactor.Address(), tokenAddr, allowance, poolThreshold)
		}
	}
}

// handleBatchJobEvent currently only handle cases when server was down
// while deposit jobs were being submitted but not recorded
func (p *Processor) handleBatchJobEvent(txHash ctype.Hash) {
	p.unrecordedDepositsLock.Lock()
	defer p.unrecordedDepositsLock.Unlock()
	if len(p.unrecordedDeposits) == 0 {
		return
	}
	found, err := p.dal.HasDepositTxHash(txHash.Hex())
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Error(err)
		return
	}
	if found {
		return
	}

	cids, receivers, amounts := p.parseDepositInBatchTransaction(txHash)
	if len(cids) == 0 {
		return
	}

	var txTimes []string
	for txtime, _ := range p.unrecordedDeposits {
		txTimes = append(txTimes, txtime)
	}
	sort.Strings(txTimes)

	var uuids []string
	for i, _ := range cids {
		key := chanDepositJobKey(cids[i], (receivers[i] != p.nodeConfig.GetOnChainAddr()))
		for _, txtime := range txTimes {
			deposit, ok := p.unrecordedDeposits[txtime][key]
			if ok && deposit.amount == amounts[i] {
				for _, job := range deposit.jobs {
					uuids = append(uuids, job.UUID)
				}
				delete(p.unrecordedDeposits[txtime], key)
				if len(p.unrecordedDeposits[txtime]) == 0 {
					delete(p.unrecordedDeposits, txtime)
				}
				if len(p.unrecordedDeposits) == 0 {
					p.unrecordedDeposits = nil
				}
				break
			}
		}
	}
	if len(uuids) > 0 {
		err = p.dal.UpdateDepositsStateAndTxHash(uuids, structs.DepositState_SUCCEEDED, txHash.Hex())
		if err != nil {
			metrics.IncDepositErrCnt()
			log.Errorln(err, uuids)
		}
	}

}

const (
	txFuncName                      = "depositInBatch"
	txFuncInput_ChannelIds          = "_channelIds"
	txFuncInput_Receivers           = "_receivers"
	txFuncInput_TransferFromAmounts = "_transferFromAmounts"
)

func (p *Processor) parseDepositInBatchTransaction(txHash ctype.Hash) ([]ctype.CidType, []ctype.Addr, []*big.Int) {
	txInfo, err := ledgerview.GetOnChainTxByHash(txHash, p.nodeConfig)
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Error(err, txHash.Hex())
		return nil, nil, nil
	}
	if txInfo.From != p.transactor.Address() {
		// return if tx was not sent by self
		return nil, nil, nil
	}
	log.Infoln(txInfo.FuncName, "tx msg from", ctype.Addr2Hex(txInfo.From), "to", ctype.Addr2Hex(txInfo.To))
	if txInfo.FuncName != txFuncName {
		return nil, nil, nil
	}
	channelIds, ok := txInfo.FuncInput[txFuncInput_ChannelIds].([][32]byte)
	if !ok {
		metrics.IncDepositErrCnt()
		log.Errorf("got data of type %T but wanted [][32]byte", txInfo.FuncInput[txFuncInput_ChannelIds])
		return nil, nil, nil
	}
	if len(channelIds) == 0 {
		return nil, nil, nil
	}
	var cids []ctype.CidType
	for _, channelId := range channelIds {
		cids = append(cids, ctype.Bytes2Cid(channelId[:]))
	}
	receivers, ok := txInfo.FuncInput[txFuncInput_Receivers].([]ctype.Addr)
	if !ok {
		metrics.IncDepositErrCnt()
		log.Errorf("got data of type %T but wanted []ctype.Addr", txInfo.FuncInput[txFuncInput_Receivers])
		return nil, nil, nil
	}
	amounts, ok := txInfo.FuncInput[txFuncInput_TransferFromAmounts].([]*big.Int)
	if !ok {
		metrics.IncDepositErrCnt()
		log.Errorf("got data of type %T but wanted []*big.Int", txInfo.FuncInput[txFuncInput_TransferFromAmounts])
		return nil, nil, nil
	}
	return cids, receivers, amounts
}

func (p *Processor) clearUnrecordedDeposits() {
	time.Sleep(30 * time.Minute)
	p.unrecordedDepositsLock.Lock()
	defer p.unrecordedDepositsLock.Unlock()
	if p.unrecordedDeposits == nil {
		return
	}
	var uuids []string
	for _, deps := range p.unrecordedDeposits {
		for _, dep := range deps {
			for _, job := range dep.jobs {
				uuids = append(uuids, job.UUID)
			}
		}
	}
	err := p.dal.UpdateDepositsErrMsg(uuids, "unrecorded deposit job timeout")
	if err != nil {
		metrics.IncDepositErrCnt()
		log.Error(err, uuids)
	}
	p.unrecordedDeposits = nil
	log.Infoln("cleared unrecorded deposit jobs", uuids)
}

type sortedDepositJobs []*structs.DepositJob

func (s sortedDepositJobs) Len() int {
	return len(s)
}

func (s sortedDepositJobs) Less(i, j int) bool {
	return s[i].Deadline.Before(s[j].Deadline)
}

func (s sortedDepositJobs) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func SortDepositJobs(input []*structs.DepositJob) []*structs.DepositJob {
	sorted := sortedDepositJobs(input)
	sort.Sort(sorted)
	return sorted
}

type sortedChannelDeposits []*channelDeposit

func (s sortedChannelDeposits) Len() int {
	return len(s)
}

func (s sortedChannelDeposits) Less(i, j int) bool {
	return s[i].deadline.Before(s[j].deadline)
}

func (s sortedChannelDeposits) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func sortChannelDeposits(input []*channelDeposit) []*channelDeposit {
	sorted := sortedChannelDeposits(input)
	sort.Sort(sorted)
	return sorted
}
