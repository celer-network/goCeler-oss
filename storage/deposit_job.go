// Copyright 2018-2019 Celer Network

package storage

import "github.com/celer-network/goCeler-oss/cnode/jobs"

const (
	depositJob           = "dj"
	depositTxHashToJobID = "dthtji"
)

func getDepositJob(st Storage, jobID string) (*jobs.DepositJob, error) {
	var job jobs.DepositJob
	err := st.Get(depositJob, jobID, &job)
	return &job, err
}

func putDepositJob(st Storage, jobID string, job *jobs.DepositJob) error {
	return st.Put(depositJob, jobID, job)
}

func hasDepositJob(st Storage, jobID string) (bool, error) {
	return st.Has(depositJob, jobID)
}

func deleteDepositJob(st Storage, jobID string) error {
	return st.Delete(depositJob, jobID)
}

func getAllDepositJobKeys(st Storage) ([]string, error) {
	return st.GetKeysByPrefix(depositJob, "")
}

func (d *DAL) GetDepositJob(jobID string) (*jobs.DepositJob, error) {
	return getDepositJob(d.st, jobID)
}

func (d *DAL) PutDepositJob(jobID string, job *jobs.DepositJob) error {
	return putDepositJob(d.st, jobID, job)
}

func (d *DAL) HasDepositJob(jobID string) (bool, error) {
	return hasDepositJob(d.st, jobID)
}

func (d *DAL) DeleteDepositJob(jobID string) error {
	return deleteDepositJob(d.st, jobID)
}

func (d *DAL) GetAllDepositJobKeys() ([]string, error) {
	return getAllDepositJobKeys(d.st)
}

func (dtx *DALTx) GetDepositJob(jobID string) (*jobs.DepositJob, error) {
	return getDepositJob(dtx.stx, jobID)
}

func (dtx *DALTx) PutDepositJob(jobID string, job *jobs.DepositJob) error {
	return putDepositJob(dtx.stx, jobID, job)
}

func (dtx *DALTx) HasDepositJob(jobID string) (bool, error) {
	return hasDepositJob(dtx.stx, jobID)
}

func (dtx *DALTx) DeleteDepositJob(jobID string) error {
	return deleteDepositJob(dtx.stx, jobID)
}

func (dtx *DALTx) GetAllDepositJobKeys() ([]string, error) {
	return getAllDepositJobKeys(dtx.stx)
}

func getDepositTxHashToJobID(st Storage, txHash string) (string, error) {
	var jobID string
	err := st.Get(depositTxHashToJobID, txHash, &jobID)
	return jobID, err
}

func putDepositTxHashToJobID(st Storage, txHash string, jobID string) error {
	return st.Put(depositTxHashToJobID, txHash, jobID)
}

func hasDepositTxHashToJobID(st Storage, txHash string) (bool, error) {
	return st.Has(depositTxHashToJobID, txHash)
}

func deleteDepositTxHashToJobID(st Storage, txHash string) error {
	return st.Delete(depositTxHashToJobID, txHash)
}

func (d *DAL) GetDepositTxHashToJobID(txHash string) (string, error) {
	return getDepositTxHashToJobID(d.st, txHash)
}

func (d *DAL) PutDepositTxHashToJobID(txHash string, jobID string) error {
	return putDepositTxHashToJobID(d.st, txHash, jobID)
}

func (d *DAL) HasDepositTxHashToJobID(txHash string) (bool, error) {
	return hasDepositTxHashToJobID(d.st, txHash)
}

func (d *DAL) DeleteDepositTxHashToJobID(txHash string) error {
	return deleteDepositTxHashToJobID(d.st, txHash)
}

func (dtx *DALTx) GetDepositTxHashToJobID(txHash string) (string, error) {
	return getDepositTxHashToJobID(dtx.stx, txHash)
}

func (dtx *DALTx) PutDepositTxHashToJobID(txHash string, jobID string) error {
	return putDepositTxHashToJobID(dtx.stx, txHash, jobID)
}

func (dtx *DALTx) HasDepositTxHashToJobID(txHash string) (bool, error) {
	return hasDepositTxHashToJobID(dtx.stx, txHash)
}

func (dtx *DALTx) DeleteDepositTxHashToJobID(txHash string) error {
	return deleteDepositTxHashToJobID(dtx.stx, txHash)
}
