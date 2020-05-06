// Copyright 2018-2020 Celer Network

package metrics

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	handler := GetPromExporter()

	mockRecordMetrics()

	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	time.Sleep(1 * time.Second)
	handler.ServeHTTP(resp, req)
	s := resp.Body.String()

	if ok := checkExportedMetrics(s); !ok {
		t.Error("wrong stats info: ", s)
	}
}

func mockRecordMetrics() {

	IncSvrAdminSendTokenCnt("attempt", "type1")
	IncSvrAdminSendTokenCnt("succeed", "type2")
	IncSvrAdminSendTokenCnt("succeed", "type1")

	IncSvrFwdMsgCnt(1, "attempt")
	IncSvrFwdMsgCnt(1, "succeed")
	IncSvrFwdMsgCnt(2, "succeed")

	IncCommonErrCnt(errors.New("error1"))
	IncCommonErrCnt(errors.New("error2"))
	IncCommonErrCnt(errors.New("error2"))

	IncDelegatePayCnt(DelegatePayAttempt)
	IncDelegatePayCnt(DelegatePaySucceed)

	IncDispatcherMsgCnt("msg1")
	IncDispatcherMsgCnt("msg2")
	IncDispatcherMsgCnt("msg1")

	start := time.Now()
	time.Sleep(25 * time.Millisecond)
	IncDispatcherMsgProcDur(start, "msg1")
	time.Sleep(50 * time.Millisecond)
	IncDispatcherMsgProcDur(start, "msg1")

	IncDispatcherErrCnt("msg1")
	IncDispatcherErrCnt("msg2")
	IncDispatcherErrCnt("msg1")

	IncCoopWithdrawEventCnt()
	IncCoopWithdrawEventCnt()

	IncDisputeSettleEventCnt("state1")
	IncDisputeSettleEventCnt("state2")

	IncDisputeWithdrawEventCnt("state1")
	IncDisputeWithdrawEventCnt("state2")

	IncDepositEventCnt()
	IncDepositEventCnt()

	IncCNodeOpenChanEventCnt(CNodeTcbChan, CNodeOpenChanOK)
	IncCNodeOpenChanEventCnt(CNodeRegularChan, CNodeOpenChanErr)
}

func checkExportedMetrics(s string) bool {
	if strings.Index(s, `celer_svr_admin_sendtoken_count{note="type1",state="attempt"} 1`) < 0 ||
		strings.Index(s, `celer_svr_admin_sendtoken_count{note="type1",state="succeed"} 1`) < 0 ||
		strings.Index(s, `celer_svr_admin_sendtoken_count{note="type2",state="succeed"} 1`) < 0 ||
		strings.Index(s, `celer_svr_forward_message_count{state="attempt",type="1"} 1`) < 0 ||
		strings.Index(s, `celer_svr_forward_message_count{state="succeed",type="1"} 1`) < 0 ||
		strings.Index(s, `celer_svr_forward_message_count{state="succeed",type="2"} 1`) < 0 ||
		strings.Index(s, `celer_common_error_count{error="error1"} 1`) < 0 ||
		strings.Index(s, `celer_common_error_count{error="error2"} 2`) < 0 ||
		strings.Index(s, `celer_delegate_pay_count{state="attempt"} 1`) < 0 ||
		strings.Index(s, `celer_delegate_pay_count{state="succeed"} 1`) < 0 ||
		strings.Index(s, `celer_dispatchers_message_count{type="msg1"} 2`) < 0 ||
		strings.Index(s, `celer_dispatchers_message_count{type="msg2"} 1`) < 0 ||
		strings.Index(s, `celer_dispatchers_message_processing_duration_bucket{type="msg1",le="0"} 0`) < 0 ||
		strings.Index(s, `celer_dispatchers_message_processing_duration_bucket{type="msg1",le="50"} 1`) < 0 ||
		strings.Index(s, `celer_dispatchers_message_processing_duration_bucket{type="msg1",le="100"} 2`) < 0 ||
		strings.Index(s, `celer_dispatchers_message_processing_duration_bucket{type="msg1",le="+Inf"} 2`) < 0 ||
		strings.Index(s, `celer_dispatchers_message_processing_duration_count{type="msg1"} 2`) < 0 ||
		strings.Index(s, `celer_dispatchers_error_handling_count{type="msg1"} 2`) < 0 ||
		strings.Index(s, `celer_dispatchers_error_handling_count{type="msg2"} 1`) < 0 ||
		strings.Index(s, `celer_cooperativewithdraw_event_count 2`) < 0 ||
		strings.Index(s, `celer_dispute_settle_event_count{state="state1"} 1`) < 0 ||
		strings.Index(s, `celer_dispute_settle_event_count{state="state2"} 1`) < 0 ||
		strings.Index(s, `celer_dispute_withdraw_event_count{state="state1"} 1`) < 0 ||
		strings.Index(s, `celer_dispute_withdraw_event_count{state="state2"} 1`) < 0 ||
		strings.Index(s, `celer_deposit_event_count 2`) < 0 ||
		strings.Index(s, `celer_cnode_openchannel_event_count{state="OK",type="tcb"} 1`) < 0 ||
		strings.Index(s, `celer_cnode_openchannel_event_count{state="ERROR",type="regular"} 1`) < 0 {
		return false
	}

	return true
}
func TestPush(t *testing.T) {
	// need to run docker of pushgateway locally and expose the port to localhost:9091
	// and then run prometheus locally to scrape the localhost:9091/metrics
	// Test passed
	PushMetricsToGateway("localhost:9091", "test")
}
