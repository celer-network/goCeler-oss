// Copyright 2018-2020 Celer Network

package ledgerview

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const (
	OnChainStatus_UNINITIALIZED uint8 = 0
	OnChainStatus_OPERABLE      uint8 = 1
	OnChainStatus_SETTLING      uint8 = 2
	OnChainStatus_CLOSED        uint8 = 3
	OnChainStatus_MIGRATED      uint8 = 4
)

func ChanStatusName(status uint8) string {
	switch status {
	case OnChainStatus_UNINITIALIZED:
		return "Uninitialized"
	case OnChainStatus_OPERABLE:
		return "Operable"
	case OnChainStatus_SETTLING:
		return "Settling"
	case OnChainStatus_CLOSED:
		return "Close"
	case OnChainStatus_MIGRATED:
		return "Migrated"
	default:
		return "Invalid"
	}
}

func GetOnChainChannelStatusOnLedger(cid ctype.CidType, nodeConfig common.GlobalNodeConfig, ledgerAddr ctype.Addr) (uint8, error) {
	contract, err := ledger.NewCelerLedgerCaller(ledgerAddr, nodeConfig.GetEthConn())
	if err != nil {
		return 0, fmt.Errorf("GetOnChainChannelStatusOnLedger new caller failed %w", err)
	}
	status, err := contract.GetChannelStatus(&bind.CallOpts{}, cid)
	if err != nil {
		return status, fmt.Errorf("GetOnChainChannelStatusLedger get status failed %w", err)
	}
	return status, nil
}

