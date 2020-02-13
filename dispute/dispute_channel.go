// Copyright 2018-2019 Celer Network

package dispute

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler-oss/chain"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/metrics"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
)

func (p *Processor) IntendSettlePaymentChannel(cid ctype.CidType) error {
	log.Infoln("Intend settle payment channel", cid.Hex())
	err := p.dal.Transactional(fsm.OnPscIntendSettle, cid)
	if err != nil {
		log.Error(err)
		return err
	}
	self := p.nodeConfig.GetOnChainAddr()
	peer, err := p.dal.GetPeer(cid)
	if err != nil {
		return fmt.Errorf("IntendSettle get channel peer failed %s", err)
	}

	var stateArray chain.SignedSimplexStateArray
	simplexSelf, stateSelf, err := p.dal.GetSimplexPaymentChannel(cid, self)
	if err == nil {
		if len(stateSelf.SigOfPeerFrom) > 0 && len(stateSelf.SigOfPeerTo) > 0 {
			sigSortedStateSelf, err2 := SigSortedSimplexState(stateSelf)
			if err2 == nil {
				stateArray.SignedSimplexStates = append(stateArray.SignedSimplexStates, sigSortedStateSelf)
			} else {
				log.Error(err2, "cid", cid.Hex())
				return err2
			}
		}
	} else {
		log.Error(err, "cid", cid.Hex())
		return err
	}
	_, statePeer, err := p.dal.GetSimplexPaymentChannel(cid, peer)
	if err == nil {
		if len(statePeer.SigOfPeerFrom) > 0 && len(statePeer.SigOfPeerTo) > 0 {
			sigSortedStatePeer, err2 := SigSortedSimplexState(statePeer)
			if err2 == nil {
				stateArray.SignedSimplexStates = append(stateArray.SignedSimplexStates, sigSortedStatePeer)
			} else {
				log.Error(err2, "cid", cid.Hex())
				return err2
			}
		}
	} else {
		log.Error(err, "cid", cid.Hex())
		return err
	}

	// handle empty channel state
	if len(stateArray.SignedSimplexStates) == 0 {
		simplexByte, err2 := proto.Marshal(simplexSelf)
		if err2 != nil {
			log.Error(err2, "cid", cid.Hex())
			return err2
		}
		simplexState := &chain.SignedSimplexState{
			SimplexState: simplexByte,
		}
		simplexState.Sigs = append(simplexState.Sigs, stateSelf.SigOfPeerFrom)
		stateArray.SignedSimplexStates = append(stateArray.SignedSimplexStates, simplexState)
	}

	stateArrayBytes, err := proto.Marshal(&stateArray)
	if err != nil {
		log.Error(err, "cid", cid.Hex())
		return err
	}
	_, err = p.transactorPool.SubmitAndWaitMinedWithGenericHandler(
		"intend settle payment channel",
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(p.nodeConfig.GetLedgerContract().GetAddr(), transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.IntendSettle(opts, stateArrayBytes)
		})
	if err != nil {
		log.Errorln("intend settle payment channel error", err, "cid", cid.Hex())
		return err
	}
	return err
}

func (p *Processor) ConfirmSettlePaymentChannel(cid ctype.CidType) error {
	log.Infoln("Confirm settle payment channel", cid.Hex())
	_, err := p.transactorPool.SubmitAndWaitMinedWithGenericHandler(
		"confirm settle payment channel",
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(p.nodeConfig.GetLedgerContract().GetAddr(), transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.ConfirmSettle(opts, cid)
		})
	if err != nil {
		log.Errorln("confirm settle payment channel error", err, "cid", cid.Hex())
		return err
	}

	err = p.dal.Transactional(p.HandleConfirmSettleEventTx, cid)
	if err != nil {
		log.Error(err, "cid", cid)
	}
	return err
}

func (p *Processor) HandleConfirmSettleEventTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	err := fsm.OnPscConfirmSettle(tx, cid)
	if err != nil {
		// err includes cid, we don't repeat in Errorln.
		log.Errorln("confirm settle:", err)
		return err
	}
	peer, err := tx.GetPeer(cid)
	if err != nil {
		log.Errorln("get peer:", err, "cid:", cid.Hex())
		return err
	}
	simplex, _, err := tx.GetSimplexPaymentChannel(cid, peer)
	if err != nil {
		log.Errorln("get simplex:", err, "cid:", cid.Hex())
		return err
	}
	token := simplex.GetTransferToPeer().GetToken()
	tokenAddr := ctype.Bytes2Hex(token.GetTokenAddress())
	if token.GetTokenType() == entity.TokenType_ETH {
		tokenAddr = common.EthContractAddr
	}
	err = tx.DeleteRoute(peer, tokenAddr)
	if err != nil {
		log.Errorln("delete route:", err, "cid:", cid.Hex())
		return err
	}
	return tx.PutCidForPeerAndToken(ctype.Hex2Bytes(peer), token, ctype.ZeroCid)
}

