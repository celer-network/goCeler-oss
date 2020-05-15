// Copyright 2020 Celer Network

package cli

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/deposit"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/golang/protobuf/proto"
)

const (
	DaySeconds = 86400
)

func (p *Processor) ViewChannel() {
	fmt.Println()
	if *chanid != "" || *peeraddr != "" {
		p.printSingleChannel()
		return
	}
	token := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(*tokenaddr))
	if *balance {
		p.printTokenBalance(token, *chanstate)
		return
	}
	if *count || *list || *detail {
		seconds := *inactivesec + (*inactiveday)*DaySeconds
		p.printChannels(token, *chanstate, seconds)
	}
}

func (p *Processor) ViewPay() {
	fmt.Println()
	if *payid == "" {
		log.Fatal("pay ID not specified")
	}
	p.printPayInfo(ctype.Hex2PayID(*payid))
}

func (p *Processor) ViewDeposit() {
	fmt.Println()
	if *depositid != "" {
		depositJob, found, err := p.dal.GetDepositJob(*depositid)
		if err != nil {
			log.Fatalf("GetDepositJob err: %s", err)
		}
		if !found {
			log.Fatalf("deposit job %s not found", *depositid)
		}
		fmt.Println("deposit:", deposit.PrintDepositJob(depositJob))
	} else if *chanid != "" {
		cid := ctype.Hex2Cid(*chanid)
		jobs, err := p.dal.GetAllDepositJobsByCid(cid)
		if err != nil {
			log.Fatalf("GetAllDepositJobsByCid err: %s", err)
		}
		jobnum := len(jobs)
		if jobnum == 0 {
			log.Infoln("no deposit jobs found for channel", *chanid)
			return
		}
		jobs = deposit.SortDepositJobs(jobs)
		fmt.Println("deposit jobs")
		for i := range jobs {
			job := jobs[jobnum-1-i]
			fmt.Println(i, deposit.PrintDepositJob(job))
		}
	}
}

func (p *Processor) ViewRoute() {
	fmt.Println()
	dest := ctype.Hex2Addr(*destaddr)
	token := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(*tokenaddr))
	p.printRoute(dest, token)
}

func (p *Processor) printSingleChannel() {
	var cid ctype.CidType
	if *chanid != "" {
		cid = ctype.Hex2Cid(*chanid)
	} else if *peeraddr != "" {
		peer := ctype.Hex2Addr(*peeraddr)
		token := ctype.Hex2Addr(*tokenaddr)
		var found bool
		var err error
		cid, found, err = p.dal.GetCidByPeerToken(peer, utils.GetTokenInfoFromAddress(token))
		if err != nil {
			log.Fatalf("GetCidByPeerToken %x %x error %s", peer, token, err)
		}
		if !found {
			log.Fatalf("peer %x token %x not found", peer, token)
		}
	} else {
		log.Fatal("must specify -cid or -peer")
	}
	p.printChannelInfo(cid)
	p.printInFlightChannelMessages(cid)
	if *payhistory {
		p.printAllPayInfo(cid)
	}
}

func (p *Processor) printChannelInfo(cid ctype.CidType) {
	fmt.Println("-- channel ID:", ctype.Cid2Hex(cid))

	state, statets, opents, chaninit, bal, selfsimplex, peersimplex, found, err := p.dal.GetChanViewInfoByID(cid)
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		p.printClosedChannelInfo(cid)
		return
	}
	fmt.Println("-- channel state:", fsm.ChanStateName(state), "| opened:", opents.UTC(), "| last activity:", statets.UTC())
	if chaninit != nil {
		fmt.Println("-- channel initializer:", utils.PrintChannelInitializer(chaninit))
	}
	fmt.Println("-- simplex from self:", utils.PrintSimplexChannel(selfsimplex))
	fmt.Println("-- simplex from peer:", utils.PrintSimplexChannel(peersimplex))
	fmt.Println("-- onchain balance: self deposit", bal.MyDeposit, "self withdrawal", bal.MyWithdrawal)
	fmt.Println("-- onchain balance: peer deposit", bal.PeerDeposit, "peer withdrawal", bal.PeerWithdrawal)
	if bal.PendingWithdrawal != nil {
		if bal.PendingWithdrawal.Amount.Cmp(big.NewInt(0)) == 1 {
			fmt.Println("-- pending withdraw amount", bal.PendingWithdrawal.Amount,
				"receiver", ctype.Addr2Hex(bal.PendingWithdrawal.Receiver), "deadline", bal.PendingWithdrawal.Deadline)
		}
	}
	blknum := ^uint64(0)
	header, err := p.nodeConfig.GetEthConn().HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Error(err)
	} else {
		blknum = header.Number.Uint64()
	}
	balance := ledgerview.ComputeBalance(selfsimplex, peersimplex, bal,
		ctype.Bytes2Addr(selfsimplex.GetPeerFrom()), ctype.Bytes2Addr(peersimplex.GetPeerFrom()), blknum)
	fmt.Println("-- self free balance:", balance.MyFree, "locked balance:", balance.MyLocked)
	fmt.Println("-- peer free balance:", balance.PeerFree, "locked balance:", balance.PeerLocked)
}

