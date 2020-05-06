// Copyright 2018-2020 Celer Network

package ledgerview

import (
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

func GetBalance(dal *storage.DAL, cid ctype.CidType, myAddr ctype.Addr, blkNum uint64) (*common.ChannelBalance, error) {
	var balance *common.ChannelBalance
	err := dal.Transactional(getBalanceTx, cid, myAddr, blkNum, &balance)
	return balance, err
}

func getBalanceTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	myAddr := args[1].(ctype.Addr)
	blkNum := args[2].(uint64)
	balance := args[3].(**common.ChannelBalance)
	bal, err := GetBalanceTx(tx, cid, myAddr, blkNum)
	*balance = bal
	return err
}

func GetBalanceTx(tx *storage.DALTx, cid ctype.CidType, myAddr ctype.Addr, blkNum uint64) (*common.ChannelBalance, error) {
	peer, onChainBalance, baseSeq, lastAckedSeq, selfSimplex, peerSimplex, found, err := tx.GetChanForBalance(cid)
	if err != nil {
		return nil, fmt.Errorf("GetChanForBalance err: %w", err)
	}
	if !found {
		return nil, common.ErrChannelNotFound
	}
	mySimplex, err := GetBaseSimplex(tx, cid, selfSimplex, baseSeq, lastAckedSeq)
	if err != nil {
		return nil, fmt.Errorf("GetBaseSimplex err: %w", err)
	}

	balance := ComputeBalance(mySimplex, peerSimplex, onChainBalance, myAddr, peer, blkNum)
	return balance, nil
}

func GetBaseSimplex(
	tx *storage.DALTx,
	cid ctype.CidType,
	selfSimplex *entity.SimplexPaymentChannel,
	baseSeq, lastAckedSeq uint64) (*entity.SimplexPaymentChannel, error) {

	if baseSeq > lastAckedSeq {
		msg, found, err := tx.GetChanMessage(cid, baseSeq)
		if err != nil || !found {
			return nil, fmt.Errorf("GetChanMessage failed: cid %x, seq %d, err %w", cid, baseSeq, err)
		}
		var simplexChannel entity.SimplexPaymentChannel
		var simplexState *rpc.SignedSimplexState
		if msg.GetCondPayRequest() != nil {
			simplexState = msg.GetCondPayRequest().GetStateOnlyPeerFromSig()
		} else if msg.GetPaymentSettleRequest() != nil {
			simplexState = msg.GetPaymentSettleRequest().GetStateOnlyPeerFromSig()
		} else {
			log.Errorln(common.ErrInvalidMsgType, msg)
			return nil, common.ErrInvalidMsgType
		}
		err = proto.Unmarshal(simplexState.GetSimplexState(), &simplexChannel)
		if err != nil {
			return nil, err
		}
		return &simplexChannel, nil
	}
	return selfSimplex, nil
}

func ComputeBalance(
	selfSimplex, peerSimplex *entity.SimplexPaymentChannel,
	onChainBalance *structs.OnChainBalance,
	myAddr, peerAddr ctype.Addr,
	blkNum uint64) *common.ChannelBalance {

	myLockedAmt := new(big.Int).SetBytes(selfSimplex.TotalPendingAmount)
	toPeerAmt := utils.BytesToBigInt(selfSimplex.TransferToPeer.Receiver.Amt)

	peerLockedAmt := new(big.Int).SetBytes(peerSimplex.TotalPendingAmount)
	fromPeerAmt := utils.BytesToBigInt(peerSimplex.TransferToPeer.Receiver.Amt)

	// myFree = myDeposit + fromPeerAmt - myWithdrawal - toPeerAmt - myLockedAmt - myPendingWithdrawal
	myFree := new(big.Int).Add(onChainBalance.MyDeposit, fromPeerAmt)
	myFree.Sub(myFree, onChainBalance.MyWithdrawal)
	myFree.Sub(myFree, toPeerAmt)
	myFree.Sub(myFree, myLockedAmt)

	// peerFree = peerDeposit + toPeerAmt - peerWithdrawal - fromPeerAmt - peerLockedAmt - peerPendingWithdrawl
	peerFree := new(big.Int).Add(onChainBalance.PeerDeposit, toPeerAmt)
	peerFree.Sub(peerFree, onChainBalance.PeerWithdrawal)
	peerFree.Sub(peerFree, fromPeerAmt)
	peerFree.Sub(peerFree, peerLockedAmt)

	if blkNum <= onChainBalance.PendingWithdrawal.Deadline+config.WithdrawTimeoutSafeMargin {
		if onChainBalance.PendingWithdrawal.Receiver == myAddr {
			myFree.Sub(myFree, onChainBalance.PendingWithdrawal.Amount)
		} else {
			peerFree.Sub(peerFree, onChainBalance.PendingWithdrawal.Amount)
		}
	}

	balance := &common.ChannelBalance{
		MyAddr:     myAddr,
		MyFree:     myFree,
		MyLocked:   myLockedAmt,
		PeerAddr:   peerAddr,
		PeerFree:   peerFree,
		PeerLocked: peerLockedAmt,
	}
	return balance
}
