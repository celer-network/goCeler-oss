// Copyright 2018-2020 Celer Network

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ec "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	profile = flag.String("profile", "", "OSP backend profile")
	ropsten = flag.Bool("ropsten", false, "use ropsten instead of mainnet")
)

var (
	openChannelInfo   = make(map[string]*route.Channel) // store all open channel logs
	openChannelNum    = 0
	settleChannelInfo = make(map[string]bool) // store all settle channel logs
	settleChannelNum  = 0
	data              route.Data // data to be stored in the file
)

const blockDelay = 5
const blkSizePerQuery = 10000

func main() {
	flag.Parse()
	if *profile == "" {
		log.Error("Please specify the profile path")
		return
	}

	cp := common.ParseProfile(*profile)
	client, err := ethclient.Dial(cp.ETHInstance)
	if err != nil {
		log.Fatal(err)
	}

	// prepare the contract related objects
	ledgerABI := ledger.CelerLedgerABI
	ledgerAddr := ctype.Hex2Addr(cp.LedgerAddr)
	parsedABI, err := abi.JSON(strings.NewReader(ledgerABI))
	if err != nil {
		log.Fatal(err)
	}
	contract := bindLedgerContract(client, ledgerAddr)

	// get the current block number
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	currentBlk := header.Number
	endBlk := calculateEndBlock(currentBlk)
	startBlk := new(big.Int).SetUint64(8068820) // mainnet contract deployed block number
	if *ropsten {
		startBlk = new(big.Int).SetUint64(5902515) // ropsten contract deployed block number
	}
	// record the end block number
	data.EndBlockNumber = endBlk.Uint64()

	openChanEv, ok := parsedABI.Events[event.OpenChannel]
	openChanEvHash := openChanEv.ID()
	openChanString := openChanEvHash.Hex()
	if !ok {
		log.Fatalf("Unknown event name: %s", event.OpenChannel)
		return
	}
	settleChanEv, ok := parsedABI.Events[event.ConfirmSettle]
	settleChanEvHash := settleChanEv.ID()
	settleChanString := settleChanEvHash.Hex()
	if !ok {
		log.Fatalf("Unknown event name: %s", event.ConfirmSettle)
	}

	// fetch logs in limited size per query
	from := startBlk.Uint64()
	to := endBlk.Uint64()

	for i := from; i <= to; i += blkSizePerQuery {
		start := i
		end := i + blkSizePerQuery - 1
		if end > to {
			end = to
		}

		log.Infof("Fetching logs from %v to %v...", start, end)
		// make filter query
		q := ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(start),
			ToBlock:   new(big.Int).SetUint64(end),
			Addresses: []ec.Address{
				ledgerAddr,
			},
			Topics: [][]ec.Hash{
				{openChanEvHash, settleChanEvHash},
			},
		}

		logs, fetchErr := client.FilterLogs(context.Background(), q)
		if fetchErr != nil {
			log.Fatal(fetchErr)
		}

		for _, eLog := range logs {
			switch eLog.Topics[0].Hex() {
			case openChanString:
				e := &ledger.CelerLedgerOpenChannel{}
				if err = contract.ParseEvent(event.OpenChannel, eLog, e); err != nil {
					log.Error(err)
				}
				handleOpenChannel(e)
			case settleChanString:
				e := &ledger.CelerLedgerConfirmSettle{}
				if err = contract.ParseEvent(event.ConfirmSettle, eLog, e); err != nil {
					log.Error(err)
				}
				handleSettleChannel(e)
			}
		}
	}

	// exclude settled channels, only keep the valid channels
	keepValidChannel()

	// write data into json file
	log.Info("Writing data into json file...")
	b, encodingErr := json.MarshalIndent(data, "	", "  ")
	if encodingErr != nil {
		log.Fatal(encodingErr)
	}

	file := fmt.Sprintf("%s-snapshot.json", time.Now().Format("2006-01-02"))
	err = ioutil.WriteFile(file, b, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func keepValidChannel() {
	var channels []*route.Channel

	log.Infof("The number of opened channel is: %v", openChannelNum)
	log.Infof("The number of settled channel is: %v", settleChannelNum)

	for cid, channel := range openChannelInfo {
		if _, ok := settleChannelInfo[cid]; ok {
			continue
		}

		channels = append(channels, channel)
	}

	log.Infof("The number of valid open channel is: %v", len(channels))
	data.Channels = channels
}

func handleOpenChannel(e *ledger.CelerLedgerOpenChannel) {
	cid := ctype.CidType(e.ChannelId).Hex()
	edge := &route.Channel{
		P1:    e.PeerAddrs[0],
		P2:    e.PeerAddrs[1],
		Cid:   ctype.CidType(e.ChannelId),
		Token: e.TokenAddress,
	}
	openChannelInfo[cid] = edge
	openChannelNum++
}

func handleSettleChannel(e *ledger.CelerLedgerConfirmSettle) {
	cid := ctype.CidType(e.ChannelId).Hex()
	settleChannelInfo[cid] = true
	settleChannelNum++
}

func bindLedgerContract(conn *ethclient.Client, address ec.Address) *chain.BoundContract {
	ledgerABI := ledger.CelerLedgerABI
	contract, _ := chain.NewBoundContract(conn, address, ledgerABI)

	return contract
}

func calculateEndBlock(current *big.Int) *big.Int {
	blkDelay := new(big.Int).SetUint64(blockDelay)
	return current.Sub(current, blkDelay)
}
