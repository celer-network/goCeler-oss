package cobj

import (
	"testing"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/virtresolver"
	"github.com/celer-network/goCeler/chain/channel-eth-go/wallet"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/route/routerregistry"
)

// any address is fine
const onchainAddrStr string = "ffffffffffffffffffffffffffffffffffffffff"

func TestGnrLedgers(t *testing.T) {
	onchainAddr := ctype.Hex2Addr(onchainAddrStr)
	oldLedgerAddrStr := "1111111111111111111111111111111111111111"
	oldLedgerAddr := ctype.Hex2Addr(oldLedgerAddrStr)
	newLedgerAddrStr := "2222222222222222222222222222222222222222"
	newLedgerAddr := ctype.Hex2Addr(newLedgerAddrStr)
	ledgers := make(map[string]string)
	ledgers[oldLedgerAddrStr] = "value doesn't matter for now"
	ledgers[newLedgerAddrStr] = "value doesn't matter for now"
	zeroAddrStr := ctype.Addr2Hex(ctype.ZeroAddr)
	profile := &common.CProfile{
		WalletAddr: zeroAddrStr,
		LedgerAddr: newLedgerAddrStr,
		Ledgers:    ledgers,
	}
	chanDAL := &chanTestDAL{
		ledgerAddrToReturn: ctype.Hex2Addr(oldLedgerAddrStr),
	}
	gnr := NewCelerGlobalNodeConfig(
		onchainAddr,
		nil, /*ethconn*/
		profile,
		wallet.CelerWalletABI,
		ledger.CelerLedgerABI,
		virtresolver.VirtContractResolverABI,
		payresolver.PayResolverABI,
		payregistry.PayRegistryABI,
		routerregistry.RouterRegistryABI,
		chanDAL,
	)

	// Test Assertions
	if gnr.GetLedgerContractOn(oldLedgerAddr).GetAddr() != oldLedgerAddr {
		t.Fatalf("GetLedgerContractOn: expect %x, actual %x", oldLedgerAddr, gnr.GetLedgerContractOn(oldLedgerAddr).GetAddr())
	}
	if gnr.GetLedgerContractOn(newLedgerAddr).GetAddr() != newLedgerAddr {
		t.Fatalf("GetLedgerContractOn: expect %x, actual %x", newLedgerAddr, gnr.GetLedgerContractOn(newLedgerAddr).GetAddr())
	}
	// expect to be same as chanTestDAL.ledgerAddrToReturn
	if gnr.GetLedgerContractOf(ctype.ZeroCid).GetAddr() != oldLedgerAddr {
		t.Fatalf("GetLedgerContractOf: expect %x, actual %x", oldLedgerAddr, gnr.GetLedgerContractOf(ctype.ZeroCid).GetAddr())
	}
	ledgersInGnr := gnr.GetAllLedgerContracts()
	if len(ledgersInGnr) != 2 {
		t.Fatalf("GetAllLedgerContracts: wrong size, expect 2, actual %d", len(ledgersInGnr))
	}
	oldLedgerInGnr, found := ledgersInGnr[oldLedgerAddr]
	if !found || oldLedgerInGnr.GetAddr() != oldLedgerAddr {
		t.Fatalf("GetAllLedgerContracts: old contract found %t expect %x, actual %x", found, oldLedgerAddr, oldLedgerInGnr.GetAddr())
	}
	newLedgerInGnr, found := ledgersInGnr[newLedgerAddr]
	if !found || newLedgerInGnr.GetAddr() != newLedgerAddr {
		t.Fatalf("GetAllLedgerContracts: new contract found %t expect %x, actual %x", found, newLedgerAddr, newLedgerInGnr.GetAddr())
	}
}

type chanTestDAL struct {
	ledgerAddrToReturn ctype.Addr
}

func (d *chanTestDAL) GetChanLedger(cid ctype.CidType) (ctype.Addr, bool, error) {
	if d.ledgerAddrToReturn == ctype.ZeroAddr {
		return ctype.ZeroAddr, false, nil
	}
	return d.ledgerAddrToReturn, true, nil
}
