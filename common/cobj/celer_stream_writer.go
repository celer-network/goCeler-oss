// Copyright 2018-2020 Celer Network

package cobj

import (
	"time"

	"github.com/celer-network/goCeler/ctype"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goutils/log"
)

type CelerStreamWriter struct {
	connManager *rpc.ConnectionManager
}

func NewCelerStreamWriter(connManager *rpc.ConnectionManager) *CelerStreamWriter {
	return &CelerStreamWriter{connManager: connManager}
}

func (w *CelerStreamWriter) WriteCelerMsg(peerTo ctype.Addr, msg *rpc.CelerMsg) error {
	sendTimeout := int64(rtconfig.GetStreamSendTimeoutSecond())
	sendChan := make(chan error, 1)
	celerStream := w.connManager.GetCelerStream(peerTo)
	if celerStream == nil {
		return common.ErrNoCelerStream
	}
	go func() {
		err := celerStream.SafeSend(msg)
		if err != nil {
			log.Error(err)
		}
		sendChan <- err
	}()

	timer := time.NewTimer(time.Duration(sendTimeout) * time.Second)
	select {
	case <-timer.C:
		w.connManager.CleanupStreams(peerTo)
		log.Errorln("CelerMsg send timed out:", peerTo.Hex())
		return common.ErrCelerMsgTimeout
	case err := <-sendChan:
		timer.Stop()
		return err
	}
}
