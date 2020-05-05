// Copyright 2018-2020 Celer Network

package cnode

import (
	"fmt"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goCeler/utils/lease"
	"github.com/celer-network/goutils/log"
)

func (c *CNode) GetChannelIdForPeer(peer, tokenAddr ctype.Addr) (ctype.CidType, error) {
	tokenInfo := utils.GetTokenInfoFromAddress(tokenAddr)
	cid, found, err := c.dal.GetCidByPeerToken(peer, tokenInfo)
	if err != nil {
		return ctype.ZeroCid, err
	}
	if !found {
		return ctype.ZeroCid, fmt.Errorf("No cid found for the peer and token pair")
	}
	return cid, nil
}

func (c *CNode) GetBalance(cid ctype.CidType) (*common.ChannelBalance, error) {
	blkNum := c.GetCurrentBlockNumber().Uint64()
	return ledgerview.GetBalance(c.dal, cid, c.nodeConfig.GetOnChainAddr(), blkNum)
}

// GetJoinStatusForNode gets the join status of an endpoint
// CAVEAT: Note that this will break if we decide to set a default route
// so that LookupNextChannelOnToken always returns a nextHop for any query.
// TODO: May not rely on LookupNextChannelOnToken in the future(yilun)
func (c *CNode) GetJoinStatusForNode(dst, tokenAddr ctype.Addr) rpc.JoinCelerStatus {
	// look up next hop channel
	_, nxtHopAddr, err := c.routeForwarder.LookupNextChannelOnToken(dst, tokenAddr)
	if err != nil {
		return rpc.JoinCelerStatus_NOT_JOIN
	}

	if dst == nxtHopAddr {
		return rpc.JoinCelerStatus_LOCAL
	}

	return rpc.JoinCelerStatus_REMOTE
}

func (c *CNode) registerEventListener() error {
	log.Infoln("register event listener", c.nodeConfig.GetSvrName())
	deadline := time.Now().Add(config.EventListenerLeaseTimeout)
	var err error
	for time.Now().Before(deadline) {
		err = lease.Acquire(c.dal, config.EventListenerLeaseName, c.nodeConfig.GetSvrName(), config.EventListenerLeaseTimeout)
		if err == nil {
			return nil
		}
		log.Warnf("register event listener failed (%s), retry every 10 seconds until %s", err, deadline.UTC())
		time.Sleep(10 * time.Second)
	}
	log.Error(err)
	return fmt.Errorf("register event listener error: %w", err)
}

func (c *CNode) keepAliveEventListener() {
	ticker := time.NewTicker(config.EventListenerLeaseRenewInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.quit:
			return
		case <-ticker.C:
			err := lease.Renew(c.dal, config.EventListenerLeaseName, c.nodeConfig.GetSvrName())
			if err != nil {
				log.Fatalln("failed to renew event listener lease:", err)
			}
		}
	}
}

// SignState signs the data using cnode crypto and return result
func (c *CNode) SignState(in []byte) []byte {
	ret, _ := c.signer.SignEthMessage(in)
	return ret
}
