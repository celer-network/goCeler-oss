// Copyright 2018-2019 Celer Network

package config

import (
	"time"

	"google.golang.org/grpc/keepalive"
)

var ChannelDisputeTimeout = uint64(10000)

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
	// AllowedTimeWindow is the window in seconds we allow client side timestamp to be different from server
	// ie. if server time is t, [t-window, t+window] are valid
	// if we want to be more strict, we'll need
	// a global map of addr->last auth ts and ensure it increments
	AllowedTimeWindow = 60
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
