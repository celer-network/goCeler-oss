// Copyright 2018-2019 Celer Network

package route

import (
	"math/big"
	"testing"

	rt "github.com/celer-network/goCeler-oss/chain/channel-eth-go/routerregistry"
	"github.com/celer-network/goCeler-oss/chain"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/monitor"
	"github.com/ethereum/go-ethereum/core/types"
)

type routerEvent = rt.RouterRegistryRouterUpdated

func TestMonitorRouterUpdatedEvent(t *testing.T) {
	p := NewMockProcessor()
	events := generateMockOnchainEvents()
	for _, event := range events {
		p.processRouterUpdatedEvent(event.e, event.blkNum)
	}

	if ok := p.keeper.IsOspExisting(addr1); ok {
		t.Errorf("This address should not exist: %s", addr1.Hex())
	}

	ospInfo := p.keeper.GetOspInfo()
	if blk, ok := ospInfo[addr2]; !ok || blk != 4 {
		t.Errorf("This address does not exist or block number is wrong: [%s, %v]", addr2.Hex(), blk)
	}

}

func TestCalculateStartBlockNumber(t *testing.T) {
	p := NewMockProcessor()
	currentBlk := p.monitorService.GetCurrentBlockNumber()

	startBlk := p.calculateStartBlockNumber() // currentBlk - 46500 -5

	if startBlk.Uint64() != (currentBlk.Uint64() - expireIntervalBlock - 5) {
		t.Errorf("Fail to calculate the correct start block number: %v", startBlk.Uint64())
	}
}

func TestRemoveExpiredRouters(t *testing.T) {
	p := NewMockProcessor()
	events := generateMockOnchainEvents()
	for _, event := range events {
		p.processRouterUpdatedEvent(event.e, event.blkNum)
	}

	p.removeExpiredRouters()

	ospInfo := p.keeper.GetOspInfo()
	if blk, ok := ospInfo[addr2]; ok {
		t.Errorf("This address should not exist: [%s, %v]", addr1.Hex(), blk)
	}

	if ok := p.keeper.IsOspExisting(addr3); !ok {
		t.Errorf("This address should exist: %s", addr3.Hex())
	}
}

var (
	addr1 = ctype.Hex2Addr("0000000000000000000000000000000000000001")
	addr2 = ctype.Hex2Addr("0000000000000000000000000000000000000002")
	addr3 = ctype.Hex2Addr("0000000000000000000000000000000000000003")
)

type MockOnchainEvent struct {
	e      *routerEvent
	blkNum uint64
}

func NewMockProcessor() *RouterProcessor {
	m := NewMockMonitorService()
	k := NewMockRoutingInfoKeeper()
	p := NewRouterProcessor(
		nil, // nodeConfig
		nil, // transactor
		m,   // monitorService
		k,   // keeper
		5,   // blockDelay
		"",  // adminWebHostAndPort
	)
	p.isRegisteredOnchain = true
	return p
}

func generateMockOnchainEvents() []*MockOnchainEvent {
	e1 := &MockOnchainEvent{
		e: &routerEvent{
			Op:            0, // Add
			RouterAddress: addr1,
			Raw:           types.Log{},
		},
		blkNum: 1,
	}

	e2 := &MockOnchainEvent{
		e: &routerEvent{
			Op:            0, //  Add
			RouterAddress: addr2,
			Raw:           types.Log{},
		},
		blkNum: 2,
	}

	e3 := &MockOnchainEvent{
		e: &routerEvent{
			Op:            1, // Remove
			RouterAddress: addr1,
			Raw:           types.Log{},
		},
		blkNum: 3,
	}

	e4 := &MockOnchainEvent{
		e: &routerEvent{
			Op:            2, // Refresh
			RouterAddress: addr2,
			Raw:           types.Log{},
		},
		blkNum: 4,
	}

	e5 := &MockOnchainEvent{
		e: &routerEvent{
			Op:            0, // Add
			RouterAddress: addr3,
			Raw:           types.Log{},
		},
		blkNum: 50000,
	}

	return []*MockOnchainEvent{e1, e2, e3, e4, e5}
}

type MockMonitorService struct{}

func NewMockMonitorService() *MockMonitorService {
	return &MockMonitorService{}
}

func (c *MockMonitorService) GetCurrentBlockNumber() *big.Int {
	return big.NewInt(70000)
}

func (c *MockMonitorService) RegisterDeadline(deadline monitor.Deadline) monitor.CallbackID { return 0 }

func (c *MockMonitorService) Monitor(
	eventName string,
	contract chain.Contract,
	startBlock *big.Int,
	endBlock *big.Int,
	quickCatch bool,
	reset bool,
	callback func(monitor.CallbackID, types.Log)) (monitor.CallbackID, error) {
	return 0, nil
}

func (c *MockMonitorService) MonitorEvent(monitor.Event, bool) (monitor.CallbackID, error) {
	return 0, nil
}

func (c *MockMonitorService) RemoveDeadline(id monitor.CallbackID) {}

func (c *MockMonitorService) RemoveEvent(id monitor.CallbackID) {}

type MockRoutingInfoKeeper struct {
	ospInfo map[ctype.Addr]uint64
}

func NewMockRoutingInfoKeeper() *MockRoutingInfoKeeper {
	k := &MockRoutingInfoKeeper{
		ospInfo: make(map[ctype.Addr]uint64),
	}

	return k
}

func (k *MockRoutingInfoKeeper) GetOspInfo() map[ctype.Addr]uint64 {
	return k.ospInfo
}

func (k *MockRoutingInfoKeeper) MarkOsp(osp ctype.Addr, blknum uint64) {
	k.ospInfo[osp] = blknum
}

func (k *MockRoutingInfoKeeper) UnmarkOsp(osp ctype.Addr) {
	delete(k.ospInfo, osp)
}

func (k *MockRoutingInfoKeeper) IsOspExisting(osp ctype.Addr) bool {
	_, ok := k.ospInfo[osp]

	return ok
}
