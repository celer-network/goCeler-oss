// Copyright 2018-2019 Celer Network

package celersdkintf

// do NOT re-number existing fields,
// these fields are exposed to sdk
const (
	PAY_STATUS_INVALID                 = 0
	PAY_STATUS_PENDING                 = 1
	PAY_STATUS_PAID                    = 2
	PAY_STATUS_PAID_RESOLVED_ONCHAIN   = 3
	PAY_STATUS_UNPAID                  = 4
	PAY_STATUS_UNPAID_EXPIRED          = 5
	PAY_STATUS_UNPAID_REJECTED         = 6
	PAY_STATUS_UNPAID_DEST_UNREACHABLE = 7
	PAY_STATUS_INITIALIZING            = 8 // before pending
)

type Payment struct {
	Sender       string
	Receiver     string
	TokenAddr    string
	AmtWei       string
	UID          string // unique id used for query etc
	PayJSON      string
	Status       int
	PayNoteType  string
	PayNoteJSON  string
	PayTimestamp int64 // in millisecond
}

// PaymentList returns an array of payment
type PaymentList struct {
	PayList []*Payment
}

type E struct {
	Reason string
	Code   int
}
