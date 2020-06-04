// Copyright 2018-2020 Celer Network

package utils

import (
	"fmt"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goutils/eth"
	"github.com/golang/protobuf/proto"
)

func PrintByteArrays(array [][]byte) string {
	s := ""
	for _, a := range array {
		s += fmt.Sprintf("%x,", a)
	}
	return s
}

func PrintPayIdList(paylist *entity.PayIdList) string {
	length := len(paylist.GetPayIds())
	if length > 0 {
		pl := fmt.Sprintf("%d pending_pays:", len(paylist.GetPayIds()))
		pl += PrintByteArrays(paylist.GetPayIds())
		return pl
	}
	return "0 pending_pays,"
}

func PrintSimplexChannel(simplex *entity.SimplexPaymentChannel) string {
	return fmt.Sprintf("cid: %x, from: %x, seq: %d, transfer: [%s], %s last_pay_deadline: %d, pending_amount: %s",
		simplex.GetChannelId(),
		simplex.GetPeerFrom(),
		simplex.GetSeqNum(),
		PrintTokenTransfer(simplex.GetTransferToPeer()),
		PrintPayIdList(simplex.GetPendingPayIds()),
		simplex.GetLastPayResolveDeadline(),
		big.NewInt(0).SetBytes(simplex.GetTotalPendingAmount()).String())
}

func PrintTokenTransfer(transfer *entity.TokenTransfer) string {
	return fmt.Sprintf("%s, amount:%s",
		PrintTokenInfo(transfer.GetToken()),
		big.NewInt(0).SetBytes(transfer.GetReceiver().GetAmt()).String())
}

func PrintTokenInfo(token *entity.TokenInfo) string {
	if token.GetTokenType() == entity.TokenType_ETH {
		return "token_type: ETH"
	} else if token.GetTokenType() == entity.TokenType_ERC20 {
		return fmt.Sprintf("token_address: %x", token.GetTokenAddress())
	}
	return "invalid_token_type"
}

func PrintAccountAmtPair(pair *entity.AccountAmtPair) string {
	return fmt.Sprintf("acct %x amt %s", pair.GetAccount(), big.NewInt(0).SetBytes(pair.GetAmt()).String())
}

func PrintTokenDistribution(dist *entity.TokenDistribution) string {
	token := PrintTokenInfo(dist.GetToken())
	distribution := ""
	for i, d := range dist.GetDistribution() {
		distribution += PrintAccountAmtPair(d)
		if i != len(dist.GetDistribution())-1 {
			distribution += " "
		}
	}
	return fmt.Sprintf("%s, %s", token, distribution)
}

func PrintChannelInitializer(initializer *entity.PaymentChannelInitializer) string {
	return fmt.Sprintf("init_distribution: [%s], open_deadline: %d, dispute_timeout: %d, msg_value_receiver: %d",
		PrintTokenDistribution(initializer.GetInitDistribution()),
		initializer.GetOpenDeadline(),
		initializer.GetDisputeTimeout(),
		initializer.GetMsgValueReceiver())
}

func PrintCondition(cond *entity.Condition) string {
	if cond.GetConditionType() == entity.ConditionType_HASH_LOCK {
		return fmt.Sprintf("<hashlock: %x>", cond.GetHashLock())
	} else if cond.GetConditionType() == entity.ConditionType_DEPLOYED_CONTRACT {
		return fmt.Sprintf("<deployed_addr: %x, args: f:%x o:%x>",
			cond.GetDeployedContractAddress(), cond.GetArgsQueryFinalization(), cond.GetArgsQueryOutcome())
	} else if cond.GetConditionType() == entity.ConditionType_VIRTUAL_CONTRACT {
		return fmt.Sprintf("<virtual_addr: %x, args: f:%x o:%x>",
			cond.GetVirtualContractAddress(), cond.GetArgsQueryFinalization(), cond.GetArgsQueryOutcome())
	}
	return "invalid_condition_type"
}

func PrintConditions(conds []*entity.Condition) string {
	condstr := ""
	for i, c := range conds {
		condstr += PrintCondition(c)
		if i != len(conds)-1 {
			condstr += ","
		}
	}
	return condstr
}

func PrintTransferFunc(transfer *entity.TransferFunction) string {
	if transfer.GetLogicType() == entity.TransferFunctionType_BOOLEAN_AND {
		return PrintTokenTransfer(transfer.GetMaxTransfer())
	}
	return fmt.Sprintf("invalid_transfer_type_%d", transfer.GetLogicType())
}

func PrintConditionalPay(pay *entity.ConditionalPay) string {
	return fmt.Sprintf("timestamp: %d, src:%x, dst:%x, conditions: [%s], transfer: [%s], deadline:%d, resolve_timeout:%d, pay_resolver:%x",
		pay.GetPayTimestamp(),
		pay.GetSrc(),
		pay.GetDest(),
		PrintConditions(pay.GetConditions()),
		PrintTransferFunc(pay.GetTransferFunc()),
		pay.GetResolveDeadline(),
		pay.GetResolveTimeout(),
		pay.GetPayResolver())
}

func PrintCooperativeWithdrawInfo(withdrawal *entity.CooperativeWithdrawInfo) string {
	info := fmt.Sprintf("cid: %s, seq: %d, receiver: %s, deadline: %d",
		ctype.Bytes2Hex(withdrawal.GetChannelId()),
		withdrawal.GetSeqNum(),
		PrintAccountAmtPair(withdrawal.GetWithdraw()),
		withdrawal.GetWithdrawDeadline())
	if ctype.Bytes2Cid(withdrawal.GetRecipientChannelId()) != ctype.ZeroCid {
		info += fmt.Sprintf(", recipient cid: %s", ctype.Bytes2Hex(withdrawal.GetRecipientChannelId()))
	}
	return info
}

func PrintRoutingUpdate(update *rpc.RoutingUpdate) string {
	channels := ""
	for _, ch := range update.GetChannels() {
		channels += ch.Cid + ":" + ch.Balance + ","
	}
	return fmt.Sprintf("origin:%s, ts:%s, chs:%d:%s",
		update.GetOrigin(),
		time.Unix(int64(update.GetTs()), 0).UTC().Format("2006-01-02 15:04:05"),
		len(update.GetChannels()), channels)
}

func PrintPayPath(payPath *rpc.PayPath, payID ctype.PayIDType) string {
	path := ""
	nhop := len(payPath.GetHops())
	if nhop > 1 {
		path = fmt.Sprintf("%d reported hops | ", nhop)
	}
	for i, signedHop := range payPath.GetHops() {
		signer := eth.RecoverSigner(signedHop.GetPayHopBytes(), signedHop.GetSig())
		var payHop rpc.PayHop
		err := proto.Unmarshal(signedHop.GetPayHopBytes(), &payHop)
		if err != nil {
			return fmt.Sprintf("%s proto.Unmarshal err: %s", path, err)
		}
		hop := fmt.Sprintf("signer:%x, prev:%x, next:%x", signer, payHop.GetPrevHopAddr(), payHop.GetNextHopAddr())
		if payHop.GetErr() != nil {
			hop += fmt.Sprintf(", err:%s", payHop.GetErr())
		}
		if ctype.Bytes2PayID(payHop.GetPayId()) != payID {
			hop += fmt.Sprintf(", invalid_payid:%x", payHop.GetPayId())
		}
		path += hop
		if i < nhop-1 {
			path += " | "
		}
	}
	return path
}
