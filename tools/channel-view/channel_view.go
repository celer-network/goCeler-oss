// Copyright 2018-2020 Celer Network

package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/deposit"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goCeler/tools/toolsetup"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/golang/protobuf/proto"
)

const (
	DaySeconds = 86400
)

var (
	pjson       = flag.String("profile", "", "OSP profile")
	storesql    = flag.String("storesql", "", "sql database URL")
	storedir    = flag.String("storedir", "", "local database directory")
	dbview      = flag.String("dbview", "", "database view command")
	chainview   = flag.String("chainview", "", "onchain view command")
	channelID   = flag.String("cid", "", "channel ID")
	chanState   = flag.Int("chanstate", 0, "channel state")
	peerAddr    = flag.String("peer", "", "peer address")
	tokenAddr   = flag.String("token", "", "token address")
	inactiveDay = flag.Int("inactiveday", 0, "days of being inactive")
	inactiveSec = flag.Int("inactivesec", 0, "seconds of being inactive")
	allIds      = flag.Bool("allids", false, "in addition to stats count, get all ids of entries")
	allPays     = flag.Bool("allpays", false, "list all pays related to this channel (heavy db operation)")
	payID       = flag.String("payid", "", "pay ID")
	depositID   = flag.String("depositid", "", "deposit job ID")
	destAddr    = flag.String("dest", "", "destination address")
	txhash      = flag.String("txhash", "", "on-chain transaction hash")
	appAddr     = flag.String("appaddr", "", "app onchain address")
	argFinalize = flag.String("finalize", "", "arg for query finalized")
	argOutcome  = flag.String("outcome", "", "arg for query outcome")
	argDecode   = flag.Bool("decode", false, "decode arg according to multisession app format")
)

type processor struct {
	myAddr     ctype.Addr
	nodeConfig common.GlobalNodeConfig
	dal        *storage.DAL
}

func main() {
	flag.Parse()
	var p processor
	p.setup()

	switch *dbview {
	case "":

	case "channel":
		var cid ctype.CidType
		if *channelID != "" {
			cid = ctype.Hex2Cid(*channelID)
		} else if *peerAddr != "" {
			peer := ctype.Hex2Addr(*peerAddr)
			token := ctype.Hex2Addr(*tokenAddr)
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
		if *allPays {
			p.printAllPayInfo(cid)
		}

	case "pay":
		if *payID == "" {
			log.Fatal("pay ID not specified")
		}
		pid := ctype.Hex2PayID(*payID)
		p.printPayInfo(pid)

	case "deposit":
		if *depositID != "" {
			depositJob, found, err := p.dal.GetDepositJob(*depositID)
			if err != nil {
				log.Fatalf("GetDepositJob err: %s", err)
			}
			if !found {
				log.Fatalf("deposit job %s not found", *depositID)
			}
			fmt.Println("")
			fmt.Println("deposit:", deposit.PrintDepositJob(depositJob))
		} else if *channelID != "" {
			cid := ctype.Hex2Cid(*channelID)
			jobs, err := p.dal.GetAllDepositJobsByCid(cid)
			if err != nil {
				log.Fatalf("GetAllDepositJobsByCid err: %s", err)
			}
			jobnum := len(jobs)
			if jobnum == 0 {
				log.Infoln("no deposit jobs found for channel", *channelID)
				return
			}
			jobs = deposit.SortDepositJobs(jobs)
			fmt.Println("deposit jobs")
			for i := range jobs {
				job := jobs[jobnum-1-i]
				fmt.Println(i, deposit.PrintDepositJob(job))
			}
		}

	case "route":
		dest := ctype.Hex2Addr(*destAddr)
		token := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(*tokenAddr))
		p.printRoute(dest, token)

	case "allchan": // list all channels of a given token, heavy db operation
		token := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(*tokenAddr))
		seconds := *inactiveSec + (*inactiveDay)*DaySeconds
		p.printAllChannelsByToken(token, seconds)

	case "balance": // print the total balance of all channels, heavy db operation
		token := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(*tokenAddr))
		p.printTokenBalance(token)

	case "stats":
		token := utils.GetTokenInfoFromAddress(ctype.Hex2Addr(*tokenAddr))
		seconds := *inactiveSec + (*inactiveDay)*DaySeconds
		if *chanState != 0 {
			p.printCidsByTokenAndState(token, *chanState, seconds)
		}

	default:
		log.Fatal("unsupported dbview command", *dbview)
	}

	switch *chainview {
	case "":

	case "channel":
		if *channelID == "" {
			log.Fatal("channel ID not specified")
		}
		cid := ctype.Hex2Cid(*channelID)
		p.printChannelLedgerInfo(cid)

	case "pay":
		if *payID == "" {
			log.Fatal("pay ID not specified")
		}
		pid := ctype.Hex2PayID(*payID)
		p.printPayRegistryInfo(pid)

	case "tx":
		txInfo, err := ledgerview.GetOnChainTxByHash(ctype.Hex2Hash(*txhash), p.nodeConfig)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("%s tx msg from %x to %x", txInfo.FuncName, txInfo.From, txInfo.To)
		// TODO: print txInfo.FuncInput

	case "app":
		if *appAddr == "" || *argOutcome == "" {
			log.Fatal("app address or arg for query outcome not specified")
		}
		addr := ctype.Hex2Addr(*appAddr)
		argF := ctype.Hex2Bytes(*argFinalize)
		argO := ctype.Hex2Bytes(*argOutcome)
		p.printAppBooleanOutcome(addr, argF, argO)

	default:
		log.Fatal("unsupported chainview command", *chainview)
	}
}

