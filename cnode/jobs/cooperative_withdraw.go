// Copyright 2018-2019 Celer Network

package jobs

type CooperativeWithdrawState int

const (
	CooperativeWithdrawWaitResponse CooperativeWithdrawState = iota
	CooperativeWithdrawWaitTx                                = 1
	CooperativeWithdrawSucceeded                             = 2
	CooperativeWithdrawFailed                                = 3
)

type CooperativeWithdrawJob struct {
	WithdrawHash string
	State        CooperativeWithdrawState
	TxHash       string
	Error        string
}

func NewCooperativeWithdrawJob(withdrawHash string) *CooperativeWithdrawJob {
	return &CooperativeWithdrawJob{
		WithdrawHash: withdrawHash,
		State:        CooperativeWithdrawWaitResponse,
	}
}
