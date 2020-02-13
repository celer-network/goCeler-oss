// Copyright 2018-2019 Celer Network

package common

import "errors"

// err constants for various errors
var (
	ErrInvalidMsgType              = errors.New("invalid message type")
	ErrPendingSimplex              = errors.New("previous send is still pending")
	ErrInvalidArg                  = errors.New("invalid arguments")
	ErrUnknownTokenType            = errors.New("unknown token type")
	ErrDeadlinePassed              = errors.New("deadline already passed")
	ErrZeroConditions              = errors.New("condpay has 0 conds")
	ErrNoChannel                   = errors.New("no available channel")
	ErrPayNotFound                 = errors.New("payment not found")
	ErrPayStateNotFound            = errors.New("payment state not found")
	ErrPayDestMismatch             = errors.New("pay dest and self mismatch")
	ErrPaySrcMismatch              = errors.New("pay src and self mismatch")
	ErrSimplexParse                = errors.New("cannot parse simplex from storage")
	ErrRateLimited                 = errors.New("rate limited, please try again later")
	ErrInvalidSig                  = errors.New("invalid signature")
	ErrInvalidSeqNum               = errors.New("invalid sequence number")
	ErrInvalidPendingPays          = errors.New("invalid pending pay list")
	ErrInvalidTokenAddress         = errors.New("invalid token address")
	ErrInvalidAccountAddress       = errors.New("invalid account address")
	ErrInvalidAmount               = errors.New("invalid amount")
	ErrInsufficentDepositCapacity  = errors.New("insufficient deposit capacity")
	ErrChannelDescriptorNotInclude = errors.New("my address not included in the channel descriptor")
	ErrOpenEventOnWrongState       = errors.New("open event on wrong state")
	ErrUnparsable                  = errors.New("unparsable")
	ErrInvalidChannelID            = errors.New("channel ID mismatch")
	ErrInvalidChannelPeerFrom      = errors.New("channel peerFrom mismatch")
	ErrInvalidTransferAmt          = errors.New("invalid transfer amount")
	ErrInvalidPendingAmt           = errors.New("invalid total pending amount")
	ErrInvalidPayDeadline          = errors.New("invalid pay resolve deadline")
	ErrInvalidLastPayDeadline      = errors.New("invalid last pay resolve deadline")
	ErrInvalidSettleReason         = errors.New("invalid payment settle reason")
	ErrTooManyPendingPays          = errors.New("too many pending payments")
	ErrNoEnoughBalance             = errors.New("balance not enough")
	ErrInvalidPayResolver          = errors.New("invalid pay resolver address")
	ErrSecretNotRevealed           = errors.New("hash lock secret not revealed")
	ErrEgressPayNotCanceled        = errors.New("egress payment not canceled")
	ErrEgressPayPaid               = errors.New("egress payment already paid")
	ErrPayOnChainResolved          = errors.New("pay already onchain resolved")
	ErrPayAlreadyPending           = errors.New("pay already exists in pending pay list")
	ErrPayRouteLoop                = errors.New("pay route loop")
	ErrInvalidPaySrc               = errors.New("invalid pay source")
	ErrInvalidPayDst               = errors.New("invalid pay destination")
	ErrRouteNotFound               = errors.New("no route to destination")
	ErrPeerNotFound                = errors.New("no peer found for the given cid")
	ErrSimplexStateNotFound        = errors.New("channel simplex state not found")
)

type E struct {
	Reason string
	Code   int
}
