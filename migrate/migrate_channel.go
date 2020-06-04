// Copyright 2019-2020 Celer Network

package migrate

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	enums "github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
)

const (
	chanMigrationDeadline uint64 = uint64(365 * 24 * 60 * 60 / 13) // estimation of block numbers produced in one year
	chanMigrationInterval uint64 = uint64(30 * 24 * 60 * 60 / 13)  // estimation of block numbers produced in one month

	// migration state for channel
	MigrationStateInitialized int = 0
	MigrationStateSubmitted   int = 1
)

// MigrateChannelProcessor defines the structure of channel's CelerLedger migration processor
type MigrateChannelProcessor struct {
	nodeConfig     common.GlobalNodeConfig
	signer         eth.Signer
	dal            *storage.DAL
	connectionMgr  *rpc.ConnectionManager
	monitorService intfs.MonitorService
	isOsp          bool
}

// NewMigrateChannelProcessor returns a migrate-ledger processor
func NewMigrateChannelProcessor(
	nodeConfig common.GlobalNodeConfig,
	signer eth.Signer,
	dal *storage.DAL,
	connectionMgr *rpc.ConnectionManager,
	monitorService intfs.MonitorService,
	isOsp bool) *MigrateChannelProcessor {
	p := &MigrateChannelProcessor{
		nodeConfig:     nodeConfig,
		signer:         signer,
		dal:            dal,
		connectionMgr:  connectionMgr,
		monitorService: monitorService,
		isOsp:          isOsp,
	}

	if isOsp {
		p.monitorOnDeprecatedLedgers()
	}

	return p
}

// CheckPeerChannelMigration gets all channels of peer and migrates each channel
// (TODO): consider that everytime client gets online, need to check migration process(?performance?)
func (p *MigrateChannelProcessor) CheckPeerChannelMigration(peer ctype.Addr) {
	cids, found, err := p.dal.GetPeerCids(peer)
	if err != nil {
		log.Error(err)
		return
	}
	if !found {
		log.Debugln("No channels found for peer", peer.Hex())
		return
	}

	for _, cid := range cids {
		err = p.checkChannelMigration(peer, cid)
		if err != nil {
			log.Error(err)
		}
	}
}

// checkChannelMigration migrates channel's outdated ledger to the latest ledger
func (p *MigrateChannelProcessor) checkChannelMigration(peer ctype.Addr, cid ctype.CidType) error {
	latestLedgerAddr := p.nodeConfig.GetLedgerContract().GetAddr()
	state, currentLedger, found, err := p.dal.GetChanForMigration(cid)
	if err != nil {
		return fmt.Errorf("Fail to get channel(%x) current ledger: %w", cid, err)
	}
	if !found {
		return fmt.Errorf("No channel found: %w", common.ErrChannelNotFound)
	}
	log.Debugf("current ledger is: %x, latest ledger is: %x", currentLedger, latestLedgerAddr)
	// check if migration init process is already done
	if state != enums.ChanState_OPENED {
		return nil
	}

	if currentLedger == latestLedgerAddr {
		return nil
	}

	currentBlk := p.monitorService.GetCurrentBlockNumber().Uint64()
	deadline, state, _, found, err := p.dal.GetChanMigration(cid, latestLedgerAddr)
	if err != nil {
		return fmt.Errorf("Fail to get channel(%x) migration info: %w", cid, err)
	}
	// if migration info already exists and state is submitted or deadline is still valid
	if found && (state == MigrationStateSubmitted || deadline > currentBlk) {
		return nil
	}

	// if no migration info found for channel,
	// then we need to begin migration initialization
	log.Infof("Start migrating channel: %x for peer: %x", cid, peer)
	deadline = currentBlk + chanMigrationDeadline
	migrationInfo := &entity.ChannelMigrationInfo{
		ChannelId:         cid.Bytes(),
		FromLedgerAddress: currentLedger.Bytes(),
		ToLedgerAddress:   latestLedgerAddr.Bytes(),
		MigrationDeadline: currentBlk + chanMigrationDeadline,
	}

	migrationInfoBytes, err := proto.Marshal(migrationInfo)
	if err != nil {
		return fmt.Errorf("Fail to marshal migration info: %w", err)
	}

	sig, err := p.signer.SignEthMessage(migrationInfoBytes)
	if err != nil {
		return fmt.Errorf("Fail to sign migration info: %w", err)
	}

	req := &rpc.MigrateChannelRequest{
		ChannelMigrationInfo: migrationInfoBytes,
		RequesterSig:         sig,
	}

	c, err := p.connectionMgr.GetClient(peer)
	if err != nil {
		return fmt.Errorf("Fail to get rpc client for peer: %w", err)
	}

	resp, err := c.CelerMigrateChannel(context.Background(), req)
	if err != nil {
		return fmt.Errorf("CelerMigrateChannel rpc error: %w", err)
	}

	// check sig from peer
	if !eth.SigIsValid(peer, migrationInfoBytes, resp.GetApproverSig()) {
		return fmt.Errorf("Signature is invalid: peer: %x", peer)
	}

	onchainReq := p.newOnchainChannelMigrationReq(peer, migrationInfoBytes, sig, resp.GetApproverSig())
	err = p.dal.Transactional(p.processChannelMigrationTx, cid, latestLedgerAddr, &deadline, MigrationStateInitialized, onchainReq)
	if err != nil {
		return fmt.Errorf("Fail to process tx for channel(%x) migration info: %w", cid, err)
	}

	log.Infof("Migrate channel initiation process finished for cid(%x), from old ledger(%x) to new ledger(%x)", cid, currentLedger, latestLedgerAddr)
	return nil
}

