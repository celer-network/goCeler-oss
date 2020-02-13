// Copyright 2018-2019 Celer Network

package ledgerview

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/structs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func GetOnChainChannelStatus(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (uint8, error) {
	contract, err := ledger.NewCelerLedgerCaller(
		nodeConfig.GetLedgerContract().GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return 0, fmt.Errorf("GetOnChainChannelStatus new caller failed %s", err)
	}
	status, err := contract.GetChannelStatus(&bind.CallOpts{}, cid)
	if err != nil {
		return status, fmt.Errorf("GetOnChainChannelStatus get status failed %s", err)
	}
	return status, nil
}

func GetOnChainSettleFinalizedTime(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (*big.Int, error) {
	contract, err := ledger.NewCelerLedgerCaller(
		nodeConfig.GetLedgerContract().GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return big.NewInt(0), fmt.Errorf("GetOnChainSettleFinalizedTime new caller failed %s", err)
	}
	time, err := contract.GetSettleFinalizedTime(&bind.CallOpts{}, cid)
	if err != nil {
		return time, fmt.Errorf("GetOnChainSettleFinalizedTime get status failed %s", err)
	}
	return time, nil
}

func GetOnChainWithdrawIntent(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (
	ctype.Addr, *big.Int, uint64, ctype.CidType, error) {

	contract, err := ledger.NewCelerLedgerCaller(
		nodeConfig.GetLedgerContract().GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return ctype.ZeroAddr, big.NewInt(0), 0, ctype.ZeroCid, fmt.Errorf("GetWithdrawIntent new caller failed %s", err)
	}
	receiver, amount, requestTime, recipientCid, err := contract.GetWithdrawIntent(&bind.CallOpts{}, cid)
	if err != nil {
		return ctype.ZeroAddr, big.NewInt(0), 0, ctype.ZeroCid, fmt.Errorf("GetWithdrawIntent new caller failed %s", err)
	}
	return receiver, amount, requestTime.Uint64(), recipientCid, nil
}

func GetOnChainDisputeTimeout(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (uint64, error) {
	contract, err := ledger.NewCelerLedgerCaller(
		nodeConfig.GetLedgerContract().GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return 0, fmt.Errorf("GetOnChainSettleFinalizedTime new caller failed %s", err)
	}
	timeout, err := contract.GetDisputeTimeout(&bind.CallOpts{}, cid)
	if err != nil {
		return 0, fmt.Errorf("GetOnChainSettleFinalizedTime get status failed %s", err)
	}
	return timeout.Uint64(), nil
}

// SyncOnChainBalance updates local on-chain balances for the given channel
func SyncOnChainBalance(
	dal *storage.DAL,
	cid ctype.CidType,
	nodeConfig common.GlobalNodeConfig) error {
	log.Debugln("sync onchain balance", cid.Hex())
	contract, err := ledger.NewCelerLedgerCaller(
		nodeConfig.GetLedgerContract().GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return err
	}
	peers, deposits, withdrawals, err := contract.GetBalanceMap(nil, cid)
	if err != nil {
		return err
	}

	if len(peers) != 2 || len(deposits) != 2 || len(withdrawals) != 2 {
		return errors.New("on chain balances length not match")
	}
	myAddr := nodeConfig.GetOnChainAddr()
	var myIndex int
	if peers[0] == ctype.Hex2Addr(myAddr) {
		myIndex = 0
	} else if peers[1] == ctype.Hex2Addr(myAddr) {
		myIndex = 1
	} else {
		return errors.New("on chain balances address not match")
	}

	onChainBalance := &structs.OnChainBalance{
		MyDeposit:      deposits[myIndex],
		MyWithdrawal:   withdrawals[myIndex],
		PeerDeposit:    deposits[1-myIndex],
		PeerWithdrawal: withdrawals[1-myIndex],
	}
	updateBalanceTx := func(tx *storage.DALTx, args ...interface{}) error {
		exist, err2 := tx.HasOnChainBalance(cid)
		if err2 != nil {
			return fmt.Errorf("HasOnChainBalance err %s", err2)
		}
		if exist {
			balance, err3 := tx.GetOnChainBalance(cid)
			if err3 != nil {
				return fmt.Errorf("GetOnChainBalance err %s", err3)
			}
			onChainBalance.PendingWithdrawal = balance.PendingWithdrawal
		}
		return tx.PutOnChainBalance(cid, onChainBalance)
	}
	return dal.Transactional(updateBalanceTx)
}
