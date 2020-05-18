// Copyright 2018-2020 Celer Network

package route

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

// Edge describes event happening to a channel.
type Edge = structs.Edge

type OspInfo struct {
	// block number of last onchain routerRegistry update
	RegistryBlock uint64
	// time of last route message update
	UpdateTime time.Time
}

type NeighborInfo struct {
	// time of last route message update
	UpdateTime time.Time
	// tokens and cids of connected channels
	TokenCids map[ctype.Addr]ctype.CidType
}

type OspEdge struct {
	edge *Edge
	// balance reported by edge.P1
	balance1 *big.Int
	// balance reported by edge.P2
	balance2 *big.Int
	// latest report of P1 or P2
	// since stream is bidirectional, we assume
	// the edge is alive as long as one peer reports
	updateTime time.Time
}

type edgeMap = map[ctype.CidType]*Edge
type accessOspSet = map[ctype.Addr]bool

type routingTableBuilder struct {
	myAddr      ctype.Addr
	dal         *storage.DAL
	edges       map[ctype.Addr]edgeMap                      // tokenAddr -> { cid -> edge }, including non-OSP edges
	ospEdges    map[ctype.CidType]*OspEdge                  // cid -> OspEdge, only OSP-to-OSP edges
	osps        map[ctype.Addr]*OspInfo                     // ospAddr -> OspInfo
	neighbors   map[ctype.Addr]*NeighborInfo                // neighborAddr -> NeighborInfo
	accessOsps  map[ctype.Addr]map[ctype.Addr]accessOspSet  // tokenAddr, clientAddr -> set of ospAddrs
	nextHopCids map[ctype.Addr]map[ctype.Addr]ctype.CidType // tokenAddr, dstOspAddr -> cid
	graphLock   sync.RWMutex                                // protect edges, ospEdges, osps, neighbors
	routeLock   sync.RWMutex                                // protect accessOsps, nextHopCids
	buildLock   sync.Mutex                                  // serialize building processes
}

// newRoutingTableBuilder creates a routing table builder and init it.
func newRoutingTableBuilder(myAddr ctype.Addr, dal *storage.DAL) *routingTableBuilder {
	b := &routingTableBuilder{
		myAddr:      myAddr,
		dal:         dal,
		edges:       make(map[ctype.Addr]edgeMap),
		ospEdges:    make(map[ctype.CidType]*OspEdge),
		osps:        make(map[ctype.Addr]*OspInfo),
		neighbors:   make(map[ctype.Addr]*NeighborInfo),
		accessOsps:  make(map[ctype.Addr]map[ctype.Addr]accessOspSet),
		nextHopCids: make(map[ctype.Addr]map[ctype.Addr]ctype.CidType),
	}

	// init edges.
	edges, err := dal.GetAllEdges()
	if err != nil {
		log.Errorln(err)
		return nil
	}

	for _, e := range edges {
		token := e.Token
		if _, ok := b.edges[token]; !ok {
			b.edges[token] = make(edgeMap)
		}
		b.edges[token][e.Cid] = e
	}

	routes, err := dal.GetAllRoutingCids()
	if err != nil {
		log.Errorln(err)
		return nil
	}
	b.nextHopCids = routes

	accessOsps, err := dal.GetAllDestTokenOsps()
	if err != nil {
		log.Errorln(err)
		return nil
	}
	b.accessOsps = accessOsps

	return b
}

func (b *routingTableBuilder) addEdge(p1, p2 ctype.Addr, cid ctype.CidType, tokenAddr ctype.Addr) error {
	b.graphLock.Lock()
	defer b.graphLock.Unlock()
	log.Infof("Adding edge cid %x", cid.Bytes())
	token := utils.GetTokenInfoFromAddress(tokenAddr)
	err := b.dal.InsertEdge(token, cid, p1, p2)
	if err != nil {
		log.Errorln(err)
		return err
	}
	if b.edges[tokenAddr] == nil {
		b.edges[tokenAddr] = make(edgeMap)
	}
	e := &Edge{P1: p1, P2: p2, Cid: cid, Token: tokenAddr}
	b.edges[tokenAddr][cid] = e

	myEdge, peerAddr := b.isMyEdge(e)
	// update neigbhbor
	if myEdge && b.osps[peerAddr] != nil {
		if b.neighbors[peerAddr] == nil {
			err = fmt.Errorf("neighbors map for peer %x not intialized", peerAddr)
			log.Error(err)
			return err
		}
		b.neighbors[peerAddr].UpdateTime = now()
		b.neighbors[peerAddr].TokenCids[tokenAddr] = cid
	}

	return nil
}