func (p *Processor) handleIntendSettleEventTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	seqNums := args[1].([2]*big.Int)
	needRespond := args[2].(*bool)
	*needRespond = false
	hasState, err := tx.HasChannelState(cid)
	if err != nil {
		log.Errorln("IntendSettle unrecognized channel:", cid.Hex())
		return err
	}
	if !hasState {
		// For case of not having state, we do want to return nil as it's a valid case.
		// That means th event is about channel opened with OSP with different address.
		log.Debugln("IntendSettle for other OSP addr. cid:", cid.Hex())
		return nil
	}
	state, _, err := tx.GetChannelState(cid)
	if err != nil {
		log.Errorln(err, "Can't get state", cid.Hex())
		return err
	}
	if state != fsm.PscOpen {
		// For setup of multi-server osp listening separately where several servers may get this event and try to respond.
		// Thanks to transaction, we can avoid respond twice by checking the state.
		log.Debugln("psc", cid.Hex(), "is not in open state")
		return nil
	}
	err = fsm.OnPscIntendSettle(tx, cid)
	if err != nil {
		log.Errorln("intend settle:", err, "cid:", cid.Hex())
		return err
	}

	peer, err := tx.GetPeer(cid)
	if err != nil {
		log.Errorln(err, "handle intend settle cid:", cid.Hex())
		return err
	}
	// Figure out which (seqNum, addr) pair, seqNums are sorted by addr.
	peerAddr := ctype.Hex2Bytes(peer)
	myAddr := p.nodeConfig.GetOnChainAddrBytes()

	var peerSimplexSeq, mySimplexSeq *big.Int
	if bytes.Compare(peerAddr, myAddr) == -1 {
		peerSimplexSeq, mySimplexSeq = seqNums[0], seqNums[1]
	} else {
		mySimplexSeq, peerSimplexSeq = seqNums[0], seqNums[1]
	}
	// Verify seqNums are correct for every simplex so that we need to send tx to respond.
	peerSimplex, _, err := tx.GetSimplexPaymentChannel(cid, peer)
	if err != nil {
		log.Errorln(err, "Can't get peer simplex cid:", cid.Hex(), "peer:", peer)
		return err
	}
	mySimplex, _, err := tx.GetSimplexPaymentChannel(cid, p.nodeConfig.GetOnChainAddr())
	if err != nil {
		log.Errorln(err, "Can't get my simplex cid:", cid.Hex(), "peer:", peer)
		return err
	}
	if peerSimplex.SeqNum > uint64(peerSimplexSeq.Int64()) {
		*needRespond = true
	}
	if mySimplex.SeqNum > uint64(mySimplexSeq.Int64()) {
		*needRespond = true
	}
	return nil
}

func (p *Processor) monitorPaymentChannelSettleEvent() {
	_, monErr := p.monitorService.Monitor(
		event.IntendSettle,
		p.nodeConfig.GetLedgerContract(),
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerIntendSettle{}
			if err := p.nodeConfig.GetLedgerContract().ParseEvent(event.IntendSettle, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			txHash := fmt.Sprintf("%x", eLog.TxHash)
			log.Infoln("Seeing IntendSettle event, cid:", cid.Hex(), "tx hash:", txHash, "callback id:", id, "blkNum:", eLog.BlockNumber)
			needRespond := false
			err := p.dal.Transactional(p.handleIntendSettleEventTx, cid, e.SeqNums, &needRespond)
			if err != nil {
				log.Errorln(err, "cid", cid.Hex())
				return
			}
			// Update data of routing table calculation
			if p.rtBuilder != nil {
				p.rtBuilder.RemoveEdge(cid)
			}
			if !needRespond {
				log.Debugln("No need to respond IntendSettle cid:", cid.Hex())
				return
			}
			log.Debugln("Responding IntendSettle cid:", cid.Hex())
			err = p.IntendSettlePaymentChannel(cid)
			if err != nil {
				log.Errorln(err, "Can't IntendSettlePaymentChannel cid:", cid.Hex())
			}
			metrics.IncDisputeSettleEventCnt(event.IntendSettle)
		})
	if monErr != nil {
		log.Error(monErr)
	}

	_, monErr = p.monitorService.Monitor(
		event.ConfirmSettle,
		p.nodeConfig.GetLedgerContract(),
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerConfirmSettle{}
			if err := p.nodeConfig.GetLedgerContract().ParseEvent(event.ConfirmSettle, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			hasState, err := p.dal.HasChannelState(cid)
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
