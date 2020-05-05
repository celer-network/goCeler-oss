// Copyright 2018-2020 Celer Network

package fsm

import (
	"fmt"

	"github.com/celer-network/goCeler/common"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/storage"
)

func PayStateName(state int) string {
	switch state {
	case enums.PayState_ONESIG_PENDING:
		return "PAY_ONESIG_PENDING"
	case enums.PayState_COSIGNED_PENDING:
		return "PAY_COSIGNED_PENDING"
	case enums.PayState_SECRET_REVEALED:
		return "PAY_SECRET_REVEALED"
	case enums.PayState_ONESIG_PAID:
		return "PAY_ONESIG_PAID"
	case enums.PayState_COSIGNED_PAID:
		return "COSIGNED_PAID"
	case enums.PayState_ONESIG_CANCELED:
		return "PAY_ONESIG_CANCELED"
	case enums.PayState_COSIGNED_CANCELED:
		return "PAY_COSIGNED_CANCELED"
	case enums.PayState_NACKED:
		return "PAY_NACKED"
	case enums.PayState_INGRESS_REJECTED:
		return "PAY_INGRESS_REJECTED"
	default:
		return "PAY_NULL"
	}
}

// return exist, newstate, err
func OnCondPayRequestSent(
	tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType, directPay bool) (bool, int, error) {
	newstate := enums.PayState_ONESIG_PENDING
	if directPay {
		newstate = enums.PayState_ONESIG_PAID
	}
	egcid, egstate, found, err := tx.GetPayEgress(payID)
	if err != nil {
		return false, newstate, fmt.Errorf("OnCondPayRequestSent err %w, payID %x", err, payID)
	}
	if found {
		if egcid != ctype.ZeroCid {
			if cid != egcid {
				return found, newstate, fmt.Errorf("OnCondPayRequestSent err: conflict cid. payID %x current cid %x new cid %x", payID, egcid, cid)
			}
			if newstate != egstate {
				return found, newstate, fmt.Errorf("OnCondPayRequestSent err: invalid state %x %x %s", payID, cid, PayStateName(egstate))
			}
		}
		return found, newstate, tx.UpdatePayEgress(payID, cid, newstate)
	}
	return found, newstate, nil
}

func OnPayEgressOneSigPaid(tx *storage.DALTx, payID ctype.PayIDType, egstate int) error {
	switch egstate {
	case enums.PayState_ONESIG_PENDING, enums.PayState_COSIGNED_PENDING, enums.PayState_SECRET_REVEALED, enums.PayState_ONESIG_CANCELED:
		return tx.UpdatePayEgressState(payID, enums.PayState_ONESIG_PAID)
	case enums.PayState_ONESIG_PAID:
		return nil
	default:
		return fmt.Errorf("OnPayEgressOneSigPaid err: invalid state %x %s", payID, PayStateName(egstate))
	}
}

func OnPayEgressOneSigCanceled(tx *storage.DALTx, payID ctype.PayIDType, egstate int) error {
	switch egstate {
	case enums.PayState_ONESIG_PENDING, enums.PayState_COSIGNED_PENDING, enums.PayState_SECRET_REVEALED:
		return tx.UpdatePayEgressState(payID, enums.PayState_ONESIG_CANCELED)
	case enums.PayState_ONESIG_CANCELED:
		return nil
	default:
		return fmt.Errorf("OnPayEgressOneSigCanceled err: invalid status %x %s", payID, PayStateName(egstate))
	}
}

func OnPayIngressSecretRevealed(tx *storage.DALTx, payID ctype.PayIDType, state int) error {
	switch state {
	case enums.PayState_COSIGNED_PENDING:
		return tx.UpdatePayIngressState(payID, enums.PayState_SECRET_REVEALED)
	case enums.PayState_SECRET_REVEALED, enums.PayState_ONESIG_PAID, enums.PayState_ONESIG_CANCELED:
		return nil
	case enums.PayState_COSIGNED_PAID, enums.PayState_COSIGNED_CANCELED:
		return common.ErrPayOffChainResolved
	default:
		return fmt.Errorf("OnPayIngressSecretRevealed err: invalid status %x %s", payID, PayStateName(state))
	}
}

func OnPayEgressSecretRevealed(tx *storage.DALTx, payID ctype.PayIDType, state int) error {
	switch state {
	case enums.PayState_COSIGNED_PENDING:
		return tx.UpdatePayEgressState(payID, enums.PayState_SECRET_REVEALED)
	case enums.PayState_SECRET_REVEALED, enums.PayState_ONESIG_PAID, enums.PayState_ONESIG_CANCELED:
		return nil
	case enums.PayState_COSIGNED_PAID, enums.PayState_COSIGNED_CANCELED:
		return common.ErrPayOffChainResolved
	default:
		return fmt.Errorf("OnPayEgressSecretRevealed err: invalid status %x %s", payID, PayStateName(state))
	}
}

func OnPayIngressCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType, state int) error {
	switch state {
	case enums.PayState_COSIGNED_PENDING, enums.PayState_SECRET_REVEALED, enums.PayState_ONESIG_PAID:
		return tx.UpdatePayIngressState(payID, enums.PayState_COSIGNED_PAID)
	case enums.PayState_COSIGNED_PAID:
		return nil
	default:
		return fmt.Errorf("OnPayIngressCoSignedPaid err: invalid status %x %s", payID, PayStateName(state))
	}
}

