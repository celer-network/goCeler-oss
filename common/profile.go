// Copyright 2018-2019 Celer Network

package common

import (
	"encoding/json"
	"io/ioutil"
)

// Defines what new profile json looks like. Note if we need to
// output profile json keys begin w/ lowercase, we'll have to split each fields
// into its own line so tag like `json:"version"` can work. otherwise
// tag is applied to all fields defined in same line and json.Marshal fails

// ProfileJSON handles new profile json schema
type ProfileJSON struct {
	// schema version, ignored for now but will be useful
	// when need to handle incompatible schema in the future
	Version  string
	Ethereum ProfileEthereum
	Osp      ProfileOsp
}

type ProfileEthereum struct {
	Gateway                                  string
	ChainId, BlockIntervalSec, BlockDelayNum uint64
	Contracts                                ProfileContracts
}

type ProfileContracts struct {
	Wallet, Ledger, VirtResolver, EthPool, PayResolver, PayRegistry, RouterRegistry string
}

type ProfileOsp struct {
	Host, Address string
}

func (pj *ProfileJSON) ToCProfile() *CProfile {
	cp := &CProfile{
		ChainId:            int64(pj.Ethereum.ChainId),
		ETHInstance:        pj.Ethereum.Gateway,
		BlockDelayNum:      pj.Ethereum.BlockDelayNum,
		PollingInterval:    pj.Ethereum.BlockIntervalSec,
		WalletAddr:         pj.Ethereum.Contracts.Wallet,
		LedgerAddr:         pj.Ethereum.Contracts.Ledger,
		VirtResolverAddr:   pj.Ethereum.Contracts.VirtResolver,
		EthPoolAddr:        pj.Ethereum.Contracts.EthPool,
		PayResolverAddr:    pj.Ethereum.Contracts.PayResolver,
		PayRegistryAddr:    pj.Ethereum.Contracts.PayRegistry,
		RouterRegistryAddr: pj.Ethereum.Contracts.RouterRegistry,
		SvrETHAddr:         pj.Osp.Address,
		SvrRPC:             pj.Osp.Host,
	}
	return cp
}

// ParseProfile parses file content at path and returns CProfile
// supports both old and new schema
func ParseProfile(path string) *CProfile {
	raw, _ := ioutil.ReadFile(path)
	return Bytes2Profile(raw)
}

// Bytes2Profile does json.Unmarshal and return CProfile
func Bytes2Profile(data []byte) *CProfile {
	pj := new(ProfileJSON)
	json.Unmarshal(data, pj)
	// fallback to support old schema
	if pj == nil || pj.Version == "" {
		cp := new(CProfile)
		json.Unmarshal(data, cp)
		return cp
	}
	return pj.ToCProfile()
}