func (p *Processor) printClosedChannelInfo(cid ctype.CidType) {
	peer, token, opents, closets, found, err := p.dal.GetClosedChan(cid)
	if err != nil {
		log.Error(err)
		return
	}
	if !found {
		fmt.Println("Channel not found")
		return
	}
	tk := utils.PrintToken(token)
	fmt.Println("-- peer:", ctype.Addr2Hex(peer), "| token:", tk, "| opened:", opents.UTC(), "| closed:", closets.UTC())
}

func (p *Processor) printInFlightChannelMessages(cid ctype.CidType) {
	// get all channel pending pay
	msgs, err := p.dal.GetAllChanMessages(cid)
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(msgs) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("------------------- inflight channel messages -------------------")
	for _, msg := range msgs {
		if msg != nil {
			if msg.GetCondPayRequest() != nil {
				condpayBytes := msg.GetCondPayRequest().GetCondPay()
				payID := ctype.PayBytes2PayID(condpayBytes)
				var condpay entity.ConditionalPay
				err = proto.Unmarshal(condpayBytes, &condpay)
				if err != nil {
					log.Fatal(err)
				}
				seqnum, err2 := parseSimplexSeq(msg.GetCondPayRequest().GetStateOnlyPeerFromSig())
				if err2 != nil {
					log.Fatal(err2)
				}
				fmt.Println("-- condPayReq: seq:", seqnum,
					"payID:", ctype.PayID2Hex(payID),
					utils.PrintConditionalPay(&condpay),
				)

			} else if msg.GetPaymentSettleRequest() != nil {
				for _, settledPay := range msg.GetPaymentSettleRequest().GetSettledPays() {
					payID := ctype.Bytes2PayID(settledPay.GetSettledPayId())
					reason := settledPay.GetReason()
					amt := new(big.Int).SetBytes(settledPay.GetAmount())
					seqnum, err2 := parseSimplexSeq(msg.GetPaymentSettleRequest().GetStateOnlyPeerFromSig())
					if err2 != nil {
						log.Fatal(err2)
					}
					fmt.Println("-- settlePayReq: seq:", seqnum,
						"payID:", ctype.PayID2Hex(payID),
						"reason:", reason,
						"amount:", amt,
					)
				}
			}
		}
	}
}

func parseSimplexSeq(simplexState *rpc.SignedSimplexState) (uint64, error) {
	var simplex entity.SimplexPaymentChannel
	err := proto.Unmarshal(simplexState.GetSimplexState(), &simplex)
	if err != nil {
		return 0, err
	}
	return simplex.GetSeqNum(), nil
}

func (p *Processor) printChannels(token *entity.TokenInfo, state int, inactiveSeconds int) {
	var count int
	var err error
	var inactiveTime time.Time
	if inactiveSeconds > 0 {
		inactiveTime = time.Now().Add(time.Duration(-inactiveSeconds) * time.Second).UTC()
		count, err = p.dal.CountInactiveCidsByTokenAndState(token, state, inactiveTime)
	} else {
		count, err = p.dal.CountCidsByTokenAndState(token, state)
	}
	if err != nil {
		log.Fatal(err)
	}
	suffix := ""
	if inactiveSeconds > 0 {
		suffix = fmt.Sprintf(", inactive since %s", inactiveTime)
	}
	chs := "channels"
	if count < 2 {
		chs = "channel"
	}
	fmt.Printf("-- %d %s at state %s, token %s%s\n\n", count, chs, fsm.ChanStateName(state), utils.PrintToken(token), suffix)
	if *list && count > 0 {
		if *detail {
			p.printDetailedChannels(token, state, inactiveSeconds, inactiveTime)
			return
		}
		var cids []ctype.CidType
		if inactiveSeconds > 0 {
			cids, err = p.dal.GetInactiveCidsByTokenAndState(token, state, inactiveTime)
		} else {
			cids, err = p.dal.GetCidsByTokenAndState(token, state)
		}
		if err != nil {
			log.Fatal(err)
		}
		for _, cid := range cids {
			fmt.Println(ctype.Cid2Hex(cid))
		}
	}
}

