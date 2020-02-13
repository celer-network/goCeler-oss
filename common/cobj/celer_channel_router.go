// Copyright 2018-2019 Celer Network

package cobj

import (
	"errors"
	"flag"
	"strings"

	"github.com/celer-network/goCeler-oss/entity"

	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/storage"
)

var (
	DefaultEdgeOsp = flag.String("defaultedgeosp", "", "default edge osp when can't find a route to a client")
)

type CelerChannelRouter struct {
	policy     common.RoutingPolicy
	dal        *storage.DAL
	gatewayOsp ctype.Addr
}

func NewCelerChannelRouter(policy common.RoutingPolicy, dal *storage.DAL, gatewayOsp ctype.Addr) *CelerChannelRouter {
	if policy != common.GateWayPolicy && policy != common.ServiceProviderPolicy {
		log.Fatalln("Routing policy NOT specified")
	}
	r := &CelerChannelRouter{policy: policy, dal: dal, gatewayOsp: gatewayOsp}
	return r
}

func (r *CelerChannelRouter) LookupNextChannelOnToken(dst string, tokenAddr string) (ctype.CidType, string, error) {
	cid := ctype.ZeroCid
	var err error
	tokenInfo := &entity.TokenInfo{
		TokenAddress: ctype.Hex2Bytes(tokenAddr),
		TokenType:    entity.TokenType_ERC20,
	}
	if tokenAddr == common.EthContractAddr {
		tokenInfo.TokenType = entity.TokenType_ETH
	}
	if r.policy == common.GateWayPolicy {
		cid = r.dal.GetCidByPeerAndToken(r.gatewayOsp.Bytes(), tokenInfo)
	} else {
		// first try to lookup locally and use it if found.
		cid = r.dal.GetCidByPeerAndToken(ctype.Hex2Bytes(dst), tokenInfo)
		if cid == ctype.ZeroCid {
			// not having direct channel with me, use two-step route lookup
			token := ctype.Hex2Addr(tokenAddr)
			// two step 1: check serving osp
			servingOsps, servingOspErr := r.dal.GetServingOsps(ctype.Hex2Addr(dst), token)
			if servingOspErr == nil && len(servingOsps) != 0 {
				// has serving osps
				// two step 2: check route to one of serving osp
				for osp := range servingOsps {
					cid, err = r.dal.GetRoute(ctype.Addr2Hex(osp), tokenAddr)
					if err == nil {
						break
					}
				}
			} else {
				// Not having serving osp, try to look up straight in route table.
				// The dst may be an osp which doesn't have serving osp.
				cid, err = r.dal.GetRoute(dst, tokenAddr)
			}
			// If nothing found in two-step lookup, use default route if any.
			if cid == ctype.ZeroCid && *DefaultEdgeOsp != "" {
				cid, err = r.dal.GetRoute(*DefaultEdgeOsp, tokenAddr)
				if cid == ctype.ZeroCid {
					cid = r.dal.GetCidByPeerAndToken(ctype.Hex2Bytes(*DefaultEdgeOsp), tokenInfo)
				}
			}
		}
	}
	if cid == ctype.ZeroCid {
		return ctype.ZeroCid, "", common.ErrRouteNotFound
	}

	peer, err := r.dal.GetPeer(cid)
	if err != nil {
		log.Errorln(err, ctype.Cid2Hex(cid))
		return cid, "", common.ErrPeerNotFound
	}
	return cid, peer, nil
}

func (r *CelerChannelRouter) LookupIngressChannelOnPay(
	payID ctype.PayIDType) (ctype.CidType, string, error) {

	cid, _, _, err := r.dal.GetPayIngressState(payID)
	if err != nil {
		return ctype.ZeroCid, "", common.ErrPayStateNotFound
	}
	peer, err := r.dal.GetPeer(cid)
	if err != nil {
		return cid, "", common.ErrPeerNotFound
	}
	return cid, peer, nil
}

func (r *CelerChannelRouter) LookupEgressChannelOnPay(
	payID ctype.PayIDType) (ctype.CidType, string, error) {

	cid, _, _, err := r.dal.GetPayEgressState(payID)
	if err != nil {
		return ctype.ZeroCid, "", common.ErrPayStateNotFound
	}
	peer, err := r.dal.GetPeer(cid)
	if err != nil {
		return cid, "", common.ErrPeerNotFound
	}
	return cid, peer, nil
}

// LookupNextChannel is deprecated
func (r *CelerChannelRouter) LookupNextChannel(dst string) (ctype.CidType, error) {
	if r.policy == common.GateWayPolicy {
		return lookupChannel(r.dal, "*")
	}
	return lookupChannel(r.dal, dst)
}

func lookupChannel(dal *storage.DAL, lookupKey string) (ctype.CidType, error) {
	rt, err := dal.GetAllRoutingTableKeysToDest(lookupKey)
	if err != nil || len(rt) == 0 {
		return ctype.ZeroCid, err
	}
	// Routing table entry: "dest@token"
	parsed := strings.Split(rt[0], common.RoutingTableDestTokenSpliter)
	if len(parsed) == 2 {
		return dal.GetRoute(lookupKey, parsed[1])
	}
	return ctype.ZeroCid, errors.New("Routing Table Entry Incorrect:" + rt[0])
}
