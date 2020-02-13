// Copyright 2018-2019 Celer Network

package route

import (
	"testing"

	"github.com/celer-network/goCeler-oss/clog"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
)

func TestBasicRouting(t *testing.T) {
	// Test set up.
	myEth := ctype.Hex2Addr("ba756d65a1a03f07d205749f35e2406e4a8522a1")
	osp1 := ctype.Hex2Addr("6a6d2a97da1c453a4e099e8054865a0a59728862")
	osp2 := ctype.Hex2Addr("6a6d2a97da1c453a4e099e8054865a0a59728863")
	c1OnOsp1 := ctype.Hex2Addr("ba756d65a1a03f07d205749f35e2406e4a8522a3")
	c2OnMe := ctype.Hex2Addr("6a6d2a97da1c453a4e099e8054865a0a59728864")
	c3OnOsp2 := ctype.Hex2Addr("6a6d2a97da1c453a4e099e8054865a0a59728865")
	cid1 := ctype.Hex2Cid("23a548990ef70278cdb6519b3646a04646408e9aec09b19c8f16e8ae9ad30871")
	cid2 := ctype.Hex2Cid("23a548990ef70278cdb6519b3646a04646408e9aec09b19c8f16e8ae9ad30872")
	cid3 := ctype.Hex2Cid("23a548990ef70278cdb6519b3646a04646408e9aec09b19c8f16e8ae9ad30873")
	cid4 := ctype.Hex2Cid("23a548990ef70278cdb6519b3646a04646408e9aec09b19c8f16e8ae9ad30874")
	cid5 := ctype.Hex2Cid("23a548990ef70278cdb6519b3646a04646408e9aec09b19c8f16e8ae9ad30875")
	ethContractAddr := ctype.Hex2Addr(common.EthContractAddr)
	edges := []*Edge{
		&Edge{
			P1:        myEth,
			P2:        osp1,
			Cid:       cid1,
			TokenAddr: ethContractAddr,
		},
		&Edge{
			P1:        c1OnOsp1,
			P2:        osp1,
			Cid:       cid2,
			TokenAddr: ethContractAddr,
		},
		&Edge{
			P1:        myEth,
			P2:        c2OnMe,
			Cid:       cid3,
			TokenAddr: ethContractAddr,
		},
		&Edge{
			P1:        osp2,
			P2:        c3OnOsp2,
			Cid:       cid4,
			TokenAddr: ethContractAddr,
		},
		&Edge{
			P1:        osp2,
			P2:        osp1,
			Cid:       cid5,
			TokenAddr: ethContractAddr,
		},
	}
	osps := []ctype.Addr{osp1, myEth, osp2}

	// Test execution
	dal := NewTestRoutingBuilderDAL()
	b := NewRoutingTableBuilder(myEth /*src*/, dal)
	for _, e := range edges {
		b.AddEdge(e.P1, e.P2, e.Cid, e.TokenAddr)
	}
	for _, osp := range osps {
		b.MarkOsp(osp, 1 /* block num */)
	}

	_, err := b.Build(ethContractAddr /*tokenAddr*/)
	rt := dal.rt[ethContractAddr]
	if len(b.GetAllTokens()) != 1 || !b.GetAllTokens()[ethContractAddr] {
		t.Fatal("token is wrong")
	}
	if err != nil {
		t.Fatal(err)
	}
	clog.Infoln(beautifyRT(rt))
	if dal.servingOsps[c1OnOsp1] != osp1 {
		t.Fatal("osp1 should serve c1")
	}
	if dal.servingOsps[c2OnMe] != myEth {
		t.Fatal("I should serve c2")
	}
	if dal.servingOsps[c3OnOsp2] != osp2 {
		t.Fatal("osp2 should serve c3")
	}
	if len(rt) != 2 {
		t.Fatal("rt should have 2 entry, has", len(rt))
	}
	if rt[osp1] != cid1 {
		t.Fatal("should has c1 as route to osp1")
	}
	if rt[osp2] != cid1 {
		t.Fatal("should has c1 as route to osp2")
	}

	// Remove c2OnMe and recalculate
	clog.Infoln("Removing edge c2OnMe<->me, every osp should be reachable, but I no longer serve c2")
	b.RemoveEdge(edges[2].Cid)
	_, err = b.Build(ethContractAddr)
	rt = dal.rt[ethContractAddr]
	// clog.Infoln(beautifyRT(rt))
	if len(dal.servingOsps) != 2 {
		t.Fatal("servingOsps should have 2 entries, has", len(dal.servingOsps))
	}
	if dal.servingOsps[c1OnOsp1] != osp1 {
		t.Fatal("osp1 should serve c1")
	}
	if dal.servingOsps[c3OnOsp2] != osp2 {
		t.Fatal("osp2 should serve c3")
	}
	if len(rt) != 2 {
		t.Fatal("rt should have 2 entries, has", len(rt))
	}
	if rt[osp1] != cid1 {
		t.Fatal("should has c1 as route to osp1")
	}
	if rt[osp2] != cid1 {
		t.Fatal("should has c1 as route to osp2")
	}

	clog.Infoln("Removing edge osp1<->osp2, osp2 should not be reachable")
	b.RemoveEdge(edges[4].Cid)
	_, err = b.Build(ethContractAddr)
	rt = dal.rt[ethContractAddr]
	if len(dal.servingOsps) != 2 {
		t.Fatal("servingOsps should have 2 entries, has", len(dal.servingOsps))
	}
	if dal.servingOsps[c1OnOsp1] != osp1 {
		t.Fatal("osp1 should serve c1")
	}
	if dal.servingOsps[c3OnOsp2] != osp2 {
		t.Fatal("osp2 should serve c3")
	}
	if len(rt) != 1 {
		t.Fatal("rt should have 1 entry, has", len(rt))
	}
	if rt[osp1] != cid1 {
		t.Fatal("should has c1 as route to osp1")
	}
}

