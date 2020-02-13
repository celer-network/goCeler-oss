// Copyright 2018-2019 Celer Network

package handlers

import (
	"github.com/celer-network/goCeler-oss/common"
)

type ForwardToServerCallback func(dest string, msg interface{}) (bool, error)

type CelerMsgHandler interface {
	GetMsgName() string
	CelerMsgRunnable
}

type CelerMsgRunnable interface {
	Run(msg *common.MsgFrame) error
}
