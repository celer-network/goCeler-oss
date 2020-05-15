// Copyright 2020 Celer Network

package cli

import (
	"context"
	"fmt"

	"github.com/celer-network/goCeler/app"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/chain/channel-eth-go/payregistry"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/golang/protobuf/proto"
)

func (p *Processor) ViewChannelOnChain() {
	if *chanid == "" {
		log.Fatal("channel ID not specified")
	}
	cid := ctype.Hex2Cid(*chanid)
	p.printChannelLedgerInfo(cid)
}

func (p *Processor) ViewPayOnChain() {
	if *payid == "" {
		log.Fatal("pay ID not specified")
	}
	p.printPayRegistryInfo(ctype.Hex2PayID(*payid))
}

func (p *Processor) ViewTxOnChain() {
	txInfo, err := ledgerview.GetOnChainTxByHash(ctype.Hex2Hash(*txhash), p.nodeConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("%s tx msg from %x to %x", txInfo.FuncName, txInfo.From, txInfo.To)
	// TODO: print txInfo.FuncInput
}

func (p *Processor) ViewAppOnChain() {
	if *appaddr == "" || *argoutcome == "" {
		log.Fatal("app address or arg for query outcome not specified")
	}
	addr := ctype.Hex2Addr(*appaddr)
	argF := ctype.Hex2Bytes(*argfinalize)
	argO := ctype.Hex2Bytes(*argoutcome)
	p.printAppBooleanOutcome(addr, argF, argO)
}

func (p *Processor) printChannelLedgerInfo(cid ctype.CidType) {
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

func (p Processor) printPayRegistryInfo(payID ctype.PayIDType) {
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

func (p Processor) printAppBooleanOutcome(appAddr ctype.Addr, argF, argO []byte) {
	fmt.Println("")
	if *argdecode {
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
