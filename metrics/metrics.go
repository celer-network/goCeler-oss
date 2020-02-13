// Copyright 2018-2019 Celer Network

package metrics

import (
	"context"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Metrics, prefix m (or M if needs to be exposed)
var (
	mSvrAdminSendTokenCnt = stats.Int64("celer/server/admin_sendtoken_count", "Number of requests to osp admin/sendtoken interface with various states", stats.UnitDimensionless)

	// Metrics for common
	mCommonPayPendingTimeoutCnt = stats.Int64("celer/common/pay_pending_timeout_count", "Number of timeout pending payments which have been reset", stats.UnitDimensionless)
	mCommonErrCnt               = stats.Int64("celer/common/error_count", "Number of errors which are defined in common", stats.UnitDimensionless)

	// Metrics for dispatchers
	mDispatcherMsgCnt     = stats.Int64("celer/dispatchers/message_count", "Number of messages with various types", stats.UnitDimensionless)
	mDispatcherMsgProcDur = stats.Float64("celer/dispatchers/message_processing_duration", "Duration of message processing with various handlers", stats.UnitMilliseconds)
	mDispatcherErrCnt     = stats.Int64("celer/dispatchers/error_handling_count", "Number of errors happened after  handling messages", stats.UnitDimensionless)

	// Metrics for cooperativewithdraw
	mCoopWithdrawEventCnt = stats.Int64("celer/cooperativewithdraw/event_count", "Number of cooperative withdraw events handled", stats.UnitDimensionless)

	// Metrics for dispute
	mDisputeSettleEventCnt   = stats.Int64("celer/dispute/settle_event_count", "Number of settle events handled with various states", stats.UnitDimensionless)
	mDisputeWithdrawEventCnt = stats.Int64("celer/dispute/withdraw_event_count", "Number of withdraw events handled with various states", stats.UnitDimensionless)

	// Metrics for cnode
	mCNodeDepositEventCnt  = stats.Int64("celer/cnode/deposit_event_count", "Number of deposit events handled", stats.UnitDimensionless)
	mCNodeOpenChanEventCnt = stats.Int64("celer/cnode/openchannel_event_count", "Number of openchannel events handled with various states", stats.UnitDimensionless)
)

// tag keys, prefix tk
var (
	// note type in SendTokenRequest, value is AnyMessageName(note)
	// tag key for server admin service
	tkSvrAdminNoteType, _ = tag.NewKey("note") // label to indicate the note type in mSvrAdminSendTokenCnt
	tkSvrAdminSendStat, _ = tag.NewKey("stat") // label to indicate the send token state in mSvrAdminSendTokenCnt

	// tag key for common
	tkCommonErrType, _ = tag.NewKey("error") // label to indicate the error type in mCommonErrCnt

	// tag key for dispatchers
	tkDispatcherMsgType, _ = tag.NewKey("type") // label to indicate the message type in mDispatcherMsgCnt

	// tag key for dispute
	tkDisputeEventStat, _ = tag.NewKey("stat") // label to indicate the event state in mDisputeSettleEventCnt

	// tag key for cnode
	tkCNodeChanType, _ = tag.NewKey("type") // label to indicate the openchannel event type in mCNodeChanEventCnt
	tkCNodeChanStat, _ = tag.NewKey("stat") // label to indicate status of openchannel event in mCNodeChanEventCnt
)

// views, prefix view
var (
	viewSvrAdminSendTokenCnt = &view.View{
		Name:        "admin/sendtoken_count",
		Description: "Number of valid requests to osp admin/sendtoken interface",
		TagKeys:     []tag.Key{tkSvrAdminSendStat, tkSvrAdminNoteType},
		Measure:     mSvrAdminSendTokenCnt,
		Aggregation: view.Count(),
	}

	// view for common
	viewCommonPayPendingTimeoutCnt = &view.View{
		Name:        "common/pay_pending_timeout_count",
		Description: "Number of timeout pending payments which have been reset",
		Measure:     mCommonPayPendingTimeoutCnt,
		Aggregation: view.Count(),
	}

	viewCommonErrCnt = &view.View{
		Name:        "common/error_count",
		Description: "Number of errors which are defined in common",
		TagKeys:     []tag.Key{tkCommonErrType},
		Measure:     mCommonErrCnt,
		Aggregation: view.Count(),
	}

	// view for dispatchers
	viewDispatcherMsgCnt = &view.View{
		Name:        "dispatchers/message_count",
		Description: "Number of messages with various types",
		TagKeys:     []tag.Key{tkDispatcherMsgType},
		Measure:     mDispatcherMsgCnt,
		Aggregation: view.Count(),
	}

	viewDispatcherMsgProcDur = &view.View{
		Name:        "dispatchers/message_processing_duration",
		Description: "Duration of message processing with various handlers",
		TagKeys:     []tag.Key{tkDispatcherMsgType},
		Measure:     mDispatcherMsgProcDur,
		Aggregation: view.Distribution(0, 50, 100),
	}

	viewDispatcherErrCnt = &view.View{
		Name:        "dispatchers/error_handling_count",
		Description: "Number of errors happened after  handling messages",
		TagKeys:     []tag.Key{tkDispatcherMsgType},
		Measure:     mDispatcherErrCnt,
		Aggregation: view.Count(),
	}

	// view for cooperativewithdraw
	viewCoopWithdrawEventCnt = &view.View{
		Name:        "cooperativewithdraw/event_count",
		Description: "Number of cooperative withdraw events handled",
		Measure:     mCoopWithdrawEventCnt,
		Aggregation: view.Count(),
	}

	// view for disput
	viewDisputeSettleEventCnt = &view.View{
		Name:        "dispute/settle_event_count",
		Description: "Number of settle events handled with various states",
		TagKeys:     []tag.Key{tkDisputeEventStat},
		Measure:     mDisputeSettleEventCnt,
		Aggregation: view.Count(),
	}

	viewDisputeWithdrawEventCnt = &view.View{
		Name:        "dispute/withdraw_event_count",
		Description: "Number of withdraw events handled with various states",
		TagKeys:     []tag.Key{tkDisputeEventStat},
		Measure:     mDisputeWithdrawEventCnt,
		Aggregation: view.Count(),
	}

	// view for cnode
	viewCNodeDepositEventCnt = &view.View{
		Name:        "cnode/deposit_event_count",
		Description: "Number of deposit events handled",
		Measure:     mCNodeDepositEventCnt,
		Aggregation: view.Count(),
	}

	viewCNodeOpenChanEventCnt = &view.View{
		Name:        "cnode/openchannel_event_count",
		Description: "Number of openchannel evnets handled with various states",
		TagKeys:     []tag.Key{tkCNodeChanStat, tkCNodeChanType},
		Measure:     mCNodeOpenChanEventCnt,
		Aggregation: view.Count(),
	}
)

const (
	// For sever admin service, send token states
	SvrAdminSendAttempt = "attempt"
	SvrAdminSendSucceed = "succeed"

	// For cnode, channel type and openchannel state
	CNodeStandardChan = "standard-chan"
	CNodeOpenChanOK   = "OK"
	CNodeOpenChanErr  = "ERROR"

	// one ether in wei
	ether = "1000000000000000000"
)

// exporter for outputing metrics, opencensus supports various exporters
var promExporter *prometheus.Exporter
var promRegistry *prom.Registry

// Init setup metrics and return http handler for prometheus scraping
func init() {
	// register view, more to be added. ignore errs
	view.Register(viewSvrAdminSendTokenCnt,
		viewCommonPayPendingTimeoutCnt,
		viewCommonErrCnt,
		viewDispatcherMsgCnt,
		viewDispatcherMsgProcDur,
		viewDispatcherErrCnt,
		viewCoopWithdrawEventCnt,
		viewDisputeSettleEventCnt,
		viewDisputeWithdrawEventCnt,
		viewCNodeDepositEventCnt,
		viewCNodeOpenChanEventCnt,
	)

	promRegistry = prom.NewRegistry()
	promExporter, _ = prometheus.NewExporter(prometheus.Options{
		Namespace: "celer",
		Registry:  promRegistry,
	})
	// Register the Prometheus exporter.
	view.RegisterExporter(promExporter)
}

// GetPromExporter would return the prometheus exporter
func GetPromExporter() *prometheus.Exporter {
	return promExporter
}

// GetPromRegistry would return the prometheus resgistry
func GetPromRegistry() *prom.Registry {
	return promRegistry
}

// IncSvrAdminSendTokenCnt records one for mAdminSendTokenCount
func IncSvrAdminSendTokenCnt(sendstat, notetype string) {
	ctx, _ := tag.New(context.Background(),
		tag.Insert(tkSvrAdminSendStat, sendstat),
		tag.Insert(tkSvrAdminNoteType, notetype))
	stats.Record(ctx, mSvrAdminSendTokenCnt.M(1))
}

// IncCommonPayPendingTimeoutCnt records one for mCommonPayPendingTimeoutCnt
func IncCommonPayPendingTimeoutCnt() {
	stats.Record(context.Background(), mCommonPayPendingTimeoutCnt.M(1))
}

// IncCommonErrCnt records one for mCommonErrCnt
func IncCommonErrCnt(err error) {
	ctx, _ := tag.New(context.Background(), tag.Insert(tkCommonErrType, err.Error()))
	stats.Record(ctx, mCommonErrCnt.M(1))
}

// IncDispatcherMsgCnt records one for mDispatcherMsgCnt
func IncDispatcherMsgCnt(msgtype string) {
	ctx, _ := tag.New(context.Background(), tag.Insert(tkDispatcherMsgType, msgtype))
	stats.Record(ctx, mDispatcherMsgCnt.M(1))
}

// IncDispatcherMsgProcDur records duration in millisecond for mDispatcherMsgProcDur
func IncDispatcherMsgProcDur(startTime time.Time, msgtype string) {
	ctx, _ := tag.New(context.Background(), tag.Insert(tkDispatcherMsgType, msgtype))

	stats.Record(ctx, mDispatcherMsgProcDur.M(sinceInMilliseconds(startTime)))
}

// IncDispatcherErrCnt records one for mDispatcherErrCnt
func IncDispatcherErrCnt(msgtype string) {
	ctx, _ := tag.New(context.Background(), tag.Insert(tkDispatcherMsgType, msgtype))

	stats.Record(ctx, mDispatcherErrCnt.M(1))
}

// IncCoopWithdrawEventCnt records one for mCoopWithdrawEventCnt
func IncCoopWithdrawEventCnt() {
	stats.Record(context.Background(), mCoopWithdrawEventCnt.M(1))
}

// IncDisputeSettleEventCnt records one for mDisputeSettleEventCnt
func IncDisputeSettleEventCnt(eventstate string) {
	ctx, _ := tag.New(context.Background(), tag.Insert(tkDisputeEventStat, eventstate))
	stats.Record(ctx, mDisputeSettleEventCnt.M(1))
}

// IncDisputeWithdrawEventCnt records one for mDisputeWithdrawEventCnt
func IncDisputeWithdrawEventCnt(eventstate string) {
	ctx, _ := tag.New(context.Background(), tag.Insert(tkDisputeEventStat, eventstate))
	stats.Record(ctx, mDisputeWithdrawEventCnt.M(1))
}

// IncCNodeDepositEventCnt records one for mCNodeDepositEventCnt
func IncCNodeDepositEventCnt() {
	stats.Record(context.Background(), mCNodeDepositEventCnt.M(1))
}

// IncCNodeOpenChanEventCnt records one for mCNodeOpenChanEvengCnt
func IncCNodeOpenChanEventCnt(chantype, state string) {
	ctx, _ := tag.New(context.Background(), tag.Insert(tkCNodeChanStat, state),
		tag.Insert(tkCNodeChanType, chantype))

	stats.Record(ctx, mCNodeOpenChanEventCnt.M(1))
}

// sinceInMilliseconds calculate time duration in milliseconds from start time
func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}
