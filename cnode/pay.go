// Copyright 2019-2020 Celer Network

package cnode

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
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
	if utils.GetTokenAddr(newPay.TransferFunc.MaxTransfer.Token) == ctype.InvalidTokenAddr {
		return ctype.ZeroPayID, common.ErrUnknownTokenType
	}

	newPay.PayTimestamp = uint64(time.Now().UnixNano())
	directPay := c.messager.IsDirectPay(newPay, ctype.ZeroAddr)

	// Skip prepending HL if it's already there or for direct-pay.
	var hashStr, secretStr string
	skipPrependHL := directPay || (len(newPay.GetConditions()) > 0 && newPay.Conditions[0].GetHashLock() != nil)
	if !skipPrependHL {
		// Prepend HL condition if there is none.
		hlCond, secret, hash := newHashLockCond()
		hashStr = ctype.Bytes2Hex(hash)
		secretStr = ctype.Bytes2Hex(secret)
		newPay.Conditions = append([]*entity.Condition{hlCond}, newPay.Conditions...)
	}

	newPay.PayResolver = c.nodeConfig.GetPayResolverContract().GetAddr().Bytes()
	payID := ctype.Pay2PayID(newPay)
	if !skipPrependHL {
		err := c.dal.InsertSecret(hashStr, secretStr, payID)
		if err != nil {
			log.Errorln("InsertSecret err", hashStr, secretStr, payID.Hex(), err)
			return ctype.ZeroPayID, fmt.Errorf("InsertSecret err %w", err)
		}
	}

	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_SEND_TOKEN_API
	logEntry.PayId = ctype.PayID2Hex(payID)
	logEntry.Src = ctype.Addr2Hex(c.nodeConfig.GetOnChainAddr())
	logEntry.Dst = ctype.Bytes2Hex(newPay.GetDest())
	logEntry.DirectPay = directPay

	newPayBytes, err := proto.Marshal(newPay)
	if err != nil {
		return ctype.ZeroPayID, err
	}
	err = c.messager.SendCondPayRequest(newPayBytes, note, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
		payID = ctype.ZeroPayID
	}
	pem.CommitPem(logEntry)
	return payID, err
}

func (c *CNode) ConfirmBooleanPay(payID ctype.PayIDType) error {
	pay, _, found, err := c.dal.GetPayment(payID)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrPayNotFound
	}
	amt := new(big.Int).SetBytes(pay.TransferFunc.MaxTransfer.Receiver.Amt)
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_CONFIRM_BOOLEAN_PAY_API
	logEntry.PayId = ctype.PayID2Hex(payID)
	err = c.messager.SendOnePaySettleRequest(pay, amt, rpc.PaymentSettleReason_PAY_PAID_MAX, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}
	pem.CommitPem(logEntry)
	return err
}

func (c *CNode) RejectBooleanPay(payID ctype.PayIDType) error {
	pay, _, found, err := c.dal.GetPayment(payID)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrPayNotFound
	}
	if !bytes.Equal(pay.Dest, c.EthAddress.Bytes()) {
		// only destination can cancel pay
		return common.ErrPayDestMismatch
	}
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_REJECT_BOOLEAN_PAY_API
	logEntry.PayId = ctype.PayID2Hex(payID)
	err = c.messager.SendOnePaySettleProof(payID, rpc.PaymentSettleReason_PAY_REJECTED, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}
	pem.CommitPem(logEntry)
	return err
}

func (c *CNode) SettleOnChainResolvedPay(payID ctype.PayIDType) error {
	pay, _, found, err := c.dal.GetPayment(payID)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrPayNotFound
	}
	if !bytes.Equal(pay.Dest, c.EthAddress.Bytes()) {
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
	selfSimplex, _, peerSimplex, _, found, err := c.dal.GetDuplexChannel(cid)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrChannelNotFound
	}
	expiredPays, _, err := c.getExpiredPays(selfSimplex)
	if err != nil {
		return err
	}

	if len(expiredPays) > 0 {
		err = c.sendSettleRequestForExpiredPays(expiredPays)
		if err != nil {
			return err
		}
	}

	_, expiredPayIDs, err := c.getExpiredPays(peerSimplex)
	if err != nil {
		return err
	}
	if len(expiredPayIDs) > 0 {
		err = c.sendSettleProofForExpiredPays(expiredPayIDs)
	}

	return err
}

