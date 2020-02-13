// Copyright 2018-2019 Celer Network

package cobj

import (
	"errors"
	"time"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/celer-network/goCeler-oss/rtconfig"
)

type CelerStreamWriter struct {
	connManager *rpc.ConnectionManager
}

func NewCelerStreamWriter(connManager *rpc.ConnectionManager) *CelerStreamWriter {
	return &CelerStreamWriter{connManager: connManager}
}

func (w *CelerStreamWriter) WriteCelerMsg(peerTo string, msg *rpc.CelerMsg) error {
	sendTimeout := int64(rtconfig.GetStreamSendTimeoutSecond())
	sendChan := make(chan error, 1)
	// One celer stream implementation.
	celerStream := w.connManager.GetCelerStream(peerTo)
	if celerStream == nil {
		return errors.New("NO_CELER_MSG_STREAM")
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
		log.Errorln("CelerMsg send timed out:", peerTo)
		return errors.New("SEND_MSG_TIMEOUT")
	case err := <-sendChan:
		timer.Stop()
		return err
	}
}
