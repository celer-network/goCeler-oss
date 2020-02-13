// Copyright 2018-2019 Celer Network

package route

import (
	"errors"
	"fmt"
	"sync"

	"github.com/RyanCarrier/dijkstra"
	log "github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/route/graph"
	"github.com/ethereum/go-ethereum/common"
)

// Edge describes event happenning to a channel.
type Edge = graph.Edge

// RoutingTableBuilderDAL defines DAL layer that to support route calculation.
type RoutingTableBuilderDAL interface {
	// PutLocalPeerCid(ctype.Addr /*localPeer*/, ctype.Addr /*tokenAddr*/, ctype.CidType) error
	// DeleteLocalPeerCid(ctype.Addr /*localPeer*/, ctype.Addr /*tokenAddr*/) error
	// tokenAddr, clientAddr -> map[ospAddr]bool
	GetAllServingOsps() (map[ctype.Addr]map[ctype.Addr]servingOspMap, error)
	PutServingOsp(ctype.Addr /*clientAddr*/, ctype.Addr /*tokenAddr*/, ctype.Addr /*ospAddr*/) error
	DeleteServingOsp(ctype.Addr /*clientAddr*/, ctype.Addr /*tokenAddr*/, ctype.Addr /*ospAddr*/) error

	// tokenAddr, dstOsp -> cid
	GetAllRoutes() (map[ctype.Addr]map[ctype.Addr]ctype.CidType, error)
	PutRoute(string, string, ctype.CidType) error
	DeleteRoute(string, string) error
	PutEdge(ctype.Addr, ctype.CidType, *Edge) error
	GetEdges(ctype.Addr) (map[ctype.CidType]*Edge, error)
	DeleteEdge(ctype.Addr, ctype.CidType) error
	GetAllEdgeTokens() (map[ctype.Addr]bool, error)
}

type edgeMap = map[ctype.CidType]*Edge
type servingOspMap = map[ctype.Addr]bool
type routingTable = map[ctype.Addr]common.Hash

type RoutingTableBuilder struct {
	// tokenAddr, clientAddr -> set of osps
	lastServingOsps map[ctype.Addr]map[ctype.Addr]servingOspMap
	// tokenAddr, dstOspAddr -> cid
	lastNextHopCidToOsp map[ctype.Addr]map[ctype.Addr]ctype.CidType
	routeLock           sync.RWMutex
	dal                 RoutingTableBuilderDAL
	edges               map[ctype.Addr]edgeMap
	ospInfo             map[ctype.Addr]uint64

	// My address
	srcAddr ctype.Addr
}

type addrToCidMap = map[ctype.Addr]ctype.CidType

// NewRoutingTableBuilder creates a routing table builder and init it.
func NewRoutingTableBuilder(srcAddr ctype.Addr, dal RoutingTableBuilderDAL) *RoutingTableBuilder {
	b := &RoutingTableBuilder{
		lastServingOsps:     make(map[ctype.Addr]map[ctype.Addr]servingOspMap),
		lastNextHopCidToOsp: make(map[ctype.Addr]map[ctype.Addr]ctype.CidType),
		srcAddr:             srcAddr,
		dal:                 dal,
		edges:               make(map[ctype.Addr]edgeMap),
		ospInfo:             make(map[ctype.Addr]uint64),
	}

	// init edges.
	tks, err := dal.GetAllEdgeTokens()
	if err != nil {
		log.Errorln(err)
		return nil
	}
	routes, getRouteErr := dal.GetAllRoutes()
	if getRouteErr != nil {
		log.Errorln(getRouteErr)
		return nil
	}
	b.lastNextHopCidToOsp = routes

	servingOsps, getServingOspErr := dal.GetAllServingOsps()
	if getServingOspErr != nil {
		log.Errorln(getServingOspErr)
		return nil
	}
	b.lastServingOsps = servingOsps
	for token := range tks {
		tokenEdges, edgeErr := dal.GetEdges(token)
		if edgeErr != nil {
			log.Errorln(edgeErr)
			return nil
		}

		if tokenEdges != nil {
			b.edges[token] = tokenEdges
		}
	}

	return b
}