func (c *CNode) getExpiredPays(simplex *entity.SimplexPaymentChannel) (
	[]*entity.ConditionalPay, []ctype.PayIDType, error) {

	currBlock := c.GetCurrentBlockNumber().Uint64()
	var expiredPays []*entity.ConditionalPay
	var payIDs []ctype.PayIDType
	for _, payID := range simplex.PendingPayIds.PayIds {
		payID := ctype.Bytes2PayID(payID)
		pay, _, found, err := c.dal.GetPayment(payID)
		if err != nil {
			return nil, nil, err
		}
		if !found {
			return nil, nil, fmt.Errorf("%w: %x", common.ErrPayNotFound, payID)
		}
		if currBlock > pay.ResolveDeadline+config.PaySendTimeoutSafeMargin {
			expiredPays = append(expiredPays, pay)
			payIDs = append(payIDs, payID)
		}
	}
	return expiredPays, payIDs, nil
}

func (c *CNode) ConfirmOnChainResolvedPays(cid ctype.CidType) error {
	simplex, _, found, err := c.dal.GetSelfSimplex(cid)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrChannelNotFound
	}

	var resolvedPays []*entity.ConditionalPay
	var payAmts []*big.Int
	for _, payID := range simplex.PendingPayIds.PayIds {
		payID := ctype.Bytes2PayID(payID)
		pay, _, found, err2 := c.dal.GetPayment(payID)
		if err2 != nil {
			return err2
		}
		if !found {
			return fmt.Errorf("%w: %x", common.ErrPayNotFound, payID)
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
		err = c.sendSettleRequestForOnChainResolvedPays(resolvedPays, payAmts)
	}

	return err
}

func (c *CNode) sendSettleRequestForExpiredPays(expiredPays []*entity.ConditionalPay) error {
	amts := make([]*big.Int, len(expiredPays))
	for i := range amts {
		amts[i] = new(big.Int).SetUint64(0)
	}

	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_SRC_SETTLE_EXPIRED_PAY_API
	_, err := c.messager.SendPaysSettleRequest(expiredPays, amts, rpc.PaymentSettleReason_PAY_EXPIRED, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}
	pem.CommitPem(logEntry)
	return err
}

func (c *CNode) sendSettleProofForExpiredPays(expiredPayIDs []ctype.PayIDType) error {
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_DST_SETTLE_EXPIRED_PAY_API
	err := c.messager.SendPaysSettleProof(expiredPayIDs, rpc.PaymentSettleReason_PAY_EXPIRED, logEntry)
	if err != nil {
		logEntry.Error = append(logEntry.Error, err.Error())
	}
	pem.CommitPem(logEntry)
	return err
}

func (c *CNode) sendSettleRequestForOnChainResolvedPays(
	resolvedPays []*entity.ConditionalPay, resolvedAmts []*big.Int) error {
	logEntry := pem.NewPem(c.nodeConfig.GetRPCAddr())
	logEntry.Type = pem.PayMessageType_CONFIRM_ON_CHAIN_PAY_API
	_, err := c.messager.SendPaysSettleRequest(
		resolvedPays, resolvedAmts, rpc.PaymentSettleReason_PAY_RESOLVED_ONCHAIN, logEntry)
	pem.CommitPem(logEntry)
	return err
}

