// Copyright 2018-2019 Celer Network

package storage

import "github.com/celer-network/goCeler-oss/cnode/jobs"

const cooperativeWithdrawJob = "cwj"

func getCooperativeWithdrawJob(
	st Storage, withdrawHash string) (*jobs.CooperativeWithdrawJob, error) {
	var job jobs.CooperativeWithdrawJob
	err := st.Get(cooperativeWithdrawJob, withdrawHash, &job)
	return &job, err
}

func putCooperativeWithdrawJob(
	st Storage, withdrawHash string, job *jobs.CooperativeWithdrawJob) error {
	return st.Put(cooperativeWithdrawJob, withdrawHash, job)
}

func deleteCooperativeWithdrawJob(st Storage, withdrawHash string) error {
	return st.Delete(cooperativeWithdrawJob, withdrawHash)
}

func hasCooperativeWithdrawJob(st Storage, withdrawHash string) (bool, error) {
	return st.Has(cooperativeWithdrawJob, withdrawHash)
}

func getAllCooperativeWithdrawJobKeys(st Storage) ([]string, error) {
	return st.GetKeysByPrefix(cooperativeWithdrawJob, "")
}

func (d *DAL) GetCooperativeWithdrawJob(withdrawHash string) (*jobs.CooperativeWithdrawJob, error) {
	return getCooperativeWithdrawJob(d.st, withdrawHash)
}

func (d *DAL) PutCooperativeWithdrawJob(
	withdrawHash string, job *jobs.CooperativeWithdrawJob) error {
	return putCooperativeWithdrawJob(d.st, withdrawHash, job)
}

func (d *DAL) DeleteCooperativeWithdrawJob(withdrawHash string) error {
	return deleteCooperativeWithdrawJob(d.st, withdrawHash)
}

func (d *DAL) HasCooperativeWithdrawJob(withdrawHash string) (bool, error) {
	return hasCooperativeWithdrawJob(d.st, withdrawHash)
}

func (d *DAL) GetAllCooperativeWithdrawJobKeys() ([]string, error) {
	return getAllCooperativeWithdrawJobKeys(d.st)
}

func (dtx *DALTx) GetCooperativeWithdrawJob(
	withdrawHash string) (*jobs.CooperativeWithdrawJob, error) {
	return getCooperativeWithdrawJob(dtx.stx, withdrawHash)
}

func (dtx *DALTx) PutCooperativeWithdrawJob(
	withdrawHash string, job *jobs.CooperativeWithdrawJob) error {
	return putCooperativeWithdrawJob(dtx.stx, withdrawHash, job)
}

func (dtx *DALTx) DeleteCooperativeWithdrawJob(withdrawHash string) error {
	return deleteCooperativeWithdrawJob(dtx.stx, withdrawHash)
}

func (dtx *DALTx) HasCooperativeWithdrawJob(withdrawHash string) (bool, error) {
	return hasCooperativeWithdrawJob(dtx.stx, withdrawHash)
}

func (dtx *DALTx) GetAllCooperativeWithdrawJobKeys() ([]string, error) {
	return getAllCooperativeWithdrawJobKeys(dtx.stx)
}