func (b *routingTableBuilder) removeEdge(cid ctype.CidType) error {
	b.graphLock.Lock()
	defer b.graphLock.Unlock()
	log.Infof("Removing cid %x", cid.Bytes())
	delete(b.ospEdges, cid)
	found := false
	for token, edges := range b.edges {
		if edge, ok := edges[cid]; ok {
			delete(edges, cid)
			if len(edges) == 0 {
				delete(b.edges, token)
			}
			myEdge, peerAddr := b.isMyEdge(edge)
			// update neigbhbor
			if myEdge && b.osps[peerAddr] != nil {
				if b.neighbors[peerAddr] == nil {
					log.Errorf("neighbors map for peer %x not intialized", peerAddr)
				} else if b.neighbors[peerAddr].TokenCids[edge.Token] == cid {
					delete(b.neighbors[peerAddr].TokenCids, edge.Token)
				}
			}
			err := b.dal.DeleteEdge(cid)
			if err != nil {
				log.Errorln(err)
				return err
			}
			found = true
			break
		}
	}
	if !found {
		errStr := fmt.Sprintf("cid %x doesn't exist in any token addr", cid)
		log.Errorf(errStr)
		return fmt.Errorf(errStr)
	}
	return nil
}

func (b *routingTableBuilder) isMyEdge(e *Edge) (bool, ctype.Addr) {
	var peerAddr ctype.Addr
	if e.P1 == b.myAddr {
		peerAddr = e.P2
	} else if e.P2 == b.myAddr {
		peerAddr = e.P1
	} else {
		return false, peerAddr
	}
	return true, peerAddr
}

func (b *routingTableBuilder) updateOspEdge(
	cid ctype.CidType, balance *big.Int, peerFrom ctype.Addr, timestamp time.Time) {
	b.graphLock.Lock()
	defer b.graphLock.Unlock()
	ospEdge, ok := b.ospEdges[cid]
	if !ok {
		for _, edges := range b.edges {
			edge := edges[cid]
			if edge != nil {
				ospEdge = &OspEdge{edge: edge}
				b.ospEdges[cid] = ospEdge
				break
			}
		}
	}
	if ospEdge == nil {
		log.Warnf("tried to add non-existing edge %x as OSP edge", cid)
		return
	}
	if timestamp.After(ospEdge.updateTime) {
		if peerFrom == ospEdge.edge.P1 {
			ospEdge.balance1 = balance
		} else if peerFrom == ospEdge.edge.P2 {
			ospEdge.balance2 = balance
		} else {
			log.Warnf("OSP edge %x peer not match %x", cid, peerFrom)
			return
		}
		ospEdge.updateTime = timestamp
	}
}

// markOsp marks an Osp as a router and records its block number
func (b *routingTableBuilder) markOsp(ospAddr ctype.Addr, blknum uint64) {
	b.graphLock.Lock()
	defer b.graphLock.Unlock()
	log.Infof("markOsp: %x", ospAddr)
	if _, ok := b.osps[ospAddr]; ok {
		b.osps[ospAddr].RegistryBlock = blknum
	} else {
		now := now()
		b.osps[ospAddr] = &OspInfo{
			RegistryBlock: blknum,
			UpdateTime:    now,
		}
		log.Debugf("add osp %x to neighbor map", ospAddr)
		cids, tokens, err := b.dal.GetCidTokensByPeer(ospAddr)
		if err != nil {
			log.Errorln("GetCidTokensByPeer err:", err)
			return
		}
		tokencids := make(map[ctype.Addr]ctype.CidType)
		if len(cids) > 0 && len(cids) == len(tokens) {
			for i, cid := range cids {
				tokencids[tokens[i]] = cid
			}
		}
		b.neighbors[ospAddr] = &NeighborInfo{
			UpdateTime: now,
			TokenCids:  tokencids,
		}
	}
}