func (b *RoutingTableBuilder) AddEdge(
	p1 ctype.Addr, p2 ctype.Addr, cid ctype.CidType, tokenAddr ctype.Addr) error {
	b.routeLock.Lock()
	defer b.routeLock.Unlock()
	log.Infof("Adding cid %x", cid.Bytes())
	e := &Edge{P1: p1, P2: p2, Cid: cid, TokenAddr: tokenAddr}
	err := b.dal.PutEdge(tokenAddr, cid, e)
	if err != nil {
		log.Errorln(err)
		return err
	}
	if b.edges[tokenAddr] == nil {
		b.edges[tokenAddr] = make(edgeMap)
	}
	b.edges[tokenAddr][cid] = e
	return nil
}
func (b *RoutingTableBuilder) RemoveEdge(cid ctype.CidType) error {
	b.routeLock.Lock()
	defer b.routeLock.Unlock()
	log.Infof("Removing cid %x", cid.Bytes())
	var token ctype.Addr
	found := false
	for token, edges := range b.edges {
		if _, ok := edges[cid]; ok {
			delete(edges, cid)
			delErr := b.dal.DeleteEdge(token, cid)
			if delErr != nil {
				log.Errorln(delErr)
				return delErr
			}
			found = true
			break
		}
	}
	if !found {
		errStr := fmt.Sprintf("cid %s doesn't exist in any token addr", ctype.Cid2Hex(cid))
		log.Errorf(errStr)
		return errors.New(errStr)
	}
	if len(b.edges[token]) == 0 {
		delete(b.edges, token)
	}
	return nil
}

// MarkOsp marks an Osp as a router and records its block number
func (b *RoutingTableBuilder) MarkOsp(osp ctype.Addr, blknum uint64) {
	b.routeLock.Lock()
	defer b.routeLock.Unlock()
	log.Infof("%x joining", osp.Bytes())
	b.ospInfo[osp] = blknum
}

// UnmarkOsp removes an Osp in routers
func (b *RoutingTableBuilder) UnmarkOsp(osp ctype.Addr) {
	b.routeLock.Lock()
	defer b.routeLock.Unlock()
	log.Infof("%x leaving", osp.Bytes())
	delete(b.ospInfo, osp)
}

// GetAllTokens returns all tokens ever seen by edge series.
func (b *RoutingTableBuilder) GetAllTokens() map[ctype.Addr]bool {
	b.routeLock.RLock()
	defer b.routeLock.RUnlock()
	tks := make(map[ctype.Addr]bool)
	for k := range b.edges {
		tks[k] = true
	}
	return tks
}

// GetOspInfo copies the all osp info and returns them back
func (b *RoutingTableBuilder) GetOspInfo() map[ctype.Addr]uint64 {
	allOspInfo := make(map[ctype.Addr]uint64)

	b.routeLock.RLock()
	defer b.routeLock.RUnlock()

	for addr := range b.ospInfo {
		allOspInfo[addr] = b.ospInfo[addr]
	}

	return allOspInfo
}

// IsOspExisting returns whether an Osp is marked or not
func (b *RoutingTableBuilder) IsOspExisting(osp ctype.Addr) bool {
	b.routeLock.RLock()
	defer b.routeLock.RUnlock()

	_, ok := b.ospInfo[osp]

	return ok
}

