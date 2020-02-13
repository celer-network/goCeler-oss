// Copyright 2018-2019 Celer Network

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

	IncCommonPayPendingTimeoutCnt()
	IncCommonPayPendingTimeoutCnt()

	IncCommonErrCnt(errors.New("error1"))
	IncCommonErrCnt(errors.New("error2"))
	IncCommonErrCnt(errors.New("error2"))

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

	IncCNodeDepositEventCnt()
	IncCNodeDepositEventCnt()

	IncCNodeOpenChanEventCnt(CNodeStandardChan, CNodeOpenChanErr)
}

func checkExportedMetrics(s string) bool {
	if strings.Index(s, `celer_admin_sendtoken_count{note="type1",stat="attempt"} 1`) < 0 ||
		strings.Index(s, `celer_admin_sendtoken_count{note="type1",stat="succeed"} 1`) < 0 ||
		strings.Index(s, `celer_admin_sendtoken_count{note="type2",stat="succeed"} 1`) < 0 ||
		strings.Index(s, `celer_common_pay_pending_timeout_count 2`) < 0 ||
		strings.Index(s, `celer_common_error_count{error="error1"} 1`) < 0 ||
		strings.Index(s, `celer_common_error_count{error="error2"} 2`) < 0 ||
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
		strings.Index(s, `celer_dispute_settle_event_count{stat="state1"} 1`) < 0 ||
		strings.Index(s, `celer_dispute_settle_event_count{stat="state2"} 1`) < 0 ||
		strings.Index(s, `celer_dispute_withdraw_event_count{stat="state1"} 1`) < 0 ||
		strings.Index(s, `celer_dispute_withdraw_event_count{stat="state2"} 1`) < 0 ||
		strings.Index(s, `celer_cnode_deposit_event_count 2`) < 0 {
		return false
	}

	return true
}