// unmarkOsp removes an Osp in routers
func (b *routingTableBuilder) unmarkOsp(ospAddr ctype.Addr) {
	b.graphLock.Lock()
	defer b.graphLock.Unlock()
	log.Infof("unmarkOsp: %x", ospAddr)
	delete(b.osps, ospAddr)
	delete(b.neighbors, ospAddr)
}

// getAllTokens returns all tokens ever seen by edge series.
func (b *routingTableBuilder) getAllTokens() map[ctype.Addr]bool {
	b.graphLock.RLock()
	defer b.graphLock.RUnlock()
	tks := make(map[ctype.Addr]bool)
	for k := range b.edges {
		tks[k] = true
	}
	return tks
}

// getAllOsps copies the all osp info and returns them back
func (b *routingTableBuilder) getAllOsps() map[ctype.Addr]*OspInfo {
	allOspInfo := make(map[ctype.Addr]*OspInfo)

	b.graphLock.RLock()
	defer b.graphLock.RUnlock()

	for addr := range b.osps {
		allOspInfo[addr] = b.osps[addr]
	}

	return allOspInfo
}

// hasOsp returns whether an Osp is marked or not
func (b *routingTableBuilder) hasOsp(ospAddr ctype.Addr) bool {
	b.graphLock.RLock()
	defer b.graphLock.RUnlock()

	_, ok := b.osps[ospAddr]

	return ok
}

func (b *routingTableBuilder) keepOspAlive(ospAddr ctype.Addr, timestamp time.Time) {
	b.graphLock.Lock()
	defer b.graphLock.Unlock()
	if _, ok := b.osps[ospAddr]; ok {
		if timestamp.After(b.osps[ospAddr].UpdateTime) {
			b.osps[ospAddr].UpdateTime = timestamp
		}
	}
}

func (b *routingTableBuilder) keepNeighborAlive(neighborAddr ctype.Addr) {
	b.graphLock.Lock()
	defer b.graphLock.Unlock()
	if _, ok := b.neighbors[neighborAddr]; ok {
		b.neighbors[neighborAddr].UpdateTime = now()
	}
}