func (p *Processor) printDetailedChannels(token *entity.TokenInfo, state int, inactiveSeconds int, inactiveTime time.Time) {
	var cids []ctype.CidType
	var peers []ctype.Addr
	var tokens []*entity.TokenInfo
	var stateTses, openTses []*time.Time
	var balances []*structs.OnChainBalance
	var err error
	var selfSimplexes, peerSimplexes []*entity.SimplexPaymentChannel
	if inactiveSeconds > 0 {
		cids, peers, tokens, stateTses, openTses, balances, selfSimplexes, peerSimplexes, err = p.dal.GetInactiveChansByTokenAndState(token, state, inactiveTime)
	} else {
		cids, peers, tokens, stateTses, openTses, balances, selfSimplexes, peerSimplexes, err = p.dal.GetAllChansByTokenAndState(token, state)
	}
	if err != nil {
		log.Fatal(err)
	}

	blknum := ^uint64(0)
	header, err := p.nodeConfig.GetEthConn().HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Error(err)
	} else {
		blknum = header.Number.Uint64()
	}
	for i, _ := range cids {
		fmt.Println("-- channel ID:", ctype.Cid2Hex(cids[i]))
		tk := utils.PrintToken(tokens[i])
		fmt.Println("-- peer:", ctype.Addr2Hex(peers[i]), "token:", tk)
		fmt.Println("-- channel state:", fsm.ChanStateName(state), "| opened:", openTses[i].UTC(), "| last activity:", stateTses[i].UTC())
		fmt.Println("-- onchain balance: self deposit", balances[i].MyDeposit, "self withdrawal", balances[i].MyWithdrawal)
		fmt.Println("-- onchain balance: peer deposit", balances[i].PeerDeposit, "peer withdrawal", balances[i].PeerWithdrawal)
		balance := ledgerview.ComputeBalance(selfSimplexes[i], peerSimplexes[i], balances[i],
			ctype.Bytes2Addr(selfSimplexes[i].GetPeerFrom()), ctype.Bytes2Addr(peerSimplexes[i].GetPeerFrom()), blknum)
		fmt.Println("-- self free balance:", balance.MyFree, "locked balance:", balance.MyLocked)
		fmt.Println("-- peer free balance:", balance.PeerFree, "locked balance:", balance.PeerLocked)
		fmt.Println()
	}
}

func (p *Processor) printTokenBalance(token *entity.TokenInfo, state int) {
	count, err := p.dal.CountCidsByTokenAndState(token, state)
	if err != nil {
		log.Fatal(err)
	}
	chs := "channels"
	if count < 2 {
		chs = "channel"
	}
	fmt.Printf("-- %d %s at state %s, token %s\n", count, chs, fsm.ChanStateName(state), utils.PrintToken(token))
	if count == 0 {
		return
	}
	cids, _, _, _, _, balances, selfSimplexes, peerSimplexes, err := p.dal.GetAllChansByTokenAndState(token, state)
	if err != nil {
		log.Fatal(err)
	}
	blknum := ^uint64(0)
	header, err := p.nodeConfig.GetEthConn().HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Error(err)
	} else {
		blknum = header.Number.Uint64()
	}
	freeBalance := new(big.Int).SetUint64(0)
	lockedBalance := new(big.Int).SetUint64(0)
	for i, _ := range cids {
		balance := ledgerview.ComputeBalance(selfSimplexes[i], peerSimplexes[i], balances[i],
			ctype.Bytes2Addr(selfSimplexes[i].GetPeerFrom()), ctype.Bytes2Addr(peerSimplexes[i].GetPeerFrom()), blknum)
		freeBalance = freeBalance.Add(freeBalance, balance.MyFree)
		lockedBalance = lockedBalance.Add(lockedBalance, balance.MyLocked)
	}
	fmt.Println("-- total free balance:", freeBalance, "locked balance:", lockedBalance)
}

