// Copyright 2020 Celer Network

package main

import (
	"flag"

	"github.com/celer-network/goCeler/tools/osp-cli/cli"
	"github.com/celer-network/goutils/log"
)

var (
	registerstream  = flag.Bool("registerstream", false, "register stream with a peer OSP")
	openchannel     = flag.Bool("openchannel", false, "open a channel with a peer OSP")
	sendtoken       = flag.Bool("sendtoken", false, "make an off-chain payment")
	deposit         = flag.Bool("deposit", false, "make an on-chain deposit to a channel")
	querydeposit    = flag.Bool("querydeposit", false, "query the status of a deposit job")
	querypeerosps   = flag.Bool("querypeerosps", false, "query info of peer OSPs")
	intendsettle    = flag.Bool("intendsettle", false, "intend unilaterally settle channel")
	confirmsettle   = flag.Bool("confirmsettle", false, "confirm unilaterally settle channel")
	intendwithdraw  = flag.Bool("intendwithdraw", false, "intend unilaterally withdraw from channel")
	confirmwithdraw = flag.Bool("confirmwithdraw", false, "confirm unilaterally withdraw from channel")
	dbview          = flag.String("dbview", "", "database view command")
	dbupdate        = flag.String("dbupdate", "", "database update command")
	onchainview     = flag.String("onchainview", "", "onchain view command")
	ethpooldeposit  = flag.Bool("ethpooldeposit", false, "deposit ETH to ethpool")
	ethpoolwithdraw = flag.Bool("ethpoolwithdraw", false, "withdraw ETH from ethpool")
	register        = flag.Bool("register", false, "register OSP as a state channel router")
	deregister      = flag.Bool("deregister", false, "deregister OSP as a state channel router")
)

func main() {
	flag.Parse()
	cli.CheckFlags()

	if *registerstream {
		cli.RegisterStream()
		return
	}
	if *openchannel {
		cli.OpenChannel()
		return
	}
	if *sendtoken {
		cli.SendToken()
		return
	}
	if *deposit {
		cli.MakeDeposit()
		return
	}
	if *querydeposit {
		cli.QueryDeposit()
		return
	}
	if *querypeerosps {
		cli.QueryPeerOsps()
		return
	}

	var p cli.Processor
	if *intendsettle || *confirmsettle || *intendwithdraw || *confirmwithdraw || *dbview != "" || *dbupdate != "" {
		p.Setup(true, false, true) // connect to db, not enforcig osp keystore, set disputer if keystore is provided
	} else if *ethpoolwithdraw || *register || *deregister {
		p.Setup(false, true, false) // no db, enforce using osp keystore, no disputer
	} else if *ethpooldeposit || *onchainview != "" {
		p.Setup(false, false, false) // no db, not enforcig osp keystore, no disputer
	} else {
		return
	}

	if *intendsettle {
		p.IntendSettle()
		return
	}
	if *confirmsettle {
		p.ConfirmSettle()
		return
	}
	if *intendwithdraw {
		p.IntendWithdraw()
		return
	}
	if *confirmwithdraw {
		p.ConfirmWithdraw()
		return
	}

	switch *dbview {
	case "":
	case "channel":
		p.ViewChannel()
	case "pay":
		p.ViewPay()
	case "deposit":
		p.ViewDeposit()
	case "route":
		p.ViewRoute()
	default:
		log.Fatalln("unsupported dbview command", *dbview)
	}

	switch *dbupdate {
	case "":
	case "netid":
		p.UpdateNetId()
	case "netbridge":
		p.UpdateNetBridge()
	case "bridgerouting":
		p.UpdateBridgeRouting()
	case "nettoken":
		p.UpdateNetToken()
	case "xnet":
		p.ConfigXnet()
	default:
		log.Fatalln("unsupported dbupdate command", *dbupdate)
	}

	switch *onchainview {
	case "":
	case "channel":
		p.ViewChannelOnChain()
	case "pay":
		p.ViewPayOnChain()
	case "tx":
		p.ViewTxOnChain()
	case "app":
		p.ViewAppOnChain()
	default:
		log.Fatalln("unsupported chainview command", *onchainview)
	}

	if *ethpooldeposit {
		p.EthPoolDeposit()
	} else if *ethpoolwithdraw {
		p.EthPoolWithdraw()
	}
	if *register {
		p.RegisterRouter()
	} else if *deregister {
		p.DeregisterRouter()
	}
}
