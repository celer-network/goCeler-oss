// Copyright 2020 Celer Network

package cli

import (
	"flag"

	"github.com/celer-network/goutils/log"
)

var (
	// configurations
	adminhostport = flag.String("adminhostport", "", "the server admin http host:port")
	pjson         = flag.String("profile", "", "OSP profile")
	storesql      = flag.String("storesql", "", "cockroachDB URL")
	storedir      = flag.String("storedir", "", "sqlite store directory")
	ksfile        = flag.String("ks", "", "key store file")
	blkdelay      = flag.Int("blkdelay", 0, "block delay for wait mined")
	nopassword    = flag.Bool("nopassword", false, "empty password for keystores")
	// parameters
	chanid       = flag.String("cid", "", "channel id")
	peeraddr     = flag.String("peer", "", "peer eth address")
	tokenaddr    = flag.String("token", "", "token address")
	chanstate    = flag.Int("chanstate", 3, "channel state") // default: 3 - opened
	payhistory   = flag.Bool("payhistory", false, "list all pays made to and from this channel")
	balance      = flag.Bool("balance", false, "get total balance of all channels for given token and state")
	count        = flag.Bool("count", false, "count channels for given input")
	list         = flag.Bool("list", false, "list channels for given input")
	detail       = flag.Bool("detail", false, "list detailed info of channels for given input")
	peerhostport = flag.String("peerhostport", "", "peer host and port")
	peerdeposit  = flag.Float64("peerdeposit", 0, "channel deposit to be made by peer")
	selfdeposit  = flag.Float64("selfdeposit", 0, "channel deposit to be made by self")
	receiver     = flag.String("receiver", "", "receiver eth address")
	amount       = flag.Float64("amount", 0, "amount in unit of 1e18")
	topeer       = flag.Bool("topeer", false, "deposit to the peer side of the channel")
	maxwaitsec   = flag.Int("maxwaitsec", 0, "time (in sec) allowed to be wait for deposit job batching")
	depositid    = flag.String("depositid", "", "deposit job id")
	payid        = flag.String("payid", "", "payment id")
	destaddr     = flag.String("dest", "", "destination address")
	txhash       = flag.String("txhash", "", "on-chain transaction hash")
	inactiveday  = flag.Int("inactiveday", 0, "days of being inactive")
	inactivesec  = flag.Int("inactivesec", 0, "seconds of being inactive")
	batchfile    = flag.String("file", "", "file path for batch operation")
	withdrawto   = flag.String("withdrawto", "", "channel ID if withdraw fund to another channel")
	appaddr      = flag.String("appaddr", "", "app onchain address")
	argfinalize  = flag.String("finalize", "", "arg for query finalized")
	argoutcome   = flag.String("outcome", "", "arg for query outcome")
	argdecode    = flag.Bool("decode", false, "decode arg according to multisession app format")
	// for db update
	bridgeaddr = flag.String("bridgeaddr", "", "net bridge address")
	netid      = flag.Uint64("netid", 0, "net id")
	localtoken = flag.String("localtoken", "", "local token address")
)

func CheckFlags() {
	if *amount < 0 || *peerdeposit < 0 || *selfdeposit < 0 || *maxwaitsec < 0 {
		log.Fatal("incorrect parameters")
	}
}
