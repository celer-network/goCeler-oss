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
	p.setNetId(xnet.NetId)
	for bridge, netid := range xnet.NetBridge {
		p.setNetBridge(bridge, netid)
	}
	for netid, bridge := range xnet.BridgeRouting {
		p.setBridgeRouting(netid, bridge)
	}
	for local, remote := range xnet.NetToken {
		for netid, token := range remote {
			p.setNetToken(netid, token, local)
		}
	}
}

func (p *Processor) SetNetId() {
	p.setNetId(*netid)
}

func (p *Processor) SetNetBridge() {
	p.setNetBridge(*bridgeaddr, *netid)
}

func (p *Processor) SetBridgeRouting() {
	p.setBridgeRouting(*netid, *bridgeaddr)
}

func (p *Processor) SetNetToken() {
	p.setNetToken(*netid, *tokenaddr, *localtoken)
}

func (p *Processor) DeleteNetBridge() {
	log.Infoln("Delete netbridge", *bridgeaddr)
	err := p.dal.DeleteNetBridge(ctype.Hex2Addr(*bridgeaddr))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) DeleteBridgeRouting() {
	log.Infoln("Delete bridge routing for dest net id", *netid)
	err := p.dal.DeleteBridgeRouting(*netid)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) DeleteNetToken() {
	log.Infof("Delete net token for net id: %d, token :%s", *netid, *tokenaddr)
	err := p.dal.DeleteNetToken(*netid, utils.GetTokenInfoFromAddress(ctype.Hex2Addr(*tokenaddr)))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) setNetId(netid uint64) {
	log.Infoln("Update net id", netid)
	err := p.dal.PutNetId(netid)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) setNetBridge(bridgeAddr string, netId uint64) {
	log.Infof("Update netbridge addr: %s, net id: %d", bridgeAddr, netId)
	err := p.dal.UpsertNetBridge(ctype.Hex2Addr(bridgeAddr), netId)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) setBridgeRouting(netId uint64, bridgeAddr string) {
	log.Infof("Update bridge routing dest net id: %d, bridge addr: %s", netId, bridgeAddr)
	err := p.dal.UpsertBridgeRouting(netId, ctype.Hex2Addr(bridgeAddr))
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Processor) setNetToken(netId uint64, tokenAddr, localToken string) {
	log.Infof("Update net token for net id: %d, net token %s, local token :%s", netId, tokenAddr, localToken)
	err := p.dal.UpsertNetToken(netId,
		utils.GetTokenInfoFromAddress(ctype.Hex2Addr(tokenAddr)),
		utils.GetTokenInfoFromAddress(ctype.Hex2Addr(localToken)))
	if err != nil {
		log.Fatal(err)
	}
}