func GetOnChainChannelStatus(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (uint8, error) {
	chanLedger := nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		return 0, fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	return GetOnChainChannelStatusOnLedger(cid, nodeConfig, chanLedger.GetAddr())
}

func GetOnChainSettleFinalizedTime(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (*big.Int, error) {
	chanLedger := nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		return big.NewInt(0), fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	contract, err := ledger.NewCelerLedgerCaller(chanLedger.GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return big.NewInt(0), fmt.Errorf("GetOnChainSettleFinalizedTime new caller failed %w", err)
	}
	time, err := contract.GetSettleFinalizedTime(&bind.CallOpts{}, cid)
	if err != nil {
		return time, fmt.Errorf("GetOnChainSettleFinalizedTime get status failed %w", err)
	}
	return time, nil
}

func GetOnChainWithdrawIntent(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (
	ctype.Addr, *big.Int, uint64, ctype.CidType, error) {
	chanLedger := nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		return ctype.ZeroAddr, big.NewInt(0), 0, ctype.ZeroCid, fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	contract, err := ledger.NewCelerLedgerCaller(chanLedger.GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return ctype.ZeroAddr, big.NewInt(0), 0, ctype.ZeroCid, fmt.Errorf("GetWithdrawIntent new caller failed %w", err)
	}
	receiver, amount, requestTime, recipientCid, err := contract.GetWithdrawIntent(&bind.CallOpts{}, cid)
	if err != nil {
		return ctype.ZeroAddr, big.NewInt(0), 0, ctype.ZeroCid, fmt.Errorf("GetWithdrawIntent new caller failed %w", err)
	}
	return receiver, amount, requestTime.Uint64(), recipientCid, nil
}

func GetOnChainDisputeTimeout(cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (uint64, error) {
	chanLedger := nodeConfig.GetLedgerContractOf(cid)
	if chanLedger == nil {
		return 0, fmt.Errorf("Fail to get ledger for channel: %x", cid)
	}
	contract, err := ledger.NewCelerLedgerCaller(chanLedger.GetAddr(), nodeConfig.GetEthConn())
	if err != nil {
		return 0, fmt.Errorf("GetOnChainSettleFinalizedTime new caller failed %w", err)
	}
	timeout, err := contract.GetDisputeTimeout(&bind.CallOpts{}, cid)
	if err != nil {
		return 0, fmt.Errorf("GetOnChainSettleFinalizedTime get status failed %w", err)
	}
	return timeout.Uint64(), nil
}

// SyncOnChainBalance updates local on-chain balances for the given channel
func SyncOnChainBalance(dal *storage.DAL, cid ctype.CidType, nodeConfig common.GlobalNodeConfig) error {
	log.Debugln("sync onchain balance", cid.Hex())
	ledgerAddr, found, err := dal.GetChanLedger(cid)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Can't find cid %x", cid)
	}
	onChainBalance, err := GetOnChainChannelBalance(cid, ledgerAddr, nodeConfig)
	if err != nil {
		return err
	}
	updateBalanceTx := func(tx *storage.DALTx, args ...interface{}) error {
		balance, found, err3 := tx.GetOnChainBalance(cid)
		if err3 != nil {
			return fmt.Errorf("GetOnChainBalance err %w", err3)
		}
		if !found {
			return common.ErrChannelNotFound
		}
		onChainBalance.PendingWithdrawal = balance.PendingWithdrawal
		return tx.UpdateOnChainBalance(cid, onChainBalance)
	}
	return dal.Transactional(updateBalanceTx)
}

// GetOnChainChannelBalance queries the ledger contract for channel balance
func GetOnChainChannelBalance(cid ctype.CidType, ledgerAddr ctype.Addr, nodeConfig common.GlobalNodeConfig) (*structs.OnChainBalance, error) {
	log.Debugln("query contract for channel balance", cid.Hex())
	contract, err := ledger.NewCelerLedgerCaller(ledgerAddr, nodeConfig.GetEthConn())
	if err != nil {
		return nil, err
	}
	peers, deposits, withdrawals, err := contract.GetBalanceMap(nil, cid)
	if err != nil {
		return nil, err
	}

	if len(peers) != 2 || len(deposits) != 2 || len(withdrawals) != 2 {
		return nil, errors.New("on chain balances length not match")
	}
	myAddr := nodeConfig.GetOnChainAddr()
	var myIndex int
	if peers[0] == myAddr {
		myIndex = 0
	} else if peers[1] == myAddr {
		myIndex = 1
	} else {
		return nil, errors.New("on chain balances address not match")
	}
	return &structs.OnChainBalance{
		MyDeposit:      deposits[myIndex],
		MyWithdrawal:   withdrawals[myIndex],
		PeerDeposit:    deposits[1-myIndex],
		PeerWithdrawal: withdrawals[1-myIndex],
	}, nil
}

// GetMigratedTo queries the ledger contract for channel migration info.
// When channel status is "Migrated", it would return the new ledger to which the channel migrated.
// When channel status is not "Migrated", it would return address 0x0.
// Usually call GetOnChainChannelStatus before call this
func GetMigratedTo(dal *storage.DAL, cid ctype.CidType, nodeConfig common.GlobalNodeConfig) (ctype.Addr, error) {
	log.Debugln("query contract for ledger channel migrated to", cid.Hex())
	currentLedger, found, err := dal.GetChanLedger(cid)
	if err != nil {
		return ctype.ZeroAddr, fmt.Errorf("Fail to get ledger for channel %x, err: %w", cid, err)
	}
	if !found {
		return ctype.ZeroAddr, common.ErrChannelNotFound
	}
	// Suppose all the ledger has compatible new caller function
	contract, err := ledger.NewCelerLedgerCaller(currentLedger, nodeConfig.GetEthConn())
	if err != nil {
		return ctype.ZeroAddr, fmt.Errorf("Fail to new ledger caller: %w", err)
	}
	toLedger, err := contract.GetMigratedTo(&bind.CallOpts{}, cid)
	if err != nil {
		return ctype.ZeroAddr, fmt.Errorf("Failt to get migration info: %w", err)
	}

	return toLedger, nil
}

type TxInfo struct {
	From      ctype.Addr
	To        ctype.Addr
	Pending   bool
	FuncName  string
	FuncInput map[string]interface{}
}

func GetOnChainTxByHash(txhash ctype.Hash, nodeConfig common.GlobalNodeConfig) (*TxInfo, error) {
	txInfo := &TxInfo{
		FuncInput: make(map[string]interface{}),
	}
	tx, pending, err := nodeConfig.GetEthConn().TransactionByHash(context.Background(), txhash)
	if err != nil {
		return nil, fmt.Errorf("TransactionByHash err: %w", err)
	}
	txInfo.Pending = pending

	msg, err := tx.AsMessage(ethtypes.NewEIP155Signer(tx.ChainId()))
	if err != nil {
		return nil, fmt.Errorf("AsMessage err: %w", err)
	}
	txInfo.To = *msg.To()
	txInfo.From = msg.From()

	ledgerABI, err := abi.JSON(strings.NewReader((nodeConfig.GetLedgerContract().GetABI())))
	if err != nil {
		return nil, fmt.Errorf("get ABI err: %w", err)
	}
	if len(msg.Data()) < 4 {
		return nil, fmt.Errorf("invalid msg data")
	}
	method, err := ledgerABI.MethodById(msg.Data()[:4])
	if err != nil {
		return nil, fmt.Errorf("MethodById err: %w", err)
	}
	txInfo.FuncName = method.Name

	err = method.Inputs.UnpackIntoMap(txInfo.FuncInput, msg.Data()[4:])
	if err != nil {
		return nil, fmt.Errorf("UnpackIntoMap err: %w", err)
	}
	return txInfo, nil
}
