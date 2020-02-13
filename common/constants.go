// Copyright 2018 Celer Network

package common

const (
	PendingSpPrefix = "pendingSp:"
)

type RoutingPolicy int

const (
	NoRoutingPolicy       RoutingPolicy = 1 << iota
	GateWayPolicy         RoutingPolicy = 1 << iota
	ServiceProviderPolicy RoutingPolicy = 1 << iota
)

const RoutingTableDestTokenSpliter = "@"

const (
	EthContractAddr = "0000000000000000000000000000000000000000"
)
