// Copyright 2018-2019 Celer Network

package common

import (
	"math/big"

	"github.com/celer-network/goCeler-oss/chain"
	"github.com/celer-network/goCeler-oss/ctype"
	"github.com/celer-network/goCeler-oss/pem"
	"github.com/celer-network/goCeler-oss/rpc"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CProfile contains configurations for CelerClient/OSP
type CProfile struct {
	ETHInstance        string `json:"ethInstance"`
	SvrETHAddr         string `json:"svrEthAddr"`
	WalletAddr         string `json:"walletAddr"`
	LedgerAddr         string `json:"ledgerAddr"`
	VirtResolverAddr   string `json:"virtResolverAddr"`
	EthPoolAddr        string `json:"ethPoolAddr"`
	PayResolverAddr    string `json:"payResolverAddr"`
	PayRegistryAddr    string `json:"payRegistryAddr"`
	RouterRegistryAddr string `json:"routerRegistryAddr"`
	SvrRPC             string `json:"svrRpc"`
	SelfRPC            string `json:"selfRpc,omitempty"`
	StoreDir           string `json:"storeDir,omitempty"`
	StoreSql           string `json:"storeSql,omitempty"`
	WebPort            string `json:"webPort,omitempty"`
	WsOrigin           string `json:"wsOrigin,omitempty"`
	ChainId            int64  `json:"chainId"`
	BlockDelayNum      uint64 `json:"blockDelayNum"`
	IsOSP              bool   `json:"isOsp,omitempty"`
	ListenOnChain      bool   `json:"listenOnChain,omitempty"`
	PollingInterval    uint64 `json:"pollingInterval"`
	DisputeTimeout     uint64 `json:"disputeTimeout"`
}

type GlobalNodeConfig interface {
	GetOnChainAddr() string
	GetOnChainAddrBytes() []byte
	GetEthPoolAddr() ctype.Addr
	GetEthConn() *ethclient.Client
	GetRPCAddr() string
	GetWalletContract() chain.Contract
	GetLedgerContract() chain.Contract
	GetVirtResolverContract() chain.Contract
	GetPayResolverContract() chain.Contract
	GetPayRegistryContract() chain.Contract
	GetRouterRegistryContract() chain.Contract
}

type StreamWriter interface {
	WriteCelerMsg(peer string, celerMsg *rpc.CelerMsg) error
}

type Signer interface {
	Sign(data []byte) ([]byte, error)
}
type SigValidator interface {
	SigIsValid(signer string, data []byte, sig []byte) bool
}
type Crypto interface {
	Signer
	SigValidator
}

type StateChannelRouter interface {
	// Return Channel ID and peer
	LookupNextChannelOnToken(dst string, tokenAddr string) (ctype.CidType, string, error)
	LookupIngressChannelOnPay(payID ctype.PayIDType) (ctype.CidType, string, error)
	LookupEgressChannelOnPay(payID ctype.PayIDType) (ctype.CidType, string, error)
	// deprecated
	LookupNextChannel(dst string) (ctype.CidType, error)
}

type StateCallback interface {
	OnDispute(seqNum int)
}

type ChannelBalance struct {
	MyAddr     string
	MyFree     *big.Int
	MyLocked   *big.Int
	PeerAddr   string
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
