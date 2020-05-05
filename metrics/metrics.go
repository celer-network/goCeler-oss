// Copyright 2018-2020 Celer Network

package metrics

import (
	"context"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/celer-network/goutils/log"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Metrics, prefix m (or M if needs to be exposed)
var (
	// Metrics for server
	mSvrAdminSendTokenCnt = stats.Int64("celer/server/admin_sendtoken_count", "Number of requests to osp admin/sendtoken interface with various states", stats.UnitDimensionless)
	mSvrFwdMsgCnt         = stats.Int64("celer/server/forward_message_count", "Number of forward message requests to osp with various states", stats.UnitDimensionless)

	// Metrics for common
	mCommonErrCnt = stats.Int64("celer/common/error_count", "Number of errors which are defined in common", stats.UnitDimensionless)

	// Metrics for delegate
	mDelegatePayCnt = stats.Int64("celer/delegate/pay_count", "Number of payments handled in delegate with various states", stats.UnitDimensionless)

	// Metrics for dispatchers
	mDispatcherMsgCnt     = stats.Int64("celer/dispatchers/message_count", "Number of messages with various types", stats.UnitDimensionless)
	mDispatcherMsgProcDur = stats.Float64("celer/dispatchers/message_processing_duration", "Duration of message processing with various handlers", stats.UnitMilliseconds)
	mDispatcherErrCnt     = stats.Int64("celer/dispatchers/error_handling_count", "Number of errors happened after handling messages", stats.UnitDimensionless)

	// Metrics for cooperativewithdraw
	mCoopWithdrawEventCnt = stats.Int64("celer/cooperativewithdraw/event_count", "Number of cooperative withdraw events handled", stats.UnitDimensionless)

	// Metrics for dispute
	mDisputeSettleEventCnt   = stats.Int64("celer/dispute/settle_event_count", "Number of settle events handled with various states", stats.UnitDimensionless)
	mDisputeWithdrawEventCnt = stats.Int64("celer/dispute/withdraw_event_count", "Number of withdraw events handled with various states", stats.UnitDimensionless)

	// Metrics for deposit
	mDepositEventCnt     = stats.Int64("celer/deposit/event_count", "Number of deposit onchain events handled", stats.UnitDimensionless)
	mDepositJobCnt       = stats.Int64("celer/deposit/job_count", "Number of deposit jobs handled", stats.UnitDimensionless)
	mDepositTxCnt        = stats.Int64("celer/deposit/tx_count", "Number of deposit onchain tx submitted", stats.UnitDimensionless)
	mDepositErrCnt       = stats.Int64("celer/deposit/err_count", "Number of deposit errors", stats.UnitDimensionless)
	mDepositPoolAlertCnt = stats.Int64("celer/deposit/pool_alert_count", "Number of deposit pool low balance alerts", stats.UnitDimensionless)

	// Metrics for cnode
	mCNodeOpenChanEventCnt = stats.Int64("celer/cnode/openchannel_event_count", "Number of openchannel events handled with various states", stats.UnitDimensionless)
)

// tag keys, prefix tk
var (
	// note type in SendTokenRequest, value is AnyMessageName(note)
	// tag key for server admin service
	tkSvrAdminNoteType, _  = tag.NewKey("note")  // label to indicate the note type in mSvrAdminSendTokenCnt
	tkSvrAdminSendState, _ = tag.NewKey("state") // label to indicate the send token state in mSvrAdminSendTokenCnt

	tkSvrFwdMsgType, _ = tag.NewKey("type")  // label to indicate the message type in mSvrFwdMsgCnt
	tkSvrFwdState, _   = tag.NewKey("state") // label to indicate the forward state in mSvrFwdMsgCnt

	// tag key for common
	tkCommonErrType, _ = tag.NewKey("error") // label to indicate the error type in mCommonErrCnt

	// tag key for delegate
	tkDelegatePayState, _ = tag.NewKey("state") // label to indicate the payment state in mDelegatePayCnt

	// tag key for dispatchers
	tkDispatcherMsgType, _ = tag.NewKey("type") // label to indicate the message type in mDispatcherMsgCnt

	// tag key for dispute
	tkDisputeEventState, _ = tag.NewKey("state") // label to indicate the event state in mDisputeSettleEventCnt

	// tag key for cnode
	tkCNodeChanType, _  = tag.NewKey("type")  // label to indicate if it is tcb channel in mCNodeChanEventCnt
	tkCNodeChanState, _ = tag.NewKey("state") // label to indicate status of openchannel event in mCNodeChanEventCnt
)

// views, prefix view
var (
	// view for server
	viewSvrAdminSendTokenCnt = &view.View{
		Name:        "svr/admin_sendtoken_count",
		Description: "Number of valid requests to osp admin/sendtoken interface",
		TagKeys:     []tag.Key{tkSvrAdminSendState, tkSvrAdminNoteType},
		Measure:     mSvrAdminSendTokenCnt,
		Aggregation: view.Count(),
	}

	viewSvrFwdMsgCnt = &view.View{
		Name:        "svr/forward_message_count",
		Description: "Number of forward message requests to osp with various states",
		TagKeys:     []tag.Key{tkSvrFwdMsgType, tkSvrFwdState},
		Measure:     mSvrFwdMsgCnt,
		Aggregation: view.Count(),
	}

	// view for common
	viewCommonErrCnt = &view.View{
		Name:        "common/error_count",
		Description: "Number of errors which are defined in common",
		TagKeys:     []tag.Key{tkCommonErrType},
		Measure:     mCommonErrCnt,
		Aggregation: view.Count(),
	}

	// view for delegate
	viewDelegatePayCnt = &view.View{
		Name:        "delegate/pay_count",
		Description: "Number of payments handled in delegate with various states",
		TagKeys:     []tag.Key{tkDelegatePayState},
		Measure:     mDelegatePayCnt,
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

	// view for dispute
	viewDisputeSettleEventCnt = &view.View{
		Name:        "dispute/settle_event_count",
		Description: "Number of settle events handled with various states",
		TagKeys:     []tag.Key{tkDisputeEventState},
		Measure:     mDisputeSettleEventCnt,
		Aggregation: view.Count(),
	}

	viewDisputeWithdrawEventCnt = &view.View{
		Name:        "dispute/withdraw_event_count",
		Description: "Number of withdraw events handled with various states",
		TagKeys:     []tag.Key{tkDisputeEventState},
		Measure:     mDisputeWithdrawEventCnt,
		Aggregation: view.Count(),
	}

	// view for deposit
	viewDepositEventCnt = &view.View{
		Name:        "deposit/event_count",
		Description: "Number of onchain deposit events handled",
		Measure:     mDepositEventCnt,
		Aggregation: view.Count(),
	}

	viewDepositJobtCnt = &view.View{
		Name:        "deposit/job_count",
		Description: "Number of deposit jobs handled",
		Measure:     mDepositJobCnt,
		Aggregation: view.Count(),
	}

	viewDepositTxtCnt = &view.View{
		Name:        "deposit/job_count",
		Description: "Number of deposit onchain tx submitted",
		Measure:     mDepositJobCnt,
		Aggregation: view.Count(),
	}

	viewDepositErrtCnt = &view.View{
		Name:        "deposit/err_count",
		Description: "Number of deposit errors",
		Measure:     mDepositJobCnt,
		Aggregation: view.Count(),
	}

	viewDepositPoolAlertCnt = &view.View{
		Name:        "deposit/pool_alert_count",
		Description: "Number of deposit pool low balance alerts",
		Measure:     mDepositJobCnt,
		Aggregation: view.Count(),
	}

	// view for cnode
	viewCNodeOpenChanEventCnt = &view.View{
		Name:        "cnode/openchannel_event_count",
		Description: "Number of openchannel evnets handled with various states",
		TagKeys:     []tag.Key{tkCNodeChanState, tkCNodeChanType},
		Measure:     mCNodeOpenChanEventCnt,
		Aggregation: view.Count(),
	}
)

const (
	// For server admin service, send token states
	SvrAdminSendAttempt = "attempt"
	SvrAdminSendSucceed = "succeed"

	// For server forward message
	SvrFwdMsgAttempt = "attempt"
	SvrFwdMsgSucceed = "succeed"

	// For delegate, state info for each payment
	DelegatePayAttempt = "attempt"
	DelegatePaySucceed = "succeed"

	// For cnode, channel type and openchannel state
	CNodeTcbChan     = "tcb"
	CNodeRegularChan = "regular"
	CNodeOpenChanOK  = "OK"
	CNodeOpenChanErr = "ERROR"

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
		viewSvrFwdMsgCnt,
		viewCommonErrCnt,
		viewDelegatePayCnt,
		viewDispatcherMsgCnt,
		viewDispatcherMsgProcDur,
		viewDispatcherErrCnt,
		viewCoopWithdrawEventCnt,
		viewDisputeSettleEventCnt,
		viewDisputeWithdrawEventCnt,
		viewCNodeOpenChanEventCnt,
		viewDepositEventCnt,
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

// IncSvrAdminSendTokenCnt records one for mAdminSendTokenCnt
func IncSvrAdminSendTokenCnt(sendstat, notetype string) {
	ctx, err := tag.New(context.Background(),
		tag.Insert(tkSvrAdminSendState, sendstat),
		tag.Insert(tkSvrAdminNoteType, notetype))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mSvrAdminSendTokenCnt.M(1))
}

// IncSvrFwdMsgCnt records one for mSvrFwdMshCnt
func IncSvrFwdMsgCnt(msgtype int32, fwdstat string) {
	ctx, err := tag.New(context.Background(),
		tag.Insert(tkSvrFwdMsgType, strconv.Itoa(int(msgtype))),
		tag.Insert(tkSvrFwdState, fwdstat))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mSvrFwdMsgCnt.M(1))
}

// IncCommonErrCnt records one for mCommonErrCnt
func IncCommonErrCnt(err error) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkCommonErrType, err.Error()))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mCommonErrCnt.M(1))
}

