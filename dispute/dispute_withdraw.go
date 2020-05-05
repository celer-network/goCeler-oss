// Copyright 2018-2020 Celer Network

package dispute

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/structs"
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
)

func (p *Processor) IntendWithdraw(cidFrom ctype.CidType, amount *big.Int, cidTo ctype.CidType) error {
	log.Infoln("Intend withdraw", cidFrom.Hex(), amount)
	state, found, err := p.dal.GetChanState(cidFrom)
	if err != nil {
		return fmt.Errorf("%x IntendWithdraw err, err %w", cidFrom, err)
	}
	if !found {
		return fmt.Errorf("%x IntendWithdraw err, channel not found", cidFrom)
	}
	if state != structs.ChanState_OPENED {
		return fmt.Errorf("%x IntendWithdraw err, invalid channel state %s", cidFrom, fsm.ChanStateName(state))
	}
	receiver, _, _, _, err := ledgerview.GetOnChainWithdrawIntent(cidFrom, p.nodeConfig)
	if receiver != ctype.ZeroAddr {
		return fmt.Errorf("previous withdraw still pending")
	}

	blkNum := p.monitorService.GetCurrentBlockNumber().Uint64()
	balance, err := ledgerview.GetBalance(p.dal, cidFrom, p.nodeConfig.GetOnChainAddr(), blkNum)
	if err != nil {
		log.Error(err)
		return err
	}
	if balance.MyFree.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance: %s", balance.MyFree)
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
			chanLedger := p.nodeConfig.GetLedgerContractOf(cidFrom)
			if chanLedger == nil {
				return nil, fmt.Errorf("Fail to get ledger for channel: %x", cidFrom)
			}
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(chanLedger.GetAddr(), transactor)
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
	if receiver != p.nodeConfig.GetOnChainAddr() {
		err2 := fmt.Errorf("withdraw receiver not match %s %x", ctype.Addr2Hex(receiver), p.nodeConfig.GetOnChainAddr())
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
			chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
			if chanLedger == nil {
				return nil, fmt.Errorf("Fail to get ledger for channel: %x", cid)
			}
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(chanLedger.GetAddr(), transactor)
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
		err2 := fmt.Errorf("SyncOnChainBalance error: %w", err)
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
			chanLedger := p.nodeConfig.GetLedgerContractOf(cid)
			if chanLedger == nil {
				return nil, fmt.Errorf("Fail to get ledger for channel: %x", cid)
			}
			contract, err2 :=
				ledger.NewCelerLedgerTransactor(chanLedger.GetAddr(), transactor)
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

func (p *Processor) monitorNoncooperativeWithdrawEvent(ledgerContract chain.Contract) {
	_, monErr := p.monitorService.Monitor(
		event.IntendWithdraw,
		ledgerContract,
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			// CAVEAT!!!: suppose we have the same struct of event.
			// If event struct changes, this monitor does not work.
			e := &ledger.CelerLedgerIntendWithdraw{}
			if err := ledgerContract.ParseEvent(event.IntendWithdraw, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			txHash := fmt.Sprintf("%x", eLog.TxHash)
			log.Infoln("Seeing IntendWithdraw event, cid:", cid.Hex(), "receiver", ctype.Addr2Hex(e.Receiver),
				"amount", e.Amount, "tx hash:", txHash, "callback id:", id, "blkNum:", eLog.BlockNumber)
			_, exist, err := p.dal.GetChanState(cid)
			if err != nil {
				log.Error(err)
				return
			}
			if exist {
				// OSP always veto withdraw if receiver is not itself
				if e.Receiver != p.nodeConfig.GetOnChainAddr() {
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
		ledgerContract,
		p.monitorService.GetCurrentBlockNumber(),
		nil,   /*endBlock*/
		false, /*quickCatch*/
		false, /*reset*/
		func(id monitor.CallbackID, eLog types.Log) {
			// CAVEAT!!!: suppose we have the same struct of event.
			// If event struct changes, this monitor does not work.
			e := &ledger.CelerLedgerConfirmWithdraw{}
			if err := ledgerContract.ParseEvent(event.ConfirmWithdraw, eLog, e); err != nil {
				log.Error(err)
				return
			}
			cid := ctype.CidType(e.ChannelId)
			txHash := fmt.Sprintf("%x", eLog.TxHash)
			log.Infoln("Seeing ConfirmWithdraw event, cid:", cid.Hex(), "receiver", ctype.Addr2Hex(e.Receiver),
				"amount", e.WithdrawnAmount, "tx hash:", txHash, "callback id:", id, "blkNum:", eLog.BlockNumber)
			peer, exist, err := p.dal.GetChanPeer(cid)
			if err != nil {
				log.Error(err, cid.Hex())
				return
			}
			if exist {
				self := p.nodeConfig.GetOnChainAddr()
				receiver := e.Receiver
				if receiver != self && receiver != peer {
					return
				}
				if len(e.Deposits) != 2 || len(e.Withdrawals) != 2 {
					log.Error("on chain balances length not match")
					return
				}
				var myIndex int
				if bytes.Compare(self.Bytes(), peer.Bytes()) < 0 {
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
					balance, found, err2 := tx.GetOnChainBalance(cid)
					if err2 != nil {
						return fmt.Errorf("GetOnChainBalance err %w", err2)
					}
					if !found {
						return fmt.Errorf("GetOnChainBalance err %w", common.ErrChannelNotFound)
					}
					onChainBalance.PendingWithdrawal = balance.PendingWithdrawal
					return tx.UpdateOnChainBalance(cid, onChainBalance)
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
