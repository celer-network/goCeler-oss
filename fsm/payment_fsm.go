// Copyright 2018-2019 Celer Network

package fsm

import (
	"fmt"

	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/storage"
)

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

// OnPayXXX returns cid, status, error after the event process

func OnPayEgressOneSigPending(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType) (ctype.CidType, string, error) {
	return onPayOneSigPending(tx, payID, cid, false)
}

func OnPayIngressCoSignedPending(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType) (ctype.CidType, string, error) {
	return onPayCoSignedPending(tx, payID, cid, true)
}

func OnPayEgressCoSignedPending(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType) (ctype.CidType, string, error) {
	return onPayCoSignedPending(tx, payID, cid, false)
}

func OnPayIngressHashLockRevealed(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayHashLockRevealed(tx, payID, true)
}

func OnPayEgressHashLockRevealed(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayHashLockRevealed(tx, payID, false)
}

func OnPayEgressOneSigPaid(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayOneSigPaid(tx, payID, false)
}

func OnPayEgressDirectOneSigPaid(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType) (ctype.CidType, string, error) {
	return onPayDirectOneSigPaid(tx, payID, cid, false)
}

func OnPayIngressCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayCoSignedPaid(tx, payID, true)
}

func OnPayIngressDirectCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType) (ctype.CidType, string, error) {
	return onPayDirectCoSignedPaid(tx, payID, cid, true)
}

func OnPayEgressCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayCoSignedPaid(tx, payID, false)
}

func OnPayEgressDirectCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType) (ctype.CidType, string, error) {
	return onPayDirectCoSignedPaid(tx, payID, cid, false)
}

func OnPayEgressOneSigCanceled(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayOneSigCanceled(tx, payID, false)
}

func OnPayIngressCoSignedCanceled(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayCoSignedCanceled(tx, payID, true)
}

func OnPayEgressCoSignedCanceled(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	return onPayCoSignedCanceled(tx, payID, false)
}