// IncDelegatePayCnt records one for mDelegatePayCnt
func IncDelegatePayCnt(paystat string) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkDelegatePayState, paystat))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mDelegatePayCnt.M(1))
}

// IncDispatcherMsgCnt records one for mDispatcherMsgCnt
func IncDispatcherMsgCnt(msgtype string) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkDispatcherMsgType, msgtype))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mDispatcherMsgCnt.M(1))
}

// IncDispatcherMsgProcDur records duration in millisecond for mDispatcherMsgProcDur
func IncDispatcherMsgProcDur(startTime time.Time, msgtype string) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkDispatcherMsgType, msgtype))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mDispatcherMsgProcDur.M(sinceInMilliseconds(startTime)))
}

// IncDispatcherErrCnt records one for mDispatcherErrCnt
func IncDispatcherErrCnt(msgtype string) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkDispatcherMsgType, msgtype))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mDispatcherErrCnt.M(1))
}

// IncCoopWithdrawEventCnt records one for mCoopWithdrawEventCnt
func IncCoopWithdrawEventCnt() {
	stats.Record(context.Background(), mCoopWithdrawEventCnt.M(1))
}

// IncDisputeSettleEventCnt records one for mDisputeSettleEventCnt
func IncDisputeSettleEventCnt(eventstate string) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkDisputeEventState, eventstate))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mDisputeSettleEventCnt.M(1))
}

