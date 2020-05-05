// Copyright 2019-2020 Celer Network

package delegate

import (
	"math/big"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

type delegateProcess interface {
	AddBooleanPay(pay *entity.ConditionalPay, note *any.Any) (ctype.PayIDType, error)
	GetCurrentBlockNumber() *big.Int
}

type DelegateManager struct {
	myAddr  ctype.Addr
	dal     *storage.DAL
	process delegateProcess
}

type DelegateEvent struct {
	PayID       ctype.PayIDType
	Pay         *entity.ConditionalPay
	Note        *any.Any
	SendSuccess bool
}

type lumpSum struct {
	amt   *big.Int
	token ctype.Addr
	note  *PayOriginNote
}

func NewDelegateManager(myAddr ctype.Addr, dal *storage.DAL, process delegateProcess) *DelegateManager {
	return &DelegateManager{
		myAddr:  myAddr,
		dal:     dal,
		process: process,
	}
}

func (m *DelegateManager) NotifyNewStream(peer ctype.Addr) error {
	delegatedPays, err := m.dal.GetDelegatedPaysOnStatus(peer, structs.DelegatedPayStatus_RECVD)
	if err != nil {
		log.Errorln("get delegated pays:", err, peer.Hex())
		return err
	}
	if len(delegatedPays) == 0 {
		log.Debugln("no delegated pays for peer", peer.Hex())
		return nil
	}

	lumpsums := make(map[ctype.Addr]*lumpSum)
	for payID, pay := range delegatedPays {
		amt, token := getAmtTokenPair(pay)
		if _, ok := lumpsums[token]; !ok {
			// First time seeing the tokenAddr
			lumpsums[token] = &lumpSum{
				amt:   big.NewInt(0),
				token: token,
				note:  &PayOriginNote{},
			}
		}

		lumpsums[token].amt.Add(lumpsums[token].amt, amt)
		origin := &OriginalPay{
			PayId:  payID.Bytes(),
			PaySrc: token.Bytes(),
			PayAmt: amt.Bytes(),
		}
		lumpsums[token].note.OriginalPays = append(lumpsums[token].note.OriginalPays, origin)
	}

	log.Infoln("delegate sending delegated pay(s) to", ctype.Addr2Hex(peer))
	for _, lumpsum := range lumpsums {
		err = m.sendToken(peer, lumpsum)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *DelegateManager) NotifyPaySendFinalize(in *DelegateEvent) error {
	if in == nil {
		log.Errorln("nil input")
		return common.ErrInvalidArg
	}
	log.Debugln("PaySendComplete:", ctype.PayID2Hex(in.PayID),
		utils.PrintConditionalPay(in.Pay), "note:", in.Note, "success:", in.SendSuccess)
	note := &PayOriginNote{}
	if ptypes.Is(in.Note, note) {
		err := ptypes.UnmarshalAny(in.Note, note)
		if err != nil {
			log.Errorln("unmarshal agent note:", err)
			return err
		}
	} else {
		log.Errorln("Receiving unrecognized pay agent note ", note)
		return common.ErrInvalidArg
	}

	if note.IsRefund {
		log.Debugln("finishing refund for pay:", note)
		return nil
	}
	for _, originalPay := range note.OriginalPays {
		// Iterate pay hash that has been fufilled by the completed pay (lump sump pay).
		payID := ctype.Bytes2PayID(originalPay.GetPayId())
		newStatus := structs.DelegatedPayStatus_RECVD
		if in.SendSuccess {
			newStatus = structs.DelegatedPayStatus_DONE
		}
		err := m.dal.UpdateDelegatedPayStatus(payID, newStatus)
		if err != nil {
			log.Errorln("update delegated pay", err, payID.Hex(), "to status", newStatus)
		}
	}
	return nil
}

func (m *DelegateManager) sendToken(dst ctype.Addr, lumpsum *lumpSum) error {
	transfer := &entity.TokenTransfer{
		Token: utils.GetTokenInfoFromAddress(lumpsum.token),
		Receiver: &entity.AccountAmtPair{
			Account: dst.Bytes(),
			Amt:     lumpsum.amt.Bytes(),
		},
	}
	note, err := ptypes.MarshalAny(lumpsum.note)
	if err != nil {
		log.Errorln(err, note)
		return err
	}
	pay := &entity.ConditionalPay{
		Src:  m.myAddr.Bytes(),
		Dest: dst.Bytes(),
		TransferFunc: &entity.TransferFunction{
			LogicType:   entity.TransferFunctionType_BOOLEAN_AND,
			MaxTransfer: transfer,
		},
		ResolveDeadline: m.process.GetCurrentBlockNumber().Uint64() + config.AdminSendTokenTimeout,
		ResolveTimeout:  config.PayResolveTimeout,
	}

	payID, err := m.process.AddBooleanPay(pay, note)
	if err != nil {
		log.Errorln(err, utils.PrintConditionalPay(pay), note)
		return err
	}
	log.Infoln("send delegated pay", payID.Hex(), "to", dst.Hex())
	return nil
}

func getAmtTokenPair(pay *entity.ConditionalPay) (*big.Int, ctype.Addr) {
	amt := big.NewInt(0).SetBytes(pay.GetTransferFunc().GetMaxTransfer().GetReceiver().GetAmt())
	token := utils.GetTokenAddr(pay.GetTransferFunc().GetMaxTransfer().GetToken())
	return amt, token
}
