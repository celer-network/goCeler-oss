// Copyright 2018-2019 Celer Network

package cobj

import (
	"github.com/celer-network/goCeler-oss/chain"
	"github.com/celer-network/goCeler-oss/common"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/ethereum/go-ethereum/ethclient"
)

type CelerGlobalNodeConfig struct {
	onchainAddr            string
	rpcAddr                string
	ethPoolAddr            ctype.Addr
	ethConn                *ethclient.Client
	walletContract         chain.Contract
	ledgerContract         chain.Contract
	virtResolverContract   chain.Contract
	payResolverContract    chain.Contract
	payRegistryContract    chain.Contract
	routerRegistryContract chain.Contract
}

func NewCelerGlobalNodeConfig(
	onchainAddr string,
	ethconn *ethclient.Client,
	profile *common.CProfile,
	walletABI string,
	ledgerABI string,
	virtResolverABI string,
	payResolverABI string,
	payRegistryABI string,
	routerRegistryABI string) *CelerGlobalNodeConfig {
	walletContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.WalletAddr), walletABI)
	ledgerContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.LedgerAddr), ledgerABI)
	virtResolverContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.VirtResolverAddr), virtResolverABI)
	payResolverContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.PayResolverAddr), payResolverABI)
	payRegistryContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.PayRegistryAddr), payRegistryABI)
	routerRegistryContract, _ := chain.NewBoundContract(
		ethconn, ctype.Hex2Addr(profile.RouterRegistryAddr), routerRegistryABI)

	gnr := &CelerGlobalNodeConfig{
		onchainAddr:            onchainAddr,
		rpcAddr:                profile.SelfRPC,
		ethPoolAddr:            ctype.Hex2Addr(profile.EthPoolAddr),
		ethConn:                ethconn,
		walletContract:         walletContract,
		ledgerContract:         ledgerContract,
		virtResolverContract:   virtResolverContract,
		payResolverContract:    payResolverContract,
		payRegistryContract:    payRegistryContract,
		routerRegistryContract: routerRegistryContract,
	}
	return gnr
}

func (config *CelerGlobalNodeConfig) GetOnChainAddr() string {
	return config.onchainAddr
}
func (config *CelerGlobalNodeConfig) GetOnChainAddrBytes() []byte {
	return ctype.Hex2Bytes(config.onchainAddr)
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
