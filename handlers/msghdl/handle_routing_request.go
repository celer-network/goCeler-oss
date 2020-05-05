// Copyright 2018-2020 Celer Network

package msghdl

import (
	"fmt"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/utils"
)

func (h *CelerMsgHandler) HandleRoutingRequest(frame *common.MsgFrame) error {
	msg := frame.Message.GetRoutingRequest()
	var err error
	if h.routeController == nil {
		if config.EventListenerHttp == "" {
			return fmt.Errorf("both router and EventListenerHttp are empty")
		}
		err = utils.RecvRoutingInfo(config.EventListenerHttp, msg)
	} else {
		err = h.routeController.RecvBcastRoutingInfo(msg)
	}
	if err != nil {
		return fmt.Errorf("RecvBcastRoutingInfo err: %w", err)
	}
	return nil
}
