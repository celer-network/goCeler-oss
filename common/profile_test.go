// Copyright 2018-2020 Celer Network

package common

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestProfileJSON(t *testing.T) {
	raw, err := ioutil.ReadFile("profile_test.json")
	if err != nil {
		t.Error(err)
	}
	pj := new(ProfileJSON)
	err = json.Unmarshal(raw, pj)
	if err != nil {
		t.Error(err)
	}
	// t.Logf("%+v", pj)
	chkEq(pj.Version, "0.1", t)
	chkEq(pj.Ethereum.ChainId, uint64(3), t)
	chkEq(pj.Ethereum.Contracts.Ledger, "abcdef..", t)
	chkEq(pj.Osp.Address, "c5b5..", t)

	cp := pj.ToCProfile()
	chkEq(cp.ChainId, int64(3), t)
	chkEq(cp.SvrRPC, "xxx.celer.app:10000", t)
	chkEq(cp.LedgerAddr, "abcdef..", t)

	cp2 := ParseProfile("profile_test.json")
	chkEq(cp.ChainId, cp2.ChainId, t)
	chkEq(cp.SvrRPC, cp2.SvrRPC, t)
	// empty also works
	pj2 := new(ProfileJSON)
	cp3 := pj2.ToCProfile()
	chkEq(cp3.LedgerAddr, "", t)
}

func chkEq(v, exp interface{}, t *testing.T) {
	if v != exp {
		t.Errorf("mismatch value exp: %v, got %v", exp, v)
	}
}
