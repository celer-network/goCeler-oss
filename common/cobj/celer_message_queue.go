// Copyright 2018-2019 Celer Network

package cobj

import (
	"github.com/celer-network/goCeler-oss/rpc"
)

const (
	queueSize = 1 << 16
)

// CelerMessageQueue is a shared buffer for messages
type CelerMessageQueue struct {
	Queue chan *MessageInfo
}

// MessageInfo records the message and its destination
type MessageInfo struct {
	Destination string
	Message     *rpc.CelerMsg
}

// NewCelerMessageQueue creates a new message queue
func NewCelerMessageQueue() *CelerMessageQueue {
	return &CelerMessageQueue{make(chan *MessageInfo, queueSize)}
}
