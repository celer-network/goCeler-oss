// Copyright 2018-2019 Celer Network

package event

import (
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	CooperativeWithdraw = "CooperativeWithdraw"
	Deploy              = "Deploy"
	Deposit             = "Deposit"
	IntendSettle        = "IntendSettle"
	OpenChannel         = "OpenChannel"
	ConfirmSettle       = "ConfirmSettle"
	IntendWithdraw      = "IntendWithdraw"
	ConfirmWithdraw     = "ConfirmWithdraw"
	VetoWithdraw        = "VetoWithdraw"
	RouterUpdated       = "RouterUpdated"
)

type OpenChannelCallback interface {
	HandleOpenChannelFinish(cid ctype.CidType)
	HandleOpenChannelErr(e *common.E)
}
type OnNewStreamCallback interface {
	HandleNewCelerStream(addr []byte)
}
type OnReceivingTokenCallback interface {
	HandleReceivingStart(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any)
	HandleReceivingDone(
		payID ctype.PayIDType,
		pay *entity.ConditionalPay,
		note *any.Any,
		reason rpc.PaymentSettleReason)
}
type OnSendingTokenCallback interface {
	HandleSendComplete(
		payID ctype.PayIDType,
		pay *entity.ConditionalPay,
		note *any.Any,
		reason rpc.PaymentSettleReason)
	HandleDestinationUnreachable(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any)
	HandleSendFail(payID ctype.PayIDType, pay *entity.ConditionalPay, note *any.Any, errMsg string)
}
