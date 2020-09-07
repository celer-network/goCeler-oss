// Copyright 2018-2020 Celer Network

package route

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ec "github.com/ethereum/go-ethereum/common"
)

// Channel describes a channel info
type Channel = structs.Edge

// Data describes a data structure to store all valid channels and the end of block number
type Data struct {
	EndBlockNumber uint64
	Channels       []*Channel
}

// StartRoutingRecoverProcess starts the routing recover process from existing routing data.
// If there already exists routing table in database, the process would do nothing.
func (p *Controller) startRoutingRecoverProcess(
	currentBlk *big.Int,
	routingData []byte,
	nodeConfig common.GlobalNodeConfig,
) error {
	// This judgement is used to skip the recover process
	// when the Osp does not start from scratch.
	// Or the user does not set the routing data flag
	if len(routingData) == 0 {
		return nil
	}

	if nodeConfig == nil || currentBlk == nil {
		return errors.New("Invalid input")
	}

	// Check whether the database has routing info or not
	tks := p.rtBuilder.getAllTokens()
	if len(tks) != 0 {
		// Osp already has routing info, just return
		return nil
	}

	log.Infoln("Starting to recover routing data...")

	var data Data
	err := json.Unmarshal(routingData, &data)
	if err != nil {
		log.Errorln(err)
		return err
	}
	tokens := make(map[ctype.Addr]bool)

	// Recover data from routing data
	for _, c := range data.Channels {
		err = p.AddEdge(c.P1, c.P2, c.Cid, c.Token)
		tokens[c.Token] = true
		if err != nil {
			log.Errorln(err)
			return err
		}
	}

	// fetch logs from the end block number in routing data to the current block number
	from := data.EndBlockNumber
	to := currentBlk.Uint64()

	// prepare the contract related objects
	contract := nodeConfig.GetLedgerContract()
	parsedABI, err := abi.JSON(strings.NewReader(contract.GetABI()))
	if err != nil {
		log.Errorln(err)
		return err
	}

	openChanEv, ok := parsedABI.Events[event.OpenChannel]
	openChanEvHash := openChanEv.ID
	openChanString := openChanEvHash.Hex()
	if !ok {
		log.Errorf("Unknown event name: %s", event.OpenChannel)
		return errors.New("Unknown event name")
	}
	settleChanEv, ok := parsedABI.Events[event.ConfirmSettle]
	settleChanEvHash := settleChanEv.ID
	settleChanString := settleChanEvHash.Hex()
	if !ok {
		log.Errorf("Unknown event name: %s", event.ConfirmSettle)
		return errors.New("Unknow event name")
	}

	// make filter query
	q := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Addresses: []ec.Address{
			contract.GetAddr(),
		},
		Topics: [][]ec.Hash{
			{openChanEvHash, settleChanEvHash},
		},
	}

	log.Infof("Fetching logs from %v to %v", from, to)
	logs, err := contract.GetETHClient().FilterLogs(context.Background(), q)
	// If somehow fetching failed(eg. Too large block range), just return
	// User should restart OSP via the lastest snapshot of routing data
	if err != nil {
		log.Errorln(err)
		return err
	}

	for _, eLog := range logs {
		switch eLog.Topics[0].Hex() {
		case openChanString:
			e := &ledger.CelerLedgerOpenChannel{}
			if err := contract.ParseEvent(event.OpenChannel, eLog, e); err != nil {
				log.Errorln(err)
				return err
			}
			if len(e.PeerAddrs) == 2 {
				if err := p.AddEdge(e.PeerAddrs[0], e.PeerAddrs[1], ctype.CidType(e.ChannelId), e.TokenAddress); err != nil {
					tokens[e.TokenAddress] = true
					log.Errorln(err)
					return err
				}
			}
		case settleChanString:
			e := &ledger.CelerLedgerConfirmSettle{}
			if err := contract.ParseEvent(event.ConfirmSettle, eLog, e); err != nil {
				log.Errorln(err)
				return err
			}
			if err := p.RemoveEdge(ctype.CidType(e.ChannelId)); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}

	for token := range tokens {
		p.BuildTable(token)
	}
	log.Infoln("Routing recovery done..")

	return nil
}
