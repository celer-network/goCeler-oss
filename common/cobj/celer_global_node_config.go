// Copyright 2018-2020 Celer Network

package cobj

import (
	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/ethclient"
)

type chanLedgerDAL interface {
	GetChanLedger(cid ctype.CidType) (ctype.Addr, bool, error)
}

type CelerGlobalNodeConfig struct {
	onchainAddr            ctype.Addr
	rpcAddr                string
	svrName                string
	ethPoolAddr            ctype.Addr
	ethConn                *ethclient.Client
	walletContract         chain.Contract
	ledgerContract         chain.Contract
	virtResolverContract   chain.Contract
	payResolverContract    chain.Contract
	payRegistryContract    chain.Contract
	routerRegistryContract chain.Contract
	ledgers                map[ctype.Addr]chain.Contract
	chanDAL                chanLedgerDAL
	checkInterval          map[string]uint64 // copy from CProfile
}

func NewCelerGlobalNodeConfig(
	onchainAddr ctype.Addr,
	ethconn *ethclient.Client,
	profile *common.CProfile,
	walletABI string,
	ledgerABI string,
	virtResolverABI string,
	payResolverABI string,
	payRegistryABI string,
	routerRegistryABI string,
	chanDAL chanLedgerDAL) *CelerGlobalNodeConfig {
	walletContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.WalletAddr), walletABI)
	ledgers := make(map[ctype.Addr]chain.Contract)
	for ledgerAddr := range profile.Ledgers {
		ledgers[ctype.Hex2Addr(ledgerAddr)], _ = chain.NewBoundContract(
			ethconn, ctype.Hex2Addr(ledgerAddr), ledgerABI)
	}
	latestLedgerAddr := ctype.Hex2Addr(profile.LedgerAddr)
	if _, hasLatestLedger := ledgers[latestLedgerAddr]; !hasLatestLedger {
		ledgers[latestLedgerAddr], _ = chain.NewBoundContract(
			ethconn, latestLedgerAddr, ledgerABI)
	}
	ledgerContract := ledgers[latestLedgerAddr]
	virtResolverContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.VirtResolverAddr), virtResolverABI)
	payResolverContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.PayResolverAddr), payResolverABI)
	payRegistryContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.PayRegistryAddr), payRegistryABI)
	routerRegistryContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.RouterRegistryAddr), routerRegistryABI)

	gnc := &CelerGlobalNodeConfig{
		onchainAddr:            onchainAddr,
		rpcAddr:                profile.SelfRPC,
		svrName:                profile.SvrName,
		ethPoolAddr:            ctype.Hex2Addr(profile.EthPoolAddr),
		ethConn:                ethconn,
		walletContract:         walletContract,
		ledgerContract:         ledgerContract,
		virtResolverContract:   virtResolverContract,
		payResolverContract:    payResolverContract,
		payRegistryContract:    payRegistryContract,
		routerRegistryContract: routerRegistryContract,
		ledgers:                ledgers,
		chanDAL:                chanDAL,
		checkInterval:          profile.CheckInterval,
	}
	return gnc
}
func (config *CelerGlobalNodeConfig) GetOnChainAddr() ctype.Addr {
	return config.onchainAddr
}
func (config *CelerGlobalNodeConfig) GetEthPoolAddr() ctype.Addr {
	return config.ethPoolAddr
}
func (config *CelerGlobalNodeConfig) GetEthConn() *ethclient.Client {
	return config.ethConn
}
func (config *CelerGlobalNodeConfig) GetRPCAddr() string {
	return config.rpcAddr
}
func (config *CelerGlobalNodeConfig) GetSvrName() string {
	return config.svrName
}
func (config *CelerGlobalNodeConfig) GetWalletContract() chain.Contract {
	return config.walletContract
}
func (config *CelerGlobalNodeConfig) GetLedgerContract() chain.Contract {
	return config.ledgerContract
}
func (config *CelerGlobalNodeConfig) GetVirtResolverContract() chain.Contract {
	return config.virtResolverContract
}
func (config *CelerGlobalNodeConfig) GetPayResolverContract() chain.Contract {
	return config.payResolverContract
}
func (config *CelerGlobalNodeConfig) GetPayRegistryContract() chain.Contract {
	return config.payRegistryContract
}

// GetRouterRegistryContract gets the router registry contract info including ethclient, address and ABI
func (config *CelerGlobalNodeConfig) GetRouterRegistryContract() chain.Contract {
	return config.routerRegistryContract
}

// GetLedgerContractOn returns ledger contract on addr. The addr must exist in profile ledger address map.
// It will return nil otherwise.
func (config *CelerGlobalNodeConfig) GetLedgerContractOn(addr ctype.Addr) chain.Contract {
	return config.ledgers[addr]
}

// GetAllLedgerContracts returns a map with key being ledger addresses in profile and ledger contract bound to to the address.
func (config *CelerGlobalNodeConfig) GetAllLedgerContracts() map[ctype.Addr]chain.Contract {
	return config.ledgers
}

// GetLedgerContractOf returns ledger contract object of which address is used by the cid.
func (config *CelerGlobalNodeConfig) GetLedgerContractOf(cid ctype.CidType) chain.Contract {
	ledgerAddr, found, err := config.chanDAL.GetChanLedger(cid)
	if err != nil {
		log.Errorf("%v: fetching ledger addr for %x", err, cid)
		return nil
	}
	if !found {
		log.Errorf("ledger not found for cid %x", cid)
		return nil
	}
	if _, has := config.ledgers[ledgerAddr]; !has {
		log.Errorf("ledger contract not found for cid %x, on addr %x", cid, ledgerAddr)
		return nil
	}
	return config.ledgers[ledgerAddr]
}

// GetCheckInterval returns interval if set in profile or 0
// monitor will treat 0 as check log every blockIntervalSec
func (cfg *CelerGlobalNodeConfig) GetCheckInterval(eventName string) uint64 {
	return cfg.checkInterval[eventName]
}