func beautifyRT(rt map[ctype.Addr]ctype.CidType) map[string]string {
	// Beautify routing table for debugging.
	beautifulRT := make(map[string]string)
	for k, v := range rt {
		beautifulRT[ctype.Addr2Hex(k)] = ctype.Cid2Hex(v)
	}
	return beautifulRT
}

type TestRoutingBuilderDAL struct {
	edges  map[ctype.Addr]map[ctype.CidType]*Edge
	ospSet map[ctype.Addr]bool
	// tokenAddr, localPeer-> cid
	// Note keys in this implementation are in reverse order of API to avoid small map object if there are many clients in test.
	localPeerCid map[ctype.Addr]map[ctype.Addr]ctype.CidType
	// clientAddr -> ospAddr
	servingOsps map[ctype.Addr]ctype.Addr
	// tokenAddr, dst -> cid
	rt map[ctype.Addr]map[ctype.Addr]ctype.CidType
}

func NewTestRoutingBuilderDAL() *TestRoutingBuilderDAL {
	return &TestRoutingBuilderDAL{
		rt:           make(map[ctype.Addr]map[ctype.Addr]ctype.CidType),
		edges:        make(map[ctype.Addr]map[ctype.CidType]*Edge),
		ospSet:       make(map[ctype.Addr]bool),
		localPeerCid: make(map[ctype.Addr]map[ctype.Addr]ctype.CidType),
		servingOsps:  make(map[ctype.Addr]ctype.Addr),
	}
}
func (dal *TestRoutingBuilderDAL) PutLocalPeerCid(localPeer ctype.Addr, tokenAddr ctype.Addr, cid ctype.CidType) error {
	if dal.localPeerCid[tokenAddr] == nil {
		dal.localPeerCid[tokenAddr] = make(map[ctype.Addr]ctype.CidType)
	}
	dal.localPeerCid[tokenAddr][localPeer] = cid
	return nil
}
func (dal *TestRoutingBuilderDAL) DeleteLocalPeerCid(localPeer ctype.Addr, tokenAddr ctype.Addr) error {
	if dal.localPeerCid[tokenAddr] == nil {
		return nil
	}
	delete(dal.localPeerCid[tokenAddr], localPeer)
	if len(dal.localPeerCid[tokenAddr]) == 0 {
		delete(dal.localPeerCid, tokenAddr)
	}
	return nil
}
func (dal *TestRoutingBuilderDAL) PutServingOsp(clientAddr ctype.Addr, tokenAddr ctype.Addr, ospAddr ctype.Addr) error {
	dal.servingOsps[clientAddr] = ospAddr
	return nil
}
func (dal *TestRoutingBuilderDAL) DeleteServingOsp(clientAddr ctype.Addr, tokenAddr ctype.Addr, ospAddr ctype.Addr) error {
	delete(dal.servingOsps, clientAddr)
	return nil
}

