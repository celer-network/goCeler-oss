// Copyright 2018-2020 Celer Network

package pem

import (
	"time"

	"github.com/celer-network/goutils/log"
)

func NewPem(machine string) *PayEventMessage {
	return &PayEventMessage{
		StartTimeStamp: time.Now().UnixNano(),
		Machine:        machine,
		SeqNums:        &SimplexSeqNums{},
		Xnet:           &CrossNetInfo{},
	}
}
func CommitPem(pem *PayEventMessage) {
	pem.EndTimeStamp = time.Now().UnixNano()
	pem.ExecutionTimeMs = (float32)(pem.EndTimeStamp-pem.StartTimeStamp) / 1000000
	pem.EndTimeStamp = 0
	pem.StartTimeStamp = 0
	if zeroSeqNums(pem.SeqNums) {
		pem.SeqNums = nil
	}
	if emptyXnet(pem.Xnet) {
		pem.Xnet = nil
	}

	if len(pem.Error) > 0 {
		log.Errorln("LOGPEM:", pem)
	} else if pem.Nack != nil {
		log.Warnln("LOGPEM:", pem)
	} else if pem.Type == PayMessageType_ROUTING_REQUEST {
		log.Debugln("LOGPEM:", pem)
	} else {
		log.Infoln("LOGPEM:", pem)
	}
}

func NewOcem(machine string) *OpenChannelEventMessage {
	return &OpenChannelEventMessage{
		StartTimeStamp: time.Now().UnixNano(),
		Machine:        machine,
	}
}
func CommitOcem(ocem *OpenChannelEventMessage) {
	ocem.EndTimeStamp = time.Now().UnixNano()
	ocem.ExecutionTimeMs = (float32)(ocem.EndTimeStamp-ocem.StartTimeStamp) / 1000000

	if len(ocem.Error) > 0 {
		log.Errorln("LOGOCEM:", ocem)
	} else {
		log.Infoln("LOGOCEM:", ocem)
	}
}

func zeroSeqNums(seq *SimplexSeqNums) bool {
	if seq == nil {
		return true
	}
	if seq.Out == 0 && seq.OutBase == 0 && seq.In == 0 && seq.InBase == 0 &&
		seq.Stored == 0 && seq.Ack == 0 && seq.LastInflight == 0 {
		return true
	}
	return false
}

func emptyXnet(xnet *CrossNetInfo) bool {
	if xnet == nil {
		return true
	}
	if xnet.GetSrcNetId() == 0 && xnet.GetDstNetId() == 0 &&
		xnet.GetOriginalPayId() == "" && xnet.GetToBridgeAddr() == "" {
		return true
	}
	return false
}