func OnPayEgressDelivered(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	cid, status, _, err := tx.GetPayEgressState(payID)
	if err != nil {
		return cid, status, err
	}
	switch status {
	case PayOneSigPending:
		err = tx.PutPayEgressState(payID, cid, PayCoSignedPending)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("OnPayEgressDelivered err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedPending, nil
	default: // PayCoSignedPending, PayHashLockRevealed, PayOneSigPaid, PayCoSignedPaid, PayOneSigCanceled, PayCoSignedCanceled, PayOneSigNacked
		return cid, status, nil
	}
}

func OnPayEgressNacked(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	cid, status, _, err := tx.GetPayEgressState(payID)
	if err != nil {
		return cid, status, err
	}
	switch status {
	case PayOneSigPending, PayOneSigPaid:
		err = tx.PutPayEgressState(payID, cid, PayOneSigNacked)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("OnPayEgressNacked err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayOneSigNacked, nil
	default: // PayCoSignedPending, PayHashLockRevealed, PayOneSigPaid, PayCoSignedPaid, PayOneSigCanceled, PayCoSignedCanceled, PayOneSigNacked
		return cid, status, nil
	}
}

func OnPayEgressUpdateAfterNack(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	cid, status, _, err := tx.GetPayEgressState(payID)
	if err != nil {
		return cid, status, err
	}
	if status == PayOneSigNacked {
		err = tx.PutPayEgressState(payID, cid, PayCoSignedCanceled)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("OnPayEgressUpdateAfterNack err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedCanceled, nil
	}
	return cid, status, nil
}

// OnPayIngressRejected converts the payment status to ingress rejected after
// sending pay settle proof(cancel).
func OnPayIngressRejected(tx *storage.DALTx, payID ctype.PayIDType) (ctype.CidType, string, error) {
	cid, status, _, err := tx.GetPayIngressState(payID)
	if err != nil {
		return cid, status, err
	}

	switch status {
	case PayCoSignedPending, PayHashLockRevealed:
		err = tx.PutPayIngressState(payID, cid, PayIngressRejected)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("OnPayIngressRejected err %s, payID %x, cid %x", err, payID, cid)
		}
		return cid, PayIngressRejected, nil
	case PayIngressRejected:
		return cid, PayIngressRejected, nil
	default:
		return cid, status, nil
	}
}

func getPayState(tx *storage.DALTx, payID ctype.PayIDType, ig bool) (ctype.CidType, string, int64, error) {
	if ig {
		return tx.GetPayIngressState(payID)
	}
	return tx.GetPayEgressState(payID)
}

func putPayState(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType, status string, ig bool) error {
	if ig {
		return tx.PutPayIngressState(payID, cid, status)
	}
	return tx.PutPayEgressState(payID, cid, status)
}

func onPayOneSigPending(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType, ig bool) (ctype.CidType, string, error) {
	curCid, status, _, err := getPayState(tx, payID, ig)
	if err != nil { // put new entry
		err = putPayState(tx, payID, cid, PayOneSigPending, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("onPayOneSigPending err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayOneSigPending, nil
	}
	if cid != curCid {
		return curCid, status, fmt.Errorf(
			"onPayOneSigPending err: conflict cid. payID %x current cid %x new cid %x", payID, curCid, cid)
	}
	switch status {
	case PayOneSigPending:
		return cid, status, nil
	default: // PayCoSignedPending, PayHashLockRevealed, PayOneSigPaid, PayCoSignedPaid, PayOneSigCanceled, PayCoSignedCanceled, PayOneSigNacked
		return cid, status, fmt.Errorf("onPayOneSigPending err: invalid status %x %x %s", payID, cid, status)
	}
}

func onPayCoSignedPending(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType, ig bool) (ctype.CidType, string, error) {
	curCid, status, _, err := getPayState(tx, payID, ig)
	if err != nil { // put new entry
		err = putPayState(tx, payID, cid, PayCoSignedPending, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("CoSignedPending err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedPending, nil
	}
	if cid != curCid {
		return curCid, status, fmt.Errorf(
			"CoSignedPending err: conflict cid. payID %s current cid %x new cid %x", payID, curCid, cid)
	}
	switch status {
	case PayOneSigPending:
		err = putPayState(tx, payID, cid, PayCoSignedPending, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("CoSignedPending err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedPending, nil
	case PayCoSignedPending:
		return cid, PayCoSignedPending, nil
	default: // PayHashLockRevealed, PayOneSigPaid, PayCoSignedPaid, PayOneSigCanceled, PayCoSignedCanceled, PayOneSigNacked
		return cid, status, fmt.Errorf("CoSignedPending err: invalid status %x %x %s", payID, cid, status)
	}
}

func onPayHashLockRevealed(tx *storage.DALTx, payID ctype.PayIDType, ig bool) (ctype.CidType, string, error) {
	cid, status, _, err := getPayState(tx, payID, ig)
	if err != nil {
		return ctype.ZeroCid, "", fmt.Errorf("HashLockRevealed err %s, payID %x", err, payID)
	}
	switch status {
	case PayCoSignedPending:
		err = putPayState(tx, payID, cid, PayHashLockRevealed, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("HashLockRevealed err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayHashLockRevealed, nil
	case PayHashLockRevealed:
		return cid, PayHashLockRevealed, nil
	case PayOneSigPaid, PayCoSignedPaid, PayOneSigCanceled, PayCoSignedCanceled:
		return cid, status, nil
	default: // PayOneSigPending, PayOneSigNacked
		return cid, status, fmt.Errorf("HashLockRevealed err: invalid status %x %x %s", payID, cid, status)
	}
}

func onPayOneSigPaid(tx *storage.DALTx, payID ctype.PayIDType, ig bool) (ctype.CidType, string, error) {
	cid, status, _, err := getPayState(tx, payID, ig)
	if err != nil {
		return ctype.ZeroCid, "", fmt.Errorf("OneSigPaid err %s, payID %x", err, payID)
	}
	switch status {
	case PayOneSigPending, PayCoSignedPending, PayHashLockRevealed, PayOneSigCanceled:
		err = putPayState(tx, payID, cid, PayOneSigPaid, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("OneSigPaid err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayOneSigPaid, nil
	case PayOneSigPaid:
		return cid, PayOneSigPaid, nil
	default: // PayCoSignedPaid, PayCoSignedCanceled, PayOneSigNacked
		return cid, status, fmt.Errorf("OneSigPaid err: invalid status %x %x %s", payID, cid, status)
	}
}

func onPayDirectOneSigPaid(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType, ig bool) (ctype.CidType, string, error) {
	curCid, status, _, err := getPayState(tx, payID, ig)
	if err == nil { // already exist
		if cid != curCid {
			return curCid, status, fmt.Errorf(
				"onPayDirectOneSigPaid err: conflict cid. payID %x current cid %x new cid %x", payID, curCid, cid)
		}
		if status != PayOneSigPaid {
			return ctype.ZeroCid, "", fmt.Errorf("onPayDirectOneSigPaid err: invalid status %x %x %s", payID, cid, status)
		}
	}
	err = putPayState(tx, payID, cid, PayOneSigPaid, ig)
	if err != nil {
		return ctype.ZeroCid, "", fmt.Errorf("onPayDirectOneSigPaid err %s, payID %x cid %x", err, payID, cid)
	}
	return cid, PayOneSigPaid, nil
}

func onPayCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType, ig bool) (ctype.CidType, string, error) {
	cid, status, _, err := getPayState(tx, payID, ig)
	if err != nil {
		return ctype.ZeroCid, "", fmt.Errorf("CoSignedPaid err %s, payID %x", err, payID)
	}
	switch status {
	case PayCoSignedPending, PayHashLockRevealed, PayOneSigPaid:
		err = putPayState(tx, payID, cid, PayCoSignedPaid, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("CoSignedPaid err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedPaid, nil
	case PayCoSignedPaid:
		return cid, PayCoSignedPaid, nil
	default: // PayOneSigPending, PayOneSigCanceled, PayCoSignedCanceled, PayOneSigNacked
		return cid, status, fmt.Errorf("CoSignedPaid err: invalid status %x %x %s", payID, cid, status)
	}
}

func onPayDirectCoSignedPaid(tx *storage.DALTx, payID ctype.PayIDType, cid ctype.CidType, ig bool) (ctype.CidType, string, error) {
	curCid, status, _, err := getPayState(tx, payID, ig)
	if err != nil { // put new entry
		err = putPayState(tx, payID, cid, PayCoSignedPaid, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("DirectCoSignedPaid err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedPaid, nil
	}
	if cid != curCid {
		return curCid, status, fmt.Errorf(
			"DirectCoSignedPaid err: conflict cid. payID %x current cid %x new cid %x", payID, curCid, cid)
	}
	switch status {
	case PayOneSigPaid:
		err = putPayState(tx, payID, cid, PayCoSignedPaid, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("DirectCoSignedPaid err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedPaid, nil
	case PayCoSignedPaid:
		return cid, PayCoSignedPaid, nil
	default:
		return cid, status, fmt.Errorf("DirectCoSignedPaid err: invalid status %x %x %s", payID, cid, status)
	}
}

func onPayOneSigCanceled(tx *storage.DALTx, payID ctype.PayIDType, ig bool) (ctype.CidType, string, error) {
	cid, status, _, err := getPayState(tx, payID, ig)
	if err != nil {
		return ctype.ZeroCid, "", fmt.Errorf("OneSigCanceled err %s, payID %x", err, payID)
	}
	switch status {
	case PayOneSigPending, PayCoSignedPending, PayHashLockRevealed:
		err = putPayState(tx, payID, cid, PayOneSigCanceled, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("OneSigCanceled err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayOneSigCanceled, nil
	case PayOneSigCanceled:
		return cid, PayOneSigCanceled, nil
	default: // PayOneSigPaid, PayCoSignedPaid, PayCoSignedCanceled, PayOneSigNacked
		return cid, status, fmt.Errorf("OneSigCanceled err: invalid status %x %x %s", payID, cid, status)
	}
}

func onPayCoSignedCanceled(tx *storage.DALTx, payID ctype.PayIDType, ig bool) (ctype.CidType, string, error) {
	cid, status, _, err := getPayState(tx, payID, ig)
	if err != nil {
		return ctype.ZeroCid, "", fmt.Errorf("CoSignedCanceled err %s, payID %x", err, payID)
	}
	switch status {
	case PayOneSigPending, PayCoSignedPending, PayHashLockRevealed, PayOneSigCanceled, PayOneSigNacked, PayIngressRejected:
		err = putPayState(tx, payID, cid, PayCoSignedCanceled, ig)
		if err != nil {
			return ctype.ZeroCid, "", fmt.Errorf("CoSignedCanceled err %s, payID %x cid %x", err, payID, cid)
		}
		return cid, PayCoSignedCanceled, nil
	case PayCoSignedCanceled:
		return cid, PayCoSignedCanceled, nil
	default: // PayOneSigPaid, PayCoSignedPaid
		return cid, status, fmt.Errorf("CoSignedCanceled err: invalid status %x %x %s", payID, cid, status)
	}
}