func (b *routingTableBuilder) getNeighborAddrs() []ctype.Addr {
	b.graphLock.RLock()
	defer b.graphLock.RUnlock()
	var addrs []ctype.Addr
	for addr, neighbor := range b.neighbors {
		// neighbor needs to have opened channel
		if len(neighbor.TokenCids) > 0 {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

func (b *routingTableBuilder) getAliveNeighbors() map[ctype.Addr]*NeighborInfo {
	aliveNeighbors := make(map[ctype.Addr]*NeighborInfo)
	b.graphLock.RLock()
	defer b.graphLock.RUnlock()
	now := now()
	for addr, neighbor := range b.neighbors {
		// neighbor needs to be alive
		if neighbor.UpdateTime.Add(config.RouterAliveTimeout).After(now) {
			if len(neighbor.TokenCids) > 0 {
				aliveNeighbors[addr] = neighbor
			}
		}
	}
	return aliveNeighbors
}

// getAllOsps copies the all osp info and returns them back
func (b *routingTableBuilder) getAllNeighbors() map[ctype.Addr]*NeighborInfo {
	allNeighbors := make(map[ctype.Addr]*NeighborInfo)
	b.graphLock.RLock()
	defer b.graphLock.RUnlock()
	for addr, neighbor := range b.neighbors {
		// neighbor needs to have opened channel
		if len(neighbor.TokenCids) > 0 {
			allNeighbors[addr] = neighbor
		}
	}
	return allNeighbors
}

func (b *routingTableBuilder) buildTable(tokenAddr ctype.Addr) (map[ctype.Addr]ctype.CidType, error) {
	b.buildLock.Lock()
	defer b.buildLock.Unlock()
	// check if build is needed
	if !b.needCompute(tokenAddr) {
		return nil, nil
	}
	log.Infoln("building routing table for token", utils.PrintTokenAddr(tokenAddr))
	// compute routes
	accessOsps, nextHopCids, nextHopAddrs := b.computeRoutes(tokenAddr)
	// update routes in database
	b.updateRouteDB(tokenAddr, accessOsps, nextHopCids, nextHopAddrs)

	return nextHopCids, nil
}

func (b *routingTableBuilder) needCompute(tokenAddr ctype.Addr) bool {
	b.graphLock.RLock()
	defer b.graphLock.RUnlock()
	if len(b.edges[tokenAddr]) == 0 {
		log.Debugln("skip compute due to no edges for token", utils.PrintTokenAddr(tokenAddr))
		return false
	}
	connected := false
	for _, neighbor := range b.neighbors {
		for tk, _ := range neighbor.TokenCids {
			if tk == tokenAddr {
				connected = true
				break
			}
		}
	}
	if !connected {
		log.Debugln("skip compute due to no direct neighbors for token", utils.PrintTokenAddr(tokenAddr))
		return false
	}
	return true
}

func (b *routingTableBuilder) computeRoutes(tokenAddr ctype.Addr) (
	map[ctype.Addr]accessOspSet, map[ctype.Addr]ctype.CidType, map[ctype.Addr]ctype.Addr) {
	b.graphLock.RLock()
	defer b.graphLock.RUnlock()
	// set of active osps
	liveOspSet := make(map[ctype.Addr]bool)
	now := now()
	for ospAddr, osp := range b.osps {
		// osp needs to be alive
		if osp.UpdateTime.Add(config.RouterAliveTimeout).After(now) {
			liveOspSet[ospAddr] = true
		}
	}
	liveOspSet[b.myAddr] = true

	// client addr -> access osps
	accessOsps := make(map[ctype.Addr]accessOspSet)
	// peer addr -> channel ID
	peerToCid := make(map[ctype.Addr]ctype.CidType)

	// build osp graph
	graph := NewGraph()
	for _, edge := range b.edges[tokenAddr] {
		// record direct connected cid as value in routing table is next hop cid instead of addr.
		if b.myAddr == edge.P1 {
			peerToCid[edge.P2] = edge.Cid
		} else if b.myAddr == edge.P2 {
			peerToCid[edge.P1] = edge.Cid
		}

		if b.osps[edge.P1] == nil && b.osps[edge.P2] == nil {
			// skip if neither is registered osp.
			continue
		}

		p1Str := ctype.Addr2Hex(edge.P1)
		p2Str := ctype.Addr2Hex(edge.P2)
		// only add edge to graph when it is alive, and both osps are alive
		if liveOspSet[edge.P1] && liveOspSet[edge.P2] {
			ospEdge := b.ospEdges[edge.Cid]
			if ospEdge == nil {
				log.Debugln("osp edge not found", p1Str, p2Str)
				continue
			}
			if ospEdge.updateTime.Add(config.RouterAliveTimeout).After(now) {
				log.Debugln("adding edge", p1Str, p2Str)
				graph.addEdge(p1Str, p2Str, 1)
				graph.addEdge(p2Str, p1Str, 1)
			}
			continue
		}

		// One is osp, the other is client, save client for last hop route.
		// The edge and client is not in routing calculation though.
		if b.osps[edge.P1] != nil && b.osps[edge.P2] == nil {
			// p1 is osp, p2 is client
			if accessOsps[edge.P2] == nil {
				accessOsps[edge.P2] = make(map[ctype.Addr]bool)
			}
			accessOsps[edge.P2][edge.P1] = true
		} else if b.osps[edge.P2] != nil && b.osps[edge.P1] == nil {
			// p2 is osp, p1 is client
			if accessOsps[edge.P1] == nil {
				accessOsps[edge.P1] = make(map[ctype.Addr]bool)
			}
			accessOsps[edge.P1][edge.P2] = true
		}
	}

	// compute shortest paths
	_, paths := graph.dijkstra(ctype.Addr2Hex(b.myAddr))
	// dest osp -> next hop cid
	nextHopCids := make(map[ctype.Addr]ctype.CidType)
	// dest osp -> next hop osp
	nextHopAddrs := make(map[ctype.Addr]ctype.Addr)
	// Calculate routes from src to all ospAddrs
	for ospAddr := range liveOspSet {
		dest := ctype.Addr2Hex(ospAddr)
		if ospAddr == b.myAddr {
			// skip myself
			continue
		}
		path := paths[dest]
		if len(path) < 2 {
			log.Debugln("no path to destination", dest)
			continue
		}
		log.Debugln("shortest path:", printPath(path))
		nextHop := ctype.Hex2Addr(path[1])
		nextHopAddrs[ospAddr] = nextHop
		nextHopCids[ospAddr] = peerToCid[nextHop]
	}
	return accessOsps, nextHopCids, nextHopAddrs
}

func (b *routingTableBuilder) updateRouteDB(
	tokenAddr ctype.Addr, accessOsps map[ctype.Addr]accessOspSet,
	nextHopCids map[ctype.Addr]ctype.CidType, nextHopAddrs map[ctype.Addr]ctype.Addr) {
	b.routeLock.Lock()
	defer b.routeLock.Unlock()
	// only update DB if there is a change. Applied to both accessOsps table and routing table.
	updatedClients := make(map[ctype.Addr]bool)
	prevAccessOsps := b.accessOsps[tokenAddr]
	if prevAccessOsps == nil {
		prevAccessOsps = make(map[ctype.Addr]accessOspSet)
	}

	for client, ospAddrs := range accessOsps {
		for ospAddr := range ospAddrs {
			if prevAccessOsps[client] != nil && prevAccessOsps[client][ospAddr] {
				// osp is already in client's access osp set. No need to update db.
				continue
			}
			log.Debugln("client", ctype.Addr2Hex(client), "has new access OSP", ctype.Addr2Hex(ospAddr))
			updatedClients[client] = true
		}
	}
	for client, ospAddrs := range prevAccessOsps {
		for ospAddr := range ospAddrs {
			if accessOsps[client] != nil && accessOsps[client][ospAddr] {
				// osp is still in client's access osp set. No need to update db.
				continue
			}
			// osp is no longer serving client.
			log.Debugln("client", ctype.Addr2Hex(client), "lost access OSP", ctype.Addr2Hex(ospAddr))
			updatedClients[client] = true
		}
	}

	var err error
	tokenInfo := utils.GetTokenInfoFromAddress(tokenAddr)
	// update DB client access osp entries.
	insertNum := 0
	updateNum := 0
	deleteNum := 0
	for client := range updatedClients {
		num := len(accessOsps[client])
		if num > 0 {
			ospAddrs := make([]ctype.Addr, 0, num)
			for ospAddr := range accessOsps[client] {
				ospAddrs = append(ospAddrs, ospAddr)
			}
			if len(prevAccessOsps[client]) == 0 {
				err = b.dal.InsertDestToken(client, tokenInfo, ospAddrs, 0)
				insertNum++
			} else {
				err = b.dal.UpdateDestTokenOsps(client, tokenInfo, ospAddrs)
				updateNum++
			}
		} else {
			err = b.dal.DeleteDestToken(client, tokenInfo)
			deleteNum++
		}
		if err != nil {
			log.Errorln(err)
			// Could not update the DB, keep using the previous OSP set.
			accessOsps[client] = prevAccessOsps[client]
		}
	}
	if len(updatedClients) > 0 {
		log.Infof("client access OSPs: %d insert %d delete %d update", insertNum, deleteNum, updateNum)
	}
	b.accessOsps[tokenAddr] = accessOsps

	// update DB routing entries
	prevNextHopCids := b.nextHopCids[tokenAddr]
	if prevNextHopCids == nil {
		prevNextHopCids = make(map[ctype.Addr]ctype.CidType)
	}
	for dst, cid := range nextHopCids {
		if prevNextHopCids[dst] == cid {
			continue
		}
		action := "adding"
		if prevNextHopCids[dst] != ctype.ZeroCid {
			action = "updating"
		}
		log.Infof("%s route to %x on token %s, next hop osp %x", action, dst, utils.PrintTokenAddr(tokenAddr), nextHopAddrs[dst])
		err = b.dal.UpsertRouting(dst, tokenInfo, cid)
		if err != nil {
			log.Errorln(err)
			// Remove the route entry in memory to be sync with database so that build next time will update db again.
			delete(nextHopCids, dst)
		}
	}
	for dst, cid := range prevNextHopCids {
		if _, ok := nextHopCids[dst]; !ok {
			log.Infof("deleting route to %x on token %s", dst, utils.PrintTokenAddr(tokenAddr))
			err = b.dal.DeleteRouting(dst, tokenInfo)
			if err != nil {
				log.Errorln(err)
				// Add back route entry in memory to be sync with database so that build next time will delete db again.
				nextHopCids[dst] = cid
			}
		}
	}
	b.nextHopCids[tokenAddr] = nextHopCids
}
