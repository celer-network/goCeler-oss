// Copyright 2018-2020 Celer Network

package dispute

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/transactor"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
)

func (p *Processor) IntendSettlePaymentChannel(cid ctype.CidType, waitMined bool) error {
	log.Infoln("Intend settle payment channel", cid.Hex())
	err := p.dal.Transactional(fsm.OnChannelIntendSettle, cid)
	if err != nil {
		log.Error(err)
		return err
	}

	_, selfSimplex, _, peerSimplex, found, err := p.dal.GetDuplexChannel(cid)
	if err != nil {
		log.Errorln("GetDuplexChannel failed:", err, cid.Hex())
		return err
	}
	if !found {
		log.Errorln("GetDuplexChannel not found:", cid.Hex())
		return common.ErrChannelNotFound
	}

	var stateArray chain.SignedSimplexStateArray
	if len(selfSimplex.SigOfPeerFrom) > 0 && len(selfSimplex.SigOfPeerTo) > 0 {
		sigSortedStateSelf, err2 := SigSortedSimplexState(selfSimplex)
		if err2 == nil {
			stateArray.SignedSimplexStates = append(stateArray.SignedSimplexStates, sigSortedStateSelf)
		} else {
			log.Error(err2, "cid", cid.Hex())
			return err2
		}
	}
	if len(peerSimplex.SigOfPeerFrom) > 0 && len(peerSimplex.SigOfPeerTo) > 0 {
		sigSortedStatePeer, err2 := SigSortedSimplexState(peerSimplex)
		if err2 == nil {
			stateArray.SignedSimplexStates = append(stateArray.SignedSimplexStates, sigSortedStatePeer)
		} else {
			log.Error(err2, "cid", cid.Hex())
			return err2
		}
	}

	// handle empty channel state
	if len(stateArray.SignedSimplexStates) == 0 {
		simplexState := &chain.SignedSimplexState{
			SimplexState: selfSimplex.GetSimplexState(),
		}
		simplexState.Sigs = append(simplexState.Sigs, selfSimplex.SigOfPeerFrom)
		stateArray.SignedSimplexStates = append(stateArray.SignedSimplexStates, simplexState)
	}

	stateArrayBytes, err := proto.Marshal(&stateArray)
	if err != nil {
		log.Error(err, "cid", cid.Hex())
		return err
	}
	if waitMined {
		return p.intendSettleAndWaitMined(cid, stateArrayBytes)
	}
	return p.intendSettle(cid, stateArrayBytes)
}

func (p *Processor) intendSettleAndWaitMined(cid ctype.CidType, stateArrayBytes []byte) error {
	receipt, err := p.transactorPool.SubmitWaitMined(
		fmt.Sprintf("intend settle payment channel %x", cid),
		&transactor.TxConfig{},
		p.intendSettleTxMethod(cid, stateArrayBytes))
	if err != nil {
		log.Errorf("intend settle payment channel error %s, cid %x", err, cid)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("intend settle transaction %x failed", receipt.TxHash)
	}
	return nil
}

func (p *Processor) intendSettle(cid ctype.CidType, stateArrayBytes []byte) error {
	tx, err := p.transactorPool.Submit(
		newGenericTransactionHandler("intend settle", cid),
		&transactor.TxConfig{},
		p.intendSettleTxMethod(cid, stateArrayBytes))
	if err != nil {
		log.Errorf("intend settle payment channel error %s, cid %x", err, cid)
		return err
	}
	log.Infof("sent intend settle tx %x for cid %x", tx.Hash(), cid)
	return nil
}

func (p *Processor) ConfirmSettlePaymentChannel(cid ctype.CidType, waitMined bool) error {
	log.Infoln("Confirm settle payment channel", cid.Hex())
	state, found, err := p.dal.GetChanState(cid)
	if err != nil {
		return fmt.Errorf("GetChanState %x err: %w", cid, err)
	}
	if !found {
		return common.ErrChannelNotFound
	}
	if state != enums.ChanState_SETTLING {
		return fmt.Errorf("invalid channel %x state %s", cid, fsm.ChanStateName(state))
	}
	blkNum := p.monitorService.GetCurrentBlockNumber()
	finalizedBlknum, err := ledgerview.GetOnChainSettleFinalizedTime(cid, p.nodeConfig)
	if err != nil {
		return fmt.Errorf("GetOnChainSettleFinalizedTime err: %w", err)
	}
	if blkNum.Uint64() < finalizedBlknum.Uint64() {
		return fmt.Errorf("channel %x not finalized yet", cid)
	}
	if waitMined {
		return p.confirmSettleAndWaitMined(cid)
	}
	return p.confirmSettle(cid)
}

