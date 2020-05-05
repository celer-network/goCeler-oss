// Copyright 2018-2020 Celer Network

package config

import (
	"math/big"
	"time"

	"google.golang.org/grpc/keepalive"
)

var ChannelDisputeTimeout = uint64(10000)
var ChainID *big.Int
var BlockDelay = uint64(5)
var EventListenerHttp = ""
var RouterBcastInterval = 293 * time.Second
var RouterBuildInterval = 367 * time.Second
var RouterAliveTimeout = 900 * time.Second
var OspClearPaysInterval = 613 * time.Second

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
