// Copyright 2018-2020 Celer Network

package config

import (
	"math/big"
	"time"

	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goutils/eth"
	"google.golang.org/grpc/keepalive"
)

// NOTE: not protected by lock, only set once at initialization
var (
	ChainId               *big.Int
	ChannelDisputeTimeout = uint64(10000)
	BlockDelay            = uint64(5)
	BlockIntervalSec      = uint64(10)
	EventListenerHttp     = ""
	RouterBcastInterval   = 293 * time.Second
	RouterBuildInterval   = 367 * time.Second
	RouterAliveTimeout    = 900 * time.Second
	OspClearPaysInterval  = 613 * time.Second
	OspReportInverval     = 887 * time.Second
)

const (
	ClientCacheSize            = 1000
	ServerCacheSize            = 16
	OpenChannelTimeout         = uint64(100)
	CooperativeWithdrawTimeout = uint64(10)
	WithdrawTimeoutSafeMargin  = uint64(6) // TODO: this should be profile.blockdelay + margin
	PayResolveTimeout          = uint64(10)
	PaySendTimeoutSafeMargin   = uint64(6)
	PayRecvTimeoutSafeMargin   = uint64(4)
	AdminSendTokenTimeout      = uint64(50)
	QuickCatchBlockDelay       = uint64(2)
	TcbTimeoutInBlockNumber    = 576000

	// Protocol Version in AuthReq, >=1 support sync
	AuthProtocolVersion = uint64(1)
	// AuthAckTimeout is duration client will wait for AuthAck msg
	AuthAckTimeout = 5 * time.Second

	// grpc dial timeout second, block until 15s
	GrpcDialTimeout = 15

	EventListenerLeaseName          = "eventlistener"
	EventListenerLeaseRenewInterval = 60 * time.Second
	EventListenerLeaseTimeout       = 90 * time.Second

	// used by clients to control onchain query frequency
	QueryName_OnChainBalance      = "onchainBalance"
	QueryName_OnChainResolvedPays = "onchainResolvedPays"
)

// KeepAliveClientParams is grpc client side keeyalive parameters
// Make sure these parameters are set in coordination with the keepalive policy
// on the server, as incompatible settings can result in closing of connection
var KeepAliveClientParams = keepalive.ClientParameters{
	Time:                15 * time.Second, // send pings every 15 seconds if there is no activity
	Timeout:             3 * time.Second,  // wait 3 seconds for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// KeepAliveServerParams is grpc server side keeyalive parameters
var KeepAliveServerParams = keepalive.ServerParameters{
	Time:    20 * time.Second, // send pings every 20 seconds if there is no activity
	Timeout: 3 * time.Second,  // wait 3 seconds for ping ack before considering the connection dead
}

// KeepAliveEnforcePolicy is grpc server side policy
var KeepAliveEnforcePolicy = keepalive.EnforcementPolicy{
	MinTime:             12 * time.Second, // must be smaller than clientParam.Time
	PermitWithoutStream: true,
}

func SetGlobalConfigFromProfile(profile *common.CProfile) {
	ChainId = big.NewInt(profile.ChainId)
	BlockDelay = profile.BlockDelayNum
	if profile.PollingInterval != 0 {
		BlockIntervalSec = profile.PollingInterval
	}
	if profile.DisputeTimeout != 0 {
		ChannelDisputeTimeout = profile.DisputeTimeout
	}
}

func WaitMinedOptions() []eth.TxOption {
	return []eth.TxOption{
		eth.WithBlockDelay(BlockDelay),
		eth.WithPollingInterval(time.Duration(BlockIntervalSec) * time.Second),
		eth.WithTimeout(time.Duration(rtconfig.GetWaitMinedTxTimeout()) * time.Second),
		eth.WithQueryTimeout(time.Duration(rtconfig.GetWaitMinedTxQueryTimeout()) * time.Second),
		eth.WithQueryRetryInterval(time.Duration(rtconfig.GetWaitMinedTxQueryRetryInterval()) * time.Second),
	}
}

func TransactOptions(opts ...eth.TxOption) []eth.TxOption {
	options := []eth.TxOption{
		eth.WithMinGasGwei(rtconfig.GetMinGasGwei()),
		eth.WithMaxGasGwei(rtconfig.GetMaxGasGwei()),
		eth.WithAddGasGwei(rtconfig.GetAddGasGwei()),
	}
	options = append(options, WaitMinedOptions()...)
	return append(options, opts...)
}

func QuickTransactOptions(opts ...eth.TxOption) []eth.TxOption {
	options := TransactOptions(opts...)
	if QuickCatchBlockDelay < BlockDelay {
		// this will overwrite the previous WithBlockDelay option
		options = append(options, eth.WithBlockDelay(QuickCatchBlockDelay))
	}
	return options
}
