// Copyright 2020 Celer Network

package cli

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
)

func RegisterStream() {
	if *batchfile != "" {
		file, err := os.Open(*batchfile)
		if err != nil {
			log.Error(err)
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) != 2 {
				log.Errorln("invalid file input:", line)
				return
			}
			peerAddr := fields[0]
			peerHostPort := fields[1]
			registerStream(peerAddr, peerHostPort)
		}
	} else {
		registerStream(*peeraddr, *peerhostport)
	}
}

func registerStream(peerAddr, peerHostPort string) {
	err := utils.RequestRegisterStream(*adminhostport, ctype.Hex2Addr(peerAddr), peerHostPort)
	if err != nil {
		log.Errorf("err while registering stream to %s, addr %s: %s", peerHostPort, peerAddr, err)
		return
	}
	log.Infof("registered stream to %s, addr %s", peerHostPort, peerAddr)
}

func OpenChannel() {
	peerDepositWei := utils.Float2Wei(*peerdeposit)
	selfDepositWei := utils.Float2Wei(*selfdeposit)
	err := utils.RequestOpenChannel(
		*adminhostport, ctype.Hex2Addr(*peeraddr), ctype.Hex2Addr(*tokenaddr), peerDepositWei, selfDepositWei)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("requested to open channel with %s, token %s, deposit: <self %f peer %f>",
		*peeraddr, utils.PrintTokenAddr(ctype.Hex2Addr(*tokenaddr)), *selfdeposit, *peerdeposit)
}

func SendToken() {
	amtWei := utils.Float2Wei(*amount)
	payID, err := utils.RequestSendToken(
		*adminhostport, ctype.Hex2Addr(*receiver), ctype.Hex2Addr(*tokenaddr), amtWei, *netid)
	if err != nil {
		log.Error(err)
		return
	}
	if *netid == 0 {
		log.Infof("requested to send payment %x to %s, token %s, amount %f",
			payID, *receiver, utils.PrintTokenAddr(ctype.Hex2Addr(*tokenaddr)), *amount)
	} else {
		log.Infof("requested to send payment %x to %s, token %s, amount %f, netid %d",
			payID, *receiver, utils.PrintTokenAddr(ctype.Hex2Addr(*tokenaddr)), *amount, *netid)
	}
}

func MakeDeposit() {
	amtWei := utils.Float2Wei(*amount)
	depositID, err := utils.RequestDeposit(
		*adminhostport, ctype.Hex2Addr(*peeraddr), ctype.Hex2Addr(*tokenaddr), amtWei, *topeer, uint64(*maxwaitsec))
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("requested to make deposit %s, peer %s, token %s, amount %f, topeer %t",
		depositID, *peeraddr, utils.PrintTokenAddr(ctype.Hex2Addr(*tokenaddr)), *amount, *topeer)
}

func QueryDeposit() {
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

func QueryPeerOsps() {
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