func (p *processor) setup() {
	if *dbview == "" && *chainview == "" {
		log.Fatal("no db or onchain view commands")
	}
	profile := common.ParseProfile(*pjson)
	overrideConfig(profile)
	p.myAddr = ctype.Hex2Addr(profile.SvrETHAddr)
	if *dbview != "" {
		p.dal = toolsetup.NewDAL(profile)
	}
	ethclient := toolsetup.NewEthClient(profile)
	p.nodeConfig = toolsetup.NewNodeConfig(profile, ethclient, p.dal)
}

func (p *processor) printChannelInfo(cid ctype.CidType) {
	fmt.Println()
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

func (p *processor) printClosedChannelInfo(cid ctype.CidType) {
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

func (p *processor) printInFlightChannelMessages(cid ctype.CidType) {
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

func (p *processor) printAllChannelsByToken(token *entity.TokenInfo, inactiveSeconds int) {
	var cids []ctype.CidType
	var peers []ctype.Addr
	var tokens []*entity.TokenInfo
	var states []int
	var stateTses, openTses []*time.Time
	var balances []*structs.OnChainBalance
	var err error
	var selfSimplexes, peerSimplexes []*entity.SimplexPaymentChannel
	if inactiveSeconds > 0 {
		inactiveTime := time.Now().Add(time.Duration(-inactiveSeconds) * time.Second).UTC()
		log.Infoln("-- inactive time", inactiveTime)
		cids, peers, tokens, states, stateTses, openTses, balances, selfSimplexes, peerSimplexes, err = p.dal.GetInactiveChanInfo(token, inactiveTime)
	} else {
		cids, peers, tokens, states, stateTses, openTses, balances, selfSimplexes, peerSimplexes, err = p.dal.GetAllChanInfoByToken(token)
	}
	if err != nil {
		log.Fatal(err)
	}
	if len(cids) == 0 {
		fmt.Println("no channels")
		return
	}
	p.printChannels(cids, peers, tokens, states, stateTses, openTses, balances, selfSimplexes, peerSimplexes)
}

func (p *processor) printChannels(
	cids []ctype.CidType, peers []ctype.Addr, tokens []*entity.TokenInfo, states []int,
	stateTses, openTses []*time.Time, balances []*structs.OnChainBalance,
	selfSimplexes, peerSimplexes []*entity.SimplexPaymentChannel) {
	fmt.Printf("------------------- %d channels -------------------\n\n", len(cids))
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
		fmt.Println("-- channel state:", fsm.ChanStateName(states[i]), "| opened:", openTses[i].UTC(), "| last activity:", stateTses[i].UTC())
		fmt.Println("-- onchain balance: self deposit", balances[i].MyDeposit, "self withdrawal", balances[i].MyWithdrawal)
		fmt.Println("-- onchain balance: peer deposit", balances[i].PeerDeposit, "peer withdrawal", balances[i].PeerWithdrawal)
		balance := ledgerview.ComputeBalance(selfSimplexes[i], peerSimplexes[i], balances[i],
			ctype.Bytes2Addr(selfSimplexes[i].GetPeerFrom()), ctype.Bytes2Addr(peerSimplexes[i].GetPeerFrom()), blknum)
		fmt.Println("-- self free balance:", balance.MyFree, "locked balance:", balance.MyLocked)
		fmt.Println("-- peer free balance:", balance.PeerFree, "locked balance:", balance.PeerLocked)
		fmt.Println()
	}
}

func (p *processor) printCidsByTokenAndState(token *entity.TokenInfo, state int, inactiveSeconds int) {
	var count int
	var err error
	var inactiveTime time.Time
	var cids []ctype.CidType
	if inactiveSeconds > 0 {
		inactiveTime = time.Now().Add(time.Duration(-inactiveSeconds) * time.Second).UTC()
		log.Infoln("-- inactive time", inactiveTime)
		count, err = p.dal.CountInactiveCidsByTokenAndState(token, state, inactiveTime)
	} else {
		count, err = p.dal.CountCidsByTokenAndState(token, state)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("-- %d channels at state %s, token %s\n", count, fsm.ChanStateName(state), utils.PrintToken(token))
	if *allIds {
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

func (p *processor) printTokenBalance(token *entity.TokenInfo) {
	cids, _, _, _, _, _, balances, selfSimplexes, peerSimplexes, err := p.dal.GetAllChanInfoByToken(token)
	if err != nil {
		log.Fatal(err)
	}
	if len(cids) == 0 {
		fmt.Println("no channels")
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

func (p *processor) printAllPayInfo(cid ctype.CidType) {
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

func (p *processor) printPayInfo(pid ctype.PayIDType) {
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

func (p *processor) printRoute(dest ctype.Addr, token *entity.TokenInfo) {
	fmt.Println("")
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

func (p *processor) getCidPeerStr(cid ctype.CidType) string {
	peer, found, err := p.dal.GetChanPeer(cid)
	if err != nil {
		log.Fatalf("GetChanPeer err %s, cid %x", err, cid)
	}
	if !found {
		log.Fatalf("peer for channel %x not found", cid)
	}
	return fmt.Sprintf("channel %x peer %x", cid, peer)
}

// ===================== On Chain View Functions =====================

func (p *processor) printChannelLedgerInfo(cid ctype.CidType) {
	fmt.Println("")
	fmt.Println("-- channel ID:", ctype.Cid2Hex(cid))
	chanLedger := p.nodeConfig.GetLedgerContract()
	contract, err := ledger.NewCelerLedgerCaller(chanLedger.GetAddr(), p.nodeConfig.GetEthConn())
	if err != nil {
		log.Fatal("NewCelerLedgerCaller error", err)
	}
	status, err := contract.GetChannelStatus(&bind.CallOpts{}, cid)
	if err != nil {
		log.Fatalln("GetChannelStatus error", err)
	}
	fmt.Println("-- status:", ledgerview.ChanStatusName(status))
	if status == ledgerview.OnChainStatus_UNINITIALIZED {
		return
	}

	peers, deposits, withdrawals, err := contract.GetBalanceMap(&bind.CallOpts{}, cid)
	if err != nil {
		log.Fatalln("GetBalanceMap error", err)
	}
	fmt.Println("-- peers:", ctype.Addr2Hex(peers[0]), ctype.Addr2Hex(peers[1]))
	fmt.Println("-- deposits:", deposits[0], deposits[1])
	fmt.Println("-- withdrawals:", withdrawals[0], withdrawals[1])

	balance, err := contract.GetTotalBalance(&bind.CallOpts{}, cid)
	if err != nil {
		log.Fatalln("GetTotalBalance error", err)
	}
	fmt.Println("-- total balance:", balance)

	if status == ledgerview.OnChainStatus_SETTLING {
		var blknum int64
		header, err := p.nodeConfig.GetEthConn().HeaderByNumber(context.Background(), nil)
		if err != nil {
			log.Error(err)
		} else {
			blknum = header.Number.Int64()
		}
		finalizeBlk, err2 := contract.GetSettleFinalizedTime(&bind.CallOpts{}, cid)
		if err2 != nil {
			log.Fatalln("GetSettleFinalizedTime error", err2)
		}
		finalizeBlknum := finalizeBlk.Int64()
		fmt.Printf("-- settle finalized block %d, current block %d, diff %d\n", finalizeBlknum, blknum, finalizeBlknum-blknum)
	}
	// TODO: add other info
}

func (p processor) printPayRegistryInfo(payID ctype.PayIDType) {
	fmt.Println("")
	fmt.Println("-- pay ID:", ctype.PayID2Hex(payID))
	contract, err := payregistry.NewPayRegistryCaller(
		p.nodeConfig.GetPayRegistryContract().GetAddr(), p.nodeConfig.GetEthConn())
	if err != nil {
		log.Fatal(err)
	}
	info, err := contract.PayInfoMap(&bind.CallOpts{}, payID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-- pay registry info", info.Amount, info.ResolveDeadline.Uint64())
}

func (p processor) printAppBooleanOutcome(appAddr ctype.Addr, argF, argO []byte) {
	fmt.Println("")
	if *argDecode {
		var sq app.SessionQuery
		err := proto.Unmarshal(argO, &sq)
		if err != nil {
			log.Error(err)
		} else {
			session := ctype.Bytes2Hex(sq.GetSession())
			query := ctype.Bytes2Hex(sq.GetQuery())
			fmt.Println("-- app session", session, "query", query)
		}
	}
	contract, err := app.NewIBooleanOutcomeCaller(appAddr, p.nodeConfig.GetEthConn())
	if err != nil {
		log.Fatal(err)
	}
	finalized, err := contract.IsFinalized(&bind.CallOpts{}, argF)
	if err != nil {
		log.Fatal(err)
	}
	outcome, err := contract.GetOutcome(&bind.CallOpts{}, argO)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-- app finalized", finalized, "outcome", outcome)
}

func overrideConfig(profile *common.CProfile) {
	if *storesql != "" {
		profile.StoreSql = *storesql
	} else if *storedir != "" {
		profile.StoreDir = *storedir
	}
}
