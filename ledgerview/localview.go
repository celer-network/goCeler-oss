// Copyright 2018-2019 Celer Network

package ledgerview

import (
	"errors"
	"fmt"
	"math/big"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/golang/protobuf/proto"
)

func GetBalance(dal *storage.DAL, cid ctype.CidType, myAddr string, blkNum uint64) (*common.ChannelBalance, error) {
	var balance *common.ChannelBalance
	err := dal.Transactional(getBalanceTx, cid, myAddr, blkNum, &balance)
	return balance, err
}

func getBalanceTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	myAddr := args[1].(string)
	blkNum := args[2].(uint64)
	balance := args[3].(**common.ChannelBalance)
	bal, err := GetBalanceTx(tx, cid, myAddr, blkNum)
	*balance = bal
	return err
}

func GetBalanceTx(tx *storage.DALTx, cid ctype.CidType, myAddr string, blkNum uint64) (*common.ChannelBalance, error) {
	peerAddr, err := tx.GetPeer(cid)
	if err != nil {
		log.Errorln("GetBalanceTx: GetPeer:", err, "cid", cid.Hex())
		return nil, err
	}

	mySimplex, _, err := GetBaseSimplexChannel(tx, cid, myAddr)
	if err != nil {
		log.Errorln("GetBalanceTx: mySimplex:", err, "cid", cid.Hex())
		return nil, err
	}
	myLockedAmt := new(big.Int).SetBytes(mySimplex.TotalPendingAmount)
	toPeerAmt := utils.BytesToBigInt(mySimplex.TransferToPeer.Receiver.Amt)

	peerSimplex, _, err := tx.GetSimplexPaymentChannel(cid, peerAddr)
	if err != nil {
		log.Errorln("GetBalanceTx: peerSimplex:", err, "cid", cid.Hex())
		return nil, err
	}
	peerLockedAmt := new(big.Int).SetBytes(peerSimplex.TotalPendingAmount)
	fromPeerAmt := utils.BytesToBigInt(peerSimplex.TransferToPeer.Receiver.Amt)

	onChainBalance, err := tx.GetOnChainBalance(cid)
	if err != nil {
		log.Errorln("GetBalanceTx: on-chain balance:", err, "cid", cid.Hex())
		return nil, err
	}

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
		if onChainBalance.PendingWithdrawal.Receiver == ctype.Hex2Addr(myAddr) {
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
	return balance, nil
}

func GetBaseSimplexChannel(
	tx *storage.DALTx,
	cid ctype.CidType,
	myAddr string) (*entity.SimplexPaymentChannel, *common.ChannelSeqNums, error) {

	seqNums, err := tx.GetChannelSeqNums(cid)
	if err != nil {
		// new peer or newly upgraded code
		return GetChannelSeqNumsFromSimplexState(tx, cid, myAddr)
	}
	if seqNums.Base > seqNums.LastAcked {
		msg, err2 := tx.GetChannelMessage(cid, seqNums.Base)
		if err2 != nil {
			err3 := fmt.Errorf("GetChannelMessage failed: %s cid %x, seq %d", err2.Error(), cid, seqNums.Base)
			return nil, seqNums, err3
		}
		var simplexChannel entity.SimplexPaymentChannel
		var simplexState *rpc.SignedSimplexState
		if msg.GetCondPayRequest() != nil {
			simplexState = msg.GetCondPayRequest().GetStateOnlyPeerFromSig()
		} else if msg.GetPaymentSettleRequest() != nil {
			simplexState = msg.GetPaymentSettleRequest().GetStateOnlyPeerFromSig()
		} else {
			log.Errorln(common.ErrInvalidMsgType, msg)
			return nil, seqNums, common.ErrInvalidMsgType
		}
		err2 = proto.Unmarshal(simplexState.GetSimplexState(), &simplexChannel)
		if err2 != nil {
			return nil, seqNums, err2
		}
		return &simplexChannel, seqNums, nil
	}
	cosignedSimplex, _, err2 := tx.GetSimplexPaymentChannel(cid, myAddr)
	if err2 != nil {
		return nil, seqNums, common.ErrSimplexStateNotFound
	}
	return cosignedSimplex, seqNums, nil
}

func GetChannelSeqNumsFromSimplexState(
	tx *storage.DALTx,
	cid ctype.CidType,
	myAddr string) (*entity.SimplexPaymentChannel, *common.ChannelSeqNums, error) {
	exist, err := tx.HasChannelSeqNums(cid)
	if err != nil {
		return nil, nil, err
	}
	if exist {
		return nil, nil, errors.New("ChannelSeqNums table already exist")
	}

	simplex, _, err := tx.GetSimplexPaymentChannel(cid, myAddr)
	if err != nil {
		log.Errorln("GetChannelSeqNumsFromSimplexState err", err, cid.Hex())
		return nil, nil, common.ErrSimplexStateNotFound
	}
	seqnum := simplex.GetSeqNum()
	seqNums := common.ChannelSeqNums{
		LastAcked: seqnum,
		LastUsed:  seqnum,
		Base:      seqnum,
	}

	return simplex, &seqNums, nil
}
