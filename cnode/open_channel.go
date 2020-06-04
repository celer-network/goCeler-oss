// Copyright 2018-2020 Celer Network

package cnode

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/celer-network/goCeler/chain"
	"github.com/celer-network/goCeler/chain/channel-eth-go/ledger"
	"github.com/celer-network/goCeler/cnode/openchannelts"
	"github.com/celer-network/goCeler/common"
	"github.com/celer-network/goCeler/common/event"
	"github.com/celer-network/goCeler/common/intfs"
	"github.com/celer-network/goCeler/common/structs"
	"github.com/celer-network/goCeler/config"
	"github.com/celer-network/goCeler/ctype"
	"github.com/celer-network/goCeler/deposit"
	"github.com/celer-network/goCeler/entity"
	"github.com/celer-network/goCeler/fsm"
	"github.com/celer-network/goCeler/ledgerview"
	"github.com/celer-network/goCeler/metrics"
	"github.com/celer-network/goCeler/monitor"
	"github.com/celer-network/goCeler/pem"
	"github.com/celer-network/goCeler/route"
	"github.com/celer-network/goCeler/rpc"
	"github.com/celer-network/goCeler/rtconfig"
	"github.com/celer-network/goCeler/storage"
	"github.com/celer-network/goutils/eth"
	"github.com/celer-network/goCeler/utils"
	"github.com/celer-network/goutils/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	openChannelRpcDeadlineSec    = 5
	maxOpenChannelTimeoutMinutes = 10.0
)

var zeroAmtBytes = []byte{0}

type openChannelProcessor struct {
	nodeConfig          common.GlobalNodeConfig
	signer              eth.Signer
	transactor          *eth.Transactor
	dal                 *storage.DAL
	connectionManager   *rpc.ConnectionManager
	monitorService      intfs.MonitorService
	callbacks           map[string]event.OpenChannelCallback // TokenAddr to OpenChannelEvent callback
	callbacksLock       sync.Mutex
	masterLock          sync.Mutex
	lockPerTokenPerPeer map[string]*sync.Mutex
	routeController     *route.Controller
	depositProcessor    *deposit.Processor
	// keepMonitor describes whether this instance is constantly monitoring on-chain open channel event.
	// clients only monitors open channel when they initialize the process while OSP is constantly monitoring.
	// There is a monitor bit to persist if a client was in monitoring state before crash or restart.
	// OSP who initialize open channel process with another osp will ignore the bit.
	keepMonitor bool
}

func startOpenChannelProcessor(
	nodeConfig common.GlobalNodeConfig,
	signer eth.Signer,
	transactor *eth.Transactor,
	dal *storage.DAL,
	connectionManager *rpc.ConnectionManager,
	monitorService intfs.MonitorService,
	routeController *route.Controller,
	depositProcessor *deposit.Processor,
	keepMonitor bool) (*openChannelProcessor, error) {
	p := &openChannelProcessor{
		nodeConfig:          nodeConfig,
		signer:              signer,
		transactor:          transactor,
		dal:                 dal,
		connectionManager:   connectionManager,
		monitorService:      monitorService,
		callbacks:           make(map[string]event.OpenChannelCallback),
		routeController:     routeController,
		depositProcessor:    depositProcessor,
		lockPerTokenPerPeer: make(map[string]*sync.Mutex),
		keepMonitor:         keepMonitor,
	}
	if keepMonitor {
		// Monitor on all ledgers
		p.monitorOnAllLedgers()
	} else {
		// Restore event monitoring
		ledgerAddrs, err := p.dal.GetMonitorAddrsByEventAndRestart(event.OpenChannel, true /*restart*/)
		if err != nil {
			return nil, err
		}

		for _, ledgerAddr := range ledgerAddrs {
			contract := p.nodeConfig.GetLedgerContractOn(ledgerAddr)
			if contract != nil {
				p.monitorSingleEvent(contract, false /*reset*/)
			}
		}
	}
	return p, nil
}

func invokeErrorCallback(openCallback event.OpenChannelCallback, err *common.E) {
	if openCallback != nil {
		go func() {
			openCallback.HandleOpenChannelErr(err)
		}()
	}
}
func (p *openChannelProcessor) prepareChannelInitializer(
	peer ctype.Addr,
	amtSelf *big.Int,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
) (*entity.PaymentChannelInitializer, error) {
	if tokenInfo.GetTokenType() == entity.TokenType_ERC20 &&
		!ethcommon.IsHexAddress(ctype.Bytes2Hex(tokenInfo.GetTokenAddress())) {
		log.Errorln(common.ErrInvalidTokenAddress.Error())
		return nil, common.ErrInvalidTokenAddress
	}
	if amtSelf == nil || amtPeer == nil {
		log.Error("amt is nil")
		return nil, common.ErrInvalidAmount
	}

	selfAddress := p.nodeConfig.GetOnChainAddr().Bytes()
	// Construct request
	msgValueReceiver := uint64(0)
	lowAddrDist := &entity.AccountAmtPair{
		Account: selfAddress,
		Amt:     amtSelf.Bytes(),
	}
	highAddrDist := &entity.AccountAmtPair{
		Account: peer.Bytes(),
		Amt:     amtPeer.Bytes(),
	}
	// We need ascending order in terms of address
	if bytes.Compare(peer.Bytes(), selfAddress) == -1 {
		lowAddrDist, highAddrDist = highAddrDist, lowAddrDist
		msgValueReceiver = 1
	}
	initializer := &entity.PaymentChannelInitializer{
		InitDistribution: &entity.TokenDistribution{
			Token: tokenInfo,
			Distribution: []*entity.AccountAmtPair{
				lowAddrDist, highAddrDist,
			},
		},
		DisputeTimeout:   config.ChannelDisputeTimeout,
		MsgValueReceiver: msgValueReceiver,
	}
	return initializer, nil
}

