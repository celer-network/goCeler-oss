// Copyright 2020 Celer Network

package main

import (
	"flag"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

var (
	// operations
	registerstream = flag.Bool("registerstream", false, "register stream with a peer OSP")
	openchannel    = flag.Bool("openchannel", false, "open a channel with a peer OSP")
	sendtoken      = flag.Bool("sendtoken", false, "make an off-chain payment")
	deposit        = flag.Bool("deposit", false, "make an on-chain deposit")
	querydeposit   = flag.Bool("querydeposit", false, "query the status of a deposit job")
	querypeerosps  = flag.Bool("querypeerosps", false, "query info of peer OSPs")
	// args
	adminhostport = flag.String("adminhostport", "", "The server admin http host and port")
	peeraddr      = flag.String("peeraddr", "", "peer eth address")
	peerhostport  = flag.String("peerhostport", "", "peer host and port")
	token         = flag.String("token", "", "token address")
	peerdeposit   = flag.Float64("peerdeposit", 0, "channel deposit to be made by peer")
	selfdeposit   = flag.Float64("selfdeposit", 0, "channel deposit to be made by self")
	receiver      = flag.String("receiver", "", "receiver eth address")
	amount        = flag.Float64("amount", 0, "amount in unit of 1e18")
	topeer        = flag.Bool("topeer", false, "deposit to the peer side of the channel")
	maxwaitsec    = flag.Int("maxwaitsec", 0, "time (in sec) allowed to be wait for deposit job batching")
	depositid     = flag.String("depositid", "", "deposit job id")
)

func main() {
	flag.Parse()
	if *adminhostport == "" {
		log.Fatal("empty adminhostport")
	}
	if *amount < 0 || *peerdeposit < 0 || *selfdeposit < 0 {
		log.Fatal("incorrect amount")
	}

	if *registerstream {
		registerStream()
	}

	if *openchannel {
		openChannel()
	}

	if *sendtoken {
		sendToken()
	}

	if *deposit {
		makeDeposit()
	}

	if *querydeposit {
		queryDeposit()
	}

	if *querypeerosps {
		queryPeerOsps()
	}
}

func registerStream() {
	err := utils.RequestRegisterStream(*adminhostport, ctype.Hex2Addr(*peeraddr), *peerhostport)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("registered stream to %s, addr %s", *peerhostport, *peeraddr)
}

func openChannel() {
	peerDepositWei := utils.Float2Wei(*peerdeposit)
	selfDepositWei := utils.Float2Wei(*selfdeposit)
	err := utils.RequestOpenChannel(
		*adminhostport, ctype.Hex2Addr(*peeraddr), ctype.Hex2Addr(*token), peerDepositWei, selfDepositWei)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("channel opened with %s, token %s, deposit: <self %f peer %f>",
		*peeraddr, utils.PrintTokenAddr(ctype.Hex2Addr(*token)), *selfdeposit, *peerdeposit)
}

func sendToken() {
	amtWei := utils.Float2Wei(*amount)
	payID, err := utils.RequestSendToken(
		*adminhostport, ctype.Hex2Addr(*receiver), ctype.Hex2Addr(*token), amtWei)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("sent payment %x to %s, token %s, amount %f",
		payID, *receiver, utils.PrintTokenAddr(ctype.Hex2Addr(*token)), *amount)
}

func makeDeposit() {
	amtWei := utils.Float2Wei(*amount)
	depositID, err := utils.RequestDeposit(
		*adminhostport, ctype.Hex2Addr(*peeraddr), ctype.Hex2Addr(*token), amtWei, *topeer, uint64(*maxwaitsec))
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("made deposit %s, peer %s, token %s, amount %f, topeer %t",
		depositID, *peeraddr, utils.PrintTokenAddr(ctype.Hex2Addr(*token)), *amount, *topeer)
}

func queryDeposit() {
	res, err := utils.QueryDeposit(*adminhostport, *depositid)
	if err != nil {
		log.Error(err)
		return
	}
	if res.Error != "" {
		log.Infof("got deposit %s status %s errmsg %s",
			*depositid, rpc.DepositState_name[int32(res.DepositState)], res.Error)
		return
	}
	log.Infof("got deposit %s status %s", *depositid, rpc.DepositState_name[int32(res.DepositState)])
}

func queryPeerOsps() {
	res, err := utils.QueryPeerOsps(*adminhostport)
	if err != nil {
		log.Error(err)
		return
	}
	for _, peerOsp := range res.GetPeerOsps() {
		updateTs := time.Unix(int64(peerOsp.GetUpdateTs()), 0).UTC()
		diffTs := time.Now().UTC().Sub(updateTs)
		log.Infof("peer OSP address %s, updated %s (%s before now)",
			peerOsp.GetOspAddress(), updateTs, diffTs)
		for _, tkcid := range peerOsp.GetTokenCidPairs() {
			log.Infof("-- token %s cid %s", utils.PrintTokenAddr(ctype.Hex2Addr(tkcid.TokenAddress)), tkcid.Cid)
		}
	}
}