// ClearPaymentsInChannel confirm onchain resolved pays and clear expired pays in the channel.
// For efficieny, only expired pays are checked for possible on-chain resolvement.
// This function will replace SettleExpiredPays and ConfirmOnChainResolvedPays in the future.
func (c *CNode) ClearPaymentsInChannel(cid ctype.CidType) error {
	selfSimplex, _, peerSimplex, _, found, err := c.dal.GetDuplexChannel(cid)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrChannelNotFound
	}

	var expiredPays []*entity.ConditionalPay
	var resolvedPays []*entity.ConditionalPay
	var resolvedAmts []*big.Int

	// get all expired pays from self
	pays, payIDs, err := c.getExpiredPays(selfSimplex)
	if err != nil {
		return err
	}

	for i, pay := range pays {
		payID := payIDs[i]
		// first check if any expired pay is resolved on chain
		amt, _, err2 := c.Disputer.GetCondPayInfoFromRegistry(payID)
		if err2 != nil {
			log.Error(err2)
			continue
		}
		maxAmt := utils.BytesToBigInt(pay.TransferFunc.MaxTransfer.Receiver.Amt)
		if amt.Cmp(maxAmt) == 0 {
			resolvedPays = append(resolvedPays, pay)
			resolvedAmts = append(resolvedAmts, amt)
		} else {
			expiredPays = append(expiredPays, pay)
		}
	}

	if len(resolvedPays) > 0 {
		err = c.sendSettleRequestForOnChainResolvedPays(resolvedPays, resolvedAmts)
		if err != nil {
			log.Error(err)
		}
	}

	if len(expiredPays) > 0 {
		err = c.sendSettleRequestForExpiredPays(expiredPays)
		if err != nil {
			log.Error(err)
		}
	}

	// get all expired pays from peer
	_, expiredPayIDs, err := c.getExpiredPays(peerSimplex)
	if err != nil {
		return err
	}
	if len(expiredPayIDs) > 0 {
		err = c.sendSettleProofForExpiredPays(expiredPayIDs)
	}

	return err
}

func (c *CNode) ClearPaymentsWithPeerOsps() error {
	cids, err := c.getConnectedOspCids()
	if err != nil {
		return err
	}
	errs := make(map[string]error)
	for _, cid := range cids {
		err := c.ClearPaymentsInChannel(cid)
		if err != nil {
			errs[ctype.Cid2Hex(cid)] = err
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf(fmt.Sprint(errs))
	}
	return nil
}

func (c *CNode) SyncOnChainChannelStates(cid ctype.CidType) (int, error) {
	localState, found, err := c.dal.GetChanState(cid)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, common.ErrChannelNotFound
	}
	onchainUnitialized := uint8(0)
	onchainSettling := uint8(2)
	onchainClosed := uint8(3)
	onchainMigrated := uint8(4)
	onchainStatus, err := ledgerview.GetOnChainChannelStatus(cid, c.nodeConfig)
	if err != nil {
		return localState, err
	}

	if onchainStatus == onchainMigrated {
		ledger, err := ledgerview.GetMigratedTo(c.dal, cid, c.nodeConfig)
		if err != nil {
			return localState, err
		}
		err = c.dal.UpdateChanLedger(cid, ledger)
		if err != nil {
			return localState, err
		}

		// get channel status from new ledger and then sync with new ledger status
		onchainStatus, err = ledgerview.GetOnChainChannelStatus(cid, c.nodeConfig)
		if err != nil {
			return localState, err
		}
	}

	if onchainStatus == onchainSettling &&
		(localState == enums.ChanState_TRUST_OPENED || localState == enums.ChanState_OPENED) {
		err2 := c.dal.Transactional(fsm.OnChannelIntendSettle, cid)
		if err2 != nil {
			return localState, err2
		}
	} else if onchainStatus == onchainClosed && localState != enums.ChanState_CLOSED {
		err2 := c.dal.Transactional(c.Disputer.HandleConfirmSettleEventTx, cid)
		if err2 != nil {
			return localState, err2
		}
	}
	if onchainStatus != onchainClosed && onchainStatus != onchainUnitialized {
		err2 := ledgerview.SyncOnChainBalance(c.dal, cid, c.nodeConfig)
		if err2 != nil {
			return localState, err2
		}
	}
	return localState, nil
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

// GetPaymentState returns the ingress and egress state of a payment
func (c *CNode) GetPaymentState(payID ctype.PayIDType) (int, int) {
	inState, outState, _, _ := c.dal.GetPayStates(payID)
	return inState, outState
}