// ProcessMigrateChannelRequest processes migrate channel request
func (p *MigrateChannelProcessor) ProcessMigrateChannelRequest(req *rpc.MigrateChannelRequest) (*rpc.MigrateChannelResponse, error) {
	if req == nil {
		log.Errorln(common.ErrInvalidArg)
		return nil, common.ErrInvalidArg
	}

	// check migration info
	migrationInfoBytes := req.GetChannelMigrationInfo()
	var migrationInfo entity.ChannelMigrationInfo
	err := proto.Unmarshal(migrationInfoBytes, &migrationInfo)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	cid := ctype.Bytes2Cid(migrationInfo.GetChannelId())
	log.Infoln("Process migrate channel request for channel:", ctype.Cid2Hex(cid))

	latestLedgerAddr := p.nodeConfig.GetLedgerContract().GetAddr()
	// (TODO): client may have no ledger info in db for now, future data migration would add ledger
	currentLedger, found, err := p.dal.GetChanLedger(cid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !found {
		log.Errorln("No related ledger found for channel", cid.Hex())
		return nil, errors.New("no current ledger found")
	}

	fromLedger := migrationInfo.GetFromLedgerAddress()
	if bytes.Compare(currentLedger.Bytes(), fromLedger) != 0 {
		log.Errorf("inconsistent current ledger info: want(%x), get(%x)", currentLedger, fromLedger)
		return nil, errors.New("inconsistent current ledger info")
	}

	toLedger := migrationInfo.GetToLedgerAddress()
	if bytes.Compare(latestLedgerAddr.Bytes(), toLedger) != 0 {
		log.Errorf("inconsistent config ledger info: want(%x), get(%x)", latestLedgerAddr, toLedger)
		return nil, errors.New("inconsistent config ledger info")
	}

	currentBlk := p.monitorService.GetCurrentBlockNumber().Uint64()
	deadline := migrationInfo.GetMigrationDeadline()
	if currentBlk+chanMigrationInterval >= deadline { // have a tolerant range for deadline
		log.Errorf("Channel migration deadline check failed: current(%d), deadline(%d)", currentBlk, deadline)
		return nil, common.ErrDeadlinePassed
	}

	// check sig from peer
	peer, found, err := p.dal.GetChanPeer(cid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !found {
		log.Errorln("No peer found for channel", cid.Hex())
		return nil, common.ErrChannelNotFound
	}

	if !eth.SigIsValid(peer, migrationInfoBytes, req.GetRequesterSig()) {
		log.Errorln("Invalid requester signature of peer:", peer.Hex())
		return nil, common.ErrInvalidSig
	}

	// after all checks, sign the migration info and prepare response
	sig, err := p.signer.SignEthMessage(migrationInfoBytes)
	if err != nil {
		log.Errorln("Fail to sign migration info:", err)
		return nil, err
	}

	resp := &rpc.MigrateChannelResponse{
		ApproverSig: sig,
	}

	onchainReq := p.newOnchainChannelMigrationReq(peer, migrationInfoBytes, sig, req.GetRequesterSig())
	err = p.dal.Transactional(p.processChannelMigrationTx, cid, latestLedgerAddr, &deadline, MigrationStateInitialized, onchainReq)
	if err != nil {
		log.Errorf("Fail to process tx for channel(%x) migration info: %v", cid, err)
		return nil, err
	}

	log.Infof("Channel migration response process done for peer %x and channel %x", peer, cid)
	return resp, nil
}

// datatbase transaction to store channel migration info
func (p *MigrateChannelProcessor) processChannelMigrationTx(tx *storage.DALTx, args ...interface{}) error {
	cid := args[0].(ctype.CidType)
	toLedger := args[1].(ctype.Addr)
	deadline := args[2].(*uint64)
	state := args[3].(int)
	onchainReq := args[4].(*chain.ChannelMigrationRequest)

	// check if channel migration info exists
	_, state, _, found, err := tx.GetChanMigration(cid, toLedger)
	if err != nil {
		return fmt.Errorf("Fail to find migration info for channel(%x): %w", cid, err)
	}
	// return if channel migration info already exists and was submitted
	if found && state == MigrationStateSubmitted {
		return nil
	}

	// no channel migration info found or found channel migration state initialized
	err = tx.UpsertChanMigration(cid, toLedger, *deadline, state, onchainReq)
	if err != nil {
		return fmt.Errorf("Fail to store migration info for channel(%x): %w", cid, err)
	}

	return nil
}

// monitorOnDeprecatedLedgers starts monitor on all deprecated ledger contracts
func (p *MigrateChannelProcessor) monitorOnDeprecatedLedgers() {
	latestLedger := p.nodeConfig.GetLedgerContract().GetAddr()
	ledgers, err := p.dal.GetAllChanLedgers()
	if err != nil {
		log.Errorln("Fail to get all channel ledgers:", err)
		return
	}

	for _, ledger := range ledgers {
		// do not monitor on latest ledger
		if ledger == latestLedger {
			continue
		}
		contract := p.nodeConfig.GetLedgerContractOn(ledger)
		if contract != nil {
			go p.monitorMigrateChannelEvent(contract)
		}
	}
}

// monitorMigrateChannelEvent monitors onchain event emitted from CelerLedger
func (p *MigrateChannelProcessor) monitorMigrateChannelEvent(contract chain.Contract) {
	monitorCfg := &monitor.Config{
		EventName:  event.MigrateChannelTo,
		Contract:   contract,
		StartBlock: p.monitorService.GetCurrentBlockNumber(),
	}
	_, err := p.monitorService.Monitor(monitorCfg,
		func(id monitor.CallbackID, eLog types.Log) {
			// CAVEAT!!!: suppose we have the same struct for all migration event.
			// If migration event struct changes, this monitor does not work.
			e := &ledger.CelerLedgerMigrateChannelTo{}
			if err := contract.ParseEvent(event.MigrateChannelTo, eLog, e); err != nil {
				log.Error(err)
				return
			}

			cid := ctype.CidType(e.ChannelId)
			newLedger := e.NewLedgerAddr

			err := p.handleMigrateChannelEvent(cid, newLedger)
			if err != nil {
				log.Error(err)
			}
		},
	)
	if err != nil {
		log.Error(err)
	}
	log.Infof("start monitoring onchain channel migration events for ledger: %x", contract.GetAddr())
}

func (p *MigrateChannelProcessor) handleMigrateChannelEvent(cid ctype.CidType, newLedger ctype.Addr) error {
	// check if channel exists
	_, has, err := p.dal.GetChanLedger(cid)
	if err != nil {
		return fmt.Errorf("Fail to find channel of cid(%x): %w", cid, err)
	}
	// unrelated channel
	if !has {
		return nil
	}

	txBody := func(tx *storage.DALTx, args ...interface{}) error {
		// update old ledger to new ledger
		err = tx.UpdateChanLedger(cid, newLedger)
		if err != nil {
			return fmt.Errorf("Fail to update ledger for channel %x, err: %w", cid, err)
		}

		err = tx.DeleteChanMigration(cid)
		if err != nil {
			return fmt.Errorf("Fail to delete migration info for channel %x, err: %w", cid, err)
		}
		return nil
	}

	err = p.dal.Transactional(txBody)
	if err != nil {
		return err
	}
	return nil
}

func (p *MigrateChannelProcessor) newOnchainChannelMigrationReq(peer ctype.Addr, migrationInfo, ownSig, peerSig []byte) *chain.ChannelMigrationRequest {
	req := &chain.ChannelMigrationRequest{
		ChannelMigrationInfo: migrationInfo,
		Sigs:                 [][]byte{ownSig, peerSig},
	}

	// sigs in an ascending order based on addresses associated with these sigs
	if bytes.Compare(p.nodeConfig.GetOnChainAddr().Bytes(), peer.Bytes()) > 0 {
		req.Sigs = [][]byte{peerSig, ownSig}
	}

	return req
}
