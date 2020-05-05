// Copyright 2018-2020 Celer Network

package handlers

import (
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
)

type ForwardToServerCallback func(dest ctype.Addr, retry bool, msg interface{}) (bool, error)

type CelerMsgHandler interface {
	GetMsgName() string
	CelerMsgRunnable
}

type CelerMsgRunnable interface {
	Run(msg *common.MsgFrame) error
}
