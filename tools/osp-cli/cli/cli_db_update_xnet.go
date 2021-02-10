// Copyright 2021 Celer Network

package cli

import (
	"encoding/json"
	"io/ioutil"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

type XnetConfig struct {
	NetId         uint64                       `json:"net_id"`         // local net id
	NetBridge     map[string]uint64            `json:"net_bridge"`     // bridgeAddr -> bridgeNetId
	BridgeRouting map[uint64]string            `json:"bridge_routing"` // destNetId -> nextHopBridgeAddr
	NetToken      map[string]map[uint64]string `json:"net_token"`      // localTokenAddr -> map(remoteNetId -> remoteTokenAddr)
}

func ParseXnetConfig(path string) (*XnetConfig, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	xnet := new(XnetConfig)
	json.Unmarshal(raw, xnet)
	return xnet, nil
}

func (p *Processor) ConfigXnet() {
	if *batchfile == "" {
		log.Fatal("no config file provided")
	}
	xnet, err := ParseXnetConfig(*batchfile)
	if err != nil {
		log.Fatal(err)
	}
	p.updateNetId(xnet.NetId)
	for bridge, netid := range xnet.NetBridge {
		p.updateNetBridge(bridge, netid)
	}
	for netid, bridge := range xnet.BridgeRouting {
		p.updateBridgeRouting(netid, bridge)
	}
	for local, remote := range xnet.NetToken {
		for netid, token := range remote {
			p.updateNetToken(netid, token, local)
		}
	}
}

func (p *Processor) UpdateNetId() {
	p.updateNetId(*netid)
}

func (p *Processor) UpdateNetBridge() {
	p.updateNetBridge(*bridgeaddr, *netid)
}

func (p *Processor) UpdateBridgeRouting() {
	p.updateBridgeRouting(*netid, *bridgeaddr)
}

func (p *Processor) UpdateNetToken() {
	p.updateNetToken(*netid, *tokenaddr, *localtoken)
}

func (p *Processor) updateNetId(netid uint64) {
	log.Infoln("Update net id", netid)
	err := p.dal.PutNetId(netid)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) updateNetBridge(bridgeAddr string, netId uint64) {
	log.Infof("Update netbridge addr: %s, net id: %d", bridgeAddr, netId)
	err := p.dal.UpsertNetBridge(ctype.Hex2Addr(bridgeAddr), netId)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) updateBridgeRouting(netId uint64, bridgeAddr string) {
	log.Infof("Update bridge routing dest net id: %d, bridge addr: %s", netId, bridgeAddr)
	err := p.dal.UpsertBridgeRouting(netId, ctype.Hex2Addr(bridgeAddr))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) updateNetToken(netId uint64, tokenAddr, localToken string) {
	log.Infof("Update net token for net id: %d, net token %s, local token :%s", netId, tokenAddr, localToken)
	err := p.dal.UpsertNetToken(netId,
		utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tokenAddr)),
		utils.GetTokenInfoFromAddress(ctype.Hex2Addr(localToken)))
	if err != nil {
		log.Fatal(err)
	}
}
