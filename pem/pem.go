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
	}
}
func CommitPem(pem *PayEventMessage) {
	pem.EndTimeStamp = time.Now().UnixNano()
	pem.ExecutionTimeMs = (float32)(pem.EndTimeStamp-pem.StartTimeStamp) / 1000000

	if len(pem.Error) > 0 {
		log.Errorln("LOGPEM:", pem)
	} else if pem.Nack != nil {
		log.Warnln("LOGPEM:", pem)
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