func (p *Processor) printAllPayInfo(cid ctype.CidType) {
	payIds, pays, notes, inCids, inStates, outCids, outStates, createTses, err := p.dal.GetAllPaymentInfoByCid(cid)
	if err != nil {
		log.Fatal(err)
	}
	if len(payIds) > 0 {
		fmt.Println()
		fmt.Println("------------------- past and current pays -------------------")
	}

	for i, _ := range payIds {
		fmt.Println("-- pay ID", ctype.PayID2Hex(payIds[i]), "created at", createTses[i].UTC())
		fmt.Println("-- conditional pay", utils.PrintConditionalPay(pays[i]))
		if inCids[i] != ctype.ZeroCid {
			fmt.Println("-- ingress cid", ctype.Cid2Hex(inCids[i]), "state", fsm.PayStateName(inStates[i]))
		}
		if outCids[i] != ctype.ZeroCid {
			fmt.Println("-- egress cid", ctype.Cid2Hex(outCids[i]), "state", fsm.PayStateName(outStates[i]))
		}
		if notes[i] != nil && len(notes[i].Value) != 0 {
			notejson, _ := utils.PbToJSONString(notes[i])
			if notejson != "" {
				fmt.Println("-- pay note:", notejson)
			} else {
				fmt.Println("-- pay note:", notes[i])
			}
		}
		fmt.Println()
	}
}

func (p *Processor) printPayInfo(pid ctype.PayIDType) {
	fmt.Println("")
	pay, note, inCid, inState, outCid, outState, createTs, found, err := p.dal.GetPaymentInfo(pid)
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		log.Fatalf("payment %x not found", pid)
	}

	fmt.Println("-- pay ID", ctype.PayID2Hex(pid), "created at", createTs.UTC())
	fmt.Println("-- conditional pay", utils.PrintConditionalPay(pay))
	if inCid != ctype.ZeroCid {
		peer, found, err2 := p.dal.GetChanPeer(inCid)
		if err2 != nil {
			log.Fatal(err2)
		}
		if !found {
			log.Fatalf("peer of cid %x not found", inCid)
		}
		fmt.Println("-- ingress cid", ctype.Cid2Hex(inCid), "peer", ctype.Addr2Hex(peer), "state", fsm.PayStateName(inState))
	}
	if outCid != ctype.ZeroCid {
		peer, found, err2 := p.dal.GetChanPeer(outCid)
		if err2 != nil {
			log.Fatal(err2)
		}
		if !found {
			log.Fatalf("peer of cid %x not found", outCid)
		}
		fmt.Println("-- egress cid", ctype.Cid2Hex(outCid), "peer", ctype.Addr2Hex(peer), "state", fsm.PayStateName(outState))
	}
	if note != nil && len(note.Value) != 0 {
		notejson, _ := utils.PbToJSONString(note)
		if notejson != "" {
			fmt.Println("-- pay note:", notejson)
		} else {
			fmt.Println("-- pay note:", note)
		}
	}
}

func (p *Processor) printRoute(dest ctype.Addr, token *entity.TokenInfo) {
	fmt.Printf("-- route info for destination %x token %s\n", dest, utils.PrintToken(token))
	cid, found, err := p.dal.GetCidByPeerToken(dest, token)
	if err != nil {
		log.Fatal(err)
	}
	if found {
		fmt.Println("-- direct", p.getCidPeerStr(cid))
		return
	}
	accessOsps, err := p.dal.GetDestTokenOsps(dest, token)
	if err != nil {
		log.Fatal(err)
	}
	if len(accessOsps) != 0 {
		ospstr := ""
		for _, osp := range accessOsps {
			ospstr += ctype.Addr2Hex(osp) + ","
		}
		fmt.Println("-- access osp:", ospstr[:len(ospstr)-1])
		for _, osp := range accessOsps {
			cid, found, err = p.dal.GetRoutingCid(osp, token)
			if err != nil {
				log.Errorf("GetRoutingCid err: %s, osp: %x", err, osp)
			} else if found {
				fmt.Printf("-- next hop to osp %x: %s\n", osp, p.getCidPeerStr(cid))
				return
			}
		}
	} else {
		cid, found, err = p.dal.GetRoutingCid(dest, token)
		if err != nil {
			log.Fatal(err)
		}
		if found {
			fmt.Println("-- next hop", p.getCidPeerStr(cid))
			return
		}
	}
	fmt.Println("-- cannot find route from db, use default route if set")
}

func (p *Processor) getCidPeerStr(cid ctype.CidType) string {
	peer, found, err := p.dal.GetChanPeer(cid)
	if err != nil {
		log.Fatalf("GetChanPeer err %s, cid %x", err, cid)
	}
	if !found {
		log.Fatalf("peer for channel %x not found", cid)
	}
	return fmt.Sprintf("channel %x peer %x", cid, peer)
}
