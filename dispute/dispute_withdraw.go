// Copyright 2018-2019 Celer Network

package dispute

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/celer-network/goCeler-oss/chain/channel-eth-go/ledger"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/common/structs"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/ledgerview"
	"github.com/celer-network/goCeler-oss/metrics"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/celer-network/goCeler-oss/storage"
	"github.com/celer-network/goCeler-oss/transactor"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

func (p *Processor) IntendWithdraw(cidFrom ctype.CidType, amount *big.Int, cidTo ctype.CidType) error {
	log.Infoln("Intend withdraw", cidFrom.Hex(), amount)
	receiptChan := make(chan *types.Receipt, 1)
	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				receiptChan <- receipt
			},
		},
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(p.nodeConfig.GetLedgerContract().GetAddr(), transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.IntendWithdraw(opts, cidFrom, amount, cidTo)
		})
	if err != nil {
		log.Error(err)
		return err
	}
	receipt := <-receiptChan
	if receipt.Status != types.ReceiptStatusSuccessful {
		err2 := fmt.Errorf("IntendWithdraw transaction 0x%x failed", receipt.TxHash.String())
		log.Error(err2)
		return err2
	}
	log.Infof("IntendWithdraw transaction 0x%x succeeded", receipt.TxHash.String())
	return nil
}

func (p *Processor) ConfirmWithdraw(cid ctype.CidType) error {
	log.Infoln("Confirm withdraw", cid.Hex())
	receiver, _, requestTime, _, err := ledgerview.GetOnChainWithdrawIntent(cid, p.nodeConfig)
	if err != nil {
		log.Error("GetOnChainWithdrawIntent failed", err)
		return err
	}
	if receiver != ctype.Hex2Addr(p.nodeConfig.GetOnChainAddr()) {
		err2 := fmt.Errorf("withdraw receiver not match %s %s", ctype.Addr2Hex(receiver), p.nodeConfig.GetOnChainAddr())
		log.Error(err2)
		return err2
	}
	disputeTimeout, err := ledgerview.GetOnChainDisputeTimeout(cid, p.nodeConfig)
	if err != nil {
		log.Error("GetOnChainDisputeTimeout failed", err)
		return err
	}
	if p.monitorService.GetCurrentBlockNumber().Uint64() < requestTime+disputeTimeout {
		err2 := fmt.Errorf("withdraw disput timeout not reached")
		log.Error(err2)
		return err2
	}
	receiptChan := make(chan *types.Receipt, 1)
	_, err = p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				receiptChan <- receipt
			},
		},
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(p.nodeConfig.GetLedgerContract().GetAddr(), transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.ConfirmWithdraw(opts, cid)
		})
	if err != nil {
		log.Error(err)
		return err
	}
	receipt := <-receiptChan
	if receipt.Status != types.ReceiptStatusSuccessful {
		err2 := fmt.Errorf("ConfirmWithdraw transaction 0x%x failed", receipt.TxHash.String())
		log.Error(err2)
		return err2
	}
	log.Infof("ConfirmWithdraw transaction 0x%x succeeded", receipt.TxHash.String())
	err = ledgerview.SyncOnChainBalance(p.dal, cid, p.nodeConfig)
	if err != nil {
		err2 := fmt.Errorf("SyncOnChainBalance error: %s", err)
		log.Error(err2)
		return err2
	}
	return nil
}

func (p *Processor) VetoWithdraw(cid ctype.CidType) error {
	log.Infoln("Veto withdraw", cid.Hex())
	receiptChan := make(chan *types.Receipt, 1)
	_, err := p.transactor.Transact(
		&transactor.TransactionMinedHandler{
			OnMined: func(receipt *types.Receipt) {
				receiptChan <- receipt
			},
		},
		big.NewInt(0),
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(p.nodeConfig.GetLedgerContract().GetAddr(), transactor)
			if err2 != nil {
				return nil, err2
			}
			return contract.VetoWithdraw(opts, cid)
		})
	if err != nil {
		log.Error(err)
		return err
	}
	receipt := <-receiptChan
	if receipt.Status != types.ReceiptStatusSuccessful {
		err2 := fmt.Errorf("VetoWithdraw transaction 0x%x failed", receipt.TxHash.String())
		log.Error(err2)
		return err2
	}
	log.Infof("VetoWithdraw transaction 0x%x succeeded", receipt.TxHash.String())
	return nil
}

