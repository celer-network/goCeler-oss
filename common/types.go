// Copyright 2018-2020 Celer Network

package common

import (
	"math/big"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/rpc"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CProfile contains configurations for CelerClient/OSP
type CProfile struct {
	ETHInstance        string            `json:"ethInstance"`
	SvrETHAddr         string            `json:"svrEthAddr"`
	WalletAddr         string            `json:"walletAddr"`
	LedgerAddr         string            `json:"ledgerAddr"`
	VirtResolverAddr   string            `json:"virtResolverAddr"`
	EthPoolAddr        string            `json:"ethPoolAddr"`
	PayResolverAddr    string            `json:"payResolverAddr"`
	PayRegistryAddr    string            `json:"payRegistryAddr"`
	RouterRegistryAddr string            `json:"routerRegistryAddr"`
	SvrRPC             string            `json:"svrRpc"`
	SvrName            string            `json:"svrName,omitempty"`
	SelfRPC            string            `json:"selfRpc,omitempty"`
	StoreDir           string            `json:"storeDir,omitempty"`
	StoreSql           string            `json:"storeSql,omitempty"`
	WsOrigin           string            `json:"wsOrigin,omitempty"`
	ChainId            int64             `json:"chainId"`
	BlockDelayNum      uint64            `json:"blockDelayNum"`
	IsOSP              bool              `json:"isOsp,omitempty"`
	ListenOnChain      bool              `json:"listenOnChain,omitempty"`
	PollingInterval    uint64            `json:"pollingInterval"`
	DisputeTimeout     uint64            `json:"disputeTimeout"`
	Ledgers            map[string]string `json:"ledgers"`
	ExplorerUrl        string            `json:"explorerUrl,omitempty"`
	CheckInterval      map[string]uint64 `json:"checkInterval,omitempty"`
}

type GlobalNodeConfig interface {
	GetOnChainAddr() ctype.Addr
	GetEthPoolAddr() ctype.Addr
	GetEthConn() *ethclient.Client
	GetRPCAddr() string
	GetSvrName() string
	GetWalletContract() chain.Contract
	// GetLedgerContract returns latest ledger contract.
	GetLedgerContract() chain.Contract
	// GetLedgerContractOn returns ledger contract on addr. The addr must exist in profile ledger address map.
	// It will return nil otherwise.
	GetLedgerContractOn(ctype.Addr) chain.Contract
	// GetAllLedgerContracts returns a map with key being ledger addresses in profile and ledger contract bound to to the address.
	GetAllLedgerContracts() map[ctype.Addr]chain.Contract
	// GetLedgerContractOf returns ledger contract object of which address is used by the cid.
	GetLedgerContractOf(ctype.CidType) chain.Contract
	GetVirtResolverContract() chain.Contract
	GetPayResolverContract() chain.Contract
	GetPayRegistryContract() chain.Contract
	GetRouterRegistryContract() chain.Contract
	GetCheckInterval(string) uint64
}

type StreamWriter interface {
	WriteCelerMsg(peer ctype.Addr, celerMsg *rpc.CelerMsg) error
}

type StateCallback interface {
	OnDispute(seqNum int)
}

type ChannelBalance struct {
	MyAddr     ctype.Addr
	MyFree     *big.Int
	MyLocked   *big.Int
	PeerAddr   ctype.Addr
	PeerFree   *big.Int
	PeerLocked *big.Int
}

type ChannelSeqNums struct {
	Base       uint64
	LastUsed   uint64
	LastAcked  uint64
	LastNacked uint64
}

type MsgFrame struct {
	Message  *rpc.CelerMsg
	PeerAddr ctype.Addr
	LogEntry *pem.PayEventMessage
}
