// Copyright 2018-2020 Celer Network

// payment related interface for celer sdk

package celersdk

import (
	"errors"

	"github.com/celer-network/goCeler/celersdkintf"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/any"
)

const cPayTimeout = 50 // timeout in blocknum for cpay ie. no app channel condition

// noteTypeUrl should be type url of any.Any.
// noteStr should be string representation of []byte in note (any.Any)
func (mc *Client) SendETH(receiver string, amtWei string, noteTypeUrl string, noteValueByte []byte) (string, error) {
	return mc.SendToken(nil, receiver, amtWei, noteTypeUrl, noteValueByte)
}

// SendETH sends ERC20/ETH token to receiver. Caller can optionally add a note in the pay.
func (mc *Client) SendToken(tk *Token, receiver string, amtWei string, noteTypeUrl string, noteValueByte []byte) (string, error) {

	xfer := createXfer(tk, receiver, amtWei)
	note := &any.Any{
		TypeUrl: noteTypeUrl,
		Value:   noteValueByte,
	}
	payID, err := mc.c.AddBooleanPay(
		xfer, []*entity.Condition{}, mc.c.GetCurrentBlockNumberUint64()+cPayTimeout, note, 0)
	if err != nil {
		log.Errorln("SendToken:", err)
		return ctype.ZeroPayIDHex, err
	}
	ret := ctype.PayID2Hex(payID)
	log.Debugln("Sent pay:", ret)
	return ret, nil
}

func (mc *Client) SendETHWithCondition(receiver string, amtWei string, cond *BooleanCondition) (string, error) {
	return mc.SendTokenWithCondition(nil, receiver, amtWei, cond)
}

// When should we call onSent? or do we need a new callback func?
func (mc *Client) SendTokenWithCondition(tk *Token, receiver string, amtWei string, cond *BooleanCondition) (string, error) {
	xfer := createXfer(tk, receiver, amtWei)
	timeout := cPayTimeout
	if cond.TimeoutBlockNum > 0 {
		timeout = cond.TimeoutBlockNum
	}
	condition, err := bc2c(cond)
	if err != nil {
		log.Errorln("SendTokenWithCondition:", err)
		return ctype.ZeroPayIDHex, err
	}
	payID, err := mc.c.AddBooleanPay(
		xfer, []*entity.Condition{condition}, mc.c.GetCurrentBlockNumberUint64()+uint64(timeout), nil /*note*/, 0)
	if err != nil {
		log.Errorln("SendTokenWithCondition:", err)
		return ctype.ZeroPayIDHex, err
	}
	ret := ctype.PayID2Hex(payID)
	log.Debugln("Sent pay:", ret)
	return ret, nil
}

// ConfirmPay settles the condpay, ie. actually paid to pay dest
func (mc *Client) ConfirmPay(payID string) error {
	return mc.c.ConfirmBooleanPay(ctype.Hex2PayID(payID))
}

// RejectPay cancels the pay, ie. ask OSP and pay src to not pay
func (mc *Client) RejectPay(payID string) error {
	return mc.c.RejectBooleanPay(ctype.Hex2PayID(payID))
}

// RemoveExpiredPays clears pending pays that have expired, if tk is nil, means ETH
func (mc *Client) RemoveExpiredPays(tk *Token) error {
	token := sdkToken2entityToken(tk)
	return mc.c.SettleExpiredPays(token)
}

// ResolvePayOnChain settles the payment onchain and receives the payment from OSP
func (mc *Client) ResolvePayOnChain(payID string) error {
	err := mc.c.ResolveCondPayOnChain(ctype.Hex2PayID(payID))
	if err != nil {
		return err
	}
	return mc.c.SettleOnChainResolvedPay(ctype.Hex2PayID(payID))
}

// ConfirmOnChainResolvedPays confirms pays that have been onchain resolved, if tk is nil, means ETH
func (mc *Client) ConfirmOnChainResolvedPays(tk *Token) error {
	token := sdkToken2entityToken(tk)
	return mc.c.ConfirmOnChainResolvedPays(token)
}