func OnPayIngressCoSignedCanceled(tx *storage.DALTx, payID ctype.PayIDType, state int) error {
	switch state {
	case enums.PayState_ONESIG_PENDING, enums.PayState_COSIGNED_PENDING, enums.PayState_SECRET_REVEALED,
		enums.PayState_ONESIG_CANCELED, enums.PayState_NACKED, enums.PayState_INGRESS_REJECTED:
		return tx.UpdatePayIngressState(payID, enums.PayState_COSIGNED_CANCELED)
	case enums.PayState_COSIGNED_CANCELED:
		return nil
	default:
		return fmt.Errorf("OnPayIngressCoSignedCanceled err: invalid status %x %s", payID, PayStateName(state))
	}
}

func OnPayEgressCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType) error {
	cid, state, found, err := tx.GetPayEgress(payID)
	if err != nil {
		return fmt.Errorf("OnPayEgressCoSignedPaid err %w, payID %x", err, payID)
	}
	if !found {
		return fmt.Errorf("OnPayEgressCoSignedPaid err %w, payID %x", common.ErrPayNotFound, payID)
	}
	switch state {
	case enums.PayState_COSIGNED_PENDING, enums.PayState_SECRET_REVEALED, enums.PayState_ONESIG_PAID:
		return tx.UpdatePayEgressState(payID, enums.PayState_COSIGNED_PAID)
	case enums.PayState_COSIGNED_PAID:
		return nil
	default:
		return fmt.Errorf("OnPayEgressCoSignedPaid err: invalid status %x %x %s", payID, cid, PayStateName(state))
	}
}

func OnPayEgressCoSignedCanceled(tx *storage.DALTx, payID ctype.PayIDType) error {
	cid, state, found, err := tx.GetPayEgress(payID)
	if err != nil {
		return fmt.Errorf("OnPayEgressCoSignedCanceled err %w, payID %x", err, payID)
	}
	if !found {
		return fmt.Errorf("OnPayEgressCoSignedCanceled err %w, payID %x", common.ErrPayNotFound, payID)
	}
	switch state {
	case enums.PayState_ONESIG_PENDING, enums.PayState_COSIGNED_PENDING, enums.PayState_SECRET_REVEALED,
		enums.PayState_ONESIG_CANCELED, enums.PayState_NACKED, enums.PayState_INGRESS_REJECTED:
		return tx.UpdatePayEgressState(payID, enums.PayState_COSIGNED_CANCELED)
	case enums.PayState_COSIGNED_CANCELED:
		return nil
	default:
		return fmt.Errorf("OnPayEgressCoSignedCanceled err: invalid status %x %x %s", payID, cid, PayStateName(state))
	}
}

func OnPayEgressNacked(tx *storage.DALTx, payID ctype.PayIDType) error {
	_, state, found, err := tx.GetPayEgress(payID)
	if err != nil {
		return fmt.Errorf("OnPayEgressNacked err %w, payID %x", err, payID)
	}
	if !found {
		return fmt.Errorf("OnPayEgressNacked err %w, payID %x", common.ErrPayNotFound, payID)
	}
	switch state {
	case enums.PayState_ONESIG_PENDING, enums.PayState_ONESIG_PAID:
		return tx.UpdatePayEgressState(payID, enums.PayState_NACKED)
	default:
		return nil
	}
}

func OnPayEgressDelivered(tx *storage.DALTx, payID ctype.PayIDType) error {
	_, state, found, err := tx.GetPayEgress(payID)
	if err != nil {
		return fmt.Errorf("OnPayEgressDelivered err %w, payID %x", err, payID)
	}
	if !found {
		return fmt.Errorf("OnPayEgressDelivered err %w, payID %x", common.ErrPayNotFound, payID)
	}
	switch state {
	case enums.PayState_ONESIG_PENDING:
		return tx.UpdatePayEgressState(payID, enums.PayState_COSIGNED_PENDING)
	default:
		return nil
	}
}

func OnPayEgressUpdateAfterNack(tx *storage.DALTx, payID ctype.PayIDType) error {
	_, state, found, err := tx.GetPayEgress(payID)
	if err != nil {
		return fmt.Errorf("OnPayEgressUpdateAfterNack err %w, payID %x", err, payID)
	}
	if !found {
		return fmt.Errorf("OnPayEgressUpdateAfterNack err %w, payID %x", common.ErrPayNotFound, payID)
	}
	switch state {
	case enums.PayState_NACKED:
		return tx.UpdatePayEgressState(payID, enums.PayState_COSIGNED_CANCELED)
	default:
		return nil
	}
}

// OnPayIngressRejected converts the payment status to ingress rejected after
// sending pay settle proof(cancel).
func OnPayIngressRejected(tx *storage.DALTx, payID ctype.PayIDType, state int) error {
	switch state {
	case enums.PayState_COSIGNED_PENDING, enums.PayState_SECRET_REVEALED:
		return tx.UpdatePayIngressState(payID, enums.PayState_INGRESS_REJECTED)
	default:
		return nil
	}
}

// ---------------------------- Deprecated state ----------------------------

// do NOT modify the string representation of existing states
const (
	PayOneSigPending    = "ONESIG_PENDING"
	PayCoSignedPending  = "COSIGNED_PENDING"
	PayHashLockRevealed = "HASHLOCK_REVEALED" // only valid for pay src or dst
	PayOneSigPaid       = "ONESIG_PAID"
	PayCoSignedPaid     = "COSIGNED_PAID"
	PayOneSigCanceled   = "ONESIG_CANCELED"
	PayCoSignedCanceled = "COSIGNED_CANCELED"
	PayOneSigNacked     = "PAY_NACKED"       // one-sig simplex rejected by the peer_to
	PayIngressRejected  = "INGRESS_REJECTED" // after sending pay settle proof(cancel), dst should record this state
)
