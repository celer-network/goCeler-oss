// Copyright 2018-2019 Celer Network

package graph

import "github.com/celer-network/goCeler-oss/ctype"

// Edge describes evnet happenning to a channel
type Edge struct {
	P1        ctype.Addr
	P2        ctype.Addr
	Cid       ctype.CidType
	TokenAddr ctype.Addr
}