// Get incoming payment status code
func (mc *Client) GetIncomingPaymentStatus(payId string) int {
	return mc.c.GetIncomingPaymentStatus(ctype.Hex2PayID(payId))
}

// Get outgoing payment status code
func (mc *Client) GetOutgoingPaymentStatus(payId string) int {
	return mc.c.GetOutgoingPaymentStatus(ctype.Hex2PayID(payId))
}

func (mc *Client) GetOnChainPaymentInfo(paymentID string) (*OnChainPaymentInfo, error) {
	amount, resolveDeadline, err := mc.c.GetCondPayInfoFromRegistry(ethcommon.HexToHash(paymentID))
	if err != nil {
		return nil, err
	}
	return &OnChainPaymentInfo{Amount: amount.String(), ResolveDeadline: resolveDeadline}, nil
}

func (mc *Client) ResolveIncomingPaymentOnChain(payId string) error {
	return mc.c.ResolveCondPayOnChain(ctype.Hex2PayID(payId))
}

func (mc *Client) SettleOnChainResolvedIncomingPayment(payId string) error {
	return mc.c.SettleOnChainResolvedPay(ctype.Hex2PayID(payId))
}

func (mc *Client) SendConditionalPayment(
	tokenInfo *TokenInfo,
	destination string,
	amount string,
	transferLogicType TransferLogicType,
	conditions []*Condition,
	timeout int64,
	note *any.Any) (string, error) {
	if transferLogicType != transferLogicTypeBooleanAnd {
		return "", errors.New("Unsupported transfer logic type")
	}
	token := &entity.TokenInfo{
		TokenType:    entity.TokenType(int32(tokenInfo.TokenType)),
		TokenAddress: ctype.Hex2Bytes(tokenInfo.TokenAddress),
	}
	transfer := &entity.TokenTransfer{
		Token: token,
		Receiver: &entity.AccountAmtPair{
			Account: ctype.Hex2Bytes(destination),
			Amt:     utils.Wei2BigInt(amount).Bytes(),
		},
	}
	entityConditions := make([]*entity.Condition, len(conditions))
	for i, condition := range conditions {
		entityConditions[i] = conditionToEntityCondition(condition)
	}
	payID, err := mc.c.AddBooleanPay(
		transfer,
		entityConditions,
		mc.c.GetCurrentBlockNumberUint64()+uint64(timeout),
		note, 0)
	if err != nil {
		log.Error(err)
		return ctype.ZeroPayIDHex, err
	}
	ret := ctype.PayID2Hex(payID)
	log.Debugln("Sent pay:", ret)
	return ret, nil
}

func (mc *Client) SettleExpiredPayments(tokenInfo *TokenInfo) error {
	return mc.c.SettleExpiredPays(&entity.TokenInfo{
		TokenType:    entity.TokenType(int32(tokenInfo.TokenType)),
		TokenAddress: ctype.Hex2Bytes(tokenInfo.TokenAddress),
	})
}

// GetPayment returns the related payment info of a specified payment ID
func (mc *Client) GetPayment(paymentID string) (*celersdkintf.Payment, error) {
	return mc.c.GetPayment(ethcommon.HexToHash(paymentID))
}

// GetAllPayments returns all payments info.
// **CAUTION**: This function costs heavy lookup on several tables and joins
// information from those tables, please take performance into consideration
// before using this.
// PaymentList.PayList is list of all payments. But due to gomobile limitation (no return list).
// mobile app needs to do following
// payList = GetAllPayments()
// for i=0; i<payList.Length; i++ {
//     pay = payList.Get(i)
// }
func (mc *Client) GetAllPayments() (*celersdkintf.PaymentList, error) {
	allpays, err := mc.c.GetAllPayments()
	if err != nil {
		return nil, err
	}

	return &celersdkintf.PaymentList{
		Length:  len(allpays),
		PayList: allpays,
	}, nil
}
