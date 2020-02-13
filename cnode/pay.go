// Copyright 2019 Celer Network

package cnode

import (
	"bytes"
	"errors"
	"math/big"
	"math/rand"
	"time"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/common/event"
	"github.com/celer-network/goCeler-oss/config"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/entity"
	"github.com/celer-network/goCeler-oss/fsm"
	"github.com/celer-network/goCeler-oss/ledgerview"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

func newHashLockCond() (*entity.Condition, []byte, []byte) {
	secret, hash := getHashLockBytes()
	cond := &entity.Condition{
		ConditionType: entity.ConditionType_HASH_LOCK,
		HashLock:      hash,
	}
	return cond, secret, hash
}

func getHashLockBytes() (preimg, hash []byte) {
	preimg = make([]byte, 32)
	rand.Read(preimg)
	hash = crypto.Keccak256(preimg)
	return
}

func (c *CNode) OnSendToken(sendCallback event.OnSendingTokenCallback) {
	c.celerMsgDispatcher.OnSendingToken(sendCallback)
}

// Similar to EstablishCondPayOnToken. This will add hash lock condition to pay condition and set time stamp.
func (c *CNode) AddBooleanPay(newPay *entity.ConditionalPay, note *any.Any) (ctype.PayIDType, error) {
	tokenAddr := utils.GetTokenAddrStr(newPay.TransferFunc.MaxTransfer.Token)
	if tokenAddr == "" {
		return ctype.ZeroPayID, common.ErrUnknownTokenType
	}

	newPay.PayTimestamp = uint64(time.Now().UnixNano())
	directPay := c.messager.IsDirectPay(newPay, "")

	// Skip prepending HL if it's already there or for direct-pay.
	skipPrependHL := directPay || (len(newPay.GetConditions()) > 0 && newPay.Conditions[0].GetHashLock() != nil)
	if !skipPrependHL {
		// Prepend HL condition if there is none.
		hlCond, secret, hash := newHashLockCond()
		err := c.dal.PutSecretRegistry(ctype.Bytes2Hex(hash), ctype.Bytes2Hex(secret))
		if err != nil {
			return ctype.ZeroPayID, errors.New("SECRET_EXISTS")
		}
		newPay.Conditions = append([]*entity.Condition{hlCond}, newPay.Conditions...)
	}

	newPay.PayResolver = c.nodeConfig.GetPayResolverContract().GetAddr().Bytes()
	payID := ctype.Pay2PayID(newPay)

	newPayBytes, err := proto.Marshal(newPay)
	if err != nil {
		return ctype.ZeroPayID, err
	}
	err = c.dal.PutConditionalPay(newPayBytes)
	if err != nil {
		return ctype.ZeroPayID, err
	}
	if note != nil {
		err = c.dal.PutPayNote(newPay, note)
		if err != nil {
			return ctype.ZeroPayID, err
		}
	}
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_SEND_TOKEN_API
	logEntry.PayId = ctype.PayID2Hex(ctype.Pay2PayID(newPay))
	logEntry.Dst = ctype.Bytes2Hex(newPay.GetDest())
	logEntry.DirectPay = directPay
	err = c.messager.SendCondPayRequest(newPay, note, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
		payID = ctype.ZeroPayID
	}
	pem.CommitPem(logEntry)
	return payID, err
}

func (c *CNode) ConfirmBooleanPay(payID ctype.PayIDType) error {
	pay, _, err := c.dal.GetConditionalPay(payID)
	if err != nil {
		return err
	}
	amt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.PayId = ctype.PayID2Hex(payID)
	return c.messager.SendOnePaySettleRequest(pay, amt, rpc.PaymentSettleReason_PAY_PAID_MAX, logEntry)
}

func (c *CNode) RejectBooleanPay(payID ctype.PayIDType) error {
	pay, _, err := c.dal.GetConditionalPay(payID)
	if err != nil {
		return err
	}
	if !bytes.Equal(pay.Dest, ctype.Hex2Bytes(c.EthAddress)) {
		// only destination can cancel pay
		return common.ErrPayDestMismatch
	}
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_REJECT_BOOLEAN_PAY_API
	err = c.messager.SendOnePaySettleProof(payID, rpc.PaymentSettleReason_PAY_REJECTED, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}
	pem.CommitPem(logEntry)
	return err
}

func (c *CNode) SettleOnChainResolvedPay(payID ctype.PayIDType) error {
	pay, _, err := c.dal.GetConditionalPay(payID)
	if err != nil {
		return err
	}
	if !bytes.Equal(pay.Dest, ctype.Hex2Bytes(c.EthAddress)) {
		// only destination can send onchain resolved pay settle proof
		return common.ErrPayDestMismatch
	}
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_SETTLE_ON_CHAIN_RESOLVED_PAY_API
	err = c.messager.SendOnePaySettleProof(payID, rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}
	pem.CommitPem(logEntry)
	return err
}

func (c *CNode) SettleExpiredPays(cid ctype.CidType) error {
	channelSelf, _, err := c.dal.GetSimplexPaymentChannel(cid, c.nodeConfig.GetOnChainAddr())
	if err != nil {
		return err
	}
	expiredPays, _, err := c.getExpiredPays(channelSelf)
	if err != nil {
		return err
	}

	if len(expiredPays) > 0 {
		amts := make([]*big.Int, len(expiredPays))
		for i := range amts {
			amts[i] = new(big.Int).SetUint64(0)
		}

		logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
		logEntry.Type = pem.PayMessageType_SRC_SETTLE_EXPIRED_PAY_API
		_, err = c.messager.SendPaysSettleRequest(expiredPays, amts, rpc.PaymentSettleReason_PAY_EXPIRED, logEntry)
		if err != nil {
			logEntry.Error = append(logEntry.Error, err.Error())
		}
		pem.CommitPem(logEntry)
		if err != nil {
			return err
		}
	}

	peer, err := c.dal.GetPeer(cid)
	if err != nil {
		log.Error(err)
		return err
	}
	channelPeer, _, err := c.dal.GetSimplexPaymentChannel(cid, peer)
	if err != nil {
		return err
	}
	_, expiredPayIDs, err := c.getExpiredPays(channelPeer)
	if err != nil {
		return err
	}
	if len(expiredPayIDs) > 0 {
		logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
		logEntry.Type = pem.PayMessageType_DST_SETTLE_EXPIRED_PAY_API
		err := c.messager.SendPaysSettleProof(expiredPayIDs, rpc.PaymentSettleReason_PAY_EXPIRED, logEntry)
		if err != nil {
			logEntry.Error = append(logEntry.Error, err.Error())
		}
		pem.CommitPem(logEntry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CNode) getExpiredPays(
	simplex *entity.SimplexPaymentChannel) ([]*entity.ConditionalPay, []ctype.PayIDType, error) {
	currBlock := c.GetCurrentBlockNumber().Uint64()
	var expiredPays []*entity.ConditionalPay
	var payIDs []ctype.PayIDType
	for _, payID := range simplex.PendingPayIds.PayIds {
		payID := ctype.Bytes2PayID(payID)
		pay, _, err2 := c.dal.GetConditionalPay(payID)
		if err2 != nil {
			return nil, nil, err2
		}
		if currBlock > pay.ResolveDeadline+config.PaySendTimeoutSafeMargin {
			expiredPays = append(expiredPays, pay)
			payIDs = append(payIDs, payID)
		}
	}
	return expiredPays, payIDs, nil
}

func (c *CNode) ConfirmOnChainResolvedPays(cid ctype.CidType) error {
	simplex, _, err := c.dal.GetSimplexPaymentChannel(cid, c.nodeConfig.GetOnChainAddr())
	if err != nil {
		return err
	}

	var resolvedPays []*entity.ConditionalPay
	var payAmts []*big.Int
	for _, payID := range simplex.PendingPayIds.PayIds {
		payID := ctype.Bytes2PayID(payID)
		pay, _, err2 := c.dal.GetConditionalPay(payID)
		if err2 != nil {
			return err2
		}
		amt, _, err2 := c.Disputer.GetCondPayInfoFromRegistry(payID)
		if err2 != nil {
			return err2
		}
		maxAmt := utils.BytesToBigInt(pay.TransferFunc.MaxTransfer.Receiver.Amt)
		if amt.Cmp(maxAmt) == 0 {
			resolvedPays = append(resolvedPays, pay)
			payAmts = append(payAmts, amt)
		}
	}

	if len(resolvedPays) > 0 {
		logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
		logEntry.Type = pem.PayMessageType_CONFIRM_ON_CHAIN_PAY_API
		_, err2 := c.messager.SendPaysSettleRequest(
			resolvedPays, payAmts, rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN, logEntry)
		if err2 != nil {
			return err2
		}
	}

	return nil
}

func (c *CNode) SyncOnChainChannelStates(cid ctype.CidType) (string, error) {
	localStatus, _, err := c.dal.GetChannelState(cid)
	if err != nil {
		return localStatus, err
	}
	onchainUnitialized := uint8(0)
	onchainSettling := uint8(2)
	onchainClosed := uint8(3)
	onchainStatus, err := ledgerview.GetOnChainChannelStatus(cid, c.nodeConfig)
	if err != nil {
		return localStatus, err
	}
	if onchainStatus == onchainSettling && localStatus == fsm.PscOpen {
		err2 := c.dal.Transactional(fsm.OnPscIntendSettle, cid)
		if err2 != nil {
			return localStatus, err2
		}
	} else if onchainStatus == onchainClosed && localStatus != fsm.PscClosed {
		err2 := c.dal.Transactional(c.Disputer.HandleConfirmSettleEventTx, cid)
		if err2 != nil {
			return localStatus, err2
		}
	}
	if onchainStatus != onchainClosed && onchainStatus != onchainUnitialized {
		err2 := ledgerview.SyncOnChainBalance(c.dal, cid, c.nodeConfig)
		if err2 != nil {
			return localStatus, err2
		}
	}
	return localStatus, nil
}

func (c *CNode) IntendWithdraw(cidFrom ctype.CidType, amount *big.Int, cidTo ctype.CidType) error {
	return c.Disputer.IntendWithdraw(cidFrom, amount, cidTo)
}

func (c *CNode) ConfirmWithdraw(cid ctype.CidType) error {
	return c.Disputer.ConfirmWithdraw(cid)
}

func (c *CNode) VetoWithdraw(cid ctype.CidType) error {
	return c.Disputer.VetoWithdraw(cid)
}

// SignState signs the data using cnode crypto and return result
func (c *CNode) SignState(in []byte) []byte {
	ret, _ := c.crypto.Sign(in)
	return ret
}