func (p *Processor) monitorNoncooperativeWithdrawEvent() {
	_, monErr := p.monitorService.Monitor(
		event.IntendWithdraw,
		p.nodeConfig.GetLedgerContract(),
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerIntendWithdraw{}
			if err := p.nodeConfig.GetLedgerContract().ParseEvent(event.IntendWithdraw, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			txHash := fmt.Sprintf("%x", eLog.TxHash)
			log.Infoln("Seeing IntendWithdraw event, cid:", cid.Hex(), "receiver", ctype.Addr2Hex(e.Receiver),
				"amount", e.Amount, "tx hash:", txHash, "callback id:", id, "blkNum:", eLog.BlockNumber)
			exist, err := p.dal.HasChannelState(cid)
			if err != nil {
				log.Error(err)
				return
			}
			if exist {
				// OSP always veto withdraw if receiver is not itself
				if e.Receiver != ctype.Hex2Addr(p.nodeConfig.GetOnChainAddr()) {
					p.VetoWithdraw(cid)
					return
				}
			} else {
				return
			}
			metrics.IncDisputeWithdrawEventCnt(event.IntendWithdraw)
		})
	if monErr != nil {
		log.Error(monErr)
	}

	_, monErr = p.monitorService.Monitor(
		event.ConfirmWithdraw,
		p.nodeConfig.GetLedgerContract(),
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			e := &ledger.CelerLedgerConfirmWithdraw{}
			if err := p.nodeConfig.GetLedgerContract().ParseEvent(event.ConfirmWithdraw, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			txHash := fmt.Sprintf("%x", eLog.TxHash)
			log.Infoln("Seeing ConfirmWithdraw event, cid:", cid.Hex(), "receiver", ctype.Addr2Hex(e.Receiver),
				"amount", e.WithdrawnAmount, "tx hash:", txHash, "callback id:", id, "blkNum:", eLog.BlockNumber)
			exist, err := p.dal.HasChannelState(cid)
			if err != nil {
				log.Error(err)
				return
			}
			if exist {
				selfStr := p.nodeConfig.GetOnChainAddr()
				peerStr, err := p.dal.GetPeer(cid)
				if err != nil {
					log.Error(err)
					return
				}
				receiver := e.Receiver
				if receiver != ctype.Hex2Addr(selfStr) && receiver != ctype.Hex2Addr(peerStr) {
					return
				}
				if len(e.Deposits) != 2 || len(e.Withdrawals) != 2 {
					log.Error("on chain balances length not match")
					return
				}
				var myIndex int
				if strings.Compare(selfStr, peerStr) < 0 {
					myIndex = 0
				} else {
					myIndex = 1
				}
				onChainBalance := &structs.OnChainBalance{
					MyDeposit:      e.Deposits[myIndex],
					MyWithdrawal:   e.Withdrawals[myIndex],
					PeerDeposit:    e.Deposits[1-myIndex],
					PeerWithdrawal: e.Withdrawals[1-myIndex],
				}
				updateBalanceTx := func(tx *storage.DALTx, args ...interface{}) error {
					balance, err2 := tx.GetOnChainBalance(cid)
					if err2 != nil {
						return fmt.Errorf("GetOnChainBalance err %s", err2)
					}
					onChainBalance.PendingWithdrawal = balance.PendingWithdrawal
					return tx.PutOnChainBalance(cid, onChainBalance)
				}
				if err := p.dal.Transactional(updateBalanceTx); err != nil {
					log.Error(err)
					return
				}
			} else {
				return
			}
			metrics.IncDisputeWithdrawEventCnt(event.ConfirmWithdraw)
		})
	if monErr != nil {
		log.Error(monErr)
	}
}
