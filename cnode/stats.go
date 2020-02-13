// Copyright 2018-2019 Celer Network

package cnode

import (
	"errors"

	"github.com/celer-network/goCeler-oss/utils"

	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/ledgerview"
)

func (c *CNode) GetChannelIdForPeer(peer string, tokenAddr string) (ctype.CidType, error) {
	tokenInfo := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tokenAddr))
	cid := c.dal.GetCidByPeerAndToken(ctype.Hex2Bytes(peer), tokenInfo)
	if cid == ctype.ZeroCid {
		return ctype.ZeroCid, errors.New("No cid found for the peer and token pair")
	}
	return cid, nil
}

func (c *CNode) GetBalance(cid ctype.CidType) (*common.ChannelBalance, error) {
	blkNum := c.GetCurrentBlockNumber().Uint64()
	return ledgerview.GetBalance(c.dal, cid, c.nodeConfig.GetOnChainAddr(), blkNum)
}