// https://docs.google.com/document/d/1ho-FHUkgvWa2Rmr_qFbPRa9rUb6WW8zHQECPKxF-y7Y/edit#
// PRD: https://docs.google.com/document/d/1ltS2YwtAMdRoESeyBpx52fr4puIKactCwQYS1eKknEo/edit?usp=drive_web&ouid=115746433510597393504
func (p *openChannelProcessor) tcbOpenChannel(
	peer ctype.Addr,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
	openCallback event.OpenChannelCallback,
	ocem *pem.OpenChannelEventMessage) error {
	initializer, err := p.prepareChannelInitializer(peer, big.NewInt(0) /*amtSelf*/, amtPeer, tokenInfo)
	tokenAddr := utils.GetTokenAddr(tokenInfo)
	ocem.Peer = ctype.Addr2Hex(peer)
	ocem.TokenAddr = ctype.Addr2Hex(tokenAddr)
	if err != nil {
		invokeErrorCallback(openCallback, &common.E{Reason: err.Error(), Code: 1})
		return err
	}
	// Record call back
	if openCallback != nil {
		p.callbacksLock.Lock()
		p.callbacks[ctype.Addr2Hex(tokenAddr)] = openCallback
		p.callbacksLock.Unlock()
	}
	// deadline according to CORE-622
	initializer.OpenDeadline = p.monitorService.GetCurrentBlockNumber().Uint64() + config.TcbTimeoutInBlockNumber
	ocem.ReadableInitializer = utils.PrintChannelInitializer(initializer)
	initializerBytes, err := proto.Marshal(initializer)
	if err != nil {
		openCallback.HandleOpenChannelErr(&common.E{Reason: err.Error()})
		return err
	}
	sig, err := p.signer.SignEthMessage(initializerBytes)
	if err != nil {
		openCallback.HandleOpenChannelErr(&common.E{Reason: err.Error()})
		return err
	}
	tcbReq := &rpc.OpenChannelRequest{
		ChannelInitializer: initializerBytes,
		RequesterSig:       sig,
	}
	cid, cidErr := p.computePscID(initializerBytes, p.nodeConfig.GetLedgerContract().GetAddr(), p.nodeConfig.GetWalletContract().GetAddr())
	if cidErr != nil {
		log.Errorln("Can't compute cid", cidErr)
		openCallback.HandleOpenChannelErr(&common.E{Reason: cidErr.Error()})
		return cidErr
	}
	rc, err := p.connectionManager.GetClient(peer)
	if err != nil {
		openCallback.HandleOpenChannelErr(&common.E{Reason: err.Error()})
		return err
	}
	rpcDeadline := time.Now().Add(time.Duration(openChannelRpcDeadlineSec) * time.Second)
	ctx, _ := context.WithDeadline(context.Background(), rpcDeadline)
	tcbResp, err := rc.CelerOpenTcbChannel(ctx, tcbReq)
	if err != nil {
		log.Errorln("CelerOpenTcbChannel rpc error", err)
		openCallback.HandleOpenChannelErr(&common.E{Reason: err.Error()})
		return err
	}
	log.Debugln("OpenTcbChannelResponse status", tcbResp.Status)
	if tcbResp.Status != rpc.OpenChannelStatus_OPEN_CHANNEL_TCB_OPENED {
		errMsg := fmt.Sprintf("Wrong Status in TcbResponse: %d", tcbResp.Status)
		log.Errorln(errMsg)
		openCallback.HandleOpenChannelErr(&common.E{Reason: errMsg})
		return err
	}
	if bytes.Compare(cid.Bytes(), tcbResp.GetPaymentChannelId()) != 0 {
		errMsg := fmt.Sprintf("Wrong Channel ID in TcbResponse: %x, expected %x", tcbResp.GetPaymentChannelId(), cid.Bytes())
		openCallback.HandleOpenChannelErr(&common.E{Reason: errMsg})
		return errors.New(errMsg)
	}
	// internal initialization
	var selfAddr ctype.Addr
	selfAddr = p.nodeConfig.GetOnChainAddr()

	// Internal initialization
	channelDescriptor := &openedChannelDescriptor{
		cid:          cid,
		tokenType:    tokenInfo.GetTokenType(),
		tokenAddress: tokenAddr,
		participants: [2]ctype.Addr{peer, selfAddr},
		initDeposits: [2]*big.Int{amtPeer, big.NewInt(0)},
	}
	p.maybeHandleEvent(channelDescriptor, structs.ChanState_TRUST_OPENED, ocem)
	err = p.dal.UpdateChanOpenResp(cid, tcbResp)
	if err != nil {
		log.Errorln("tcbOpenChannel:", err, "can't save initializer", cid.Hex())
		ocem.Error = append(ocem.Error, err.Error())
	}
	return nil
}