func (b *RoutingTableBuilder) Build(tokenAddr ctype.Addr) (addrToCidMap, error) {
	log.Infof("building routing table for %x", tokenAddr)

	// client->serving osps
	servingOsps := make(map[ctype.Addr]servingOspMap)
	b.routeLock.RLock()
	defer b.routeLock.RUnlock()
	allMarkedOsps := make(map[ctype.Addr]bool)
	for osp := range b.ospInfo {
		allMarkedOsps[osp] = true
	}
	allMarkedOsps[b.srcAddr] = true

	graph, peerToCid := b.buildOspGraph(tokenAddr, servingOsps, allMarkedOsps)

	nextHopCidToOsp, err := b.calcRoutesAmongOsps(graph, allMarkedOsps, peerToCid)
	if err != nil {
		return nil, err
	}

	// optimization here: only update DB if there is a change. Applied to both servingOsps table and routing table.
	for client, osps := range servingOsps {
		for osp := range osps {
			lastServingOspOnToken, ok := b.lastServingOsps[tokenAddr]
			if ok {
				servingOspSet, ok := lastServingOspOnToken[client]
				if ok && servingOspSet[osp] {
					// osp is already in client's serving osp set. No need to update db.
					continue
				}
			}
			log.Debugln("puting serving osps", ctype.Addr2Hex(client), ctype.Addr2Hex(osp))
			putErr := b.dal.PutServingOsp(client, tokenAddr, osp)
			if putErr != nil {
				log.Errorln(putErr)
				// delete osp from map in case of db update failure so that it has chance to be updated next time.
				delete(osps, osp)
			}
		}
	}
	for client, osps := range b.lastServingOsps[tokenAddr] {
		for osp := range osps {
			if servingOsps[client] != nil && servingOsps[client][osp] {
				// osp is still in client's serving osp set. No need to update db.
				continue
			} else {
				// osp is no longer serving client.
				log.Debugln("deleting serving osps", ctype.Addr2Hex(client), ctype.Addr2Hex(osp))
				delErr := b.dal.DeleteServingOsp(client, tokenAddr, osp)
				if delErr != nil {
					log.Errorln(delErr)
					// add back osp in case of db delete failure so that it has chance to be deleted next time.
					servingOsps[client][osp] = true
				}
			}
		}
	}
	b.lastServingOsps[tokenAddr] = servingOsps
	for dst, cid := range nextHopCidToOsp {
		lastNextHopCidToOspOnToken, ok := b.lastNextHopCidToOsp[tokenAddr]
		if ok && lastNextHopCidToOspOnToken[dst] == cid {
			continue
		}
		log.Debugf("adding route to %x on token %x", dst.Bytes(), tokenAddr.Bytes())
		putErr := b.dal.PutRoute(ctype.Addr2Hex(dst), ctype.Addr2Hex(tokenAddr), cid)
		if putErr != nil {
			log.Errorln(putErr)
			// Remove the route entry in memory to be sync with database so that build next time will update db again.
			delete(nextHopCidToOsp, dst)
		}
	}
	for dst, cid := range b.lastNextHopCidToOsp[tokenAddr] {
		log.Debugf("old table has route to %x", dst.Bytes())
		if _, ok := nextHopCidToOsp[dst]; !ok {
			log.Debugf("Deleting route to %x on token %x", dst.Bytes(), tokenAddr.Bytes())
			delErr := b.dal.DeleteRoute(ctype.Addr2Hex(dst), ctype.Addr2Hex(tokenAddr))
			if delErr != nil {
				log.Errorln(delErr)
				// Add back route entry in memory to be sync with database so that build next time will delete db again.
				nextHopCidToOsp[dst] = cid
			}
		}
	}
	b.lastNextHopCidToOsp[tokenAddr] = nextHopCidToOsp
	ret := make(addrToCidMap)
	// copy result to return in case of someone modifying the returned value.
	for k, v := range nextHopCidToOsp {
		ret[k] = v
	}
	return ret, nil
}

type ospClientMap = map[ctype.Addr]map[ctype.Addr]bool

