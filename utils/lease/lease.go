// Copyright 2018-2020 Celer Network

package lease

import (
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/storage"
)

func Acquire(dal *storage.DAL, id, owner string, timeout time.Duration) error {
	return dal.Transactional(AcquireTx, id, owner, timeout)
}

func AcquireTx(tx *storage.DALTx, args ...interface{}) error {
	id := args[0].(string)
	owner := args[1].(string)
	timeout := args[2].(time.Duration)

	currentOwner, updateTs, found, err := tx.GetLease(id)
	if err != nil {
		return err
	}
	if !found {
		return tx.InsertLease(id, owner)
	}
	if currentOwner == owner {
		return tx.UpdateLeaseTimestamp(id, owner)
	}
	if time.Now().UTC().After(updateTs.Add(timeout)) {
		return tx.UpdateLeaseOwner(id, owner)
	}

	return common.ErrLeaseAcquired
}

func Renew(dal *storage.DAL, id, owner string) error {
	return dal.UpdateLeaseTimestamp(id, owner)
}

func Release(dal *storage.DAL, id, owner string) error {
	return dal.DeleteLeaseOwner(id, owner)
}

func CheckOwner(dal *storage.DAL, id, owner string) bool {
	currentOwner, found, err := dal.GetLeaseOwner(id)
	if err != nil || !found {
		return false
	}
	return owner == currentOwner
}
