// Copyright 2018-2020 Celer Network

package route

import (
	"flag"
	"fmt"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

var (
	defaultRoute = flag.String("defaultroute", "", "default access osp when can't find a route")
)

type RoutingPolicy int

const (
	NoRoutingPolicy       RoutingPolicy = 1 << iota
	GateWayPolicy         RoutingPolicy = 1 << iota
	ServiceProviderPolicy RoutingPolicy = 1 << iota
)

type Forwarder struct {
	policy     RoutingPolicy
	dal        *storage.DAL
	gatewayOsp ctype.Addr
}

func NewForwarder(policy RoutingPolicy, dal *storage.DAL, gatewayOsp ctype.Addr) *Forwarder {
	if policy != GateWayPolicy && policy != ServiceProviderPolicy {
		log.Fatalln("Routing policy NOT specified")
	}
	r := &Forwarder{policy: policy, dal: dal, gatewayOsp: gatewayOsp}
	return r
}

func (f *Forwarder) LookupNextChannelOnToken(dest ctype.Addr, token ctype.Addr) (ctype.CidType, ctype.Addr, error) {
	cid := ctype.ZeroCid
	var found bool
	var err error
	tokenInfo := utils.GetTokenInfoFromAddress(token)

	// --------------------- GateWayPolicy ----------------------

	if f.policy == GateWayPolicy {
		cid, found, err = f.dal.GetCidByPeerToken(f.gatewayOsp, tokenInfo)
		if err != nil {
			return ctype.ZeroCid, ctype.ZeroAddr, fmt.Errorf("GetCidByPeerToken err: %w", err)
		}
		if !found {
			return ctype.ZeroCid, ctype.ZeroAddr, common.ErrRouteNotFound
		}
		return f.getCidAndPeer(cid)
	}

	// ------------------ ServiceProviderPolicy -----------------

	// first try to lookup locally and use it if found.
	cid, found, err = f.dal.GetCidByPeerToken(dest, tokenInfo)
	if err != nil {
		return ctype.ZeroCid, ctype.ZeroAddr, fmt.Errorf("GetCidByPeerToken err: %w", err)
	}
	if found {
		return f.getCidAndPeer(cid)
	}

	// not having direct channel with me, use two-step route lookup
	// step 1/2: check access osp
	accessOsps, err := f.dal.GetDestTokenOsps(dest, tokenInfo)
	if err != nil {
		return ctype.ZeroCid, ctype.ZeroAddr, fmt.Errorf("GetDestTokenOsps err: %w", err)
	}
	// step 2/2: check route to one of access osp, only take effect if len(accessOsps) > 0
	for _, osp := range accessOsps {
		cid, found, err = f.dal.GetRoutingCid(osp, tokenInfo)
		if err != nil {
			return ctype.ZeroCid, ctype.ZeroAddr, fmt.Errorf("GetRoutingCid on access osp err: %w", err)
		}
		if found {
			return f.getCidAndPeer(cid)
		}
	}

	// not having access osp, try to look up directly in route table.
	// The dst may be an osp which doesn't have access osp.
	cid, found, err = f.dal.GetRoutingCid(dest, tokenInfo)
	if err != nil {
		return ctype.ZeroCid, ctype.ZeroAddr, fmt.Errorf("GetRoutingCid err: %w", err)
	}
	if found {
		return f.getCidAndPeer(cid)
	}

	// If nothing found in two-step lookup, use default route if any.
	if *defaultRoute != "" {
		cid, found, err = f.dal.GetRoutingCid(ctype.Hex2Addr(*defaultRoute), tokenInfo)
		if err != nil {
			return ctype.ZeroCid, ctype.ZeroAddr, fmt.Errorf("GetRoutingCid on default route err: %w", err)
		}
		if found {
			return f.getCidAndPeer(cid)
		}

		cid, found, err = f.dal.GetCidByPeerToken(ctype.Hex2Addr(*defaultRoute), tokenInfo)
		if err != nil {
			return ctype.ZeroCid, ctype.ZeroAddr, fmt.Errorf("GetCidByPeerToken on default route err: %w", err)
		}
		if found {
			return f.getCidAndPeer(cid)
		}
	}

	return ctype.ZeroCid, ctype.ZeroAddr, common.ErrRouteNotFound
}

func (f *Forwarder) LookupIngressChannelOnPay(payID ctype.PayIDType) (ctype.CidType, ctype.Addr, error) {
	cid, found, err := f.dal.GetPayIngressChannel(payID)
	if err != nil {
		return ctype.ZeroCid, ctype.ZeroAddr, err
	}
	if !found {
		return cid, ctype.ZeroAddr, common.ErrPayNotFound
	}
	if cid == ctype.ZeroCid {
		return cid, ctype.ZeroAddr, common.ErrPayNoIngress
	}
	return f.getCidAndPeer(cid)
}

func (f *Forwarder) LookupEgressChannelOnPay(payID ctype.PayIDType) (ctype.CidType, ctype.Addr, error) {
	cid, found, err := f.dal.GetPayEgressChannel(payID)
	if err != nil {
		return ctype.ZeroCid, ctype.ZeroAddr, err
	}
	if !found {
		return cid, ctype.ZeroAddr, common.ErrPayNotFound
	}
	if cid == ctype.ZeroCid {
		return cid, ctype.ZeroAddr, common.ErrPayNoEgress
	}
	return f.getCidAndPeer(cid)
}

func (f *Forwarder) getCidAndPeer(cid ctype.CidType) (ctype.CidType, ctype.Addr, error) {
	peer, found, err := f.dal.GetChanPeer(cid)
	if err != nil {
		return cid, ctype.ZeroAddr, fmt.Errorf("GetChanPeer err: %w, cid: %x", err, cid)
	}
	if !found {
		return cid, ctype.ZeroAddr, fmt.Errorf("GetChanPeer err: %w, cid: %x", common.ErrPeerNotFound, cid)
	}
	return cid, peer, nil
}