func (p *Processor) confirmSettleAndWaitMined(cid ctype.CidType) error {
	receipt, err := p.transactorPool.SubmitWaitMined(
		fmt.Sprintf("confirm settle payment channel %x", cid),
		&transactor.TxConfig{},
		p.confirmSettleTxMethod(cid))
	if err != nil {
		log.Errorf("confirm settle payment channel error %s, cid %x", err, cid)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("confirm settle transaction %x failed", receipt.TxHash)
	}

	if p.isOSP {
		// OSP event monitor will call HandleConfirmSettleEventTx
		return nil
	}
	return p.dal.Transactional(p.HandleConfirmSettleEventTx, cid)
}

func (p *Processor) confirmSettle(cid ctype.CidType) error {
	tx, err := p.transactorPool.Submit(
		newGenericTransactionHandler("confirm settle", cid),
		&transactor.TxConfig{},
		p.confirmSettleTxMethod(cid))
	if err != nil {
		log.Errorf("confirm settle payment channel error %s, cid %x", err, cid)
		return err
	}
	log.Infof("sent confirm settle tx %x for cid %x", tx.Hash(), cid)
	return nil
}

func newGenericTransactionHandler(description string, cid ctype.CidType) *transactor.TransactionStateHandler {
	return &transactor.TransactionStateHandler{
		OnMined: func(receipt *types.Receipt) {
			if receipt.Status == types.ReceiptStatusSuccessful {
				log.Infof("%s transaction %x succeeded, cid %x", description, receipt.TxHash, cid)
			} else {
				log.Errorf("%s transaction %x failed, cid %x", description, receipt.TxHash, cid)
			}
		},
	}
}

func (p *Processor) intendSettleTxMethod(cid ctype.CidType, stateArrayBytes []byte) transactor.TxMethod {
	return func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
		chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
		if chanLedger == nil {
			return nil, fmt.Errorf("Fail to get ledger for channel: %x", cid)
		}
		contract, err2 := ledger.NewCelerLedgerTransactor(chanLedger.GetAddr(), transactor)
		if err2 != nil {
			return nil, err2
		}
		return contract.IntendSettle(opts, stateArrayBytes)
	}
}

func (p *Processor) confirmSettleTxMethod(cid ctype.CidType) transactor.TxMethod {
	return func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
		chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
		if chanLedger == nil {
			return nil, fmt.Errorf("Fail to get ledger for channel: %x", cid)
		}
		contract, err2 := ledger.NewCelerLedgerTransactor(chanLedger.GetAddr(), transactor)
		if err2 != nil {
			return nil, err2
		}
		return contract.ConfirmSettle(opts, cid)
	}
}

func (p *Processor) HandleConfirmSettleEventTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	_, found, err := tx.GetChanState(cid)
	if err != nil {
		return fmt.Errorf("%x GetChanState err: %w", cid, err)
	}
	if !found {
		return nil
	}

	err = fsm.OnChannelConfirmSettle(tx, cid)
	if err != nil {
		log.Errorln("fsm OnChannelConfirmSettle err:", err)
		return err
	}
	peer, token, opents, found, err := tx.GetChanForClose(cid)
	if err != nil {
		log.Errorln("GetChanPeerToken:", err, "cid:", cid.Hex())
		return err
	}
	if !found {
		log.Errorln("GetChanPeerToken channel not found:", cid.Hex())
		return common.ErrChannelNotFound
	}
	err = tx.DeleteChan(cid)
	if err != nil {
		log.Errorln(err, cid.Hex())
		return err
	}
	err = tx.InsertClosedChan(cid, peer, token, opents, time.Now().UTC())
	if err != nil {
		log.Errorln(err, cid.Hex())
		return err
	}
	err = tx.DeleteRouting(peer, token)
	if storage.IsDbError(err) {
		log.Errorln(err, cid.Hex(), ctype.Addr2Hex(peer), ctype.Bytes2Hex(token.GetTokenAddress()))
		return err
	}
	return nil
}