func (b *RoutingTableBuilder) calcRoutesAmongOsps(
	graph *dijkstra.Graph, allMarkedOsps map[ctype.Addr]bool, peerToCid addrToCidMap) (addrToCidMap, error) {
	srcAddrStr := ctype.Addr2Hex(b.srcAddr)
	srcAddrMapping, srcErr := graph.GetMapping(srcAddrStr)
	if srcErr != nil {
		log.Errorln(srcErr)
		return nil, srcErr
	}
	// dest osp -> next hop osp cid
	nextHopCidToOsp := make(addrToCidMap)
	// Calculate routes from src to all osps
	for osp := range allMarkedOsps {
		if osp == b.srcAddr {
			// skip myself
			continue
		}

		destAddrStr := ctype.Addr2Hex(osp)
		destAddrMapping, destErr := graph.GetMapping(destAddrStr)
		if destErr != nil {
			log.Errorln(destErr)
			return nil, destErr
		}
		log.Debugln("calc", destAddrStr)
		// The graph library doesn't provide API to run dijkstra once for all nodes.
		best, shortErr := graph.Shortest(srcAddrMapping, destAddrMapping)
		if shortErr != nil {
			log.Errorln(shortErr, "- could be a natural result of network partition")
			continue
		}
		nextHop, shortErr := graph.GetMapped(best.Path[1])
		if shortErr != nil {
			log.Errorln(shortErr)
			return nil, shortErr
		}
		nextHopCidToOsp[osp] = peerToCid[ctype.Hex2Addr(nextHop)]
	}
	return nextHopCidToOsp, nil
}

func (b *RoutingTableBuilder) buildOspGraph(tokenAddr ctype.Addr,
	servingOsps map[ctype.Addr]servingOspMap, allMarkedOsps map[ctype.Addr]bool) (*dijkstra.Graph, addrToCidMap) {
	peerToCid := make(addrToCidMap)
	graph := dijkstra.NewGraph()
	// graph.AddMappedVertex(ctype.Addr2Hex(b.srcAddr))

	// Read each event one by one and look for event related to tokenAddr.
	// Set up osp topology.
	for _, edge := range b.edges[tokenAddr] {
		// Record cid that I connect straight as value in routing table is next hop cid, not addr.
		// This needs to be done before non-osp check.
		if b.srcAddr == edge.P1 {
			peerToCid[edge.P2] = edge.Cid
		} else if b.srcAddr == edge.P2 {
			peerToCid[edge.P1] = edge.Cid
		}
		_, p1IsOsp := allMarkedOsps[edge.P1]
		_, p2IsOsp := allMarkedOsps[edge.P2]

		if !p1IsOsp && !p2IsOsp {
			// skip routing calc if none of them is osp.
			continue
		}

		p1Str := ctype.Addr2Hex(edge.P1)
		p2Str := ctype.Addr2Hex(edge.P2)
		// Only add edge to graph when both are osps.
		if p1IsOsp && p2IsOsp {
			log.Debugln("adding", p1Str, p2Str)
			graph.AddMappedVertex(p1Str)
			graph.AddMappedVertex(p2Str)
			graph.AddMappedArc(p1Str, p2Str, 1)
			graph.AddMappedArc(p2Str, p1Str, 1)
			continue
		}

		// One is osp, the other is client, save client for final routing table setup.
		// The edge and client is not in routing calculation though. We only calculate
		// OSP1-OSP2 and use same next hop for client connecting to OSP1 or OSP2.
		if p1IsOsp {
			if servingOsps[edge.P2] == nil {
				servingOsps[edge.P2] = make(map[ctype.Addr]bool)
			}
			// p1 is osp, p2 is client
			servingOsps[edge.P2][edge.P1] = true
			log.Debugln("adding", p1Str)
			graph.AddMappedVertex(p1Str)
		} else if p2IsOsp {
			if servingOsps[edge.P1] == nil {
				servingOsps[edge.P1] = make(map[ctype.Addr]bool)
			}
			// p2 is osp, p1 is client
			servingOsps[edge.P1][edge.P2] = true
			log.Debugln("adding", p2Str)
			graph.AddMappedVertex(p2Str)
		}
	}
	return graph, peerToCid
}