func (p *openChannelProcessor) openChannel(
	peer ctype.Addr,
	amtSelf *big.Int,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
	ospToOspOpen bool,
	openCallback event.OpenChannelCallback,
	ocem *pem.OpenChannelEventMessage) error {

	latestLedger := p.nodeConfig.GetLedgerContract()
	latestLedgerAddr := latestLedger.GetAddr()
	initializer, err := p.prepareChannelInitializer(peer, amtSelf, amtPeer, tokenInfo)
	ocem.Peer = ctype.Addr2Hex(peer)
	tokenAddr := utils.GetTokenAddr(tokenInfo)
	ocem.TokenAddr = ctype.Addr2Hex(tokenAddr)
	if err != nil {
		invokeErrorCallback(openCallback, &common.E{Reason: err.Error(), Code: 1})
		return err
	}
	// Record call back
	if openCallback != nil {
		p.callbacksLock.Lock()
		p.callbacks[ctype.Addr2Hex(tokenAddr)] = openCallback
		p.callbacksLock.Unlock()
	}
	initializer.OpenDeadline = p.monitorService.GetCurrentBlockNumber().Uint64() + config.OpenChannelTimeout
	ocem.ReadableInitializer = utils.PrintChannelInitializer(initializer)
	initializerBytes, err := proto.Marshal(initializer)
	if err != nil {
		p.processOpenError(openCallback, latestLedgerAddr, err)
		return err
	}
	sig, err := p.signer.SignEthMessage(initializerBytes)
	if err != nil {
		p.processOpenError(openCallback, latestLedgerAddr, err)
		return err
	}
	openBy := rpc.OpenChannelBy_OPEN_CHANNEL_PROPOSER
	// Cold bootstrap
	if big.NewInt(0).Cmp(amtSelf) == 0 {
		openBy = rpc.OpenChannelBy_OPEN_CHANNEL_APPROVER
	}

	req := &rpc.OpenChannelRequest{
		ChannelInitializer: initializerBytes,
		RequesterSig:       sig,
		OpenBy:             openBy,
		OspToOsp:           ospToOspOpen,
	}

	rc, err := p.connectionManager.GetClient(peer)
	if err != nil {
		p.processOpenError(openCallback, latestLedgerAddr, err)
		return err
	}

	if !p.keepMonitor {
		// If the open-channel monitor (watcher) was auto-restarted after
		// a crash, there is nothing left to do.
		has, found, hasBitErr := p.dal.GetMonitorRestart(monitor.NewEventStr(latestLedgerAddr, event.OpenChannel))
		if hasBitErr != nil {
			log.Debugln("CelerOpenChannel cannot check event monitor bit:", hasBitErr)
			p.processOpenError(openCallback, latestLedgerAddr, hasBitErr)
			return hasBitErr
		} else if found && has {
			log.Debugln("CelerOpenChannel event monitor already running")
			return nil
		}

		// Before sending the request, start monitoring until config.OpenChannelTimeout
		p.monitorSingleEvent(latestLedger, true)
		putBitErr := p.dal.UpsertMonitorRestart(monitor.NewEventStr(latestLedgerAddr, event.OpenChannel), true)
		if putBitErr != nil {
			log.Debugln("CelerOpenChannel cannot put event monitor bit:", putBitErr)
			p.processOpenError(openCallback, latestLedgerAddr, putBitErr)
			return putBitErr
		}
	}
	resp, err := rc.CelerOpenChannel(context.Background(), req)
	if err != nil {
		log.Debugln("CelerOpenChannel rpc error", err)
		p.processOpenError(openCallback, latestLedgerAddr, err)
		return err
	}
	log.Debugln("OpenChannelResponse status", resp.Status)
	switch resp.Status {
	case rpc.OpenChannelStatus_UNDEFINED_OPEN_CHANNEL_STATUS:
		undefinedStatusErr := errors.New("Status undefined in response")
		p.processOpenError(openCallback, latestLedgerAddr, undefinedStatusErr)
		return undefinedStatusErr
	case rpc.OpenChannelStatus_OPEN_CHANNEL_TX_SUBMITTED:
		return nil
	case rpc.OpenChannelStatus_OPEN_CHANNEL_APPROVED:
		// Approved, send transaction
	case rpc.OpenChannelStatus_OPEN_CHANNEL_TCB_OPENED:
		wrongTcbErr := errors.New("Wrong TCB status")
		p.processOpenError(openCallback, latestLedgerAddr, wrongTcbErr)
		return wrongTcbErr
	}
	ocem.Cid = ctype.Bytes2Hex(resp.GetPaymentChannelId())
	txValue := amtSelf
	if tokenInfo.GetTokenType() == entity.TokenType_ERC20 {
		approveErr := p.approveErc20Allowance(latestLedgerAddr, amtSelf, tokenInfo, openCallback)
		if approveErr != nil {
			p.processOpenError(openCallback, latestLedgerAddr, approveErr)
			return approveErr
		}
		txValue = ctype.ZeroBigInt
	}

	log.Debug("Sending OpenChannel tx")
	return p.sendOpenChannelTransaction(
		latestLedgerAddr, resp, p.nodeConfig.GetOnChainAddr().Bytes(), peer.Bytes(), txValue, openCallback)
}
func (p *openChannelProcessor) approveErc20Allowance(ledgerAddr ctype.Addr, amtSelf *big.Int, tokenInfo *entity.TokenInfo, openCallback event.OpenChannelCallback) error {
	// Deposit ERC20
	tokenAddress := utils.GetTokenAddr(tokenInfo)
	log.Debugln("Token address:", ctype.Addr2Hex(tokenAddress))
	erc20, err := chain.NewERC20Caller(tokenAddress, p.transactor.ContractCaller())
	if err != nil {
		p.processOpenError(openCallback, ledgerAddr, err)
		return err
	}
	owner := p.nodeConfig.GetOnChainAddr()
	spender := p.nodeConfig.GetLedgerContract().GetAddr()
	allowance, err := erc20.Allowance(&bind.CallOpts{}, owner, spender)
	if err != nil {
		p.processOpenError(openCallback, ledgerAddr, err)
		return err
	}
	if allowance.Cmp(amtSelf) < 0 {
		receipt, approveErr := p.transactor.TransactWaitMined(
			"Approve",
			&eth.TxConfig{},
			func(
				transactor bind.ContractTransactor,
				opts *bind.TransactOpts) (*types.Transaction, error) {
				erc20, err := chain.NewERC20Transactor(tokenAddress, transactor)
				if err != nil {
					return nil, err
				}
				return erc20.Approve(opts, spender, amtSelf)
			})
		if approveErr == nil && receipt.Status != types.ReceiptStatusSuccessful {
			approveErr = fmt.Errorf("Approve transaction %x failed", receipt.TxHash)
		}
		if approveErr != nil {
			log.Error(approveErr)
			p.processOpenError(openCallback, ledgerAddr, approveErr)
			return approveErr
		}
	}
	return nil
}

func (p *openChannelProcessor) sendOpenChannelTransaction(
	ledgerAddr ctype.Addr,
	peerResponse *rpc.OpenChannelResponse,
	requesterAddr []byte,
	approverAddr []byte,
	txValue *big.Int,
	openCallback event.OpenChannelCallback) error {
	sig0 := peerResponse.RequesterSig
	sig1 := peerResponse.ApproverSig
	if bytes.Compare(requesterAddr, approverAddr) == 1 {
		sig0, sig1 = sig1, sig0
	}
	req := &chain.OpenChannelRequest{
		ChannelInitializer: peerResponse.ChannelInitializer,
		Sigs:               [][]byte{sig0, sig1},
	}
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		log.Debugln("Marshal OpenChannelRequest error", err)
		p.processOpenError(openCallback, ledgerAddr, err)
		return err
	}
	initializer := &entity.PaymentChannelInitializer{}
	marshalErr := proto.Unmarshal(peerResponse.ChannelInitializer, initializer)
	if marshalErr != nil {
		log.Errorf("Can't unmarshal initializer addr 0x%x", requesterAddr)
		return marshalErr
	}
	log.Infoln("Sending OpenChannel:", utils.PrintChannelInitializer(initializer))
	cid := ctype.Bytes2Cid(peerResponse.GetPaymentChannelId())
	_, err = p.transactor.Transact(
		&eth.TransactionStateHandler{
			OnMined: func(receipt *types.Receipt) {
				txHash := receipt.TxHash
				tokenAddr := utils.GetTokenAddr(initializer.GetInitDistribution().GetToken())
				if receipt.Status == types.ReceiptStatusSuccessful {
					log.Debugf("OpenChannel transaction 0x%x succeeded, addr 0x%x, token 0x%x", txHash, requesterAddr, tokenAddr)
				} else {
					errMsg := fmt.Sprintf("OpenChannel transaction 0x%x failed, addr 0x%x, token 0x%x", txHash, requesterAddr, tokenAddr)
					log.Error(errMsg)
					openErr := errors.New(errMsg)
					p.processOpenError(openCallback, ledgerAddr, openErr)
					chState, _, _ := p.dal.GetChanState(cid)
					// reset ch state back to tcb if instantiating failed
					if chState == structs.ChanState_INSTANTIATING {
						p.dal.UpdateChanState(cid, structs.ChanState_TRUST_OPENED)
					}
					recordErr := p.dal.Transactional(
						p.recordOpenChannelFinishTx, ctype.Bytes2Addr(requesterAddr), tokenAddr)
					if recordErr != nil {
						log.Errorln(recordErr, ctype.Bytes2Hex(requesterAddr))
					}
				}
			},
		},
		&eth.TxConfig{EthValue: txValue},
		func(transactor bind.ContractTransactor, opts *bind.TransactOpts) (*types.Transaction, error) {
			contract, err2 := ledger.NewCelerLedgerTransactor(ledgerAddr, transactor)
			if err2 != nil {
				p.processOpenError(openCallback, ledgerAddr, err2)
				return nil, err2
			}
			return contract.OpenChannel(opts, reqBytes)
		})
	if err != nil {
		log.Errorf("%s, Can't open channel for 0x%x", err, requesterAddr)
		p.processOpenError(openCallback, ledgerAddr, err)
	}
	return err
}