// IncDisputeWithdrawEventCnt records one for mDisputeWithdrawEventCnt
func IncDisputeWithdrawEventCnt(eventstate string) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkDisputeEventState, eventstate))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mDisputeWithdrawEventCnt.M(1))
}

// IncDepositEventCnt records one for mDepositEventCnt
func IncDepositEventCnt() {
	stats.Record(context.Background(), mDepositEventCnt.M(1))
}

// IncDepositJobCnt records one for mDepositJobCnt
func IncDepositJobCnt() {
	stats.Record(context.Background(), mDepositJobCnt.M(1))
}

// IncDepositTxCnt records one for mDepositTxCnt
func IncDepositTxCnt() {
	stats.Record(context.Background(), mDepositTxCnt.M(1))
}

// IncDepositErrCnt records one for mDepositErrCnt
func IncDepositErrCnt() {
	stats.Record(context.Background(), mDepositErrCnt.M(1))
}

// IncDepositPoolAlertCnt records one for mDepositPoolAlertCnt
func IncDepositPoolAlertCnt() {
	stats.Record(context.Background(), mDepositPoolAlertCnt.M(1))
}

// IncCNodeOpenChanEventCnt records one for mCNodeOpenChanEvengCnt
func IncCNodeOpenChanEventCnt(chantype, state string) {
	ctx, err := tag.New(context.Background(), tag.Insert(tkCNodeChanState, state),
		tag.Insert(tkCNodeChanType, chantype))
	if err != nil {
		log.Error(err)
		return
	}
	stats.Record(ctx, mCNodeOpenChanEventCnt.M(1))
}

// PushMetricsToGateway push all the metrics to the gateway url. This
// function is designed for ephemeral jobs
func PushMetricsToGateway(url, job string) {
	// ignore the error return by the push gateway
	push.New(url, job).Gatherer(GetPromRegistry()).Grouping("time", time.Now().String()).Add()
}

// sinceInMilliseconds calculate time duration in milliseconds from start time
func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}