// GetAllServingOsps only supports ETH.
func (dal *TestRoutingBuilderDAL) GetAllServingOsps() (map[ctype.Addr]map[ctype.Addr]servingOspMap, error) {
	allServingOsps := make(map[ctype.Addr]map[ctype.Addr]bool)
	for client, osp := range dal.servingOsps {
		allServingOsps[client] = make(map[ctype.Addr]bool)
		allServingOsps[client][osp] = true
	}
	ret := make(map[ctype.Addr]map[ctype.Addr]servingOspMap)
	ret[ctype.Hex2Addr(common.EthContractAddr)] = allServingOsps
	return ret, nil
}

func (dal *TestRoutingBuilderDAL) PutRoute(dstStr string, tokenAddrStr string, cid ctype.CidType) error {
	dst := ctype.Hex2Addr(dstStr)
	tokenAddr := ctype.Hex2Addr(tokenAddrStr)
	if dal.rt[tokenAddr] == nil {
		dal.rt[tokenAddr] = make(map[ctype.Addr]ctype.CidType)
	}
	dal.rt[tokenAddr][dst] = cid
	return nil
}
func (dal *TestRoutingBuilderDAL) DeleteRoute(dstStr string, tokenAddrStr string) error {
	dst := ctype.Hex2Addr(dstStr)
	tokenAddr := ctype.Hex2Addr(tokenAddrStr)
	if dal.rt[tokenAddr] == nil {
		return nil
	}
	delete(dal.rt[tokenAddr], dst)
	if len(dal.rt[tokenAddr]) == 0 {
		delete(dal.rt, tokenAddr)
	}
	return nil
}
func (dal *TestRoutingBuilderDAL) GetAllRoutes() (map[ctype.Addr]map[ctype.Addr]ctype.CidType, error) {
	return dal.rt, nil
}
func (dal *TestRoutingBuilderDAL) PutEdge(token ctype.Addr, cid ctype.CidType, edge *Edge) error {
	if dal.edges[token] == nil {
		dal.edges[token] = make(map[ctype.CidType]*Edge)
	}
	dal.edges[token][cid] = edge
	return nil
}
func (dal *TestRoutingBuilderDAL) GetEdges(token ctype.Addr) (map[ctype.CidType]*Edge, error) {
	return dal.edges[token], nil
}
func (dal *TestRoutingBuilderDAL) DeleteEdge(token ctype.Addr, cid ctype.CidType) error {
	delete(dal.edges[token], cid)
	return nil
}
func (dal *TestRoutingBuilderDAL) GetAllEdgeTokens() (map[ctype.Addr]bool, error) {
	tks := make(map[ctype.Addr]bool)
	for k := range dal.edges {
		tks[k] = true
	}
	return tks, nil
}
func (dal *TestRoutingBuilderDAL) MarkOsp(osp ctype.Addr) {
	dal.ospSet[osp] = true
}
func (dal *TestRoutingBuilderDAL) GetAllMarkedOsp() (map[ctype.Addr]bool, error) {
	return dal.ospSet, nil
}
func (dal *TestRoutingBuilderDAL) IsOspMarked(osp ctype.Addr) bool {
	return dal.ospSet[osp]
}
func (dal *TestRoutingBuilderDAL) UnmarkOsp(osp ctype.Addr) {
	delete(dal.ospSet, osp)
}