func (p *openChannelProcessor) processTcbRequest(in *rpc.OpenChannelRequest, ocem *pem.OpenChannelEventMessage) (*rpc.OpenChannelResponse, error) {
	// openchannel state for metrics
	// deferred function would use the most recent value of it when returning
	stat := metrics.CNodeOpenChanErr
	defer func() {
		metrics.IncCNodeOpenChanEventCnt(metrics.CNodeTcbChan, stat)
	}()

	errTcbResponse := &rpc.OpenChannelResponse{
		Status: rpc.OpenChannelStatus_UNDEFINED_OPEN_CHANNEL_STATUS,
	}
	pscInitializer := &entity.PaymentChannelInitializer{}
	err := proto.Unmarshal(in.GetChannelInitializer(), pscInitializer)
	if err != nil {
		return errTcbResponse, status.Error(codes.InvalidArgument, "channel initializer not parsable")
	}
	ocem.ReadableInitializer = utils.PrintChannelInitializer(pscInitializer)
	log.Infoln("process tcb openchannel request", utils.PrintChannelInitializer(pscInitializer))
	tokenInfo := pscInitializer.GetInitDistribution().GetToken()
	tokenAddr := utils.GetTokenAddr(tokenInfo)
	ocem.TokenAddr = ctype.Addr2Hex(tokenAddr)
	dist := pscInitializer.GetInitDistribution().GetDistribution()
	addr0 := ctype.Bytes2Addr(dist[0].GetAccount())
	addr1 := ctype.Bytes2Addr(dist[1].GetAccount())
	myAddr := p.nodeConfig.GetOnChainAddr()
	var peerAddr ctype.Addr
	// Figure out peer addr
	if addr0 == myAddr {
		peerAddr = addr1
	} else if addr1 == myAddr {
		peerAddr = addr0
	} else {
		return errTcbResponse, status.Error(codes.InvalidArgument, "account list wrong")
	}
	ocem.Peer = ctype.Addr2Hex(peerAddr)

	// Critical section to open channel (for each requester, token pair).
	existingCid, exist, err := p.dal.GetCidByPeerToken(peerAddr, tokenInfo)
	if err != nil {
		return errTcbResponse, status.Error(codes.Internal, err.Error())
	}
	if exist {
		return errTcbResponse, status.Error(codes.AlreadyExists, "cid="+existingCid.Hex())
	}
	// dedup inflight requests.
	allowToOpen := false
	err = p.dal.Transactional(p.recordInflightOpenChannelTx, peerAddr, tokenAddr, &allowToOpen)
	if err != nil {
		return errTcbResponse, status.Error(codes.Internal, err.Error())
	}
	if !allowToOpen {
		return errTcbResponse, status.Error(codes.AlreadyExists, "more than one inflight open channel request.")
	}

	cid, cidErr := p.computePscID(in.GetChannelInitializer(), p.nodeConfig.GetLedgerContract().GetAddr(), p.nodeConfig.GetWalletContract().GetAddr())
	if cidErr != nil {
		revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, peerAddr, tokenAddr)
		if revertErr != nil {
			log.Errorln(revertErr, peerAddr.Hex())
		}
		return errTcbResponse, status.Error(codes.Internal, "can't compute cid")
	}
	ocem.Cid = ctype.Cid2Hex(cid)
	// Internal initialization
	channelDescriptor := &openedChannelDescriptor{
		cid:          cid,
		tokenType:    tokenInfo.GetTokenType(),
		tokenAddress: tokenAddr,
		participants: [2]ctype.Addr{addr0, addr1},
		initDeposits: [2]*big.Int{new(big.Int).SetBytes(dist[0].GetAmt()), new(big.Int).SetBytes(dist[1].GetAmt())},
	}
	mySig, signErr := p.signer.SignEthMessage(in.GetChannelInitializer())
	if signErr != nil {
		revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, peerAddr, tokenAddr)
		if revertErr != nil {
			log.Errorln(revertErr, peerAddr.Hex())
		}
		return errTcbResponse, status.Error(codes.Internal, "can't sign initializer")
	}
	policy, err := RequestTcbDeposit(p.dal, p.nodeConfig, pscInitializer)
	if err != nil {
		ocem.Error = append(ocem.Error, err.Error())
	}
	if policy&AllowTcbOpenChannel == 0 {
		revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, peerAddr, tokenAddr)
		if revertErr != nil {
			log.Errorln(revertErr, peerAddr.Hex())
		}
		return errTcbResponse, status.Error(codes.InvalidArgument, "policy not allowed for this initializer set up")
	}
	p.maybeHandleEvent(channelDescriptor, structs.ChanState_TRUST_OPENED, ocem)
	resp := &rpc.OpenChannelResponse{
		ChannelInitializer: in.GetChannelInitializer(),
		RequesterSig:       in.GetRequesterSig(),
		ApproverSig:        mySig,
		Status:             rpc.OpenChannelStatus_OPEN_CHANNEL_TCB_OPENED,
		PaymentChannelId:   cid.Bytes(),
	}
	err = p.dal.UpdateChanOpenResp(cid, resp)
	if err != nil {
		// If failed, recycle balance and stop approving the tcb.
		log.Errorln("processTcbRequest:", err, "can't save response", cid.Hex())
		p.dal.Transactional(RecycleInstantiatedTcbDepositTx, channelDescriptor, p.nodeConfig.GetOnChainAddr())
		revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, peerAddr, tokenAddr)
		if revertErr != nil {
			log.Errorln(revertErr, peerAddr.Hex())
		}
		return errTcbResponse, status.Error(codes.Internal, "approver can't save response")
	}
	stat = metrics.CNodeOpenChanOK
	return resp, nil
}
func (p *openChannelProcessor) revertInflightOpenChannelTx(tx *storage.DALTx, args ...interface{}) error {
	peerAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	return tx.DeleteOpenChannelTs(peerAddr, tokenAddr)
}
func (p *openChannelProcessor) recordOpenChannelFinishTx(tx *storage.DALTx, args ...interface{}) error {
	peerAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	hasTs, err := tx.HasOpenChannelTs(peerAddr, tokenAddr)
	if err != nil {
		log.Errorln(err, "checking pending open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
		return err
	}
	if hasTs {
		timeNow := time.Now()
		ts, getErr := tx.GetOpenChannelTs(peerAddr, tokenAddr)
		if getErr != nil {
			log.Errorln(err, "getting open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
			return err
		}
		ts.FinishTs = &timeNow
		err = tx.PutOpenChannelTs(peerAddr, tokenAddr, ts)
		if err != nil {
			log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
			return err
		}
	}
	return nil
}
func (p *openChannelProcessor) recordInflightOpenChannelTx(tx *storage.DALTx, args ...interface{}) error {
	peerAddr := args[0].(ctype.Addr)
	tokenAddr := args[1].(ctype.Addr)
	allow := args[2].(*bool)
	*allow = false
	// Dedup open requests during processing open request.
	timeNow := time.Now()
	hasTs, err := tx.HasOpenChannelTs(peerAddr, tokenAddr)
	if err != nil {
		log.Errorln(err, "checking pending open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
		return err
	}
	if hasTs {
		ts, getErr := tx.GetOpenChannelTs(peerAddr, tokenAddr)
		if getErr != nil {
			log.Errorln(getErr, "checking pending open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
			return err
		}
		// No route on (peer, token) but FinishTs set, which means channel has been closed, allow.
		if ts.FinishTs != nil {
			err = tx.PutOpenChannelTs(peerAddr, tokenAddr, &openchannelts.OpenChannelTs{RequestTs: &timeNow})
			if err != nil {
				log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
				return err
			}
			*allow = true
			return nil
		}
		// Having ts but doesn't finish. Check if max deadline has passed.
		// This could happen for case e.g. user sent a open request but didn't send tx on-chain.
		if timeNow.Sub(*ts.RequestTs).Minutes() > maxOpenChannelTimeoutMinutes {
			err = tx.PutOpenChannelTs(peerAddr, tokenAddr, &openchannelts.OpenChannelTs{RequestTs: &timeNow})
			if err != nil {
				log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
				return err
			}
			*allow = true
			return nil
		}
		return nil
	}
	// No ts set, definitely allow.
	err = tx.PutOpenChannelTs(peerAddr, tokenAddr, &openchannelts.OpenChannelTs{RequestTs: &timeNow})
	if err != nil {
		log.Errorln(err, "recording open channel for peer:", peerAddr.Hex(), "token:", ctype.Addr2Hex(tokenAddr))
		return err
	}
	*allow = true
	return nil
}
func (p *openChannelProcessor) processOpenChannelRequest(req *rpc.OpenChannelRequest, ocem *pem.OpenChannelEventMessage) (*rpc.OpenChannelResponse, error) {
	var initializer entity.PaymentChannelInitializer
	err := proto.Unmarshal(req.ChannelInitializer, &initializer)
	errResp := &rpc.OpenChannelResponse{
		Status: rpc.OpenChannelStatus_UNDEFINED_OPEN_CHANNEL_STATUS,
	}

	if err != nil {
		log.Error("Cannot parse channel initializer")
		return errResp, status.Error(codes.InvalidArgument, "channel initializer not parsable")
	}
	ocem.ReadableInitializer = utils.PrintChannelInitializer(&initializer)
	log.Infoln("process openchannel request", utils.PrintChannelInitializer(&initializer))
	// distribution is sorted based on address. So we need to figure out who's requester.
	accnt0 := initializer.InitDistribution.Distribution[0].Account
	accnt1 := initializer.InitDistribution.Distribution[1].Account
	myAddr := p.nodeConfig.GetOnChainAddr().Bytes()
	if bytes.Compare(accnt0, accnt1) != -1 {
		return errResp, status.Error(codes.InvalidArgument, "wrong distribution address order")
	}
	if bytes.Compare(myAddr, accnt0) != 0 && bytes.Compare(myAddr, accnt1) != 0 {
		return errResp, status.Error(codes.InvalidArgument, "wrong channel peers")
	}
	ocem.OspToOsp = req.GetOspToOsp()
	policy, policyErr := RequestStandardDeposit(
		p.monitorService.GetCurrentBlockNumber().Uint64(), p.nodeConfig.GetOnChainAddr(), &initializer, req.GetOspToOsp(), ocem)
	if policy&AllowStandardOpenChannel == 0 {
		return errResp, status.Error(codes.InvalidArgument, "breaks policy:"+policyErr.Error())
	}
	requester := initializer.InitDistribution.Distribution[0].Account
	approver := initializer.InitDistribution.Distribution[1].Account
	if bytes.Compare(requester, myAddr) == 0 {
		requester, approver = approver, requester
	}
	ocem.Peer = ctype.Bytes2Hex(requester)
	tokenInfo := initializer.GetInitDistribution().GetToken()
	tokenAddr := utils.GetTokenAddr(tokenInfo)
	ocem.TokenAddr = ctype.Addr2Hex(tokenAddr)

	// Critical section to open channel (for each requester, token pair).
	existingCid, exist, err := p.dal.GetCidByPeerToken(ctype.Bytes2Addr(requester), tokenInfo)
	if err != nil {
		return errResp, status.Error(codes.Internal, err.Error())
	}
	if exist {
		return errResp, status.Error(codes.AlreadyExists, "cid="+existingCid.Hex())
	}
	allowToOpen := false
	err = p.dal.Transactional(p.recordInflightOpenChannelTx, ctype.Bytes2Addr(requester), tokenAddr, &allowToOpen)
	if err != nil {
		return errResp, status.Error(codes.Internal, err.Error())
	}
	if !allowToOpen {
		return errResp, status.Error(codes.AlreadyExists, "more than one inflight open channel request.")
	}

	mySig, err := p.signer.SignEthMessage(req.ChannelInitializer)
	if err != nil {
		revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, ctype.Bytes2Addr(requester), tokenAddr)
		if revertErr != nil {
			log.Errorln(revertErr, ctype.Bytes2Hex(requester), ctype.Bytes2Hex(approver))
		}
		return errResp, status.Error(
			codes.InvalidArgument, "failed to sign channel initializer")
	}
	latestLedgerAddr := p.nodeConfig.GetLedgerContract().GetAddr()
	cid, cidErr := p.computePscID(req.GetChannelInitializer(), latestLedgerAddr, p.nodeConfig.GetWalletContract().GetAddr())
	if cidErr == nil {
		ocem.Cid = ctype.Cid2Hex(cid)
	}

	resp := &rpc.OpenChannelResponse{
		ChannelInitializer: req.ChannelInitializer,
		RequesterSig:       req.RequesterSig,
		ApproverSig:        mySig,
	}
	switch req.OpenBy {
	case rpc.OpenChannelBy_OPEN_CHANNEL_PROPOSER:
		resp.Status = rpc.OpenChannelStatus_OPEN_CHANNEL_APPROVED
		return &rpc.OpenChannelResponse{
			ChannelInitializer: req.GetChannelInitializer(),
			RequesterSig:       req.GetRequesterSig(),
			ApproverSig:        mySig,
			Status:             rpc.OpenChannelStatus_OPEN_CHANNEL_APPROVED,
		}, nil
	case rpc.OpenChannelBy_OPEN_CHANNEL_APPROVER:
		resp.Status = rpc.OpenChannelStatus_OPEN_CHANNEL_TX_SUBMITTED
		err := p.sendOpenChannelTransaction(
			latestLedgerAddr, resp, requester, approver, big.NewInt(0) /*txValue*/, nil /*openCallback*/)
		if err != nil {
			revertErr := p.dal.Transactional(p.revertInflightOpenChannelTx, ctype.Bytes2Addr(requester), tokenAddr)
			if revertErr != nil {
				log.Errorln(revertErr, ctype.Bytes2Hex(requester), ctype.Bytes2Hex(approver))
			}
			return errResp, status.Error(codes.Internal, "Can't send open channel tx on-chain.")
		}
		return resp, nil
	default:
		return errResp, status.Error(codes.InvalidArgument, "OpenBy not set")
	}
}
func (p *openChannelProcessor) computePscID(
	channelInitializerBytes []byte, ledgerAdr, walletAddr ctype.Addr) (ctype.CidType, error) {
	nonce := crypto.Keccak256(channelInitializerBytes)
	// Does same thing as abi.encodePack in solidity.
	packed := make([]byte, 0, len(walletAddr.Bytes())+len(ledgerAdr.Bytes())+len(nonce))
	packed = append(packed, walletAddr.Bytes()...)
	packed = append(packed, ledgerAdr.Bytes()...)
	packed = append(packed, nonce...)
	walletID := crypto.Keccak256(packed)
	return ctype.Bytes2Cid(walletID), nil
}

type openedChannelDescriptor struct {
	cid          ctype.CidType
	participants [2]ctype.Addr
	initDeposits [2]*big.Int
	tokenType    entity.TokenType
	tokenAddress ctype.Addr
}

func (p *openChannelProcessor) maybeHandleEvent(descriptor *openedChannelDescriptor, chanState int, ocem *pem.OpenChannelEventMessage) bool {
	cid := descriptor.cid
	ocem.Cid = ctype.Cid2Hex(cid)
	log.Infoln("Handle open channel", cid.Hex(), fsm.ChanStateName(chanState))

	if len(descriptor.participants) != 2 || len(descriptor.initDeposits) != 2 {
		log.Error("on chain balances length not match")
		return false
	}
	self := p.nodeConfig.GetOnChainAddr()
	var myIndex int
	if descriptor.participants[0] == self {
		myIndex = 0
	} else if descriptor.participants[1] == self {
		myIndex = 1
	} else {
		return false
	}
	peer := descriptor.participants[1-myIndex]
	ocem.Peer = ctype.Addr2Hex(peer)
	ocem.TokenAddr = ctype.Addr2Hex(descriptor.tokenAddress)

	onChainBalance := &structs.OnChainBalance{
		MyDeposit:      descriptor.initDeposits[myIndex],
		MyWithdrawal:   big.NewInt(0),
		PeerDeposit:    descriptor.initDeposits[1-myIndex],
		PeerWithdrawal: big.NewInt(0),
	}

	txBody := func(tx *storage.DALTx, args ...interface{}) error {
		currState, found, err := tx.GetChanState(cid)
		if err != nil {
			return err
		}
		// If channel state exists, it has to be trust open, new state cannot be trust open
		if found {
			if chanState == structs.ChanState_TRUST_OPENED {
				return common.ErrOpenEventOnWrongState
			}
			if currState == structs.ChanState_TRUST_OPENED || currState == structs.ChanState_INSTANTIATING {
				log.Debugln("channle was TCB opened", cid.Hex())
				err = tx.UpdateChanState(cid, chanState)
				if err != nil {
					return err
				}
				recycleErr := RecycleInstantiatedTcbDepositTx(tx, descriptor, p.nodeConfig.GetOnChainAddr())
				if recycleErr != nil {
					log.Warnln(recycleErr, "it's fine if it's proposer of open channel")
				}
				return nil
			}
			log.Errorf("Receiving open event on state %s, cid %s", fsm.ChanStateName(currState), cid.Hex())
			return common.ErrOpenEventOnWrongState
		}

		token := utils.GetTokenInfoFromAddress(descriptor.tokenAddress)
		selfSimplex, err2 := p.emptySimplex(descriptor, self)
		if err2 != nil {
			return err2
		}
		peerSimplex, err2 := p.emptySimplex(descriptor, peer)
		if err2 != nil {
			return err2
		}
		ledgerAddr := p.nodeConfig.GetLedgerContract().GetAddr()
		err = tx.InsertChan(cid, peer, token, ledgerAddr, chanState, nil /*openResp*/, onChainBalance, 0 /*baseSeqNum*/, 0 /*lastUsedSeqNum*/, 0 /*lastAckedSeqNum*/, 0 /*lastNackedSeqNum*/, selfSimplex, peerSimplex)
		if err != nil {
			return err
		}
		return tx.UpdatePeerCid(peer, cid, true) // add cid
	}
	if err := p.dal.Transactional(txBody); err != nil {
		ocem.Error = append(ocem.Error, err.Error()+":Initializing Simplex")
		log.Error(err)
		return false
	}

	tokenAddr := ctype.Addr2Hex(descriptor.tokenAddress)
	// OSP checks if refill is needed
	if p.keepMonitor && chanState == structs.ChanState_OPENED {
		if err := p.checkBalanceRefill(cid, tokenAddr); err != nil {
			ocem.Error = append(ocem.Error, err.Error()+":refilleWarn")
		}
	}
	recordErr := p.dal.Transactional(
		p.recordOpenChannelFinishTx, peer, descriptor.tokenAddress)
	if recordErr != nil {
		ocem.Error = append(ocem.Error, recordErr.Error()+":recording finish")
		log.Errorln(recordErr, peer)
	}
	// Non-blocking notification.
	p.callbacksLock.Lock()
	cb, ok := p.callbacks[tokenAddr]
	if ok {
		go cb.HandleOpenChannelFinish(cid)
		delete(p.callbacks, tokenAddr)
	}
	p.callbacksLock.Unlock()
	return true
}

func (p *openChannelProcessor) emptySimplex(
	descriptor *openedChannelDescriptor, peerFrom ctype.Addr) (*rpc.SignedSimplexState, error) {
	self := p.nodeConfig.GetOnChainAddr()
	emptySimplex := &entity.SimplexPaymentChannel{
		ChannelId: descriptor.cid[:],
		PeerFrom:  peerFrom.Bytes(),
		SeqNum:    0,
		TransferToPeer: &entity.TokenTransfer{
			Token: &entity.TokenInfo{
				TokenType:    descriptor.tokenType,
				TokenAddress: descriptor.tokenAddress.Bytes(),
			},
			Receiver: &entity.AccountAmtPair{Amt: zeroAmtBytes},
		},
		PendingPayIds:          &entity.PayIdList{},
		LastPayResolveDeadline: 0,
		TotalPendingAmount:     zeroAmtBytes,
	}
	emptySimplexByte, err := proto.Marshal(emptySimplex)
	if err != nil {
		return nil, err
	}
	mySig, err := p.signer.SignEthMessage(emptySimplexByte)
	if err != nil {
		return nil, err
	}
	if self == peerFrom {
		return &rpc.SignedSimplexState{
			SimplexState:  emptySimplexByte,
			SigOfPeerFrom: mySig,
		}, nil
	}
	return &rpc.SignedSimplexState{
		SimplexState: emptySimplexByte,
		SigOfPeerTo:  mySig,
	}, nil
}

func (p *openChannelProcessor) checkBalanceRefill(cid ctype.CidType, tokenAddr string) error {
	refillThreshold := rtconfig.GetRefillThreshold(tokenAddr)
	blkNum := p.monitorService.GetCurrentBlockNumber().Uint64()
	balance, err := ledgerview.GetBalance(p.dal, cid, p.nodeConfig.GetOnChainAddr(), blkNum)
	if err != nil {
		log.Errorln(err, "unabled to find balance for cid", cid.Hex())
		return err
	}
	if refillThreshold.Cmp(balance.MyFree) == 1 {
		warnMsg := fmt.Sprintf("cid %x balance %s below refill threshold %s", cid, balance.MyFree, refillThreshold)
		refillAmount, maxWait := rtconfig.GetRefillAmountAndMaxWait(tokenAddr)
		depositID, err := p.depositProcessor.RequestRefill(cid, refillAmount, maxWait)
		if err == nil {
			log.Warnln(warnMsg, "refill", refillAmount, "job ID:", depositID)
		} else if errors.Is(err, common.ErrPendingRefill) {
			log.Warn(warnMsg)
		} else {
			log.Errorln(warnMsg, "refill error", err)
			return err
		}
	}
	return nil
}

func (p *openChannelProcessor) monitorOnAllLedgers() {
	ledgers := p.nodeConfig.GetAllLedgerContracts()

	for _, contract := range ledgers {
		if contract != nil {
			p.monitorEvent(contract)
		}
	}
}

func (p *openChannelProcessor) monitorEvent(ledgerContract chain.Contract) {
	monitorCfg := &monitor.Config{
		EventName:  event.OpenChannel,
		Contract:   ledgerContract,
		StartBlock: p.monitorService.GetCurrentBlockNumber(),
	}
	_, err := p.monitorService.Monitor(monitorCfg,
		func(id monitor.CallbackID, eLog types.Log) {
			ocem := pem.NewOcem(p.nodeConfig.GetRPCAddr())
			ocem.Type = pem.OpenChannelEventType_CHANNEL_MINED
			e := &ledger.CelerLedgerOpenChannel{}
			if err := ledgerContract.ParseEvent(event.OpenChannel, eLog, e); err != nil {
				log.Error(err)
			}
			channelDescriptor := &openedChannelDescriptor{
				cid:          ctype.CidType(e.ChannelId),
				participants: e.PeerAddrs,
				initDeposits: e.InitialDeposits,
				tokenAddress: e.TokenAddress,
				tokenType:    entity.TokenType(e.TokenType.Int64()),
			}
			// set channel open state to err and maybe change it after event handled
			chanOpen := metrics.CNodeOpenChanErr
			if p.maybeHandleEvent(channelDescriptor, structs.ChanState_OPENED, ocem) {
				chanOpen = metrics.CNodeOpenChanOK
				pem.CommitOcem(ocem)
			}
			if len(e.PeerAddrs) == 2 && p.routeController != nil {
				go p.routeController.AddEdge(e.PeerAddrs[0], e.PeerAddrs[1], e.ChannelId, e.TokenAddress)
			}

			metrics.IncCNodeOpenChanEventCnt(metrics.CNodeRegularChan, chanOpen)
		})
	if err != nil {
		log.Error(err)
	}
}

func (p *openChannelProcessor) monitorSingleEvent(ledgerContract chain.Contract, reset bool) {
	startBlock := p.monitorService.GetCurrentBlockNumber()
	endBlock := new(big.Int).Add(startBlock, big.NewInt(int64(config.OpenChannelTimeout)))
	monitorCfg := &monitor.Config{
		EventName:  event.OpenChannel,
		Contract:   ledgerContract,
		StartBlock: startBlock,
		EndBlock:   endBlock,
		Reset:      reset,
	}
	_, err := p.monitorService.Monitor(monitorCfg,
		func(id monitor.CallbackID, eLog types.Log) {
			ocem := pem.NewOcem(p.nodeConfig.GetRPCAddr())
			ocem.Type = pem.OpenChannelEventType_CHANNEL_MINED
			e := &ledger.CelerLedgerOpenChannel{}
			if err := ledgerContract.ParseEvent(event.OpenChannel, eLog, e); err != nil {
				log.Error(err)
			}
			channelDescriptor := &openedChannelDescriptor{
				cid:          ctype.CidType(e.ChannelId),
				participants: e.PeerAddrs,
				initDeposits: e.InitialDeposits,
				tokenAddress: e.TokenAddress,
				tokenType:    entity.TokenType(e.TokenType.Int64()),
			}
			if p.maybeHandleEvent(channelDescriptor, structs.ChanState_OPENED, ocem) {
				if len(e.PeerAddrs) == 2 && p.routeController != nil {
					go p.routeController.AddEdge(e.PeerAddrs[0], e.PeerAddrs[1], e.ChannelId, e.TokenAddress)
				}
				p.monitorService.RemoveEvent(id)
				p.dal.UpsertMonitorRestart(monitor.NewEventStr(ledgerContract.GetAddr(), event.OpenChannel), false)
				pem.CommitOcem(ocem)
			}
		})
	if err != nil {
		log.Error(err)
	}
}

func initDistributionBreaksPolicy(initDist *entity.TokenDistribution, myAddr []byte) bool {
	// Distribution is sorted based on address. Need to figure out who's me.
	requesterDist := initDist.Distribution[0]
	myDist := initDist.Distribution[1]
	if bytes.Compare(myAddr, requesterDist.Account) == 0 {
		requesterDist, myDist = myDist, requesterDist
	}
	requesterDeposit := big.NewInt(0)
	requesterDeposit.SetBytes(requesterDist.Amt)
	myDeposit := big.NewInt(0)
	myDeposit.SetBytes(myDist.Amt)

	if requesterDeposit.Cmp(big.NewInt(0)) == 0 {
		// 0-balance bootstrap
		myMax := big.NewInt(0)
		tokenType := initDist.Token.TokenType
		if tokenType == entity.TokenType_ETH {
			myMax = rtconfig.GetEthColdBootstrapDeposit()
		} else if tokenType == entity.TokenType_ERC20 {
			myMax = rtconfig.GetErc20ColdBootstrapDeposit(initDist.GetToken().GetTokenAddress())
		}
		if myMax.Cmp(myDeposit) == -1 {
			return true
		}
	} else {
		// Otherwise, don't deposit more than a multiplier
		if myDeposit.Cmp(requesterDeposit.Mul(requesterDeposit, big.NewInt(rtconfig.GetOspDepositMultiplier()))) == 1 {
			return true
		}
	}
	return false
}

func (c *CNode) OpenChannel(
	peer ctype.Addr,
	amtSelf *big.Int,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
	ospToOspOpen bool,
	openCallback event.OpenChannelCallback) error {
	ocem := pem.NewOcem(c.nodeConfig.GetRPCAddr())
	ocem.Type = pem.OpenChannelEventType_OPEN_CHANNEL_API
	err := c.openChannelProcessor.openChannel(peer, amtSelf, amtPeer, tokenInfo, ospToOspOpen, openCallback, ocem)
	if err != nil {
		ocem.Error = append(ocem.Error, err.Error())
	}
	pem.CommitOcem(ocem)
	return err
}
func (c *CNode) TcbOpenChannel(
	peer ctype.Addr,
	amtPeer *big.Int,
	tokenInfo *entity.TokenInfo,
	openCallback event.OpenChannelCallback) error {
	ocem := pem.NewOcem(c.nodeConfig.GetRPCAddr())
	ocem.Type = pem.OpenChannelEventType_TCB_API
	err := c.openChannelProcessor.tcbOpenChannel(peer, amtPeer, tokenInfo, openCallback, ocem)
	if err != nil {
		ocem.Error = append(ocem.Error, err.Error())
	}
	pem.CommitOcem(ocem)
	return err
}

func (c *CNode) ProcessOpenChannelRequest(in *rpc.OpenChannelRequest) (*rpc.OpenChannelResponse, error) {
	ocem := pem.NewOcem(c.nodeConfig.GetRPCAddr())
	ocem.Type = pem.OpenChannelEventType_OPEN_CHANNEL_REQUEST
	response, err := c.openChannelProcessor.processOpenChannelRequest(in, ocem)
	if err != nil {
		ocem.Error = append(ocem.Error, err.Error())
		log.Error(err)
	}
	pem.CommitOcem(ocem)
	return response, err
}

func (c *CNode) ProcessTcbRequest(in *rpc.OpenChannelRequest) (*rpc.OpenChannelResponse, error) {
	ocem := pem.NewOcem(c.nodeConfig.GetRPCAddr())
	ocem.Type = pem.OpenChannelEventType_TCB_REQUEST
	response, err := c.openChannelProcessor.processTcbRequest(in, ocem)
	if err != nil {
		log.Error(err)
		ocem.Error = append(ocem.Error, err.Error())
	}
	pem.CommitOcem(ocem)
	return response, err
}

func (c *CNode) InstantiateChannel(cid ctype.CidType, openCallback event.OpenChannelCallback) error {
	log.Infoln("starting instantiating channel", cid.Hex())
	err := c.openChannelProcessor.instantiateChannel(cid, openCallback)
	log.Infoln(err, "finish instantiating channel", cid.Hex())
	if err == nil {
		// nil err means tx has been sent onchain, waiting for it to be mined,
		// only set channel state to ChanState_INSTANTIATING now, and in openchan event handler
		// channel state will be set to open. or it's set back to tcb if tx fail
		return c.dal.UpdateChanState(cid, structs.ChanState_INSTANTIATING)
	}
	return err
}
func (p *openChannelProcessor) processOpenError(openCallback event.OpenChannelCallback, ledgerAddr ctype.Addr, err error) {
	if !p.keepMonitor {
		p.dal.UpsertMonitorRestart(monitor.NewEventStr(ledgerAddr, event.OpenChannel), false)
	}
	invokeErrorCallback(openCallback, &common.E{Reason: err.Error(), Code: 1})
}
func (p *openChannelProcessor) instantiateChannel(cid ctype.CidType, openCallback event.OpenChannelCallback) error {
	resp, found, err := p.dal.GetChanOpenResp(cid)
	if err != nil {
		return err
	}
	if !found {
		return common.ErrChannelNotFound
	}
	var initializer entity.PaymentChannelInitializer
	err = proto.Unmarshal(resp.GetChannelInitializer(), &initializer)
	if err != nil {
		return err
	}
	tokenInfo := initializer.GetInitDistribution().GetToken()
	depositMap := getDepositMap(initializer.GetInitDistribution().GetDistribution())
	var peer ctype.Addr
	myAddr := p.nodeConfig.GetOnChainAddr()
	amtSelf := depositMap[myAddr]
	for addr := range depositMap {
		if addr != myAddr {
			peer = addr
			break
		}
	}
	if openCallback != nil {
		p.callbacksLock.Lock()
		p.callbacks[utils.GetTokenAddrStr(tokenInfo)] = openCallback
		p.callbacksLock.Unlock()
	}
	ledgerContract := p.nodeConfig.GetLedgerContractOf(cid)
	if ledgerContract == nil {
		return fmt.Errorf("Fail to find ledger contract for channel: %x", cid)
	}
	ledgerAddr := ledgerContract.GetAddr()
	p.monitorSingleEvent(ledgerContract, true)
	err = p.dal.UpsertMonitorRestart(monitor.NewEventStr(ledgerAddr, event.OpenChannel), true)
	if err != nil {
		log.Debugln("CelerOpenChannel cannot put event monitor bit:", err)
		p.processOpenError(openCallback, ledgerAddr, err)
		return err
	}

	txValue := amtSelf
	if tokenInfo.GetTokenType() == entity.TokenType_ERC20 {
		approveErr := p.approveErc20Allowance(ledgerAddr, amtSelf, tokenInfo, openCallback)
		if approveErr != nil {
			p.processOpenError(openCallback, ledgerAddr, approveErr)
			return approveErr
		}
		txValue = ctype.ZeroBigInt
	}
	log.Debug("Instantiating Channel tx")
	return p.sendOpenChannelTransaction(
		ledgerAddr, resp, p.nodeConfig.GetOnChainAddr().Bytes(), peer.Bytes(), txValue, openCallback)
}