func (p *Processor) handleIntendSettleEventTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	seqNums := args[1].([2]*big.Int)
	needRespond := args[2].(*bool)
	*needRespond = false

	peer, state, selfSimplex, peerSimplex, found, err := tx.GetChanForIntendSettle(cid)
	if err != nil {
		log.Error(err, cid.Hex())
		return err
	}
	if !found {
		// For case of not having state, we do want to return nil as it's a valid case.
		// That means th event is about channel opened with OSP with different address.
		log.Debugf("IntendSettle for other OSP addr. cid %x", cid)
		return nil
	}
	if state != enums.ChanState_OPENED {
		// For setup of multi-server osp listening separately where several servers may get this event and try to respond.
		// Thanks to transaction, we can avoid respond twice by checking the state.
		log.Debugf("cid %x is not in open or migrating state", cid)
		return nil
	}
	err = tx.UpdateChanState(cid, enums.ChanState_SETTLING)
	if err != nil {
		log.Errorf("UpdateChanState err %s, cid %x", err, cid)
		return err
	}

	// Figure out which (seqNum, addr) pair, seqNums are sorted by addr.
	var peerSimplexSeq, selfSimplexSeq *big.Int
	if bytes.Compare(peer.Bytes(), p.nodeConfig.GetOnChainAddr().Bytes()) == -1 {
		peerSimplexSeq, selfSimplexSeq = seqNums[0], seqNums[1]
	} else {
		selfSimplexSeq, peerSimplexSeq = seqNums[0], seqNums[1]
	}
	if peerSimplex.SeqNum > uint64(peerSimplexSeq.Int64()) {
		*needRespond = true
	}
	if selfSimplex.SeqNum > uint64(selfSimplexSeq.Int64()) {
		*needRespond = true
	}
	return nil
}

func (p *Processor) monitorPaymentChannelSettleEvent(ledgerContract chain.Contract) {
	_, monErr := p.monitorService.Monitor(
		event.IntendSettle,
		ledgerContract,
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			// CAVEAT!!!: suppose we have the same struct of event.
			// If event struct changes, this monitor does not work.
			e := &ledger.CelerLedgerIntendSettle{}
			if err := ledgerContract.ParseEvent(event.IntendSettle, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			log.Infof("Seeing IntendSettle event, cid %x txhash %x blknum %d ", cid, eLog.TxHash, eLog.BlockNumber)
			needRespond := false
			err := p.dal.Transactional(p.handleIntendSettleEventTx, cid, e.SeqNums, &needRespond)
			if err != nil {
				return
			}
			// Update data of routing table calculation
			if p.routeController != nil {
				p.routeController.RemoveEdge(cid)
			}
			if !needRespond {
				log.Debugln("No need to respond IntendSettle cid:", cid.Hex())
				return
			}
			log.Debugln("Responding IntendSettle cid:", cid.Hex())
			p.IntendSettlePaymentChannel(cid, false) // errs logged within func
			metrics.IncDisputeSettleEventCnt(event.IntendSettle)
		})
	if monErr != nil {
		log.Error(monErr)
	}

	_, monErr = p.monitorService.Monitor(
		event.ConfirmSettle,
		ledgerContract,
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			// CAVEAT!!!: suppose we have the same struct of event.
			// If event struct changes, this monitor does not work.
			e := &ledger.CelerLedgerConfirmSettle{}
			if err := ledgerContract.ParseEvent(event.ConfirmSettle, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			_, hasState, err := p.dal.GetChanState(cid)
			if err != nil {
				log.Error(err, "cid", cid.Hex())
			}
			if hasState {
				log.Infoln("Seeing ConfirmSettle event cid:", cid.Hex(), "final balance:", e.SettleBalance)
				err = p.dal.Transactional(p.HandleConfirmSettleEventTx, cid)
				if err != nil {
					log.Errorln(err, "cid", cid.Hex())
				}
			}
			metrics.IncDisputeSettleEventCnt(event.ConfirmSettle)
		})
	if monErr != nil {
		log.Error(monErr)
	}
}
