// Copyright 2018-2019 Celer Network

package jobs

import "github.com/celer-network/goCeler-oss/ctype"

type DepositState int

const (
	DepositWaitApprove DepositState = iota
	DepositWaitDeposit              = 1
	DepositSucceeded                = 2
	DepositFailed                   = 3
)

type DepositJob struct {
	JobID         string
	State         DepositState
	Amount        string
	ChannelID     ctype.CidType
	ApproveTxHash string
	DepositTxHash string
	Error         string
}

func NewDepositJob(
	jobID string,
	amount string,
	channelID ctype.CidType,
	approveTxHash string,
	depositTxHash string) *DepositJob {
	var state DepositState
	if approveTxHash == "" {
		state = DepositWaitDeposit
	} else {
		state = DepositWaitApprove
	}
	return &DepositJob{
		JobID:         jobID,
		State:         state,
		Amount:        amount,
		ChannelID:     channelID,
		ApproveTxHash: approveTxHash,
		DepositTxHash: depositTxHash,
	}
}
