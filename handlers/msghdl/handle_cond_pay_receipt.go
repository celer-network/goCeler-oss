// Copyright 2018-2020 Celer Network

package msghdl

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/celer-network/goCeler/common"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/utils/hashlist"
	"github.com/celer-network/goutils/log"
)

func (h *CelerMsgHandler) HandleCondPayReceipt(frame *common.MsgFrame) error {
	receipt := frame.Message.GetCondPayReceipt()
	logEntry := frame.LogEntry
	if receipt == nil {
		return common.ErrInvalidMsgType
	}
	dst := ctype.Bytes2Addr(frame.Message.GetToAddr())
	logEntry.Dst = ctype.Addr2Hex(dst)
	payID := ctype.Bytes2PayID(receipt.PayId)
	logEntry.PayId = ctype.PayID2Hex(payID)

	// Forward Msg if not destination
	if dst != h.nodeConfig.GetOnChainAddr() {
		_, peer, err := h.routeForwarder.LookupIngressChannelOnPay(payID)
		if err != nil {
			return fmt.Errorf("LookupIngressChannelOnPay err %w", err)
		}
		logEntry.MsgTo = ctype.Addr2Hex(peer)
		log.Debugf("Forwarding cond pay receipt to %x, next hop %x", dst, peer)
		return h.messager.ForwardCelerMsg(peer, frame.Message)
	}

	pay, payBytes, egstate, found, err := h.dal.GetPayAndEgressState(payID)
	if err != nil {
		return fmt.Errorf("GetPayment err %w", err)
	}
	if !found {
		return common.ErrPayNotFound
	}
	if egstate == enums.PayState_COSIGNED_PAID || egstate == enums.PayState_COSIGNED_CANCELED {
		log.Warn(common.ErrPayOffChainResolved, payID.Hex())
		return nil
	}

	// verify pay source
	if ctype.Bytes2Addr(pay.GetSrc()) != h.nodeConfig.GetOnChainAddr() {
		return common.ErrInvalidPaySrc
	}
	var delegateDescription *rpc.DelegationDescription
	proof := receipt.GetDelegationProof()
	if proof != nil {
		description, unmarhsalErr := utils.UnmarshalDelegationDescription(proof)
		if unmarhsalErr != nil {
			return unmarhsalErr
		}
		delegateDescription = description
		delegationErr := h.verifyDelegationProof(proof, description, pay, payBytes, receipt)
		logEntry.DelegationDescription = description
		if delegationErr != nil {
			return delegationErr
		}
	} else {
		// Check signature of receipt signed by destination
		if !utils.SigIsValid(ctype.Bytes2Addr(pay.GetDest()), payBytes, receipt.PayDestSig) {
			return errors.New("RECEIPT_NOT_SIGNED_BY_DEST")
		}
	}

	// Return secret
	// The first condition is always HashLock
	if len(pay.GetConditions()) == 0 {
		log.Warnln(common.ErrZeroConditions, payID.Hex())
		return nil
	}
	secretHash := ctype.Bytes2Hex(pay.Conditions[0].GetHashLock())
	secret, found, err := h.dal.GetSecret(secretHash)
	if err != nil {
		return fmt.Errorf("GetSecret err %w hash %x", err, secretHash)
	}
	if !found {
		return fmt.Errorf("%w, hash %x", common.ErrSecretNotRevealed, secretHash)
	}
	secretBytes := ctype.Hex2Bytes(secret)
	secretMsg := &rpc.RevealSecret{
		PayId:  receipt.PayId,
		Secret: secretBytes,
	}
	toAddr := pay.GetDest()
	if delegateDescription != nil {
		toAddr = delegateDescription.GetDelegator()
		err = h.dal.InsertPayDelegator(
			payID, ctype.Bytes2Addr(pay.GetDest()), ctype.Bytes2Addr(delegateDescription.GetDelegator()))
		if err != nil {
			return fmt.Errorf("InsertPayDelegator err %w", err)
		}
	}
	celerMsg := &rpc.CelerMsg{
		ToAddr: toAddr,
		Message: &rpc.CelerMsg_RevealSecret{
			RevealSecret: secretMsg,
		},
	}
	err = h.streamWriter.WriteCelerMsg(frame.PeerAddr, celerMsg)
	if err != nil {
		return fmt.Errorf("WriteCelerMsg err %w", err)
	}

	return nil
}

func (h *CelerMsgHandler) verifyDelegationProof(
	proof *rpc.DelegationProof,
	description *rpc.DelegationDescription,
	pay *entity.ConditionalPay,
	payBytes []byte,
	receipt *rpc.CondPayReceipt) error {
	if ctype.Bytes2Addr(description.GetDelegatee()) != ctype.Bytes2Addr(pay.GetDest()) {
		return errors.New("destination is NOT delegatee in delegation description")
	}
	tokenAddr := utils.GetTokenAddr(pay.GetTransferFunc().GetMaxTransfer().GetToken())
	if !hashlist.Exist(description.GetTokenToDelegate(), tokenAddr.Bytes()) {
		return errors.New("token type not approved by destination to be delegated")
	}
	if description.GetExpiresAfterBlock() < h.monitorService.GetCurrentBlockNumber().Int64() {
		return errors.New("description expired")
	}
	if !utils.SigIsValid(ctype.Bytes2Addr(description.GetDelegator()), payBytes, receipt.GetPayDelegatorSig()) {
		return errors.New("paybytes not signed by delegator")
	}
	if len(pay.GetConditions()) != 1 || pay.GetConditions()[0].GetConditionType() != entity.ConditionType_HASH_LOCK {
		return errors.New("delegating pay having more than HL condition")
	}
	if bytes.Compare(description.GetDelegatee(), proof.GetSigner()) != 0 {
		return errors.New("delegatee in description not same as signer in proof")
	}
	return nil
}
